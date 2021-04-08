package local

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	numHostingWorkers = 4
)

var (
	validAttrNames = map[string]struct{}{
		api.HeaderContentType:             {},
		api.HeaderContentDisposition:      {},
		api.HeaderContentLanguage:         {},
		api.HeaderContentEncoding:         {},
		api.HeaderCacheControl:            {},
		api.HeaderWebsiteRedirectLocation: {},
	}
)

// Hosting is the local Realm app hosting
type Hosting struct {
	RootDir string
}

// HostingAssetClient is the hosting asset client
type HostingAssetClient interface {
	Get(url string) (*http.Response, error)
}

// FindAppHosting finds the local Realm app hosting files
func FindAppHosting(path string) (Hosting, error) {
	app, ok, err := FindApp(path)
	if err != nil {
		return Hosting{}, err
	}
	if !ok {
		return Hosting{}, nil
	}

	rootDir := filepath.Join(app.RootDir, NameHosting)

	return Hosting{rootDir}, nil
}

// HostingDiffs are the hosting asset differences between a local and remote Realm app
type HostingDiffs struct {
	Added    []realm.HostingAsset
	Deleted  []realm.HostingAsset
	Modified []ModifiedHostingAsset
}

// Cap returns the hosting diffs' total capacity
func (d HostingDiffs) Cap() int {
	return d.Size() + 3
}

// Size returns the total number of diffs
func (d HostingDiffs) Size() int {
	return len(d.Added) + len(d.Deleted) + len(d.Modified)
}

// Strings returns the hosting diffs' formatted output
func (d HostingDiffs) Strings() []string {
	diffs := make([]string, 0, d.Cap())

	if len(d.Added) > 0 {
		diffs = append(diffs, "New hosting files")
		for _, added := range d.Added {
			diffs = append(diffs, terminal.Indent+"+ "+added.FilePath)
		}
	}

	if len(d.Deleted) > 0 {
		diffs = append(diffs, "Removed hosting files")
		for _, deleted := range d.Deleted {
			diffs = append(diffs, terminal.Indent+"- "+deleted.FilePath)
		}
	}

	if len(d.Modified) > 0 {
		diffs = append(diffs, "Modified hosting files")
		for _, modified := range d.Modified {
			diffs = append(diffs, terminal.Indent+"* "+modified.FilePath)
		}
	}

	return diffs
}

// ModifiedHostingAsset is a Realm hosting asset with information about its local file changes
type ModifiedHostingAsset struct {
	realm.HostingAsset
	BodyModified  bool
	AttrsModified bool
}

// Diffs returns the local Realm app's hosting asset differences
// with the provided remote Realm app's hosting assets
func (h Hosting) Diffs(cachePath, appID string, appAssets []realm.HostingAsset) (HostingDiffs, error) {
	assets, err := readMetadata(h.RootDir)
	if err != nil {
		return HostingDiffs{}, err
	}

	assetCache, err := loadHostingAssetCache(cachePath)
	if err != nil {
		return HostingDiffs{}, err
	}
	localAssets, err := walkFiles(h.RootDir, appID, assets, assetCache)
	if err != nil {
		return HostingDiffs{}, err
	}

	if assetCache.dirty {
		if err := assetCache.save(); err != nil {
			return HostingDiffs{}, err
		}
	}

	var added, deleted []realm.HostingAsset
	var modified []ModifiedHostingAsset

	appAssetsByPath := make(map[string]realm.HostingAsset, len(appAssets))
	for _, asset := range appAssets {
		appAssetsByPath[asset.FilePath] = asset
	}
	delete(appAssetsByPath, "/") // ignore root directory

	for _, localAsset := range localAssets {
		if appAsset, ok := appAssetsByPath[localAsset.FilePath]; !ok {
			added = append(added, localAsset)
		} else {
			bodyModified := localAsset.FileHash != appAsset.FileHash
			attrsModified := !assetAttrsEquals(appAsset.Attrs, localAsset.Attrs)

			if bodyModified || attrsModified {
				modified = append(modified, ModifiedHostingAsset{
					HostingAsset:  localAsset,
					BodyModified:  bodyModified,
					AttrsModified: attrsModified,
				})
			}

			delete(appAssetsByPath, localAsset.FilePath)
		}
	}

	for _, appAsset := range appAssetsByPath {
		deleted = append(deleted, appAsset)
	}

	return HostingDiffs{added, deleted, modified}, nil
}

// UploadHostingAssets uploads the hosting assets based on the diff of that file
func (h Hosting) UploadHostingAssets(realmClient realm.Client, groupID, appID string, hostingDiffs HostingDiffs, errHandler func(err error)) error {
	var wg sync.WaitGroup

	jobCh := make(chan func())
	errCh := make(chan error)
	doneCh := make(chan struct{})

	var errs []error
	go func() {
		for err := range errCh {
			errHandler(err)
			errs = append(errs, err)
		}
		doneCh <- struct{}{}
	}()

	for n := 0; n < numHostingWorkers; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobCh {
				job()
			}
		}()
	}

	assetsDir := filepath.Join(h.RootDir, NameFiles)

	for _, added := range hostingDiffs.Added {
		asset := added // the closure otherwise sees the same value for `added` each iteration
		jobCh <- func() {
			if err := realmClient.HostingAssetUpload(groupID, appID, assetsDir, asset); err != nil {
				errCh <- fmt.Errorf("failed to add %s: %w", asset.FilePath, err)
			}
		}
	}

	for _, deleted := range hostingDiffs.Deleted {
		asset := deleted // the closure otherwise sees the same value for `added` each iteration
		jobCh <- func() {
			if err := realmClient.HostingAssetRemove(groupID, appID, asset.FilePath); err != nil {
				errCh <- fmt.Errorf("failed to remove %s: %w", asset.FilePath, err)
			}
		}
	}

	for _, modified := range hostingDiffs.Modified {
		asset := modified // the closure otherwise sees the same value for `added` each iteration
		jobCh <- func() {
			if asset.AttrsModified && !asset.BodyModified {
				if err := realmClient.HostingAssetAttributesUpdate(groupID, appID, asset.FilePath, asset.Attrs...); err != nil {
					errCh <- fmt.Errorf("failed to update attributes for %s: %w", asset.FilePath, err)
				}
			} else {
				if err := realmClient.HostingAssetUpload(groupID, appID, assetsDir, asset.HostingAsset); err != nil {
					errCh <- fmt.Errorf("failed to update %s: %w", asset.FilePath, err)
				}
			}
		}
	}

	close(jobCh)
	wg.Wait()

	close(errCh)
	<-doneCh

	if len(errs) > 0 {
		return fmt.Errorf("%d error(s) occurred while importing hosting assets", len(errs))
	}
	return nil
}

// WriteHostingAssets writes the hosting assets to disk
func WriteHostingAssets(assetClient HostingAssetClient, rootDir, groupID, appID string, appAssets []realm.HostingAsset) error {
	dir := filepath.Join(rootDir, NameHosting)

	assets := make([]hostingAsset, 0, len(appAssets))

	for _, appAsset := range appAssets {
		// if there are no asset attributes
		if len(appAsset.Attrs) == 0 {
			continue // do not add it to metadata.json
		}

		// if the asset has a single, "Content-Type" attribute
		if len(appAsset.Attrs) == 1 && appAsset.Attrs[0].Name == api.HeaderContentType {
			ext := filepath.Ext(appAsset.FilePath)
			if ext != "" {
				// and its value equals the content type specified by its file extension
				contentType, ok := api.ContentTypeByExtension(ext[1:])
				if ok && contentType == appAsset.Attrs[0].Value {
					continue // do not add it to metadata.json
				}
			}
		}

		assetAttrs := make(realm.HostingAssetAttributes, 0, len(appAsset.Attrs))
		for _, attr := range appAsset.Attrs {
			if _, ok := validAttrNames[attr.Name]; ok {
				assetAttrs = append(assetAttrs, attr)
			}
		}

		assets = append(assets, hostingAsset{appAsset.FilePath, assetAttrs})
	}

	metadata, err := json.Marshal(assets)
	if err != nil {
		return err
	}

	if err := WriteFile(
		filepath.Join(dir, NameMetadata+extJSON),
		0666,
		bytes.NewReader(metadata),
	); err != nil {
		return err
	}

	var wg sync.WaitGroup

	assetCh := make(chan realm.HostingAsset)
	errCh := make(chan error)
	doneCh := make(chan struct{})

	var errs []error

	go func() {
		for err := range errCh {
			errs = append(errs, err)
		}
		doneCh <- struct{}{}
	}()

	for n := 0; n < numHostingWorkers; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for asset := range assetCh {
				if strings.HasSuffix(asset.FilePath, "/") {
					continue
				}

				res, err := assetClient.Get(asset.URL)
				if err != nil {
					errCh <- err
					continue
				}
				if res.StatusCode != http.StatusOK {
					errCh <- api.ErrUnexpectedStatusCode{"get hosting asset", res.StatusCode}
					continue
				}

				if err := WriteFile(
					filepath.Join(dir, NameFiles, asset.FilePath),
					0666,
					res.Body,
				); err != nil {
					errCh <- err
				}
			}
		}()
	}

	for _, appAsset := range appAssets {
		assetCh <- appAsset
	}

	close(assetCh)
	wg.Wait()

	close(errCh)
	<-doneCh

	if len(errs) > 0 {
		return fmt.Errorf("%d error(s) occurred while exporting hosting assets", len(errs))
	}
	return nil
}

func assetAttrsEquals(appAssetAttrs, localAssetAttrs realm.HostingAssetAttributes) bool {
	sort.Sort(&appAssetAttrs)
	sort.Sort(&localAssetAttrs)

	if len(appAssetAttrs) != len(localAssetAttrs) {
		return false
	}

	for i, appAssetAttr := range appAssetAttrs {
		localAssetAttr := localAssetAttrs[i]
		if localAssetAttr.Name != appAssetAttr.Name || localAssetAttr.Value != appAssetAttr.Value {
			return false
		}
	}

	return true
}

type hostingAsset struct {
	Path  string                        `json:"path"`
	Attrs []realm.HostingAssetAttribute `json:"attrs"`
}

// readMetadata will parse the Realm app's hosting metadata file
// and return the assets mapped by their normalized file paths
func readMetadata(rootDir string) (map[string]hostingAsset, error) {
	f, err := os.Open(filepath.Join(rootDir, NameMetadata+extJSON))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var assets []hostingAsset
	if err := json.NewDecoder(f).Decode(&assets); err != nil {
		return nil, err
	}

	assetsByPath := make(map[string]hostingAsset, len(assets))
	for _, asset := range assets {
		path := normalizePathSeparator(asset.Path)
		assetsByPath[path] = asset
	}
	return assetsByPath, nil
}

func walkFiles(rootDir, appID string, localAssets map[string]hostingAsset, assetCache *hostingAssetCache) ([]realm.HostingAsset, error) {
	dir := filepath.Join(rootDir, NameFiles)

	var assets []realm.HostingAsset

	if err := filepath.Walk(dir, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			return nil
		}

		pathRelative, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		assetPath := "/" + normalizePathSeparator(pathRelative)

		localAsset, localAssetOK := localAssets[assetPath]

		var attrs []realm.HostingAssetAttribute
		if localAssetOK {
			attrs = localAsset.Attrs
		} else {
			attrs = resolveAttributes(assetPath)
		}

		var assetData realm.HostingAssetData

		if data, ok := assetCache.get(appID, assetPath); ok &&
			data.FileSize == fileInfo.Size() &&
			data.LastModified == fileInfo.ModTime().Unix() {
			assetData = data
		} else {
			hash, err := generateHash(path)
			if err != nil {
				return err
			}

			assetData = realm.HostingAssetData{
				FilePath:     assetPath,
				FileHash:     hash,
				FileSize:     fileInfo.Size(),
				LastModified: fileInfo.ModTime().Unix(),
			}
			assetCache.set(appID, assetData)
		}

		assets = append(assets, realm.HostingAsset{
			HostingAssetData: assetData,
			AppID:            appID,
			Attrs:            attrs,
		})
		return nil
	}); err != nil {
		return nil, err
	}

	assetsByPath := make(map[string]realm.HostingAsset, len(assets))
	for _, asset := range assets {
		assetsByPath[asset.FilePath] = asset
	}

	for k := range localAssets {
		if _, ok := assetsByPath[k]; !ok {
			return nil, fmt.Errorf("file '%s' has an entry in metadata file, but does not appear in files directory", k)
		}
	}
	return assets, nil
}

func generateHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// normalizePathSeparator returns a path with all provided path separators
// replaced with '/' uniformly
func normalizePathSeparator(path string) string {
	return strings.ReplaceAll(path, string(os.PathSeparator), "/")
}

func resolveAttributes(path string) []realm.HostingAssetAttribute {
	ext := filepath.Ext(path)
	if ext == "" {
		return []realm.HostingAssetAttribute{}
	}

	contentType, ok := api.ContentTypeByExtension(ext[1:])
	if !ok {
		return []realm.HostingAssetAttribute{}
	}

	return []realm.HostingAssetAttribute{{api.HeaderContentType, contentType}}
}

type hostingAssetCache struct {
	path    string
	dirty   bool
	entries map[string]map[string]realm.HostingAssetData
}

func loadHostingAssetCache(cachePath string) (*hostingAssetCache, error) {
	cache := hostingAssetCache{path: cachePath, entries: map[string]map[string]realm.HostingAssetData{}}

	file, err := os.Open(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &cache, nil
		}
		return nil, err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&cache); err != nil {
		return nil, err
	}
	return &cache, nil
}

func (cache hostingAssetCache) save() error {
	dir := filepath.Dir(cache.path)
	if err := mkdir(dir); err != nil {
		return err
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	file, err := os.Create(cache.path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

func (cache hostingAssetCache) get(appID, path string) (realm.HostingAssetData, bool) {
	appEntries, ok := cache.entries[appID]
	if !ok {
		return realm.HostingAssetData{}, false
	}

	entry, ok := appEntries[path]
	return entry, ok
}

func (cache *hostingAssetCache) set(appID string, entry realm.HostingAssetData) {
	if _, ok := cache.entries[appID]; !ok {
		cache.entries[appID] = map[string]realm.HostingAssetData{}
	}

	cache.dirty = true
	cache.entries[appID][entry.FilePath] = entry
}

func (cache hostingAssetCache) MarshalJSON() ([]byte, error) {
	return json.Marshal(cache.entries)
}

func (cache *hostingAssetCache) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &cache.entries)
}
