package commands

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/10gen/stitch-cli/api"
)

// Supported extensions
const (
	extZip = ".zip"
	extTar = ".tar"
	extGz  = ".gz"
	extTgz = ".tgz"
)

func ImportDependencies(groupID, appID, dir string, client api.StitchClient) error {
	fullPath, findErr := findDependenciesArchive(dir)
	if findErr != nil {
		return findErr
	}

	valErr := validateDependenciesFileFormat(fullPath)
	if valErr != nil {
		return valErr
	}

	if uploadErr := client.UploadDependencies(groupID, appID, fullPath); uploadErr != nil {
		return fmt.Errorf("failed to import dependencies: %s", uploadErr)
	}

	return nil
}

func findDependenciesArchive(dir string) (string, error) {
	archFile := filepath.Join(dir, "node_modules.*")
	matches, err := filepath.Glob(archFile)

	if err != nil {
		return "", fmt.Errorf("failed to find a node_modules archive in the '%s' directory: %s", dir, err)
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("node_modules archive not found in the '%s' directory", dir)
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("found more than one node_modules archive in the '%s' directory", dir)
	}

	return filepath.Abs(matches[0])
}

func validateDependenciesFileFormat(fullPath string) error {
	ext := path.Ext(fullPath)
	switch ext {
	case extZip, extTar, extGz, extTgz:
		return nil
	default:
		return fmt.Errorf("file '%s' has an unsupported format", fullPath)
	}
}
