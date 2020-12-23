package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestResolveAppData(t *testing.T) {
	wd, wdErr := os.Getwd()
	assert.Nil(t, wdErr)

	testRoot := wd
	projectRoot := filepath.Join(testRoot, "testdata", "project")

	t.Run("With a working directory outside of the root of a project directory", func(t *testing.T) {
		t.Run("Resolving the app directory should return an empty string", func(t *testing.T) {
			path, insideProject, err := ResolveAppDirectory(testRoot)
			assert.Nil(t, err)
			assert.False(t, insideProject, "expected to be outside project")
			assert.Equal(t, "", path)
		})

		t.Run("Resolving the app data should successfully return empty data", func(t *testing.T) {
			data, err := ResolveAppData(testRoot)
			assert.Nil(t, err)
			assert.Equal(t, AppData{}, data)
		})
	})

	t.Run("With a working directory at the root of a project directory", func(t *testing.T) {
		t.Run("Resolving the app directory should return the working directory", func(t *testing.T) {
			path, insideProject, err := ResolveAppDirectory(projectRoot)
			assert.Nil(t, err)
			assert.True(t, insideProject, "expected to be inside project")
			assert.Equal(t, projectRoot, path)
		})

		t.Run("Resolving the app data should successfully return project data", func(t *testing.T) {
			data, err := ResolveAppData(projectRoot)
			assert.Nil(t, err)
			assert.Equal(t, AppData{ID: "eggcorn-abcde", Name: "eggcorn"}, data)
		})
	})

	t.Run("With a working directory nested deeply inside a project directory", func(t *testing.T) {
		nestedRoot := filepath.Join(projectRoot, "l1", "l2", "l3")

		t.Run("Resolving the app directory should return the working directory", func(t *testing.T) {
			path, insideProject, err := ResolveAppDirectory(nestedRoot)
			assert.Nil(t, err)
			assert.True(t, insideProject, "expected to be inside project")
			assert.Equal(t, projectRoot, path)
		})

		t.Run("Resolving the app data should successfully return project data", func(t *testing.T) {
			data, err := ResolveAppData(nestedRoot)
			assert.Nil(t, err)
			assert.Equal(t, AppData{ID: "eggcorn-abcde", Name: "eggcorn"}, data)
		})

		t.Run("Resolving the app data should return empty data if it exceeds the max search depth", func(t *testing.T) {
			origMaxDirectoryContainSearchDepth := maxDirectoryContainSearchDepth
			maxDirectoryContainSearchDepth = 2
			defer func() { maxDirectoryContainSearchDepth = origMaxDirectoryContainSearchDepth }()

			data, err := ResolveAppData(nestedRoot)
			assert.Nil(t, err)
			assert.Equal(t, AppData{}, data)
		})
	})

	t.Run("Resolving the app data should fail when a project has an empty configuration", func(t *testing.T) {
		emptyProjectRoot := filepath.Join(testRoot, "testdata", "empty_project")

		expectedErr := fmt.Errorf(
			"failed to read app data at %s",
			filepath.Join(emptyProjectRoot, realm.FileAppConfig),
		)

		_, err := ResolveAppData(filepath.Join(emptyProjectRoot, "l1", "l2", "l3"))
		assert.Equal(t, expectedErr, err)
	})
}
