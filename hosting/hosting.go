package hosting

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/10gen/stitch-cli/api"

	"github.com/10gen/stitch-cli/utils"
)

// AssetMetadata ------------------------------------------------------------------------------------------------------
func listLocalAssetMetadata(appID, rootDirectory string) ([]api.AssetMetadata, error) {
	if err := validatePath(rootDirectory); err != nil {
		return nil, err
	}

	var assetMetadata []api.AssetMetadata
	err := filepath.Walk(rootDirectory, buildAssetMetadata(appID, &assetMetadata))
	if err != nil {
		return nil, err
	}

	return assetMetadata, nil
}

func buildAssetMetadata(appID string, assetMetadata *[]api.AssetMetadata) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			am, fileErr := fileToAssetMetadata(appID, path, info)
			if fileErr != nil {
				return fileErr
			}
			*assetMetadata = append(*assetMetadata, *am)
		}
		return nil
	}
}

func fileToAssetMetadata(appID string, path string, info os.FileInfo) (*api.AssetMetadata, error) {
	fileHash, err := utils.GenerateFileHash(path)
	if err != nil {
		return nil, err
	}

	//TODO STITCH-2028 add assetAttributes

	fileHashStr := utils.FileHashStr(fileHash)
	return &api.AssetMetadata{
		AppID:        appID,
		FilePath:     path,
		FileHash:     fileHashStr,
		FileSize:     info.Size(),
		LastModified: info.ModTime().Unix(),
	}, nil
}

// ValidatePath checks if a path string contains illegal path characters
func validatePath(path string) error {
	if containsDotDot(path) {
		return errors.New("invalid path")
	}
	return nil
}

func containsDotDot(v string) bool {
	return strings.Contains(v, "..")
}
