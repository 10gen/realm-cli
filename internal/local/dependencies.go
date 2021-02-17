package local

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
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

	dir, dirErr := filepath.Abs(rootDir)
	if dirErr != nil {
		return Dependencies{}, dirErr
	}

	archives, archivesErr := filepath.Glob(filepath.Join(dir, "node_modules*"))
	if archivesErr != nil {
		return Dependencies{}, archivesErr
	}
	if len(archives) == 0 {
		return Dependencies{}, fmt.Errorf("node_modules archive not found at '%s'", dir)
	}

	archivePath, archivePathErr := filepath.Abs(archives[0])
	if archivePathErr != nil {
		return Dependencies{}, archivePathErr
	}

	return Dependencies{RootDir: rootDir, ArchivePath: archivePath}, nil
}

// PrepareUpload prepares a dependencies upload package by creating a .zip file
// containing the specified archive's transpiled file contents in a tempmorary directory
// and returns that file path
func (d Dependencies) PrepareUpload() (string, error) {
	file, fileErr := os.Open(d.ArchivePath)
	if fileErr != nil {
		return "", fileErr
	}
	defer file.Close()

	archive, archiveErr := newArchiveReader(d.ArchivePath, file)
	if archiveErr != nil {
		return "", archiveErr
	}

	transpiler, transpilerErr := newDefaultTranspiler()
	if transpilerErr != nil {
		return "", transpilerErr
	}

	out, outErr := os.Create(filepath.Join(os.TempDir(), "node_modules.zip"))
	if outErr != nil {
		return "", outErr
	}
	defer out.Close()

	w := zip.NewWriter(out)

	writeFile := func(path string, data []byte) error {
		file, fileErr := w.Create(path)
		if fileErr != nil {
			return fileErr
		}
		if _, err := file.Write(data); err != nil {
			return err
		}
		return nil
	}

	var jsPaths, jsSources []string
	for {
		h, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		if h.Info.IsDir() {
			continue
		}

		path, pathErr := filepath.Rel(d.RootDir, h.Path)
		if pathErr != nil {
			path = h.Path // h.Path _is_ the relative path already
		}

		data, dataErr := ioutil.ReadAll(archive)
		if dataErr != nil {
			return "", dataErr
		}

		// separate out javascript and non-javascript
		if filepath.Ext(path) == extJS {
			jsPaths = append(jsPaths, path)
			jsSources = append(jsSources, string(data))
			continue
		}

		if err := writeFile(path, data); err != nil {
			return "", err
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	transpiledSources, transpilerErr := transpiler.Transpile(ctx, jsSources...)
	if transpilerErr != nil {
		return "", transpilerErr
	}

	for i, transpiledSource := range transpiledSources {
		if err := writeFile(jsPaths[i], []byte(transpiledSource)); err != nil {
			return "", err
		}
	}

	if err := w.Close(); err != nil {
		return "", err
	}
	return filepath.Abs(out.Name())
}
