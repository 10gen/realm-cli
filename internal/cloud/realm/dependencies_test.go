package realm_test

import (
	"archive/zip"
	"encoding/json"
	"io/ioutil"
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

	t.Run("with dependencies as a archive", func(t *testing.T) {
		client := newAuthClient(t)

		groupID := u.CloudGroupID()

		app, teardown := setupTestApp(t, client, groupID, "importexport-test")
		defer teardown()

		t.Run("should successfully import a zip node_modules archive", func(t *testing.T) {
			wd, err := os.Getwd()
			assert.Nil(t, err)

			uploadPath := filepath.Join(wd, "testdata/dependencies_upload.zip")

			_, err = client.DependenciesStatus(groupID, app.ID)
			assert.Nil(t, err)

			assert.Nil(t, client.ImportDependencies(groupID, app.ID, uploadPath))

			startStatus, err := client.DependenciesStatus(groupID, app.ID)
			assert.Nil(t, err)
			assert.Equal(t, realm.DependenciesStatus{
				State:   realm.DependenciesStateCreated,
				Message: "processing request",
			}, startStatus)

			// wait for dependencies to finish installation
			var currentStatus realm.DependenciesStatus
			for i := 0; i < 50; i++ {
				var err error
				currentStatus, err = client.DependenciesStatus(groupID, app.ID)
				assert.Nil(t, err)
				if currentStatus.State != realm.DependenciesStateCreated {
					break
				}
				time.Sleep(time.Second)
			}
			assert.Equal(t, realm.DependenciesStateSuccessful, currentStatus.State)

			t.Run("and wait for those dependencies to be deployed to the app", func(t *testing.T) {
				waitForDepDeployment(client, groupID, app.ID, t)
			})
		})

		t.Run("should successfully export a zip node_modules archive", func(t *testing.T) {
			tmpDir, teardown, tmpDirErr := u.NewTempDir("dependencies")
			assert.Nil(t, tmpDirErr)
			defer teardown()

			name, zipPkg, err := client.ExportDependenciesArchive(groupID, app.ID)
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
	})

	t.Run("with dependencies as package.json", func(t *testing.T) {
		client := newAuthClient(t)

		groupID := u.CloudGroupID()

		app, teardown := setupTestApp(t, client, groupID, "importexport-test")
		defer teardown()

		t.Run("should successfully import a package.json", func(t *testing.T) {
			wd, err := os.Getwd()
			assert.Nil(t, err)

			uploadPath := filepath.Join(wd, "testdata/package.json")

			_, err = client.DependenciesStatus(groupID, app.ID)
			assert.Nil(t, err)

			assert.Nil(t, client.ImportDependencies(groupID, app.ID, uploadPath))

			startStatus, err := client.DependenciesStatus(groupID, app.ID)
			assert.Nil(t, err)
			assert.Equal(t, realm.DependenciesStatus{
				State:   realm.DependenciesStateCreated,
				Message: "processing request",
			}, startStatus)

			// wait for dependencies to finish installation
			var currentStatus realm.DependenciesStatus
			for i := 0; i < 150; i++ {
				var err error
				currentStatus, err = client.DependenciesStatus(groupID, app.ID)
				assert.Nil(t, err)
				if currentStatus.State != realm.DependenciesStateCreated {
					break
				}
				time.Sleep(time.Second)
			}
			assert.Equal(t, realm.DependenciesStateSuccessful, currentStatus.State)

			t.Run("and wait for those dependencies to be deployed to the app", func(t *testing.T) {
				waitForDepDeployment(client, groupID, app.ID, t)
			})
		})

		t.Run("should successfully export a package.json", func(t *testing.T) {
			tmpDir, teardown, tmpDirErr := u.NewTempDir("dependencies")
			assert.Nil(t, tmpDirErr)
			defer teardown()

			name, pkgJSON, err := client.ExportDependencies(groupID, app.ID)
			assert.Nil(t, err)

			assert.Equal(t, "package.json", name)

			assert.Nil(t, local.WriteFile(filepath.Join(tmpDir, name), 0666, pkgJSON))

			actualStr, actualDepsErr := ioutil.ReadFile(filepath.Join(tmpDir, name))
			assert.Nil(t, actualDepsErr)
			var actualDeps map[string]struct{}
			assert.Nil(t, json.Unmarshal(actualStr, &actualDeps))

			t.Run("and it should match the expected package.json", func(t *testing.T) {
				expectedStr, expectedDepsErr := ioutil.ReadFile("./testdata/package.json")
				assert.Nil(t, expectedDepsErr)
				var expectedDeps map[string]struct{}
				assert.Nil(t, json.Unmarshal(expectedStr, &expectedDeps))

				// make sure same files are present
				assert.Equal(t, expectedDeps, actualDeps)
			})
		})

		t.Run("should successfully diff a package.json", func(t *testing.T) {
			wd, err := os.Getwd()
			assert.Nil(t, err)

			uploadPath := filepath.Join(wd, "testdata/updated-package.json")
			diff, err := client.DiffDependencies(groupID, app.ID, uploadPath)
			assert.Nil(t, err)
			assert.Equal(t, realm.DependenciesDiff{
				Added:    []realm.DependencyData{{"twilio", "3.35.1"}},
				Deleted:  []realm.DependencyData{{"debug", "4.3.1"}},
				Modified: []realm.DependencyDiffData{{DependencyData: realm.DependencyData{"underscore", "1.9.2"}, PreviousVersion: "1.9.1"}},
			}, diff)
		})
	})
}

func waitForDepDeployment(client realm.Client, groupID string, appID string, t *testing.T) {
	deployments, err := client.Deployments(groupID, appID)
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

		deployment, err = client.Deployment(groupID, appID, deployment.ID)
		assert.Nil(t, err)

		counter++
	}
	assert.Equal(t, realm.DeploymentStatusSuccessful, deployment.Status)
}

func getZipFileNames(t *testing.T, file *os.File) map[string]struct{} {
	t.Helper()

	fileInfo, err := file.Stat()
	assert.Nil(t, err)

	zipPkg, err := zip.NewReader(file, fileInfo.Size())
	assert.Nil(t, err)

	out := make(map[string]struct{}, len(zipPkg.File))

	for _, file := range zipPkg.File {
		if file.FileInfo().IsDir() {
			continue
		}
		out[file.Name] = struct{}{}
	}
	return out
}
