package local

import (
	"fmt"
	"os"
	"path/filepath"
)

// Dependencies holds the data related to a local Realm app's dependencies
type Dependencies struct {
	RootDir  string
	FilePath string
}

// FindNodeModules finds the Realm app dependencies as a node_modules archive
func FindNodeModules(path string) (Dependencies, error) {
	app, appOK, appErr := FindApp(path)
	if appErr != nil {
		return Dependencies{}, appErr
	}
	if !appOK {
		return Dependencies{}, nil
	}

	rootDir := filepath.Join(app.RootDir, NameFunctions)

	archives, archivesErr := filepath.Glob(filepath.Join(rootDir, nameNodeModules+"*"))
	if archivesErr != nil {
		return Dependencies{}, archivesErr
	}
	if len(archives) == 0 {
		return Dependencies{}, fmt.Errorf("node_modules archive not found at '%s'", rootDir)
	}

	archivePath, archivePathErr := filepath.Abs(archives[0])
	if archivePathErr != nil {
		return Dependencies{}, archivePathErr
	}

	return Dependencies{rootDir, archivePath}, nil
}

// FindPackageJSON finds the Realm app dependencies as a package.json file
func FindPackageJSON(path string) (Dependencies, error) {
	app, appOK, appErr := FindApp(path)
	if appErr != nil {
		return Dependencies{}, appErr
	}
	if !appOK {
		return Dependencies{}, nil
	}

	rootDir := filepath.Join(app.RootDir, NameFunctions)

	packageJSONPath := filepath.Join(rootDir, NamePackageJSON)

	_, err := os.Stat(packageJSONPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Dependencies{}, fmt.Errorf("package.json not found at '%s'", rootDir)
		}
		return Dependencies{}, err
	}

	return Dependencies{rootDir, packageJSONPath}, nil
}
