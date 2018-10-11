package hosting

import (
	"os"
	"path/filepath"
	"strings"

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
			assetPath := strings.TrimPrefix(path, rootDir)
			var assetDesc AssetDescription
			if assetDescriptions != nil {
				assetDesc = assetDescriptions[assetPath]
			}
			am, fileErr := FileToAssetMetadata(appID, path, info, rootDir, assetDesc)
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
func FileToAssetMetadata(appID string, path string, info os.FileInfo, rootDir string, desc AssetDescription) (*AssetMetadata, error) {
	fileHashStr, err := utils.GenerateFileHashStr(path)
	if err != nil {
		return nil, err
	}
	return NewAssetMetadata(appID, strings.TrimPrefix(path, rootDir), fileHashStr, info.Size(), desc.Attrs), nil
}

// DiffAssetMetadata compares a local and remote []AssetMetadata and returns a AssetMetadataDiffs
// which contains information about the difrences between the two
func DiffAssetMetadata(local, remote []AssetMetadata) AssetMetadataDiffs {
	var addedLocally []AssetMetadata
	var modifiedLocally []ModifiedAssetMetadata
	remoteAM := mapAssetMetadataByPath(remote)

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

	return *NewAssetMetadataDiffs(addedLocally, deletedLocally, modifiedLocally)
}

// mapAssetMetadataByPath returns the AssetMetadata mapped by their FilePath
func mapAssetMetadataByPath(assetsMetadata []AssetMetadata) map[string]AssetMetadata {
	mdM := make(map[string]AssetMetadata, len(assetsMetadata))
	for _, md := range assetsMetadata {
		mdM[md.FilePath] = md
	}
	return mdM
}
