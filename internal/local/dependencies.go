package local

import (
	"fmt"
	"path/filepath"
)

// Dependencies holds the data related to a local Realm app's dependencies
type Dependencies struct {
	RootDir     string
	ArchivePath string
}

// FindAppDependencies finds the Realm app dependencies archive
func FindAppDependencies(path string) (Dependencies, error) {
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
