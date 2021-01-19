package pull

import (
	"archive/zip"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestPullSetup(t *testing.T) {
	profile := mock.NewProfile(t)
	profile.SetRealmBaseURL("http://localhost:8080")

	cmd := &Command{}
	assert.Nil(t, cmd.realmClient)

	assert.Nil(t, cmd.Setup(profile, nil))
	assert.NotNil(t, cmd.realmClient)
}

func TestPullHandler(t *testing.T) {
	t.Run("should return an error if the command fails to resolve from", func(t *testing.T) {
		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return nil, errors.New("something bad happened")
		}

		cmd := &Command{
			inputs:      inputs{From: "somewhere"},
			realmClient: realmClient,
		}

		err := cmd.Handler(nil, nil)
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

		cmd := &Command{
			inputs:      inputs{From: "somewhere"},
			realmClient: realmClient,
		}

		err := cmd.Handler(nil, nil)
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

		for _, tc := range []struct {
			description string
			dryRun      bool
			test        func(t *testing.T, destination string)
		}{
			{
				description: "should not write any contents to the destination in a dry run",
				dryRun:      true,
				test: func(t *testing.T, destination string) {
					_, err := os.Stat(destination)
					assert.True(t, os.IsNotExist(err), "expected %s to not exist, but instead: %s", err)
				},
			},
			{
				description: "should write the received zip package to the destination",
				test: func(t *testing.T, destination string) {
					_, err := os.Stat(destination)
					assert.Nil(t, err)

					testData, readErr := ioutil.ReadFile(filepath.Join(destination, "test.json"))
					assert.Nil(t, readErr)
					assert.Equal(t, `{"egg":"corn"}`, string(testData))
				},
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				profile, teardown := mock.NewProfileFromTmpDir(t, "pull_handler_test")
				defer teardown()

				cmd := &Command{
					inputs:      inputs{DryRun: tc.dryRun},
					realmClient: realmClient,
				}

				assert.Nil(t, cmd.Handler(profile, nil))
				assert.Equal(t, filepath.Join(profile.WorkingDirectory, "app"), cmd.outputs.destination)
			})
		}

		t.Run("should not write any contents to the destination in a dry run", func(t *testing.T) {
			profile := mock.NewProfile(t)

			cmd := &Command{
				inputs:      inputs{DryRun: true},
				realmClient: realmClient,
			}

			assert.Nil(t, cmd.Handler(profile, nil))
			assert.Equal(t, filepath.Join(profile.WorkingDirectory, "app"), cmd.outputs.destination)
		})

		t.Run("should write the received zip package to the destination", func(t *testing.T) {
			profile, teardown := mock.NewProfileFromTmpDir(t, "pull_handler_test")
			defer teardown()

			cmd := &Command{
				inputs:      inputs{DryRun: true},
				realmClient: realmClient,
			}

			assert.Nil(t, cmd.Handler(profile, nil))
			assert.Equal(t, filepath.Join(profile.WorkingDirectory, "app"), cmd.outputs.destination)
		})
	})
}

func TestPullFeedback(t *testing.T) {
	t.Run("feedback should print a message", func(t *testing.T) {
		for _, tc := range []struct {
			description      string
			inputs           inputs
			outputs          outputs
			expectedContents string
		}{
			{
				description:      "that changes were pulled down successfully",
				expectedContents: "01:23:45 UTC INFO  Successfully pulled down Realm app to your local filesystem\n",
			},
			{
				description: "with a new app created and but in a dry run that the user should remove the dry run flag",
				inputs:      inputs{DryRun: true},
				outputs:     outputs{destination: "/dev/tmp"},
				expectedContents: strings.Join(
					[]string{
						"01:23:45 UTC INFO  No changes were written to your file system",
						"01:23:45 UTC DEBUG Contents would have been written to: /dev/tmp\n",
					},
					"\n",
				),
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				out, ui := mock.NewUI()

				cmd := &Command{inputs: tc.inputs, outputs: tc.outputs}

				err := cmd.Feedback(nil, ui)
				assert.Nil(t, err)

				assert.Equal(t, tc.expectedContents, out.String())
			})
		}
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

		cmd := &Command{
			inputs:      inputs{AppVersion: realm.AppConfigVersion20210101},
			realmClient: realmClient,
		}

		_, _, err := cmd.doExport(nil, from{groupID, appID})
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

				cmd := &Command{
					inputs:      inputs{Target: tc.targetFlag},
					realmClient: realmClient,
				}

				path, zipPkg, err := cmd.doExport(profile, from{})
				assert.Nil(t, err)
				assert.NotNil(t, zipPkg)
				assert.Equal(t, tc.expectedPath, path)
			})
		}
	})
}
