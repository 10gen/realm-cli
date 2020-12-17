package cli

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Mkdir makes the directories specified by path
func Mkdir(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory at %s: %w", path, err)
	}
	return nil
}

// WriteFile writes data to the specified filepath
func WriteFile(path string, perm os.FileMode, r io.Reader) error {
	f, openErr := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if openErr != nil {
		return fmt.Errorf("failed to open file at %s: %s", path, openErr)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("failed to write file at %s: %w", path, err)
	}
	return nil
}

// WriteZip writes the zip contents to the specified filepath
func WriteZip(basePath string, zipPkg *zip.Reader) error {
	for _, zipFile := range zipPkg.File {
		path := filepath.Join(basePath, zipFile.Name)

		if zipFile.FileInfo().IsDir() {
			if err := Mkdir(path); err != nil {
				return err
			}
			continue
		}

		data, openErr := zipFile.Open()
		if openErr != nil {
			return openErr
		}
		defer data.Close()

		if err := WriteFile(path, zipFile.Mode(), data); err != nil {
			return err
		}
	}
	return nil
}
