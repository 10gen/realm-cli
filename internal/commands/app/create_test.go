package app

import (
	"archive/zip"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

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
	t.Run("should create bare bone project when no from type is specified", func(t *testing.T) {
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

		data, readErr := ioutil.ReadFile(filepath.Join(profile.WorkingDirectory, cmd.inputs.Name, local.FileRealmConfig.String()))
		assert.Nil(t, readErr)

		var config local.AppRealmConfigJSON
		assert.Nil(t, json.Unmarshal(data, &config))
		assert.Equal(t, local.AppRealmConfigJSON{local.AppDataV2{local.AppStructureV2{
			ConfigVersion:   realm.DefaultAppConfigVersion,
			Name:            "test-app",
			Location:        realm.LocationVirginia,
			DeploymentModel: realm.DeploymentModelGlobal,
		}}}, config)
		assert.Equal(t, realm.App{
			GroupID: "test-project",
			Name:    "test-app",
			AppMeta: realm.AppMeta{
				Location:        realm.LocationVirginia,
				DeploymentModel: realm.DeploymentModelGlobal,
			},
		}, createdApp)
	})

	t.Run("should create a templated app when from is set", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()

		testApp := realm.App{
			ID:          primitive.NewObjectID().Hex(),
			GroupID:     primitive.NewObjectID().Hex(),
			ClientAppID: "from-app-abcde",
			Name:        "from-app",
		}

		var zipCloser *zip.ReadCloser
		defer func() { zipCloser.Close() }()

		var createdApp realm.App
		client := mock.RealmClient{}
		client.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{testApp}, nil
		}
		client.ExportFn = func(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error) {
			zipPkg, err := zip.OpenReader("testdata/project.zip")

			zipCloser = zipPkg
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

		data, readErr := ioutil.ReadFile(filepath.Join(profile.WorkingDirectory, cmd.inputs.From, local.FileRealmConfig.String()))
		assert.Nil(t, readErr)

		var config local.AppRealmConfigJSON
		assert.Nil(t, json.Unmarshal(data, &config))
		assert.Equal(t, local.AppRealmConfigJSON{local.AppDataV2{local.AppStructureV2{
			ConfigVersion:   realm.DefaultAppConfigVersion,
			Name:            testApp.Name,
			Location:        realm.LocationIreland,
			DeploymentModel: realm.DeploymentModelGlobal,
		}}}, config)
		assert.Equal(t, realm.App{
			GroupID: testApp.GroupID,
			Name:    testApp.Name,
			AppMeta: realm.AppMeta{
				Location:        realm.LocationIreland,
				DeploymentModel: realm.DeploymentModelGlobal,
			},
		}, createdApp)
	})
}

func TestAppCreateFeedback(t *testing.T) {
	t.Run("feedback should print a message that app creation was successful", func(t *testing.T) {
		out, ui := mock.NewUI()

		cmd := &CommandCreate{}

		err := cmd.Feedback(nil, ui)
		assert.Nil(t, err)

		assert.Equal(t, "01:23:45 UTC INFO  Successfully created app\n", out.String())
	})
}
