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

func TestAppInitSetup(t *testing.T) {
	t.Run("Should construct a Realm client with the configured base url", func(t *testing.T) {
		profile := mock.NewProfile(t)
		profile.SetRealmBaseURL("http://localhost:8080")

		cmd := &CommandInit{inputs: initInputs{
			Name: "test-app",
		}}
		assert.Nil(t, cmd.realmClient)

		assert.Nil(t, cmd.Setup(profile, nil))
		assert.NotNil(t, cmd.realmClient)
	})
}

func TestAppInitHandler(t *testing.T) {
	t.Run("Should initialize an empty project when no from type is specified", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_init_test")
		defer teardown()

		cmd := &CommandInit{inputs: initInputs{
			Name:            "test-app",
			DeploymentModel: realm.DeploymentModelLocal,
			Location:        realm.LocationSydney,
		}}

		assert.Nil(t, cmd.Handler(profile, nil))

		data, readErr := ioutil.ReadFile(filepath.Join(profile.WorkingDirectory, local.FileConfig.String()))
		assert.Nil(t, readErr)

		var config local.AppConfigJSON
		assert.Nil(t, json.Unmarshal(data, &config))
		assert.Equal(t, local.AppConfigJSON{local.AppDataV1{local.AppStructureV1{
			ConfigVersion:   realm.DefaultAppConfigVersion,
			Name:            "test-app",
			Location:        realm.LocationSydney,
			DeploymentModel: realm.DeploymentModelLocal,
		}}}, config)
	})

	t.Run("Should initialze a templated app when from type is specified to app", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_init_test")
		defer teardown()

		app := realm.App{
			ID:          primitive.NewObjectID().Hex(),
			GroupID:     primitive.NewObjectID().Hex(),
			ClientAppID: "test-app-abcde",
			Name:        "test-app",
		}

		var zipCloser *zip.ReadCloser
		defer func() { zipCloser.Close() }()

		client := mock.RealmClient{}
		client.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{app}, nil
		}
		client.ExportFn = func(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error) {
			zipPkg, err := zip.OpenReader("testdata/project.zip")

			zipCloser = zipPkg
			return "", &zipPkg.Reader, err
		}

		cmd := &CommandInit{
			inputs:      initInputs{From: "test"},
			realmClient: client,
		}

		assert.Nil(t, cmd.Handler(profile, nil))

		t.Run("Should have the expected contents in the app config file", func(t *testing.T) {
			data, readErr := ioutil.ReadFile(filepath.Join(profile.WorkingDirectory, local.FileConfig.String()))
			assert.Nil(t, readErr)

			var config local.AppConfigJSON
			assert.Nil(t, json.Unmarshal(data, &config))
			assert.Equal(t, local.AppConfigJSON{local.AppDataV1{local.AppStructureV1{
				ConfigVersion:        realm.DefaultAppConfigVersion,
				Name:                 "from-app",
				Location:             realm.LocationIreland,
				DeploymentModel:      realm.DeploymentModelGlobal,
				Security:             map[string]interface{}{},
				CustomUserDataConfig: map[string]interface{}{"enabled": false},
				Sync:                 map[string]interface{}{"development_mode_enabled": false},
			}}}, config)
		})

		t.Run("Should have the expected contents in the api key auth provider config file", func(t *testing.T) {
			config, err := ioutil.ReadFile(filepath.Join(profile.WorkingDirectory, local.NameAuthProviders, "api-key.json"))
			assert.Nil(t, err)
			assert.Equal(t, `{
    "name": "api-key",
    "type": "api-key",
    "disabled": false
}
`, string(config))
		})
	})
}

func TestAppInitFeedback(t *testing.T) {
	t.Run("Feedback should print a message that app initialization was successful", func(t *testing.T) {
		out, ui := mock.NewUI()

		cmd := &CommandInit{}

		err := cmd.Feedback(nil, ui)
		assert.Nil(t, err)

		assert.Equal(t, "01:23:45 UTC INFO  Successfully initialized app\n", out.String())
	})
}
