package commands

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/10gen/stitch-cli/api"
	"github.com/10gen/stitch-cli/dependency/transpiler"
	"github.com/10gen/stitch-cli/utils"
	"github.com/mitchellh/cli"
)

func ImportDependencies(ui cli.Ui, groupID, appID, dir string, client api.StitchClient) error {
	fullPath, err := findDependenciesLocation(dir)
	if err != nil {
		return err
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return fmt.Errorf("failed to open the dependencies file '%s': %s", fullPath, err)
	}

	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return errors.New("failed to read dependencies from " + fullPath)
	}

	archive, err := utils.NewArchiveReader(file, fullPath, fileInfo.Size())
	if err != nil {
		return err
	}

	tr := transpiler.NewExternalTranspiler(transpiler.DefaultTranspilerCommand)

	outFile, err := os.Create(filepath.Join(os.TempDir(), "node_modules.zip"))
	if err != nil {
		return err
	}
	defer outFile.Close()

	w := zip.NewWriter(outFile)

	fullNames := make([]string, 0)
	sources := make([]string, 0)
	for {
		header, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to advance to the next entry in the archive: %s", err)
		}

		if header.FileInfo().IsDir() {
			continue
		}

		fullpath, err := filepath.Rel(dir, header.FullPath)
		if err != nil {
			// header.FullPath is the relative path already
			fullpath = header.FullPath
		}

		fileContents, err := ioutil.ReadAll(archive)
		if err != nil {
			return fmt.Errorf("failed to read file '%s' in the archive: %s", fullpath, err)
		}

		ext := filepath.Ext(fullpath)
		if ext != ".js" {
			f, err := w.Create(fullpath)
			if err != nil {
				return err
			}
			_, err = f.Write(fileContents)
			if err != nil {
				return err
			}
			continue
		}

		sources = append(sources, string(fileContents))
		fullNames = append(fullNames, fullpath)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ui.Info("transpiling dependencies started.")
	transpiled, err := tr.Transpile(ctx, sources...)
	if err != nil {
		return err
	}
	for i, t := range transpiled {
		f, err := w.Create(fullNames[i])
		if err != nil {
			return err
		}
		_, err = f.Write([]byte(t.Code))
		if err != nil {
			return err
		}
	}
	ui.Info("transpiling dependencies finished.")

	err = w.Close()
	if err != nil {
		return err
	}

	fp, err := filepath.Abs(outFile.Name())
	if err != nil {
		return err
	}

	// clean up after ourselves
	defer os.Remove(fp)

	err = client.UploadDependencies(groupID, appID, fp)
	if err != nil {
		return err
	}

	return nil
}

func findDependenciesLocation(dir string) (string, error) {
	archFile := filepath.Join(dir, "node_modules*")

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
