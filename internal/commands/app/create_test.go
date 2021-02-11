package app

import (
	"archive/zip"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"github.com/Netflix/go-expect"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAppCreateSetup(t *testing.T) {
	t.Run("should construct a realm and atlas client with configured base urls", func(t *testing.T) {
		profile := mock.NewProfile(t)
		profile.SetRealmBaseURL("http://localhost:8080")
		profile.SetAtlasBaseURL("http://localhost:8888")

		cmd := &CommandCreate{inputs: createInputs{newAppInputs: newAppInputs{
			Name: "test-app",
		}}}
		assert.Nil(t, cmd.realmClient)
		assert.Nil(t, cmd.atlasClient)
		assert.Nil(t, cmd.Setup(profile, nil))
		assert.NotNil(t, cmd.realmClient)
		assert.NotNil(t, cmd.atlasClient)
	})
}

func TestAppCreateHandler(t *testing.T) {
	t.Run("should create minimal project when no from type is specified", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()

		var createdApp realm.App
		client := mock.RealmClient{}
		client.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			createdApp = realm.App{
				GroupID: groupID,
				Name:    name,
				AppMeta: meta,
			}
			return createdApp, nil
		}
		client.ImportFn = func(groupID, appID string, appData interface{}) error {
			return nil
		}

		cmd := &CommandCreate{
			inputs: createInputs{newAppInputs: newAppInputs{
				Name:            "test-app",
				Project:         "test-project",
				Location:        realm.LocationVirginia,
				DeploymentModel: realm.DeploymentModelGlobal,
			}},
			realmClient: client,
		}

		assert.Nil(t, cmd.Handler(profile, nil))

		localApp, err := local.LoadApp(filepath.Join(profile.WorkingDirectory, cmd.inputs.Name))
		assert.Nil(t, err)

		expectedAppData := local.AppRealmConfigJSON{local.AppDataV2{local.AppStructureV2{
			ConfigVersion:   realm.DefaultAppConfigVersion,
			Name:            "test-app",
			Location:        realm.LocationVirginia,
			DeploymentModel: realm.DeploymentModelGlobal,
		}}}

		assert.Equal(t, &expectedAppData, localApp.AppData)
		assert.Equal(t, realm.App{
			GroupID: "test-project",
			Name:    "test-app",
			AppMeta: realm.AppMeta{
				Location:        realm.LocationVirginia,
				DeploymentModel: realm.DeploymentModelGlobal,
			},
		}, createdApp)
	})

	t.Run("when from and project is not set should create minimal project and prompt for project", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()

		var createdApp realm.App
		rc := mock.RealmClient{}
		rc.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			createdApp = realm.App{
				GroupID: groupID,
				Name:    name,
				AppMeta: meta,
			}
			return createdApp, nil
		}
		rc.ImportFn = func(groupID, appID string, appData interface{}) error {
			return nil
		}
		ac := mock.AtlasClient{}
		ac.GroupsFn = func() ([]atlas.Group, error) {
			return []atlas.Group{{ID: "test-project"}}, nil
		}

		procedure := func(c *expect.Console) {
			c.ExpectString("Atlas Project")
			c.Send("test-project")
			c.SendLine(" ")
			c.ExpectEOF()
		}

		_, console, _, ui, consoleErr := mock.NewVT10XConsole()
		assert.Nil(t, consoleErr)
		defer console.Close()

		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)
			procedure(console)
		}()

		cmd := &CommandCreate{
			inputs: createInputs{newAppInputs: newAppInputs{
				Name:            "test-app",
				Project:         "test-project",
				Location:        realm.LocationVirginia,
				DeploymentModel: realm.DeploymentModelGlobal,
			}},
			realmClient: rc,
			atlasClient: ac,
		}

		assert.Nil(t, cmd.Handler(profile, ui))

		console.Tty().Close() // flush the writers
		<-doneCh              // wait for procedure to complete

		localApp, err := local.LoadApp(filepath.Join(profile.WorkingDirectory, cmd.inputs.Name))
		assert.Nil(t, err)

		expectedAppData := local.AppRealmConfigJSON{local.AppDataV2{local.AppStructureV2{
			ConfigVersion:   realm.DefaultAppConfigVersion,
			Name:            "test-app",
			Location:        realm.LocationVirginia,
			DeploymentModel: realm.DeploymentModelGlobal,
		}}}

		assert.Equal(t, &expectedAppData, localApp.AppData)
		assert.Equal(t, realm.App{
			GroupID: "test-project",
			Name:    "test-app",
			AppMeta: realm.AppMeta{
				Location:        realm.LocationVirginia,
				DeploymentModel: realm.DeploymentModelGlobal,
			},
		}, createdApp)
	})

	t.Run("should create a new app with a structure based on the specified from app", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()

		testApp := realm.App{
			ID:          primitive.NewObjectID().Hex(),
			GroupID:     primitive.NewObjectID().Hex(),
			ClientAppID: "from-app-abcde",
			Name:        "from-app",
		}

		zipPkg, err := zip.OpenReader("testdata/project.zip")
		assert.Nil(t, err)

		var createdApp realm.App
		client := mock.RealmClient{}
		client.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{testApp}, nil
		}
		client.ExportFn = func(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error) {
			return "", &zipPkg.Reader, err
		}
		client.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			createdApp = realm.App{
				GroupID: groupID,
				Name:    name,
				AppMeta: meta,
			}
			return createdApp, nil
		}
		client.ImportFn = func(groupID, appID string, appData interface{}) error {
			return nil
		}

		cmd := &CommandCreate{
			inputs: createInputs{newAppInputs: newAppInputs{
				From:            testApp.Name,
				Project:         testApp.GroupID,
				Location:        realm.LocationIreland,
				DeploymentModel: realm.DeploymentModelGlobal,
			}},
			realmClient: client,
		}

		assert.Nil(t, cmd.Handler(profile, nil))

		localApp, err := local.LoadApp(filepath.Join(profile.WorkingDirectory, cmd.inputs.From))
		assert.Nil(t, err)

		actualAppData := localApp.AppData.(*local.AppRealmConfigJSON)

		expectedAppData := local.AppRealmConfigJSON{local.AppDataV2{local.AppStructureV2{
			ConfigVersion:   realm.DefaultAppConfigVersion,
			Name:            testApp.Name,
			Location:        realm.LocationIreland,
			DeploymentModel: realm.DeploymentModelGlobal,
			Auth:            actualAppData.Auth,
			Sync:            actualAppData.Sync,
		}}}

		assert.Equal(t, &expectedAppData, actualAppData)
		assert.Equal(t, realm.App{
			GroupID: testApp.GroupID,
			Name:    testApp.Name,
			AppMeta: realm.AppMeta{
				Location:        realm.LocationIreland,
				DeploymentModel: realm.DeploymentModelGlobal,
			},
		}, createdApp)
	})

	t.Run("should error when resolving groupID when project is not set", func(t *testing.T) {
		profile := mock.NewProfileFromWD(t)

		client := mock.AtlasClient{}
		client.GroupsFn = func() ([]atlas.Group, error) {
			return nil, errors.New("atlas client error")
		}

		cmd := &CommandCreate{
			inputs: createInputs{newAppInputs: newAppInputs{
				Name:            "test-app",
				Location:        realm.LocationVirginia,
				DeploymentModel: realm.DeploymentModelGlobal,
			}},
			atlasClient: client,
		}

		assert.Equal(t, errors.New("atlas client error"), cmd.Handler(profile, nil))
	})

	t.Run("should error when resolving app when from is set", func(t *testing.T) {
		profile := mock.NewProfileFromWD(t)

		client := mock.RealmClient{}
		client.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return nil, errors.New("realm client error")
		}

		cmd := &CommandCreate{
			inputs: createInputs{newAppInputs: newAppInputs{
				From:            "test-app",
				Location:        realm.LocationVirginia,
				DeploymentModel: realm.DeploymentModelGlobal,
			}},
			realmClient: client,
		}

		assert.Equal(t, errors.New("realm client error"), cmd.Handler(profile, nil))
	})
}

// Add test cases to handler
// func TestAppNewAppInputsResolveProject(t *testing.T) {
// 	testApp := realm.App{
// 		ID:          primitive.NewObjectID().Hex(),
// 		GroupID:     primitive.NewObjectID().Hex(),
// 		ClientAppID: "test-app-abcde",
// 		Name:        "test-app",
// 	}

// 	for _, tc := range []struct {
// 		description     string
// 		inputs          newAppInputs
// 		expectedProject string
// 		expectedErr     error
// 	}{
// 	} {
// 		t.Run(tc.description, func(t *testing.T) {
// 			ac := mock.AtlasClient{}
// 			ac.GroupsFn = func() ([]atlas.Group, error) {
// 				return []atlas.Group{{ID: testApp.GroupID}}, tc.expectedErr
// 			}

// 			groupID, err := tc.inputs.resolveProject(nil, ac)

// 			assert.Equal(t, tc.expectedErr, err)
// 			assert.Equal(t, tc.expectedProject, groupID)
// 		})
// 	}

//
// }

func TestAppCreateFeedback(t *testing.T) {
	t.Run("feedback should print a message that app creation was successful", func(t *testing.T) {
		out, ui := mock.NewUI()

		cmd := &CommandCreate{
			outputs: createOutputs{
				clientAppID: "test-client-id",
				dir:         "/file/path/to/test-app",
				uiURL:       "https://realm.mongodb.com/groups/123/apps/123/dashboard",
				followUpCmd: "cd ./test-app && realm-cli app describe",
			},
		}

		err := cmd.Feedback(nil, ui)
		assert.Nil(t, err)

		expectedContent := strings.Join(
			[]string{
				"01:23:45 UTC INFO  Successfully created app",
				"  Info                Details                                                ",
				"  ------------------  -------------------------------------------------------",
				"  Client App ID       test-client-id                                         ",
				"  Realm Directory     /file/path/to/test-app                                 ",
				"  Realm UI            https://realm.mongodb.com/groups/123/apps/123/dashboard",
				"  Check out your app  cd ./test-app && realm-cli app describe                ",
				"",
			},
			"\n",
		)

		assert.Equal(t, expectedContent, out.String())
	})
}
