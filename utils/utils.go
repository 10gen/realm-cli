package utils

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// WriteZipToDir takes a destination and an io.Reader containing zip data and unpacks it
func WriteZipToDir(dest string, zipData io.Reader) error {
	b, err := ioutil.ReadAll(zipData)
	if err != nil {
		return err
	}

	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return err
	}

	err = os.MkdirAll(dest, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %s", dest, err)
	}

	for _, zipFile := range r.File {
		if err := processFile(filepath.Join(dest, zipFile.Name), zipFile); err != nil {
			return err
		}
	}

	return err
}

func processFile(path string, zipFile *zip.File) error {
	fileData, err := zipFile.Open()
	if err != nil {
		return fmt.Errorf("failed to extract file %s: %s", path, err)
	}
	defer fileData.Close()

	if zipFile.FileInfo().IsDir() {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create sub-directory %s: %s", path, err)
		}
	} else {
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zipFile.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file %s: %s", path, err)
		}
		defer f.Close()

		_, err = io.Copy(f, fileData)
		if err != nil {
			return fmt.Errorf("failed to extract file %s: %s", path, err)
		}
	}

	return nil
}
