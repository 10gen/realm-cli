package local

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Dependencies holds the data related to a local Realm app's dependencies
type Dependencies struct {
	RootDir     string
	FilePath    string
	isDirectory bool
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

	archive, err := selectArchive(archives)
	if err != nil {
		return Dependencies{}, fmt.Errorf("failed to find supported archive file at %s: %s", rootDir, err)
	}

	archivePath, archivePathErr := filepath.Abs(archive.path)
	if archivePathErr != nil {
		return Dependencies{}, archivePathErr
	}

	return Dependencies{rootDir, archivePath, archive.isDir}, nil
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

	if _, err := os.Stat(packageJSONPath); err != nil {
		if os.IsNotExist(err) {
			return Dependencies{}, fmt.Errorf("package.json not found at '%s'", rootDir)
		}
		return Dependencies{}, err
	}

	return Dependencies{rootDir, packageJSONPath, false}, nil
}

// PrepareUpload will prepare the dependencies for upload and returns the file path
// for the artifact to be uploaded along with a callback that will perform any cleanup
// required for that upload artifact
func (d Dependencies) PrepareUpload() (string, func(), error) {
	if !d.isDirectory {
		return d.FilePath, func() {}, nil
	}

	// a node_modules directory needs to be written to a temporary .zip and cleaned up afterwards

	dir, err := os.Open(d.FilePath)
	if err != nil {
		return "", func() {}, err
	}
	defer dir.Close()

	r, err := newDirReader(d.FilePath)
	if err != nil {
		return "", func() {}, err
	}

	tmpDir, err := ioutil.TempDir("", "") // uses os.TempDir and guarantees existence and proper permissions
	if err != nil {
		return "", func() {}, err
	}

	out, err := os.Create(filepath.Join(tmpDir, nameNodeModules+extZip))
	if err != nil {
		return "", func() {}, err
	}
	defer out.Close()

	w := zip.NewWriter(out)

	writeFile := func(path string, data []byte) error {
		file, err := w.Create(path)
		if err != nil {
			return err
		}
		if _, err := file.Write(data); err != nil {
			return err
		}
		return nil
	}

	for {
		h, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", func() {}, err
		}

		if h.info.IsDir() {
			continue
		}

		path, err := filepath.Rel(d.RootDir, h.path)
		if err != nil {
			path = h.path // h.path _is_ the relative path already
		}

		data, err := ioutil.ReadAll(r)
		if err != nil {
			return "", func() {}, err
		}

		if err := writeFile(path, data); err != nil {
			return "", func() {}, err
		}
	}

	if err := w.Close(); err != nil {
		return "", func() {}, err
	}

	tmpZip, err := filepath.Abs(out.Name())
	if err != nil {
		return "", func() {}, err
	}
	return tmpZip, func() {
		// TODO(REALMC-8369): remove the nolint directive once errcheck is fixed
		os.Remove(tmpZip) //nolint: errcheck
	}, nil
}
