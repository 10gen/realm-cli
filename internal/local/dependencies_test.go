package local

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"

	"github.com/google/go-cmp/cmp"
)

func TestDependenciesFindNodeModules(t *testing.T) {
	assert.RegisterOpts(reflect.TypeOf(Dependencies{}), cmp.AllowUnexported(Dependencies{}))

	wd, wdErr := os.Getwd()
	assert.Nil(t, wdErr)

	testRoot := filepath.Join(wd, "testdata/dependencies")

	t.Run("should return an empty data when run outside a project directory", func(t *testing.T) {
		deps, err := FindNodeModules(testRoot)
		assert.Nil(t, err)
		assert.Equal(t, Dependencies{}, deps)
	})

	t.Run("should return an error when a project has no node_modules archive", func(t *testing.T) {
		dir := filepath.Join(testRoot, "empty")

		_, err := FindNodeModules(dir)
		assert.Equal(t, fmt.Errorf("node_modules archive not found at '%s/functions'", dir), err)
	})

	for _, tc := range []struct {
		description string
		path        string
		archiveName string
		isDir       bool
	}{
		{
			description: "should find a a node_modules archive in directory format with an absolute path",
			path:        filepath.Join(testRoot, "dir"),
			archiveName: "node_modules",
			isDir:       true,
		},
		{
			description: "should find a a node_modules archive in directory format with a relative path",
			path:        "../local/testdata/dependencies/dir",
			archiveName: "node_modules",
			isDir:       true,
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

			deps, err := FindNodeModules(tc.path)
			assert.Nil(t, err)
			assert.Equal(t, Dependencies{
				filepath.Join(absPath, "functions"),
				filepath.Join(absPath, "functions", tc.archiveName),
				tc.isDir,
			}, deps)
		})
	}
}

func TestDependenciesFindPackageJSON(t *testing.T) {
	assert.RegisterOpts(reflect.TypeOf(Dependencies{}), cmp.AllowUnexported(Dependencies{}))

	wd, wdErr := os.Getwd()
	assert.Nil(t, wdErr)

	testRoot := filepath.Join(wd, "testdata/dependencies")

	t.Run("should return an empty object when run outside a project directory", func(t *testing.T) {
		deps, err := FindPackageJSON(testRoot)
		assert.Nil(t, err)
		assert.Equal(t, Dependencies{}, deps)
	})

	t.Run("should return an error when a project has no package.json file", func(t *testing.T) {
		dir := filepath.Join(testRoot, "empty")

		_, err := FindPackageJSON(dir)
		assert.NotNil(t, err)
		assert.Equal(t,
			err.Error(),
			fmt.Sprintf("package.json not found at '%s/functions'", dir),
		)
	})

	t.Run("should find a package.json", func(t *testing.T) {
		absPath, err := filepath.Abs(filepath.Join(testRoot, "json"))
		assert.Nil(t, err)

		t.Run("with an absolute path", func(t *testing.T) {
			deps, err := FindPackageJSON(filepath.Join(testRoot, "json"))
			assert.Nil(t, err)
			assert.Equal(t, Dependencies{
				filepath.Join(absPath, "functions"),
				filepath.Join(absPath, "functions", "package.json"),
				false,
			}, deps)
		})

		t.Run("with a relative path", func(t *testing.T) {
			deps, err := FindPackageJSON("../local/testdata/dependencies/json")
			assert.Nil(t, err)
			assert.Equal(t, Dependencies{
				filepath.Join(absPath, "functions"),
				filepath.Join(absPath, "functions", "package.json"),
				false,
			}, deps)
		})
	})
}

func TestDependenciesPrepare(t *testing.T) {
	wd, err := os.Getwd()
	assert.Nil(t, err)

	testRoot := filepath.Join(wd, "testdata/dependencies")

	testUpload, err := os.Open(filepath.Join(wd, "testdata/dependencies/node_modules.zip"))
	assert.Nil(t, err)

	testUploadPkg := parseZipPkg(t, testUpload)

	for _, tc := range []struct {
		description         string
		rootDir             string
		isDirectory         bool
		archiveName         string
		expectedArchiveName string
		parseArchive        func(t *testing.T, upload *os.File) map[string]string
	}{
		{
			description: "should prepare an upload from a zip archive",
			rootDir:     filepath.Join(testRoot, "zip"),
			archiveName: "node_modules.zip",
			parseArchive: func(t *testing.T, upload *os.File) map[string]string {
				return parseZipPkg(t, upload)
			},
		},
		{
			description: "should prepare an upload from a tar archive",
			rootDir:     filepath.Join(testRoot, "tar"),
			archiveName: "node_modules.tar",
			parseArchive: func(t *testing.T, upload *os.File) map[string]string {
				return parseTarPkg(t, upload)
			},
		},
		{
			description: "should prepare an upload from a tar.gz archive",
			rootDir:     filepath.Join(testRoot, "tgz"),
			archiveName: "node_modules.tar.gz",
			parseArchive: func(t *testing.T, upload *os.File) map[string]string {
				return parseTgzPkg(t, upload)
			},
		},
		{
			description:         "should prepare an upload from a directory archive",
			rootDir:             filepath.Join(testRoot, "dir"),
			isDirectory:         true,
			archiveName:         "node_modules",
			expectedArchiveName: "node_modules.zip",
			parseArchive: func(t *testing.T, upload *os.File) map[string]string {
				return parseZipPkg(t, upload)
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			dependencies := Dependencies{
				RootDir:     filepath.Join(tc.rootDir, "functions"),
				FilePath:    filepath.Join(tc.rootDir, "functions", tc.archiveName),
				isDirectory: tc.isDirectory,
			}

			uploadPath, cleanup, err := dependencies.PrepareUpload()
			assert.Nil(t, err)
			defer cleanup()

			expectedArchiveName := tc.expectedArchiveName
			if expectedArchiveName == "" {
				expectedArchiveName = tc.archiveName
			}

			assert.True(t, strings.HasSuffix(uploadPath, expectedArchiveName), "should have dependencies archive: "+expectedArchiveName)

			upload, fileErr := os.Open(uploadPath)
			assert.Nil(t, fileErr)
			defer upload.Close()

			uploadPkg := tc.parseArchive(t, upload)
			assert.Equal(t, testUploadPkg, uploadPkg)
		})
	}
}

func parseTarPkg(t *testing.T, file *os.File) map[string]string {
	t.Helper()

	tr := tar.NewReader(file)
	return parseTarFiles(t, tr)
}

func parseTgzPkg(t *testing.T, file *os.File) map[string]string {
	t.Helper()

	gzf, err := gzip.NewReader(file)
	assert.Nil(t, err)

	tr := tar.NewReader(gzf)
	return parseTarFiles(t, tr)
}

func parseTarFiles(t *testing.T, tr *tar.Reader) map[string]string {
	t.Helper()

	out := make(map[string]string)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		assert.Nil(t, err)

		if h.FileInfo().IsDir() {
			continue
		}

		data, err := ioutil.ReadAll(tr)
		assert.Nil(t, err)

		out[h.Name] = string(data)
	}
	return out
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
