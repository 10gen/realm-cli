package local

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/briandowns/spinner"
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

	archives, archivesErr := filepath.Glob(filepath.Join(rootDir, "node_modules*"))
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

// PrepareUpload prepares a dependencies upload package by creating a .zip file
// containing the specified archive's transpiled file contents in a tempmorary directory
// and returns that file path
func (d Dependencies) PrepareUpload() (string, error) {
	file, err := os.Open(d.ArchivePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	archive, err := newArchiveReader(d.ArchivePath, file)
	if err != nil {
		return "", err
	}

	transpiler, err := newDefaultTranspiler()
	if err != nil {
		return "", err
	}

	dir, err := ioutil.TempDir("", "") // uses os.TempDir and guarantees existence and proper permissions
	if err != nil {
		return "", err
	}

	out, err := os.Create(filepath.Join(dir, "node_modules.zip"))
	if err != nil {
		return "", err
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

		path, err := filepath.Rel(d.RootDir, h.Path)
		if err != nil {
			path = h.Path // h.Path _is_ the relative path already
		}

		data, err := ioutil.ReadAll(archive)
		if err != nil {
			return "", err
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

	transpiledSources, err := transpiler.Transpile(ctx, jsSources...)
	if err != nil {
		return "", err
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

func PrepareDependencies(app App, ui terminal.UI) (string, error) {
	dependencies, err := FindAppDependencies(app.RootDir)
	if err != nil {
		return "", err
	}

	s := spinner.New(terminal.SpinnerCircles, 250*time.Millisecond)
	s.Suffix = " Transpiling dependency sources..."

	prepareUpload := func() (string, error) {
		s.Start()
		defer s.Stop()

		path, err := dependencies.PrepareUpload()
		if err != nil {
			return "", err
		}

		ui.Print(terminal.NewTextLog("Transpiled dependency sources"))
		return path, nil
	}

	return prepareUpload()
}
