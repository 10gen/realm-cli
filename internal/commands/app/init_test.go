package app

import (
	"archive/zip"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAppInitHandler(t *testing.T) {
	t.Run("should initialize an empty project when no remote type is specified", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_init_test")
		defer teardown()

		out, ui := mock.NewUI()

		cmd := &CommandInit{initInputs{newAppInputs: newAppInputs{
			Name:            "test-app",
			Project:         "test-project",
			DeploymentModel: realm.DeploymentModelLocal,
			Location:        realm.LocationSydney,
			ConfigVersion:   realm.DefaultAppConfigVersion,
		}}}

		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{}))

		data, err := ioutil.ReadFile(filepath.Join(profile.WorkingDirectory, local.FileRealmConfig.String()))
		assert.Nil(t, err)

		assert.Equal(t, "Successfully initialized app\n", out.String())

		var config local.AppRealmConfigJSON
		assert.Nil(t, json.Unmarshal(data, &config))
		assert.Equal(t, local.AppRealmConfigJSON{local.AppDataV2{local.AppStructureV2{
			ConfigVersion:   realm.DefaultAppConfigVersion,
			Name:            "test-app",
			Location:        realm.LocationSydney,
			DeploymentModel: realm.DeploymentModelLocal,
		}}}, config)

		t.Run("should have the expected contents in the auth custom user data file", func(t *testing.T) {
			config, err := ioutil.ReadFile(filepath.Join(profile.WorkingDirectory, local.NameAuth, local.FileCustomUserData.String()))
			assert.Nil(t, err)
			assert.Equal(t, `{
    "enabled": false
}
`, string(config))
		})

		t.Run("should have the expected contents in the auth providers file", func(t *testing.T) {
			config, err := ioutil.ReadFile(filepath.Join(profile.WorkingDirectory, local.NameAuth, local.FileProviders.String()))
			assert.Nil(t, err)
			assert.Equal(t, `{
    "api-key": {
        "disabled": true,
        "name": "api-key",
        "type": "api-key"
    }
}
`, string(config))
		})

		t.Run("should have data sources directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(profile.WorkingDirectory, local.NameDataSources))
			assert.Nil(t, err)
		})

		t.Run("should have the expected contents in the functions config file", func(t *testing.T) {
			config, err := ioutil.ReadFile(filepath.Join(profile.WorkingDirectory, local.NameFunctions, local.FileConfig.String()))
			assert.Nil(t, err)
			assert.Equal(t, `[]
`, string(config))
		})

		t.Run("should have graphql custom resolvers directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(profile.WorkingDirectory, local.NameGraphQL, local.NameCustomResolvers))
			assert.Nil(t, err)
		})

		t.Run("should have the expected contents in the graphql config file", func(t *testing.T) {
			config, err := ioutil.ReadFile(filepath.Join(profile.WorkingDirectory, local.NameGraphQL, local.FileConfig.String()))
			assert.Nil(t, err)
			assert.Equal(t, `{
    "use_natural_pluralization": true
}
`, string(config))
		})

		t.Run("should have http endpoints directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(profile.WorkingDirectory, local.NameHTTPEndpoints))
			assert.Nil(t, err)
		})

		t.Run("should have services directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(profile.WorkingDirectory, local.NameServices))
			assert.Nil(t, err)
		})

		t.Run("should have the expected contents in the sync config file", func(t *testing.T) {
			config, err := ioutil.ReadFile(filepath.Join(profile.WorkingDirectory, local.NameSync, local.FileConfig.String()))
			assert.Nil(t, err)
			assert.Equal(t, `{
    "development_mode_enabled": false
}
`, string(config))
		})

		t.Run("should have values directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(profile.WorkingDirectory, local.NameValues))
			assert.Nil(t, err)
		})
	})

	t.Run("should initialze a templated app when remote type is specified to app", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_init_test")
		defer teardown()

		out, ui := mock.NewUI()

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

		cmd := &CommandInit{initInputs{newAppInputs: newAppInputs{RemoteApp: "test", ConfigVersion: realm.DefaultAppConfigVersion}}}

		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: client}))

		assert.Equal(t, "Successfully initialized app\n", out.String())

		t.Run("should have the expected contents in the app config file", func(t *testing.T) {
			data, readErr := ioutil.ReadFile(filepath.Join(profile.WorkingDirectory, local.FileRealmConfig.String()))
			assert.Nil(t, readErr)

			var config local.AppRealmConfigJSON
			assert.Nil(t, json.Unmarshal(data, &config))
			assert.Equal(t, local.AppRealmConfigJSON{local.AppDataV2{local.AppStructureV2{
				ConfigVersion:   realm.DefaultAppConfigVersion,
				Name:            "remote-app",
				Location:        realm.LocationIreland,
				DeploymentModel: realm.DeploymentModelGlobal,
			}}}, config)
		})

		t.Run("should have the expected contents in the auth custom user data file", func(t *testing.T) {
			config, err := ioutil.ReadFile(filepath.Join(profile.WorkingDirectory, local.NameAuth, local.FileCustomUserData.String()))
			assert.Nil(t, err)
			assert.Equal(t, `{
    "enabled": false
}
`, string(config))
		})

		t.Run("should have the expected contents in the auth providers file", func(t *testing.T) {
			config, err := ioutil.ReadFile(filepath.Join(profile.WorkingDirectory, local.NameAuth, local.FileProviders.String()))
			assert.Nil(t, err)
			assert.Equal(t, `{}
`, string(config))
		})

		t.Run("should have data sources directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(profile.WorkingDirectory, local.NameDataSources))
			assert.Nil(t, err)
		})

		t.Run("should have http endpoints directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(profile.WorkingDirectory, local.NameHTTPEndpoints))
			assert.Nil(t, err)
		})

		t.Run("should have services directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(profile.WorkingDirectory, local.NameServices))
			assert.Nil(t, err)
		})

		t.Run("should have the expected contents in the sync config file", func(t *testing.T) {
			config, err := ioutil.ReadFile(filepath.Join(profile.WorkingDirectory, local.NameSync, local.FileConfig.String()))
			assert.Nil(t, err)
			assert.Equal(t, `{
    "development_mode_enabled": false
}
`, string(config))
		})
	})
}
