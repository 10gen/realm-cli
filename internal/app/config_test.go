package app

import (
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestConfigVersionFile(t *testing.T) {
	for _, tc := range []struct {
		configVersion realm.AppConfigVersion
		file          File
	}{
		{realm.AppConfigVersionZero, FileRealmConfig},
		{realm.AppConfigVersion20180301, FileStitch},
		{realm.AppConfigVersion20200603, FileConfig},
		{realm.AppConfigVersion20210101, FileRealmConfig},
	} {
		t.Run(fmt.Sprintf("Should return a %s file for a config version of %d", tc.file, tc.configVersion), func(t *testing.T) {
			file, err := configVersionFile(tc.configVersion)
			assert.Nil(t, err)
			assert.Equal(t, tc.file, file)
		})
	}

	t.Run("Should return an error for an unknown config version", func(t *testing.T) {
		_, err := configVersionFile(realm.AppConfigVersion(1))
		assert.Equal(t, err, errUnknownConfigVersion)
	})
}

func TestFileConfigVersion(t *testing.T) {
	for _, tc := range []struct {
		configVersion realm.AppConfigVersion
		file          File
	}{
		{realm.AppConfigVersionZero, File{Name: "new_config", Ext: extJSON}},
		{realm.AppConfigVersion20180301, FileStitch},
		{realm.AppConfigVersion20200603, FileConfig},
		{realm.AppConfigVersion20210101, FileRealmConfig},
	} {
		t.Run(fmt.Sprintf("Should return a %d config version for the %s file", tc.configVersion, tc.file), func(t *testing.T) {
			assert.Equal(t, tc.configVersion, fileConfigVersion(tc.file))
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	t.Run("Should return the correct default config", func(t *testing.T) {
		app := realm.App{
			ClientAppID: "client-app-id",
			Name:        "name",
			AppMeta: realm.AppMeta{
				Location:        realm.LocationVirginia,
				DeploymentModel: realm.DeploymentModelGlobal,
			},
		}

		expectedConfig := Config{
			ConfigVersion:   realm.DefaultAppConfigVersion,
			ID:              "client-app-id",
			Name:            "name",
			Location:        realm.LocationVirginia,
			DeploymentModel: realm.DeploymentModelGlobal,
		}

		assert.Equal(t, expectedConfig, ToDefaultConfig(app))
	})
}
