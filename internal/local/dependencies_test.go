package local

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestDependenciesFind(t *testing.T) {
	wd, wdErr := os.Getwd()
	assert.Nil(t, wdErr)

	testRoot := filepath.Join(wd, "testdata/dependencies")

	t.Run("should return an empty data when run outside a project directory", func(t *testing.T) {
		deps, err := FindAppDependencies(testRoot)
		assert.Nil(t, err)
		assert.Equal(t, Dependencies{}, deps)
	})

	t.Run("should return an error when a project has no node_modules archive", func(t *testing.T) {
		dir := filepath.Join(testRoot, "empty")

		_, err := FindAppDependencies(dir)
		assert.Equal(t, fmt.Errorf("node_modules archive not found at '%s/functions'", dir), err)
	})

	for _, tc := range []struct {
		description string
		path        string
		archiveName string
	}{
		{
			description: "should find a a node_modules archive in directory format with an absolute path",
			path:        filepath.Join(testRoot, "dir"),
			archiveName: "node_modules",
		},
		{
			description: "should find a a node_modules archive in directory format with a relative path",
			path:        "../local/testdata/dependencies/dir",
			archiveName: "node_modules",
		},
		{
			description: "should find a a node_modules archive in tar format with an absolute path",
			path:        filepath.Join(testRoot, "tar"),
			archiveName: "node_modules.tar",
		},
		{
			description: "should find a a node_modules archive in tar format with a relative path",
			path:        "../local/testdata/dependencies/tar",
			archiveName: "node_modules.tar",
		},
		{
			description: "should find a a node_modules archive in tgz format with an absolute path",
			path:        filepath.Join(testRoot, "tgz"),
			archiveName: "node_modules.tar.gz",
		},
		{
			description: "should find a a node_modules archive in tgz format with a relative path",
			path:        "../local/testdata/dependencies/tgz",
			archiveName: "node_modules.tar.gz",
		},
		{
			description: "should find a a node_modules archive in zip format with an absolute path",
			path:        filepath.Join(testRoot, "zip"),
			archiveName: "node_modules.zip",
		},
		{
			description: "should find a a node_modules archive in zip format with a relative path",
			path:        "../local/testdata/dependencies/zip",
			archiveName: "node_modules.zip",
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			absPath, err := filepath.Abs(tc.path)
			assert.Nil(t, err)

			deps, err := FindAppDependencies(tc.path)
			assert.Nil(t, err)
			assert.Equal(t, Dependencies{
				filepath.Join(absPath, "functions"),
				filepath.Join(absPath, "functions", tc.archiveName),
			}, deps)
		})
	}
}

func TestDependenciesPrepare(t *testing.T) {
	wd, wdErr := os.Getwd()
	assert.Nil(t, wdErr)

	testRoot := filepath.Join(wd, "testdata/dependencies")

	testUpload, testUploadErr := os.Open(filepath.Join(wd, "testdata/dependencies/upload.zip"))
	assert.Nil(t, testUploadErr)

	testUploadPkg := getZipFileNames(t, testUpload)

	for _, tc := range []struct {
		description string
		rootDir     string
		archiveName string
		test        func(t *testing.T)
	}{
		{
			description: "should prepare an upload from a zip archive",
			rootDir:     filepath.Join(testRoot, "zip"),
			archiveName: "node_modules.zip",
		},
		{
			description: "should prepare an upload from a tar archive",
			rootDir:     filepath.Join(testRoot, "tar"),
			archiveName: "node_modules.tar",
		},
		{
			description: "should prepare an upload from a tgz archive",
			rootDir:     filepath.Join(testRoot, "tgz"),
			archiveName: "node_modules.tar.gz",
		},
		{
			description: "should prepare an upload from a dir archive",
			rootDir:     filepath.Join(testRoot, "dir"),
			archiveName: "node_modules",
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			dependencies := Dependencies{
				RootDir:     filepath.Join(tc.rootDir, "functions"),
				ArchivePath: filepath.Join(tc.rootDir, "functions", tc.archiveName),
			}

			uploadPath, err := dependencies.PrepareUpload()
			assert.Nil(t, err)
			defer os.Remove(uploadPath)

			assert.True(t, strings.HasSuffix(uploadPath, "node_modules.zip"), "should have upload path")

			upload, fileErr := os.Open(uploadPath)
			assert.Nil(t, fileErr)
			defer upload.Close()

			assert.Equal(t, testUploadPkg, getZipFileNames(t, upload))
		})
	}
}

func getZipFileNames(t *testing.T, file *os.File) map[string]bool {
	t.Helper()

	fileInfo, err := file.Stat()
	assert.Nil(t, err)

	zipPkg, err := zip.NewReader(file, fileInfo.Size())
	assert.Nil(t, err)

	fileNames := make(map[string]bool, len(zipPkg.File))

	for _, file := range zipPkg.File {
		if file.FileInfo().IsDir() {
			continue
		}
		fileNames[file.Name] = true
	}
	return fileNames
}
