package local

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestDependenciesFindNodeModules(t *testing.T) {
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

			deps, err := FindNodeModules(tc.path)
			assert.Nil(t, err)
			assert.Equal(t, Dependencies{
				filepath.Join(absPath, "functions"),
				filepath.Join(absPath, "functions", tc.archiveName),
			}, deps)
		})
	}
}

func TestDependenciesFindPackageJSON(t *testing.T) {
	wd, wdErr := os.Getwd()
	assert.Nil(t, wdErr)

	testRoot := filepath.Join(wd, "testdata/dependencies")

	t.Run("should return an empty data when run outside a project directory", func(t *testing.T) {
		deps, err := FindPackageJSON(testRoot)
		assert.Nil(t, err)
		assert.Equal(t, Dependencies{}, deps)
	})

	t.Run("should return an error when a project has no package.json file", func(t *testing.T) {
		dir := filepath.Join(testRoot, "empty")

		_, err := FindPackageJSON(dir)
		pathError := &fs.PathError{Op: "stat", Path: fmt.Sprintf("%s/functions/package.json", dir), Err: errors.New("no such file or directory")}
		assert.Equal(t, pathError, err.(*fs.PathError))
	})

	t.Run("Should find a package.json", func(t *testing.T) {
		absPath, err := filepath.Abs(filepath.Join(testRoot, "json"))
		assert.Nil(t, err)
		t.Run("with an absolute path", func(t *testing.T) {
			deps, err := FindPackageJSON(filepath.Join(testRoot, "json"))
			assert.Nil(t, err)
			assert.Equal(t, Dependencies{
				filepath.Join(absPath, "functions"),
				filepath.Join(absPath, "functions", "package.json"),
			}, deps)
		})

		t.Run("with a relative path", func(t *testing.T) {
			deps, err := FindPackageJSON("../local/testdata/dependencies/json")
			assert.Nil(t, err)
			assert.Equal(t, Dependencies{
				filepath.Join(absPath, "functions"),
				filepath.Join(absPath, "functions", "package.json"),
			}, deps)
		})
	})
}
