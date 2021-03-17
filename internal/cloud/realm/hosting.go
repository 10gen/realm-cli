package realm

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	hostingPathPattern       = appPathPattern + "/hosting"
	hostingAssetsPathPattern = hostingPathPattern + "/assets"
	hostingAssetPathPattern  = hostingAssetsPathPattern + "/asset"
	hostingCachePathPattern  = hostingPathPattern + "/cache"

	hostingQueryRecursive = "recursive"

	paramMetadata = "meta"
	paramPath     = "path"
)

// HostingAssetData is a Realm app's hosting asset data
type HostingAssetData struct {
	FilePath     string `json:"path"`
	FileHash     string `json:"hash,omitempty"`
	FileSize     int64  `json:"size,omitempty"`
	LastModified int64  `json:"last_modified,omitempty"`
}

// HostingAsset is a Realm app's hosting asset
type HostingAsset struct {
	HostingAssetData
	AppID string                 `json:"appId,omitempty"`
	Attrs HostingAssetAttributes `json:"attrs"`
	URL   string                 `json:"url,omitempty"`
}

// HostingAssetAttribute is a Realm app's hosting asset attribute
type HostingAssetAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (c *client) HostingAssets(groupID, appID string) ([]HostingAsset, error) {
	res, err := c.do(
		http.MethodGet,
		fmt.Sprintf(hostingAssetsPathPattern, groupID, appID),
		api.RequestOptions{Query: map[string]string{hostingQueryRecursive: trueVal}},
	)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, api.ErrUnexpectedStatusCode{"get hosting assets", res.StatusCode}
	}
	defer res.Body.Close()

	var assets []HostingAsset
	if err := json.NewDecoder(res.Body).Decode(&assets); err != nil {
		return nil, err
	}
	return assets, nil
}

func (c *client) HostingAssetUpload(groupID, appID, rootDir string, asset HostingAsset) error {
	file, err := os.Open(filepath.Join(rootDir, asset.FilePath))
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(HostingAsset{
		AppID: appID,
		HostingAssetData: HostingAssetData{
			FilePath: asset.FilePath,
			FileHash: asset.FileHash,
			FileSize: asset.FileSize,
		},
		Attrs: asset.Attrs,
	})
	if err != nil {
		return err
	}

	// Construct a pipe stream: the reader side will be consumed and sent as the body of the outgoing request,
	// and the writer side we can use to asynchronously populate it.
	pipeReader, pipeWriter := io.Pipe()

	bodyWriter := multipart.NewWriter(pipeWriter)
	errChan := make(chan error, 2)

	go func() {
		defer func() {
			bodyWriter.Close()
			pipeWriter.Close()
		}()
		// Create the first part and write the metadata into it
		mw, err := bodyWriter.CreateFormField(paramMetadata)
		if err != nil {
			errChan <- fmt.Errorf("failed to create metadata multipart field: %w", err)
		}

		if _, err := mw.Write(data); err != nil {
			errChan <- fmt.Errorf("failed to write metadata to body: %w", err)
		}

		fw, err := bodyWriter.CreateFormField(paramFile)
		if err != nil {
			errChan <- fmt.Errorf("failed to create file multipart field: %w", err)
		}

		if _, err := io.Copy(fw, file); err != nil {
			errChan <- fmt.Errorf("failed to write file to body: %w", err)
		}
		errChan <- nil
	}()

	res, err := c.do(
		http.MethodPut,
		fmt.Sprintf(hostingAssetPathPattern, groupID, appID),
		api.RequestOptions{
			Body:        pipeReader,
			ContentType: "multipart/mixed; boundary=" + bodyWriter.Boundary(),
		},
	)
	if err := <-errChan; err != nil {
		return err
	}
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusNoContent {
		return api.ErrUnexpectedStatusCode{"upload hosting asset", res.StatusCode}
	}
	return nil
}

func (c *client) HostingAssetRemove(groupID, appID, path string) error {
	res, err := c.do(
		http.MethodDelete,
		fmt.Sprintf(hostingAssetPathPattern, groupID, appID),
		api.RequestOptions{Query: map[string]string{paramPath: path}},
	)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusNoContent {
		return api.ErrUnexpectedStatusCode{"delete hosting asset", res.StatusCode}
	}
	return nil
}

type hostingAssetAttributesUpdateRequest struct {
	Attributes HostingAssetAttributes `json:"attributes"`
}

func (c *client) HostingAssetAttributesUpdate(groupID, appID, path string, attrs ...HostingAssetAttribute) error {
	res, err := c.doJSON(
		http.MethodPatch,
		fmt.Sprintf(hostingAssetPathPattern, groupID, appID),
		hostingAssetAttributesUpdateRequest{attrs},
		api.RequestOptions{Query: map[string]string{paramPath: path}},
	)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusNoContent {
		return api.ErrUnexpectedStatusCode{"update hosting asset attribute", res.StatusCode}
	}
	return nil
}

type hostingCacheInvalidateRequest struct {
	Invalidate bool   `json:"invalidate"`
	Path       string `json:"path"`
}

func (c *client) HostingCacheInvalidate(groupID, appID, path string) error {
	res, err := c.doJSON(
		http.MethodPut,
		fmt.Sprintf(hostingCachePathPattern, groupID, appID),
		hostingCacheInvalidateRequest{true, path},
		api.RequestOptions{},
	)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusNoContent {
		return api.ErrUnexpectedStatusCode{"invalidate hosting cache", res.StatusCode}
	}
	return nil
}

// HostingAssetAttributes is a Realm app's hosting asset attributes
type HostingAssetAttributes []HostingAssetAttribute

func (b HostingAssetAttributes) Len() int {
	return len(b)
}
func (b HostingAssetAttributes) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
func (b HostingAssetAttributes) Less(i, j int) bool {
	if b[i].Name == b[j].Name {
		return b[i].Value < b[j].Value
	}
	return b[i].Name < b[j].Name
}
