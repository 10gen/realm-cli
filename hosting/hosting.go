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
func ListLocalAssetMetadata(appID, rootDirectory string, assetDescriptions map[string]AssetDescription) ([]AssetMetadata, error) {
	var assetMetadata []AssetMetadata

	err := filepath.Walk(rootDirectory, buildAssetMetadata(appID, &assetMetadata, rootDirectory, assetDescriptions))
	if err != nil {
		return nil, err
	}

	return assetMetadata, nil
}

func buildAssetMetadata(appID string, assetMetadata *[]AssetMetadata, rootDir string, assetDescriptions map[string]AssetDescription) filepath.WalkFunc {
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
			var assetDesc AssetDescription
			if assetDescriptions != nil {
				assetDesc = assetDescriptions[assetPath]
			}
			am, fileErr := FileToAssetMetadata(appID, path, assetPath, info, assetDesc)
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
func FileToAssetMetadata(appID, path, assetPath string, info os.FileInfo, desc AssetDescription) (*AssetMetadata, error) {
	fileHashStr, err := utils.GenerateFileHashStr(path)
	if err != nil {
		return nil, err
	}

	return NewAssetMetadata(appID, assetPath, fileHashStr, info.Size(), desc.Attrs), nil
}

// MetadataFileToAssetDescriptions attempts to open the file at the path given
// and build AssetDescriptions from this file
func MetadataFileToAssetDescriptions(path string) (map[string]AssetDescription, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
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

// DiffAssetMetadata compares a local and remote []AssetMetadata and returns a AssetMetadataDiffs
// which contains information about the difrences between the two
func DiffAssetMetadata(local, remote []AssetMetadata) *AssetMetadataDiffs {
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
	for _, rAM := range remoteAM {
		deletedLocally = append(deletedLocally, rAM)
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
