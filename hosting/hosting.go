package hosting

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/10gen/stitch-cli/utils"
)

// ListLocalAssetMetadata walks all files from the rootDirectory
// and builds []AssetMetadata from those files
// returns the assetMetadata and possibly alters the assetCache
func ListLocalAssetMetadata(appID, rootDirectory string, assetDescriptions map[string]AssetDescription, assetCache AssetCache) ([]AssetMetadata, error) {
	var assetMetadata []AssetMetadata

	err := filepath.Walk(rootDirectory, buildAssetMetadata(appID, &assetMetadata, rootDirectory, assetDescriptions, assetCache))
	if err != nil {
		return nil, err
	}

	metadataOnDisk := make(map[string]AssetMetadata)

	for _, am := range assetMetadata {
		metadataOnDisk[am.FilePath] = am
	}

	for key := range assetDescriptions {
		if _, ok := metadataOnDisk[key]; !ok {
			return nil, fmt.Errorf("file '%s' has an entry in metadata file, but does not appear in files directory", key)
		}
	}

	return assetMetadata, nil
}

func buildAssetMetadata(appID string, assetMetadata *[]AssetMetadata, rootDir string, assetDescriptions map[string]AssetDescription, assetCache AssetCache) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, pathErr := filepath.Rel(rootDir, path)
			if pathErr != nil {
				return pathErr
			}
			assetPath := fmt.Sprintf("/%s", relPath)

			var desc *AssetDescription
			if assetDescriptions != nil {
				if descEntry, ok := assetDescriptions[assetPath]; ok {
					desc = &descEntry
				}
			}

			am, fileErr := FileToAssetMetadata(appID, path, assetPath, info, desc, assetCache)
			if fileErr != nil {
				return fileErr
			}

			*assetMetadata = append(*assetMetadata, *am)
		}

		return nil
	}
}

// FileToAssetMetadata generates a file hash for the given file
// and generates the assetAttributes and creates an AssetMetadata from these
// if the file hash has changed this will update the assetCache
func FileToAssetMetadata(appID, path, assetPath string, info os.FileInfo, desc *AssetDescription, assetCache AssetCache) (*AssetMetadata, error) {

	attrs := []AssetAttribute{}
	if desc != nil {
		attrs = desc.Attrs
	} else {
		// This asset doesn't have an entry in the metadata. Try to assign a Content-Type
		// based on the file extension, if possible.
		if extension := filepath.Ext(assetPath); extension != "" {
			if contentType, ok := utils.GetContentTypeByExtension(extension[1:]); ok {
				attrs = []AssetAttribute{
					{Name: "Content-Type", Value: contentType},
				}
			}
		}
	}

	// check cache for file hash
	if ace, ok := assetCache.Get(appID, assetPath); ok {
		if ace.FileSize == info.Size() && ace.LastModified == info.ModTime().Unix() {
			return NewAssetMetadata(appID, assetPath, ace.FileHash, info.Size(), attrs, info.ModTime().Unix()), nil
		}
	}

	// file hash was not cached so generate one
	generated, err := utils.GenerateFileHashStr(path)
	if err != nil {
		return nil, err
	}

	assetCache.Set(appID, AssetCacheEntry{
		assetPath,
		info.ModTime().Unix(),
		info.Size(),
		generated,
	})

	return NewAssetMetadata(appID, assetPath, generated, info.Size(), attrs, info.ModTime().Unix()), nil
}

// MetadataFileToAssetDescriptions attempts to open the file at the path given
// and build AssetDescriptions from this file
func MetadataFileToAssetDescriptions(path string) (map[string]AssetDescription, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	descs := []AssetDescription{}
	decErr := dec.Decode(&descs)
	if decErr != nil {
		return nil, decErr
	}

	descM := make(map[string]AssetDescription, len(descs))
	for _, desc := range descs {
		descM[desc.FilePath] = desc
	}

	return descM, nil
}

// CacheFileToAssetCache attempts to open the file at the path given
// and build a map of appID to a map of file path strings a AssetCache
func CacheFileToAssetCache(path string) (AssetCache, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	assetCache := basicAssetCache{}
	decErr := dec.Decode(&assetCache)
	if decErr != nil {
		return nil, decErr
	}
	return &assetCache, nil
}

// UpdateCacheFile attempts to update the file at the path given
// with the AssetCache passed in
func UpdateCacheFile(path string, assetCache AssetCache) error {
	mAssetCache, mErr := json.Marshal(assetCache)
	if mErr != nil {
		return mErr
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	_, wErr := f.Write(mAssetCache)
	if wErr != nil {
		return wErr
	}

	return f.Close()
}

// DiffAssetMetadata compares a local and remote []AssetMetadata and returns a AssetMetadataDiffs
// which contains information about the differences between the two
// if the merge paramater is true than me ignore deleted assets
func DiffAssetMetadata(local, remote []AssetMetadata, merge bool) *AssetMetadataDiffs {
	var addedLocally []AssetMetadata
	var modifiedLocally []ModifiedAssetMetadata
	remoteAM := AssetsMetadata(remote).MapByPath()

	for _, lAM := range local {
		if rAM, ok := remoteAM[lAM.FilePath]; !ok {
			addedLocally = append(addedLocally, lAM)
		} else {
			modifiedAM := GetModifiedAssetMetadata(lAM, rAM)
			if modifiedAM.BodyModified || modifiedAM.AttrModified {
				modifiedLocally = append(modifiedLocally, modifiedAM)
			}
			delete(remoteAM, lAM.FilePath)
		}
	}

	var deletedLocally []AssetMetadata
	//at this point the remoteAM map only contains AssetMetadata that were deleted locally
	//if this is a merge then just ignore files deleted locally
	if !merge {
		for _, rAM := range remoteAM {
			deletedLocally = append(deletedLocally, rAM)
		}
	}

	return NewAssetMetadataDiffs(addedLocally, deletedLocally, modifiedLocally)
}

// Diff returns a list of strings representing the diff
func (amd *AssetMetadataDiffs) Diff() []string {
	var diff []string

	if len(amd.AddedLocally) > 0 {
		diff = append(diff, "New Files:")
	}
	for _, added := range amd.AddedLocally {
		diff = append(diff, fmt.Sprintf("\t+ %s", added.FilePath))
	}

	if len(amd.DeletedLocally) > 0 {
		diff = append(diff, "Removed Files:")
	}
	for _, deleted := range amd.DeletedLocally {
		diff = append(diff, fmt.Sprintf("\t- %s", deleted.FilePath))
	}

	if len(amd.ModifiedLocally) > 0 {
		diff = append(diff, "Modified Files:")
	}
	for _, modified := range amd.ModifiedLocally {
		diff = append(diff, fmt.Sprintf("\t* %s", modified.AssetMetadata.FilePath))
	}

	return diff
}
