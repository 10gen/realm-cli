package realm_test

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestRealmDependencies(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	client := newAuthClient(t)

	groupID := u.CloudGroupID()

	app, teardown := setupTestApp(t, client, groupID, "importexport-test")
	defer teardown()

	t.Run("should successfully import a zip node_modules archive", func(t *testing.T) {
		wd, err := os.Getwd()
		assert.Nil(t, err)

		uploadPath := filepath.Join(wd, "testdata/dependencies_upload.zip")

		_, err = client.GetDependenciesStatus(groupID, app.ID)
		assert.NotNilf(t, err, "dependency status must return error before import")
		assert.Nil(t, client.ImportDependencies(groupID, app.ID, uploadPath))

		status, err := client.GetDependenciesStatus(groupID, app.ID)
		assert.Equalf(t, status, realm.DependenciesStatusSuccessful, "must wait until dependency status is successful during import")
		assert.Nil(t, err)

		t.Run("and wait for those dependencies to be deployed to the app", func(t *testing.T) {
			deployments, err := client.Deployments(groupID, app.ID)
			assert.Nil(t, err)

			if len(deployments) == 0 {
				return // no pending jobs to wait for, tests can proceed
			}

			var counter int
			deployment := deployments[0]
			for deployment.Status == realm.DeploymentStatusCreated || deployment.Status == realm.DeploymentStatusPending {
				if counter > 100 {
					t.Fatal("failed to wait for dependencies to deploy")
				}
				if counter%10 == 0 {
					t.Logf("waiting for deployment (id: %s, current status: %s)", deployment.ID, deployment.Status)
				}
				time.Sleep(time.Second)

				deployment, err = client.Deployment(groupID, app.ID, deployment.ID)
				assert.Nil(t, err)

				counter++
			}
		})
	})

	t.Run("should successfully export a zip node_modules archive", func(t *testing.T) {
		tmpDir, teardown, tmpDirErr := u.NewTempDir("dependencies")
		assert.Nil(t, tmpDirErr)
		defer teardown()

		name, zipPkg, err := client.ExportDependencies(groupID, app.ID)
		assert.Nil(t, err)

		assert.Equal(t, "node_modules.zip", name)

		assert.Nil(t, local.WriteFile(filepath.Join(tmpDir, name), 0666, zipPkg))

		actualDeps, actualDepsErr := os.Open(filepath.Join(tmpDir, name))
		assert.Nil(t, actualDepsErr)
		defer actualDeps.Close()

		t.Run("and it should match the expected zip archive", func(t *testing.T) {
			expectedDeps, expectedDepsErr := os.Open("./testdata/dependencies_upload.zip")
			assert.Nil(t, expectedDepsErr)
			defer expectedDeps.Close()

			// make sure same files are present; contents may be different because of transpilation
			assert.Equal(t, getZipFileNames(t, expectedDeps), getZipFileNames(t, actualDeps))
		})
	})

	t.Run("should successfully diff a zip node_modules archive", func(t *testing.T) {
		wd, err := os.Getwd()
		assert.Nil(t, err)

		uploadPath := filepath.Join(wd, "testdata/dependencies_diff.zip")
		diff, err := client.DiffDependencies(groupID, app.ID, uploadPath)
		assert.Nil(t, err)
		assert.Equal(t, realm.DependenciesDiff{
			Added:    []realm.DependencyData{{"twilio", "3.35.1"}},
			Deleted:  []realm.DependencyData{{"debug", "4.3.1"}},
			Modified: []realm.DependencyDiffData{{DependencyData: realm.DependencyData{"underscore", "1.9.2"}, PreviousVersion: "1.9.1"}},
		}, diff)
	})
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
