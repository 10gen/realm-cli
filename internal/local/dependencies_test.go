package local

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
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

	t.Run("Should return an empty data when a project has no node_modules archive", func(t *testing.T) {
		deps, err := FindAppDependencies(testRoot)
		assert.Nil(t, err)
		assert.Equal(t, Dependencies{}, deps)
	})

	t.Run("Should return an error when a project has no node_modules archive", func(t *testing.T) {
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
			description: "Should find a a node_modules archive in directory format with an absolute path",
			path:        filepath.Join(testRoot, "dir"),
			archiveName: "node_modules",
		},
		{
			description: "Should find a a node_modules archive in directory format with a relative path",
			path:        "../local/testdata/dependencies/dir",
			archiveName: "node_modules",
		},
		{
			description: "Should find a a node_modules archive in tar format with an absolute path",
			path:        filepath.Join(testRoot, "tar"),
			archiveName: "node_modules.tar",
		},
		{
			description: "Should find a a node_modules archive in tar format with a relative path",
			path:        "../local/testdata/dependencies/tar",
			archiveName: "node_modules.tar",
		},
		{
			description: "Should find a a node_modules archive in tgz format with an absolute path",
			path:        filepath.Join(testRoot, "tgz"),
			archiveName: "node_modules.tar.gz",
		},
		{
			description: "Should find a a node_modules archive in tgz format with a relative path",
			path:        "../local/testdata/dependencies/tgz",
			archiveName: "node_modules.tar.gz",
		},
		{
			description: "Should find a a node_modules archive in zip format with an absolute path",
			path:        filepath.Join(testRoot, "zip"),
			archiveName: "node_modules.zip",
		},
		{
			description: "Should find a a node_modules archive in zip format with a relative path",
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

	testUploadPkg := parseZipPkg(t, testUpload)

	for _, tc := range []struct {
		description string
		rootDir     string
		archiveName string
		test        func(t *testing.T)
	}{
		{
			description: "Should prepare an upload from a zip archive",
			rootDir:     filepath.Join(testRoot, "zip"),
			archiveName: "node_modules.zip",
		},
		{
			description: "Should prepare an upload from a tar archive",
			rootDir:     filepath.Join(testRoot, "tar"),
			archiveName: "node_modules.tar",
		},
		{
			description: "Should prepare an upload from a tgz archive",
			rootDir:     filepath.Join(testRoot, "tgz"),
			archiveName: "node_modules.tar.gz",
		},
		{
			description: "Should prepare an upload from a dir archive",
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

			assert.Equal(t, testUploadPkg, parseZipPkg(t, upload))
		})
	}
}

func parseZipPkg(t *testing.T, file *os.File) map[string]string {
	t.Helper()

	fileInfo, fileInfoErr := file.Stat()
	assert.Nil(t, fileInfoErr)

	zipPkg, zipErr := zip.NewReader(file, fileInfo.Size())
	assert.Nil(t, zipErr)

	out := make(map[string]string)
	for _, file := range zipPkg.File {
		if file.FileInfo().IsDir() {
			continue
		}
		out[file.Name] = parseZipFile(t, file)
	}
	return out
}

func parseZipFile(t *testing.T, file *zip.File) string {
	t.Helper()

	r, openErr := file.Open()
	assert.Nil(t, openErr)

	data, readErr := ioutil.ReadAll(r)
	assert.Nil(t, readErr)

	return string(data)
}
