package local

import (
	"fmt"
	"path/filepath"
)

// Dependencies holds the data related to a local Realm app's dependencies
type Dependencies struct {
	RootDir  string
	FilePath string
}

// FindAppDependenciesArchive finds the Realm app dependencies archive
func FindAppDependenciesArchive(path string) (Dependencies, error) {
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

// FindPackageJSON finds the Realm app dependencies JSON
func FindPackageJSON(path string) (Dependencies, error) {
	app, appOK, appErr := FindApp(path)
	if appErr != nil {
		return Dependencies{}, appErr
	}
	if !appOK {
		return Dependencies{}, nil
	}

	rootDir := filepath.Join(app.RootDir, NameFunctions)

	JSONs, err := filepath.Glob(filepath.Join(rootDir, nameJSON))
	if err != nil {
		return Dependencies{}, err
	}
	if len(JSONs) == 0 {
		return Dependencies{}, fmt.Errorf("package json not found at '%s'", rootDir)
	}

	JSONPath, err := filepath.Abs(JSONs[0])
	if err != nil {
		return Dependencies{}, err
	}

	return Dependencies{rootDir, JSONPath}, nil
}
