package pull

import (
	"archive/zip"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestPullHandler(t *testing.T) {
	t.Run("should return an error if the command fails to resolve from", func(t *testing.T) {
		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return nil, errors.New("something bad happened")
		}

		cmd := &Command{inputs{From: "somewhere"}}

		err := cmd.Handler(nil, nil, cli.Clients{Realm: realmClient})
		assert.Equal(t, errors.New("something bad happened"), err)
	})

	t.Run("should return an error if the command fails to do the export", func(t *testing.T) {
		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return nil, nil
		}
		realmClient.ExportFn = func(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error) {
			return "", nil, errors.New("something bad happened")
		}

		cmd := &Command{inputs{From: "somewhere"}}

		err := cmd.Handler(nil, nil, cli.Clients{Realm: realmClient})
		assert.Equal(t, errors.New("something bad happened"), err)
	})

	t.Run("with a successful export", func(t *testing.T) {
		zipPkg, zipErr := zip.OpenReader("testdata/test.zip")
		assert.Nil(t, zipErr)

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return nil, nil
		}
		realmClient.ExportFn = func(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error) {
			return "app_20210101", &zipPkg.Reader, nil
		}

		t.Run("should not write any contents to the destination in a dry run", func(t *testing.T) {
			profile := mock.NewProfile(t)

			out, ui := mock.NewUI()

			cmd := &Command{inputs{DryRun: true}}

			assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: realmClient}))
			destination := filepath.Join(profile.WorkingDirectory, "app")

			assert.Equal(t, fmt.Sprintf(`01:23:45 UTC INFO  No changes were written to your file system
01:23:45 UTC DEBUG Contents would have been written to: %s
`, destination), out.String())

			_, err := os.Stat(destination)
			assert.True(t, os.IsNotExist(err), "expected %s to not exist, but instead: %s", err)
		})

		t.Run("should write the received zip package to the destination", func(t *testing.T) {
			profile, teardown := mock.NewProfileFromTmpDir(t, "pull_handler_test")
			defer teardown()

			out, ui := mock.NewUI()

			cmd := &Command{}

			assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: realmClient}))
			destination := filepath.Join(profile.WorkingDirectory, "app")

			assert.Equal(t, "01:23:45 UTC INFO  Successfully pulled down Realm app to your local filesystem\n", out.String())

			_, err := os.Stat(destination)
			assert.Nil(t, err)

			testData, readErr := ioutil.ReadFile(filepath.Join(destination, "test.json"))
			assert.Nil(t, readErr)
			assert.Equal(t, "{\"egg\":\"corn\"}\n", string(testData))
		})
	})
}

func TestPullCommandDoExport(t *testing.T) {
	t.Run("should return an error if the export fails", func(t *testing.T) {
		groupID, appID := "groupID", "appID"

		var realmClient mock.RealmClient

		var capturedGroupID, capturedAppID string
		var capturedExportReq realm.ExportRequest
		realmClient.ExportFn = func(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error) {
			capturedGroupID = groupID
			capturedAppID = appID
			capturedExportReq = req
			return "", nil, errors.New("something bad happened")
		}

		cmd := &Command{inputs{AppVersion: realm.AppConfigVersion20210101}}

		_, _, err := cmd.doExport(nil, realmClient, groupID, appID)
		assert.Equal(t, errors.New("something bad happened"), err)

		t.Log("and should properly pass through the expected args")
		assert.Equal(t, groupID, capturedGroupID)
		assert.Equal(t, appID, capturedAppID)
		assert.Equal(t, realm.ExportRequest{ConfigVersion: realm.AppConfigVersion20210101}, capturedExportReq)
	})

	t.Run("should return the expected destination file path", func(t *testing.T) {
		profile := mock.NewProfile(t)

		for _, tc := range []struct {
			description  string
			targetFlag   string
			zipName      string
			expectedPath string
		}{
			{
				description:  "with a target flag set",
				targetFlag:   "../../my-project",
				expectedPath: "../../my-project",
			},
			{
				description:  "with no target flag set and the zip file name has a timestamp",
				zipName:      "app_20210101",
				expectedPath: filepath.Join(profile.WorkingDirectory, "app"),
			},
			{
				description:  "with no target flag set and the zip file name has no timestamp",
				zipName:      "app-abcde",
				expectedPath: filepath.Join(profile.WorkingDirectory, "app-abcde"),
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				var realmClient mock.RealmClient
				realmClient.ExportFn = func(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error) {
					return tc.zipName, &zip.Reader{}, nil
				}

				cmd := &Command{inputs{Target: tc.targetFlag}}

				path, zipPkg, err := cmd.doExport(profile, realmClient, "", "")
				assert.Nil(t, err)
				assert.NotNil(t, zipPkg)
				assert.Equal(t, tc.expectedPath, path)
			})
		}
	})
}
