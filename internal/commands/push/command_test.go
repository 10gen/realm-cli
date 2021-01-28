package push

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
	"github.com/AlecAivazis/survey/v2/terminal"

	"github.com/Netflix/go-expect"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPushSetup(t *testing.T) {
	profile := mock.NewProfile(t)
	profile.SetAtlasBaseURL("http://localhost:8000")
	profile.SetRealmBaseURL("http://localhost:8080")

	cmd := &Command{}
	assert.Nil(t, cmd.atlasClient)
	assert.Nil(t, cmd.realmClient)

	assert.Nil(t, cmd.Setup(profile, nil))
	assert.NotNil(t, cmd.atlasClient)
	assert.NotNil(t, cmd.realmClient)
}

func TestPushHandler(t *testing.T) {
	wd, wdErr := os.Getwd()
	assert.Nil(t, wdErr)

	testApp := local.App{
		RootDir: filepath.Join(wd, "testdata/project"),
		Config:  local.FileConfig,
		AppData: &local.AppConfigJSON{local.AppDataV1{local.AppStructureV1{
			ConfigVersion:        realm.AppConfigVersion20200603,
			ID:                   "eggcorn-abcde",
			Name:                 "eggcorn",
			Location:             realm.LocationVirginia,
			DeploymentModel:      realm.DeploymentModelGlobal,
			Security:             map[string]interface{}{},
			CustomUserDataConfig: map[string]interface{}{"enabled": true},
			Sync:                 map[string]interface{}{"development_mode_enabled": false},
		}}},
	}

	t.Run("Should return an error if the command fails to resolve to", func(t *testing.T) {
		var realmClient mock.RealmClient

		var capturedFilter realm.AppFilter
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			capturedFilter = filter
			return nil, errors.New("something bad happened")
		}

		cmd := &Command{
			inputs:      inputs{AppDirectory: "testdata/project", Project: "groupID", To: "appID"},
			realmClient: realmClient,
		}

		err := cmd.Handler(nil, nil)
		assert.Equal(t, errors.New("something bad happened"), err)

		t.Log("And should properly pass through the expected inputs")
		assert.Equal(t, realm.AppFilter{"groupID", "appID"}, capturedFilter)
	})

	t.Run("Should return an error if the command fails to resolve group id", func(t *testing.T) {
		var atlasClient mock.AtlasClient
		atlasClient.GroupsFn = func() ([]atlas.Group, error) {
			return nil, errors.New("something bad happened")
		}

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return nil, nil
		}

		cmd := &Command{
			inputs:      inputs{AppDirectory: "testdata/project"},
			atlasClient: atlasClient,
			realmClient: realmClient,
		}

		err := cmd.Handler(nil, nil)
		assert.Equal(t, errors.New("something bad happened"), err)
	})

	t.Run("Should return an error if the command fails to create a new app", func(t *testing.T) {
		out := new(bytes.Buffer)
		ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{GroupID: "groupID"}}, nil
		}

		var capturedGroupID, capturedName string
		var capturedMeta realm.AppMeta
		var i int
		realmClient.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			i++
			capturedGroupID = groupID
			capturedName = name
			capturedMeta = meta
			return realm.App{}, errors.New("something bad happened")
		}

		cmd := &Command{
			inputs:      inputs{AppDirectory: "testdata/project", To: "appID"},
			realmClient: realmClient,
		}

		err := cmd.Handler(nil, ui)
		assert.Equal(t, errors.New("something bad happened"), err)

		t.Log("And should properly pass through the expected inputs")
		assert.Equal(t, "groupID", capturedGroupID)
		assert.Equal(t, "eggcorn", capturedName)
		assert.Equal(t, realm.AppMeta{realm.LocationVirginia, realm.DeploymentModelGlobal}, capturedMeta)
	})

	t.Run("Should return an error if the command fails to get the initial diff", func(t *testing.T) {
		_, ui := mock.NewUI()

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", GroupID: "groupID"}}, nil
		}

		var capturedAppData interface{}
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			capturedAppData = appData
			return nil, errors.New("something bad happened")
		}

		cmd := &Command{
			inputs:      inputs{AppDirectory: "testdata/project", To: "appID"},
			realmClient: realmClient,
		}

		err := cmd.Handler(nil, ui)
		assert.Equal(t, errors.New("something bad happened"), err)

		t.Log("And should properly pass through the expected inputs")
		assert.Equal(t, testApp, capturedAppData)
	})

	t.Run("Should return an error if the command fails to create a new draft", func(t *testing.T) {
		out := new(bytes.Buffer)
		ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", GroupID: "groupID"}}, nil
		}
		realmClient.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return realm.App{ID: "appID", GroupID: "groupID"}, nil
		}
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			return []string{"diff1"}, nil
		}

		var capturedGroupID, capturedAppID string
		realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) {
			capturedGroupID = groupID
			capturedAppID = appID
			return realm.AppDraft{}, errors.New("something bad happened")
		}

		cmd := &Command{
			inputs:      inputs{AppDirectory: "testdata/project", To: "appID"},
			realmClient: realmClient,
		}

		err := cmd.Handler(nil, ui)
		assert.Equal(t, errors.New("something bad happened"), err)

		t.Log("And should properly pass through the expected inputs")
		assert.Equal(t, "groupID", capturedGroupID)
		assert.Equal(t, "appID", capturedAppID)
	})

	t.Run("Should return an error if the command fails to import", func(t *testing.T) {
		out := new(bytes.Buffer)
		ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", GroupID: "groupID"}}, nil
		}
		realmClient.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return realm.App{ID: "appID", GroupID: "groupID"}, nil
		}
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			return []string{"diff1"}, nil
		}
		realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) {
			return realm.AppDraft{ID: "draftID"}, nil
		}

		var capturedGroupID, capturedAppID string
		var capturedAppData interface{}
		realmClient.ImportFn = func(groupID, appID string, appData interface{}) error {
			capturedGroupID = groupID
			capturedAppID = appID
			capturedAppData = appData
			return errors.New("something bad happened")
		}

		cmd := &Command{
			inputs:      inputs{AppDirectory: "testdata/project", To: "appID"},
			realmClient: realmClient,
		}

		err := cmd.Handler(nil, ui)
		assert.Equal(t, errors.New("something bad happened"), err)

		t.Log("And should properly pass through the expected inputs")
		assert.Equal(t, "groupID", capturedGroupID)
		assert.Equal(t, "appID", capturedAppID)
		assert.Equal(t, testApp, capturedAppData)
	})

	t.Run("Should return an error if the command fails to deploy the draft", func(t *testing.T) {
		out := new(bytes.Buffer)
		ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", GroupID: "groupID"}}, nil
		}
		realmClient.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return realm.App{ID: "appID", GroupID: "groupID"}, nil
		}
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			return []string{"diff1"}, nil
		}
		realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) {
			return realm.AppDraft{ID: "draftID"}, nil
		}
		realmClient.ImportFn = func(groupID, appID string, appData interface{}) error {
			return nil
		}

		var capturedGroupID, capturedAppID, capturedDraftID string
		realmClient.DeployDraftFn = func(groupID, appID, draftID string) (realm.AppDeployment, error) {
			capturedGroupID = groupID
			capturedAppID = appID
			capturedDraftID = draftID
			return realm.AppDeployment{}, errors.New("something bad happened")
		}

		cmd := &Command{
			inputs:      inputs{AppDirectory: "testdata/project", To: "appID"},
			realmClient: realmClient,
		}

		err := cmd.Handler(nil, ui)
		assert.Equal(t, errors.New("something bad happened"), err)

		t.Log("And should properly pass through the expected inputs")
		assert.Equal(t, "groupID", capturedGroupID)
		assert.Equal(t, "appID", capturedAppID)
		assert.Equal(t, "draftID", capturedDraftID)
	})

	t.Run("Should exit early", func(t *testing.T) {
		for _, tc := range []struct {
			description  string
			groupID      string
			groupsCalled bool
		}{
			{
				description:  "And should fetch group id if to is not resolved",
				groupsCalled: true,
			},
			{
				description: "And should not fetch group id if to is resolved",
				groupID:     "groupID",
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				var atlasClient mock.AtlasClient
				var calledGroups bool
				atlasClient.GroupsFn = func() ([]atlas.Group, error) {
					calledGroups = true
					return []atlas.Group{{ID: "groupID", Name: "groupName"}}, nil
				}

				var realmClient mock.RealmClient
				realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
					return []realm.App{{GroupID: tc.groupID}}, nil
				}

				cmd := &Command{
					inputs:      inputs{AppDirectory: "testdata/project", DryRun: true, To: "appID"},
					atlasClient: atlasClient,
					realmClient: realmClient,
				}

				assert.Nil(t, cmd.Handler(nil, nil))
				assert.False(t, cmd.outputs.appCreated, "should not have created app")
				assert.Equal(t, tc.groupsCalled, calledGroups)
			})
		}
	})

	t.Run("With a user rejecting the option to create a new app", func(t *testing.T) {
		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{GroupID: "groupID"}}, nil
		}

		_, console, _, ui, consoleErr := mock.NewVT10XConsole()
		assert.Nil(t, consoleErr)
		defer console.Close()

		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)

			console.ExpectString("Do you wish to create a new app?")
			console.SendLine("")
			console.ExpectEOF()
		}()

		cmd := &Command{
			inputs:      inputs{AppDirectory: "testdata/project", To: "appID"},
			realmClient: realmClient,
		}

		err := cmd.Handler(nil, ui)

		console.Tty().Close() // flush the writers
		<-doneCh              // wait for procedure to complete

		assert.Nil(t, err)
		assert.False(t, cmd.outputs.appCreated, "should not have created app")
	})

	t.Run("With no diffs generated from the app", func(t *testing.T) {
		_, ui := mock.NewUI()

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", GroupID: "groupID"}}, nil
		}
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			return []string{}, nil
		}

		cmd := &Command{
			inputs:      inputs{AppDirectory: "testdata/project", DryRun: true, To: "appID"},
			realmClient: realmClient,
		}

		assert.Nil(t, cmd.Handler(nil, ui))
		assert.True(t, cmd.outputs.noDiffs, "diff should not exist")
	})

	t.Run("With diffs generated from the app but is a dry run", func(t *testing.T) {
		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", GroupID: "groupID"}}, nil
		}
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			return []string{"diff1", "diff2"}, nil
		}

		out, ui := mock.NewUI()

		cmd := &Command{
			inputs:      inputs{AppDirectory: "testdata/project", DryRun: true, To: "appID"},
			realmClient: realmClient,
		}

		err := cmd.Handler(nil, ui)

		assert.Nil(t, err)
		assert.False(t, cmd.outputs.noDiffs, "diff should exist")
		assert.Equal(t, `01:23:45 UTC INFO  The following reflects the proposed changes to your Realm app
diff1
diff2
`, out.String())
	})

	t.Run("With diffs generated from the app but the user rejects them", func(t *testing.T) {
		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", GroupID: "groupID"}}, nil
		}
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			return []string{"diff1", "diff2"}, nil
		}

		_, console, _, ui, consoleErr := mock.NewVT10XConsole()
		assert.Nil(t, consoleErr)
		defer console.Close()

		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)

			console.ExpectString("Please confirm the changes shown above")
			console.SendLine("")
			console.ExpectEOF()
		}()

		cmd := &Command{
			inputs:      inputs{AppDirectory: "testdata/project", To: "appID"},
			realmClient: realmClient,
		}

		err := cmd.Handler(nil, ui)

		console.Tty().Close() // flush the writers
		<-doneCh              // wait for procedure to complete

		assert.Nil(t, err)
		assert.True(t, cmd.outputs.diffRejected, "diff should be rejected")
	})
}

func TestPushFeedback(t *testing.T) {
	t.Run("Feedback should print a message", func(t *testing.T) {
		for _, tc := range []struct {
			description      string
			inputs           inputs
			outputs          outputs
			expectedContents string
		}{
			{
				description:      "That changes were pushed successfully",
				expectedContents: "01:23:45 UTC INFO  Successfully pushed app changes\n",
			},
			{
				description:      "That there is nothing to do when the diffs between app and draft do not exist",
				outputs:          outputs{noDiffs: true},
				expectedContents: "01:23:45 UTC INFO  Deployed app is identical to proposed version, nothing to do\n",
			},
			{
				description:      "That no changes were pushed when the user rejects the diff",
				outputs:          outputs{diffRejected: true},
				expectedContents: "01:23:45 UTC INFO  No changes were pushed to your Realm application\n",
			},
			{
				description:      "That no changes were pushed when the user chooses to keep its existing draft",
				outputs:          outputs{diffRejected: true},
				expectedContents: "01:23:45 UTC INFO  No changes were pushed to your Realm application\n",
			},
			{
				description: "With a new app created and but in a dry run that the user should remove the dry run flag",
				inputs:      inputs{DryRun: true},
				outputs:     outputs{appCreated: true},
				expectedContents: strings.Join(
					[]string{
						"01:23:45 UTC INFO  This is a new app. To create a new app, you must omit the 'dry-run' flag to proceed",
						"01:23:45 UTC DEBUG Try running instead: realm-cli push\n",
					},
					"\n",
				),
			},
			{
				description:      "With a new app created and but in a dry run that the user should remove the dry run flag",
				outputs:          outputs{appCreated: true},
				expectedContents: "01:23:45 UTC INFO  This is a new app. You must create a new app to proceed\n",
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

func TestPushCommandCreateNewApp(t *testing.T) {
	groupID := "groupID"
	appID := primitive.NewObjectID().Hex()

	fullPkg := &local.AppConfigJSON{local.AppDataV1{local.AppStructureV1{
		Name:            "name",
		Location:        realm.Location("location"),
		DeploymentModel: realm.DeploymentModel("deployment_model"),
	}}}

	t.Run("With a client that successfully creates apps", func(t *testing.T) {
		realmClient := mock.RealmClient{}
		realmClient.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return realm.App{
				GroupID: groupID,
				ID:      appID,
				Name:    name,
				AppMeta: meta,
			}, nil
		}

		t.Run("And a ui that is not set to auto confirm", func(t *testing.T) {
			for _, tc := range []struct {
				description string
				autoConfirm bool
				procedure   func(c *expect.Console)
				expectedApp realm.App
				test        func(t *testing.T, configPath string)
			}{
				{
					description: "Should return empty data if user does not wish to continue",
					procedure: func(c *expect.Console) {
						c.ExpectString("Do you wish to create a new app?")
						c.SendLine("")
						c.ExpectEOF()
					},
					test: func(t *testing.T, configPath string) {
						_, err := os.Stat(configPath)
						assert.True(t, os.IsNotExist(err), "expected config path to not exist, but err was: %s", err)
					},
				},
				{
					description: "Should prompt for all missing app info if user does want to continue",
					procedure: func(c *expect.Console) {
						c.ExpectString("Do you wish to create a new app?")
						c.SendLine("y")

						c.ExpectString("App Name")
						c.SendLine("testApp")

						c.ExpectString("App Location")
						c.Send(string(terminal.KeyArrowDown))
						c.SendLine("")

						c.ExpectString("App Deployment Model")
						c.Send(string(terminal.KeyArrowDown))
						c.SendLine("")

						c.ExpectEOF()
					},
					test: func(t *testing.T, configPath string) {
						configData, readErr := ioutil.ReadFile(configPath)
						assert.Nil(t, readErr)
						assert.Equal(t, `{
    "config_version": 20200603,
    "name": "testApp",
    "location": "US-OR",
    "deployment_model": "LOCAL"
}`, string(configData))
					},
					expectedApp: realm.App{
						ID:      appID,
						GroupID: groupID,
						Name:    "testApp",
						AppMeta: realm.AppMeta{
							Location:        realm.LocationOregon,
							DeploymentModel: realm.DeploymentModelLocal,
						},
					},
				},
				{
					description: "Should still prompt for name with all missing data but auto confirm set to true",
					autoConfirm: true,
					procedure: func(c *expect.Console) {
						c.ExpectString("App Name")
						c.SendLine("testApp")

						c.ExpectEOF()
					},
					expectedApp: realm.App{
						ID:      appID,
						GroupID: groupID,
						Name:    "testApp",
					},
					test: func(t *testing.T, configPath string) {
						configData, readErr := ioutil.ReadFile(configPath)
						assert.Nil(t, readErr)
						assert.Equal(t, `{
    "config_version": 20200603,
    "name": "testApp"
}`, string(configData))
					},
				},
			} {
				t.Run(tc.description, func(t *testing.T) {
					tmpDir, teardown, tmpDirErr := u.NewTempDir("push_handler")
					assert.Nil(t, tmpDirErr)
					defer teardown()

					out := new(bytes.Buffer)
					console, _, ui, consoleErr := mock.NewVT10XConsoleWithOptions(mock.UIOptions{AutoConfirm: tc.autoConfirm}, out)
					assert.Nil(t, consoleErr)
					defer console.Close()

					doneCh := make(chan (struct{}))
					go func() {
						defer close(doneCh)
						tc.procedure(console)
					}()

					cmd := &Command{
						inputs:      inputs{AppDirectory: tmpDir},
						realmClient: realmClient,
					}

					app, err := cmd.createNewApp(ui, groupID, map[string]interface{}{})

					console.Tty().Close() // flush the writers
					<-doneCh              // wait for procedure to complete

					assert.Nil(t, err)
					assert.Equal(t, tc.expectedApp, app)

					tc.test(t, filepath.Join(tmpDir, local.FileConfig.String()))
				})
			}
		})

		t.Run("And a static ui that is set to auto confirm", func(t *testing.T) {
			out := new(bytes.Buffer)
			ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

			for _, tc := range []struct {
				description     string
				appData         interface{}
				expectedAppMeta realm.AppMeta
			}{
				{
					description: "Should use the package name when present and zero values for app meta",
					appData:     local.AppConfigJSON{local.AppDataV1{local.AppStructureV1{Name: "name"}}},
				},
				{
					description:     "Should use the package name location and deployment model when present",
					appData:         fullPkg,
					expectedAppMeta: realm.AppMeta{realm.Location("location"), realm.DeploymentModel("deployment_model")},
				},
			} {
				t.Run(tc.description, func(t *testing.T) {
					tmpDir, teardown, tmpDirErr := u.NewTempDir("push_handler")
					assert.Nil(t, tmpDirErr)
					defer teardown()

					expectedApp := realm.App{
						GroupID: groupID,
						ID:      appID,
						Name:    "name",
						AppMeta: tc.expectedAppMeta,
					}

					cmd := &Command{
						inputs:      inputs{AppDirectory: tmpDir},
						realmClient: realmClient,
					}

					app, err := cmd.createNewApp(ui, "groupID", tc.appData)
					assert.Nil(t, err)
					assert.Equal(t, expectedApp, app)
				})
			}
		})

		t.Run("And an interactive ui that is set to auto confirm", func(t *testing.T) {
			for _, tc := range []struct {
				description     string
				appData         interface{}
				expectedAppMeta realm.AppMeta
			}{
				{
					description:     "Should prompt for name if not present in the package",
					appData:         local.AppConfigJSON{local.AppDataV1{local.AppStructureV1{Location: realm.Location("location"), DeploymentModel: realm.DeploymentModel("deployment_model")}}},
					expectedAppMeta: realm.AppMeta{realm.Location("location"), realm.DeploymentModel("deployment_model")},
				},
				{
					description: "Should not prompt for location and deployment model even if not present in the package",
					appData:     map[string]interface{}{},
				},
			} {
				t.Run(tc.description, func(t *testing.T) {
					tmpDir, teardown, tmpDirErr := u.NewTempDir("push_handler")
					assert.Nil(t, tmpDirErr)
					defer teardown()

					out := new(bytes.Buffer)
					console, _, ui, consoleErr := mock.NewVT10XConsoleWithOptions(mock.UIOptions{AutoConfirm: true}, out)
					assert.Nil(t, consoleErr)
					defer console.Close()

					doneCh := make(chan (struct{}))
					go func() {
						defer close(doneCh)

						console.ExpectString("App Name")
						console.SendLine("test-app")
						console.ExpectEOF()
					}()

					cmd := &Command{
						inputs:      inputs{AppDirectory: tmpDir},
						realmClient: realmClient,
					}

					app, err := cmd.createNewApp(ui, groupID, tc.appData)
					assert.Nil(t, err)

					console.Tty().Close() // flush the writers
					<-doneCh              // wait for procedure to complete

					expectedApp := realm.App{
						GroupID: groupID,
						ID:      appID,
						Name:    "test-app",
						AppMeta: tc.expectedAppMeta,
					}

					assert.Equal(t, expectedApp, app)
				})
			}
		})
	})

	t.Run("With a client that fails to create apps it should return that error", func(t *testing.T) {
		realmClient := mock.RealmClient{}
		realmClient.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return realm.App{}, errors.New("something bad happened")
		}

		out := new(bytes.Buffer)
		ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

		cmd := &Command{realmClient: realmClient}

		_, err := cmd.createNewApp(ui, "groupID", fullPkg)
		assert.Equal(t, errors.New("something bad happened"), err)
	})
}

func TestPushCommandCreateNewDraft(t *testing.T) {
	t.Run("Should create and return the draft when initially successful", func(t *testing.T) {
		groupID, appID := "groupID", "appID"
		testDraft := realm.AppDraft{ID: "id"}

		realmClient := mock.RealmClient{}

		var capturedGroupID, capturedAppID string
		realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) {
			capturedGroupID = groupID
			capturedAppID = appID
			return testDraft, nil
		}

		draft, proceed, err := createNewDraft(nil, realmClient, to{groupID, appID})
		assert.Nil(t, err)
		assert.Equal(t, testDraft, draft)
		assert.True(t, proceed, "expected draft to be created successfully")

		t.Log("And should properly pass through the expected inputs")
		assert.Equal(t, groupID, capturedGroupID)
		assert.Equal(t, appID, capturedAppID)
	})

	t.Run("Should return the error if client fails to create the draft for reasons other than it already exists", func(t *testing.T) {
		realmClient := mock.RealmClient{}
		realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) {
			return realm.AppDraft{}, errors.New("something bad happened while creating a draft")
		}

		_, _, err := createNewDraft(nil, realmClient, to{})
		assert.Equal(t, errors.New("something bad happened while creating a draft"), err)
	})

	t.Run("With a client that fails to create a draft because it already exists", func(t *testing.T) {
		errDraftAlreadyExists := realm.ServerError{Code: realm.ErrCodeDraftAlreadyExists, Message: "a draft already exists"}

		realmClient := mock.RealmClient{}
		realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) {
			return realm.AppDraft{}, errDraftAlreadyExists
		}

		t.Run("And fails to retrieve the existing draft should return the error", func(t *testing.T) {
			realmClient.DraftFn = func(groupID, appID string) (realm.AppDraft, error) {
				return realm.AppDraft{}, errors.New("something bad happened while getting a draft")
			}

			_, _, err := createNewDraft(nil, realmClient, to{})
			assert.Equal(t, errors.New("something bad happened while getting a draft"), err)
		})

		t.Run("And successfully retrieves and diffs the existing draft", func(t *testing.T) {
			draftID := "draftID"

			realmClient.DraftFn = func(groupID, appID string) (realm.AppDraft, error) {
				return realm.AppDraft{ID: draftID}, nil
			}

			realmClient.DiffDraftFn = func(groupID, appID, draftID string) (realm.AppDraftDiff, error) {
				return realm.AppDraftDiff{}, nil
			}

			t.Run("With a ui set to auto-confirm", func(t *testing.T) {
				out := new(bytes.Buffer)
				ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

				t.Run("And a client that fails to discard the draft should return the error", func(t *testing.T) {
					var capturedDraftID string
					realmClient.DiscardDraftFn = func(groupID, appID, draftID string) error {
						capturedDraftID = draftID
						return errors.New("something bad happened while discarding the draft")
					}

					_, _, err := createNewDraft(ui, realmClient, to{})
					assert.Equal(t, errors.New("something bad happened while discarding the draft"), err)

					t.Log("And should properly pass through the expected inputs")
					assert.Equal(t, draftID, capturedDraftID)
				})

				t.Run("And a client that successfully discards the existing draft", func(t *testing.T) {
					realmClient.DiscardDraftFn = func(groupID, appID, draftID string) error {
						return nil
					}

					t.Run("But still fails to create a new draft should return the error", func(t *testing.T) {
						realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) {
							return realm.AppDraft{}, errDraftAlreadyExists
						}

						_, _, err := createNewDraft(ui, realmClient, to{})
						assert.Equal(t, errDraftAlreadyExists, err)
					})

					t.Run("And successfully creates a new draft should be successful", func(t *testing.T) {
						testDraft := realm.AppDraft{ID: "id"}

						realmClient := mock.RealmClient{}

						realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) {
							return testDraft, nil
						}

						draft, proceed, err := createNewDraft(nil, realmClient, to{})
						assert.Nil(t, err)
						assert.Equal(t, testDraft, draft)
						assert.True(t, proceed, "expected draft to be created successfully")
					})
				})
			})

			t.Run("Should prompt the user to accept the diffed changes", func(t *testing.T) {
				t.Run("And mark the draft as kept in command outputs if the user selects no", func(t *testing.T) {
					_, console, _, ui, consoleErr := mock.NewVT10XConsole()
					assert.Nil(t, consoleErr)
					defer console.Close()

					doneCh := make(chan (struct{}))
					go func() {
						defer close(doneCh)

						console.ExpectString("Would you like to discard this draft?")
						console.SendLine("")
						console.ExpectEOF()
					}()

					draft, proceed, err := createNewDraft(ui, realmClient, to{})

					console.Tty().Close() // flush the writers
					<-doneCh              // wait for procedure to complete

					assert.Nil(t, err)
					assert.Equal(t, realm.AppDraft{}, draft)
					assert.False(t, proceed, "expected draft to be rejected")
				})

				t.Run("And return a newly created draft if the user selects yes", func(t *testing.T) {
					testDraft := realm.AppDraft{ID: "id"}

					realmClient.DiscardDraftFn = func(groupID, appID, draftID string) error {
						return nil
					}

					realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) {
						return testDraft, nil
					}

					_, console, _, ui, consoleErr := mock.NewVT10XConsole()
					assert.Nil(t, consoleErr)
					defer console.Close()

					doneCh := make(chan (struct{}))
					go func() {
						defer close(doneCh)

						console.ExpectString("Would you like to discard this draft?")
						console.SendLine("y")
						console.ExpectEOF()
					}()

					draft, proceed, err := createNewDraft(ui, realmClient, to{})

					console.Tty().Close() // flush the writers
					<-doneCh              // wait for procedure to complete

					assert.Nil(t, err)
					assert.Equal(t, testDraft, draft)
					assert.True(t, proceed, "expected draft to be created successfully")
				})
			})
		})
	})
}

func TestPushCommandDiffDraft(t *testing.T) {
	t.Run("With a client that fails to diff the draft should return the error", func(t *testing.T) {
		groupID, appID, draftID := "groupID", "appID", "draftID"

		var realmClient mock.RealmClient

		var capturedGroupID, capturedAppID, capturedDraftID string
		realmClient.DiffDraftFn = func(groupID, appID, draftID string) (realm.AppDraftDiff, error) {
			capturedGroupID = groupID
			capturedAppID = appID
			capturedDraftID = draftID
			return realm.AppDraftDiff{}, errors.New("something bad happened")
		}

		err := diffDraft(nil, realmClient, to{groupID, appID}, draftID)
		assert.Equal(t, errors.New("something bad happened"), err)

		t.Log("And should properly pass through the expected inputs")
		assert.Equal(t, groupID, capturedGroupID)
		assert.Equal(t, appID, capturedAppID)
		assert.Equal(t, draftID, capturedDraftID)
	})

	t.Run("Should print the expected contents", func(t *testing.T) {
		for _, tc := range []struct {
			description      string
			actualDiff       realm.AppDraftDiff
			expectedContents string
		}{
			{
				description:      "With a client that returns an empty diff",
				expectedContents: "01:23:45 UTC INFO  An empty draft already exists for your app\n",
			},
			{
				description: "With a client that returns a minimal diff",
				actualDiff:  realm.AppDraftDiff{Diffs: []string{"diff1", "diff2", "diff3"}},
				expectedContents: strings.Join(
					[]string{
						"01:23:45 UTC INFO  The following draft already exists for your app...",
						"  diff1",
						"  diff2",
						"  diff3\n",
					},
					"\n",
				),
			},
			{
				description: "With a client that returns a full diff",
				actualDiff: realm.AppDraftDiff{
					Diffs: []string{"diff1", "diff2", "diff3"},
					HostingFilesDiff: realm.HostingFilesDiff{
						Added:   []string{"hosting_added1", "hosting_added2"},
						Deleted: []string{"hosting_deleted1"},
					},
					DependenciesDiff: realm.DependenciesDiff{
						Added: []realm.DependencyData{{"dep_added1", "v1"}},
						Modified: []realm.DependencyDiffData{
							{realm.DependencyData{"dep_modified1", "v1"}, "v2"},
							{realm.DependencyData{"dep_modified2", "v2"}, "v1"},
						},
					},
					GraphQLConfigDiff: realm.GraphQLConfigDiff{[]realm.FieldDiff{{"gql_field1", "previous", "updated"}}},
					SchemaOptionsDiff: realm.SchemaOptionsDiff{
						GraphQLValidationDiffs: []realm.FieldDiff{{"gql_validation_field1", "old", "new"}},
						RestValidationDiffs:    []realm.FieldDiff{{"rest_validation_field1", "old", "new"}},
					},
				},
				expectedContents: strings.Join(
					[]string{
						"01:23:45 UTC INFO  The following draft already exists for your app...",
						"  diff1",
						"  diff2",
						"  diff3",
						"01:23:45 UTC INFO  With changes to your static hosting files...",
						"  added: hosting_added1",
						"  added: hosting_added2",
						"  deleted: hosting_deleted1",
						"01:23:45 UTC INFO  With changes to your app dependencies...",
						"  + dep_added1@v1",
						"  dep_modified1@v2 -> dep_modified1@v1",
						"  dep_modified2@v1 -> dep_modified2@v2",
						"01:23:45 UTC INFO  With changes to your GraphQL configuration...",
						"  gql_field1: previous -> updated",
						"01:23:45 UTC INFO  With changes to your app schema...",
						"  gql_validation_field1: old -> new",
						"  rest_validation_field1: old -> new",
						"",
					},
					"\n",
				),
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				var realmClient mock.RealmClient
				realmClient.DiffDraftFn = func(groupID, appID, draftID string) (realm.AppDraftDiff, error) {
					return tc.actualDiff, nil
				}

				out, ui := mock.NewUI()

				assert.Nil(t, diffDraft(ui, realmClient, to{}, ""))
				assert.Equal(t, tc.expectedContents, out.String())
			})
		}
	})
}

func TestPushCommandDeployDraftAndWait(t *testing.T) {
	groupID, appID, draftID := "groupID", "appID", "draftID"
	t.Run("Should return an error with a client that fails to deploy the draft", func(t *testing.T) {
		realmClient := mock.RealmClient{}

		var capturedGroupID, capturedAppID, capturedDraftID string
		realmClient.DeployDraftFn = func(groupID, appID, draftID string) (realm.AppDeployment, error) {
			capturedGroupID = groupID
			capturedAppID = appID
			capturedDraftID = draftID
			return realm.AppDeployment{}, errors.New("something bad happened")
		}

		err := deployDraftAndWait(nil, realmClient, to{groupID, appID}, draftID)
		assert.Equal(t, errors.New("something bad happened"), err)

		t.Log("And should properly pass through the expected inputs")
		assert.Equal(t, groupID, capturedGroupID)
		assert.Equal(t, appID, capturedAppID)
		assert.Equal(t, draftID, capturedDraftID)
	})

	t.Run("With a client that successfully deploys a draft", func(t *testing.T) {
		realmClient := mock.RealmClient{}
		realmClient.DeployDraftFn = func(groupID, appID, draftID string) (realm.AppDeployment, error) {
			return realm.AppDeployment{ID: "id", Status: realm.DeploymentStatusCreated}, nil
		}

		t.Run("But fails to get the deployment", func(t *testing.T) {
			realmClient.DeploymentFn = func(groupID, appID, deploymentID string) (realm.AppDeployment, error) {
				return realm.AppDeployment{}, errors.New("something bad happened")
			}

			for _, tc := range []struct {
				description      string
				discardDraftErr  error
				expectedContents string
			}{
				{
					description:      "Yet can successfully discard the draft should return the error",
					expectedContents: "01:23:45 UTC INFO  Checking on the status of your deployment...\n",
				},
				{
					description:     "And fails to discard the draft should return the deployment error and print a warning message",
					discardDraftErr: errors.New("failed to discard draft"),
					expectedContents: strings.Join(
						[]string{
							"01:23:45 UTC INFO  Checking on the status of your deployment...",
							"01:23:45 UTC WARN  We failed to discard the draft we created for your deployment\n",
						},
						"\n",
					),
				},
			} {
				t.Run(tc.description, func(t *testing.T) {
					realmClient.DiscardDraftFn = func(groupID, appID, draftID string) error {
						return tc.discardDraftErr
					}

					out, ui := mock.NewUI()

					err := deployDraftAndWait(ui, realmClient, to{groupID, appID}, draftID)
					assert.Equal(t, errors.New("something bad happened"), err)
					assert.Equal(t, tc.expectedContents, out.String())
				})
			}
		})

		t.Run("And successfully retrieves the deployment should eventually succeed", func(t *testing.T) {
			var polls int

			realmClient.DeploymentFn = func(groupID, appID, deploymentID string) (realm.AppDeployment, error) {
				status := realm.DeploymentStatusPending
				if polls > 1 {
					status = realm.DeploymentStatusSuccessful
				}
				polls++
				return realm.AppDeployment{ID: deploymentID, Status: status}, nil
			}

			out, ui := mock.NewUI()

			err := deployDraftAndWait(ui, realmClient, to{groupID, appID}, draftID)
			assert.Nil(t, err)

			assert.Equal(t, strings.Repeat("01:23:45 UTC INFO  Checking on the status of your deployment...\n", polls), out.String())
		})
	})
}
