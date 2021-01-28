package realm_test

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestRealmImportExport(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	groupID := u.CloudGroupID()

	t.Run("Should fail without an auth client", func(t *testing.T) {
		client := realm.NewClient(u.RealmServerURL())

		err := client.Import(groupID, primitive.NewObjectID().Hex(), nil)
		assert.Equal(t, realm.ErrInvalidSession{}, err)
	})

	t.Run("With an active session", func(t *testing.T) {
		client := newAuthClient(t)

		app, appErr := client.CreateApp(groupID, "importexport-test", realm.AppMeta{})
		assert.Nil(t, appErr)

		resetPasswordURL := "http://localhost:8080/reset_password"
		emailConfirmationURL := "http://localhost:8080/confirm_email"

		t.Run("Should import an app with auth providers", func(t *testing.T) {
			err := client.Import(groupID, app.ID, map[string]interface{}{
				local.NameAuthProviders: []map[string]interface{}{
					{"name": "api-key", "type": "api-key"},
					{"name": "local-userpass", "type": "local-userpass", "config": map[string]interface{}{
						"resetPasswordUrl":     resetPasswordURL,
						"emailConfirmationUrl": emailConfirmationURL,
					}},
				},
			})
			assert.Nil(t, err)
		})

		t.Run("Should export the same app with the imported changes included", func(t *testing.T) {
			filename, zipPkg, err := client.Export(groupID, app.ID, realm.ExportRequest{})
			assert.Nil(t, err)

			filenameMatch, matchErr := regexp.MatchString(fmt.Sprintf("%s_.*\\.zip", app.Name), filename)
			assert.Nil(t, matchErr)
			assert.True(t, filenameMatch, "expected exported filename to match '$appName_yyyymmddHHMMSS'")

			exported := parseZipPkg(t, zipPkg)

			t.Run("And the app config contents should be as expected", func(t *testing.T) {
				appConfig, appConfigOK := exported[local.FileConfig.String()]
				assert.True(t, appConfigOK, "expected exported app to have file: %s", local.FileConfig)
				assert.Equal(t, fmt.Sprintf(`{
    "app_id": %q,
    "config_version": %s,
    "name": "importexport-test",
    "location": "US-VA",
    "deployment_model": "GLOBAL",
    "security": {},
    "custom_user_data_config": {
        "enabled": false
    },
    "sync": {
        "development_mode_enabled": false
    }
}
`, app.ClientAppID, realm.DefaultAppConfigVersion), appConfig)
			})

			t.Run("And the auth provider contents should be as expected", func(t *testing.T) {
				apiKeyConfigFilepath := filepath.Join(local.NameAuthProviders, "api-key.json")
				apiKeyConfigPayload, apiKeyConfigOK := exported[apiKeyConfigFilepath]
				assert.True(t, apiKeyConfigOK, "expected exported app to have file: %s", apiKeyConfigFilepath)

				var apiKeyConfig map[string]interface{}
				assert.Nil(t, json.Unmarshal([]byte(apiKeyConfigPayload), &apiKeyConfig))
				assert.Nilf(t, apiKeyConfig["id"], "expected api-key.json to not have an 'id' field")
				assert.Equal(t, "api-key", apiKeyConfig["name"])
				assert.Equal(t, "api-key", apiKeyConfig["type"])
				assert.False(t, apiKeyConfig["disabled"], "expected api-key.json to have 'disabled' field set to false")

				localUserpassConfigFilepath := filepath.Join(local.NameAuthProviders, "local-userpass.json")
				localUserpassConfigPayload, localUserpassOK := exported[localUserpassConfigFilepath]
				assert.True(t, localUserpassOK, "expected exported app to have file: %s", localUserpassConfigFilepath)

				var localUserpassConfig map[string]interface{}
				assert.Nil(t, json.Unmarshal([]byte(localUserpassConfigPayload), &localUserpassConfig))
				assert.Nilf(t, localUserpassConfig["id"], "expected local-userpass.json to not have an 'id' field")
				assert.Equal(t, "local-userpass", localUserpassConfig["name"])
				assert.Equal(t, "local-userpass", localUserpassConfig["type"])
				assert.False(t, apiKeyConfig["disabled"], "expected local-userpass.json to have 'disabled' field set to false")

				localUserpassConfigConfig, localUserpassConfigConfigOK := localUserpassConfig["config"].(map[string]interface{})
				assert.True(t, localUserpassConfigConfigOK, "expected local-userpass.json to have 'config' field set to a nested map")
				assert.Equal(t, resetPasswordURL, localUserpassConfigConfig["resetPasswordUrl"])
				assert.Equal(t, emailConfirmationURL, localUserpassConfigConfig["emailConfirmationUrl"])
			})

			t.Run("And the graphql contents should be as expected", func(t *testing.T) {
				graphqlFilepath := "graphql/config.json"
				graphQLConfig, graphQLConfigOK := exported[graphqlFilepath]
				assert.True(t, graphQLConfigOK, "expected exported app to have file: %s", graphqlFilepath)
				assert.Equal(t, `{
    "use_natural_pluralization": true
}
`, graphQLConfig)
			})
		})
	})
}

// TestConfigVersion20210101 is responsible for testing our expectations of what
// the app structure for ConfigVersion20210101 looks like
// TODO(REALMC-7653I: this test is now redundant the current set of integration tests
// Tests should be written to verify we still export/read older config versions properly
func TestImportExportRoundTrip(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	for _, tc := range []struct {
		description   string
		configVersion realm.AppConfigVersion
		importData    func(app realm.App) local.AppData
	}{
		{
			configVersion: realm.AppConfigVersion20180301,
			importData: func(app realm.App) local.AppData {
				return &local.AppStitchJSON{appDataV1(realm.AppConfigVersion20180301, app)}
			},
		},
		{
			configVersion: realm.AppConfigVersion20200603,
			importData: func(app realm.App) local.AppData {
				return &local.AppConfigJSON{appDataV1(realm.AppConfigVersion20200603, app)}
			},
		},
		// TODO(REALMC-7653): add round-trip test for new config version
		// {
		// 	configVersion: realm.AppConfigVersion20210101,
		// 	importData: func(app realm.App) local.AppData {
		// 		return &local.AppRealmConfigJSON{appDataV2(app)}
		// 	},
		// },
	} {
		t.Run(fmt.Sprintf("Should import and export the same data for config version %d", tc.configVersion), func(t *testing.T) {
			tmpDir, tmpDirTeardown, tmpDirErr := u.NewTempDir("importexport")
			assert.Nil(t, tmpDirErr)
			defer tmpDirTeardown()

			client := newAuthClient(t)

			groupID := u.CloudGroupID()

			app, appErr := client.CreateApp(groupID, "importexport-test", realm.AppMeta{})
			assert.Nil(t, appErr)

			appData := tc.importData(app)

			assert.Nil(t, client.Import(groupID, app.ID, appData))

			filename, zipPkg, exportErr := client.Export(groupID, app.ID, realm.ExportRequest{ConfigVersion: tc.configVersion})
			assert.Nil(t, exportErr)

			filenameMatch, matchErr := regexp.MatchString(fmt.Sprintf("%s_.*\\.zip", app.Name), filename)
			assert.Nil(t, matchErr)
			assert.True(t, filenameMatch, "expected exported filename to match '$appName_yyyymmddHHMMSS'")

			wd := filepath.Join(tmpDir, filename)

			t.Log("write the exported zip package to disk")
			assert.Nil(t, local.WriteZip(wd, zipPkg))

			t.Log("read the exported app data from disk")
			localApp, localAppErr := local.LoadApp(wd)
			assert.Nil(t, localAppErr)
			assert.Equal(t, appData, localApp.AppData)
		})
	}
}

func parseZipPkg(t *testing.T, zipPkg *zip.Reader) map[string]string {
	t.Helper()

	out := make(map[string]string)
	for _, file := range zipPkg.File {
		out[file.Name] = parseZipFile(t, file)
	}
	return out
}

func parseZipFile(t *testing.T, file *zip.File) string {
	t.Helper()

	r, openErr := file.Open()
	assert.Nil(t, openErr)

	data, readErr := ioutil.ReadAll(r)
	assert.Nil(t, readErr)

	return string(data)
}
func appDataV1(configVersion realm.AppConfigVersion, app realm.App) local.AppDataV1 {
	return local.AppDataV1{local.AppStructureV1{
		ConfigVersion:        configVersion,
		ID:                   app.ClientAppID,
		Name:                 app.Name,
		Location:             app.Location,
		DeploymentModel:      app.DeploymentModel,
		Sync:                 map[string]interface{}{"development_mode_enabled": false},
		CustomUserDataConfig: map[string]interface{}{"enabled": false},
		AuthProviders: []map[string]interface{}{
			{"name": realm.AuthProviderTypeAnonymous.String(), "type": realm.AuthProviderTypeAnonymous.String(), "disabled": false},
			{"name": realm.AuthProviderTypeAPIKey.String(), "type": realm.AuthProviderTypeAPIKey.String(), "disabled": false},
		},
		Functions: []map[string]interface{}{
			{
				local.NameConfig: map[string]interface{}{
					"name":    "test",
					"private": true,
				},
				local.NameSource: `exports = function(){
console.log('got heem!');
};`,
			},
		},
		Triggers: []map[string]interface{}{
			{
				"name":          "yell",
				"type":          "SCHEDULED",
				"config":        map[string]interface{}{"schedule": "0 0 * * 1"},
				"function_name": "test",
				"disabled":      false,
			},
		},
		GraphQL: local.GraphQLStructure{
			Config:          map[string]interface{}{"use_natural_pluralization": true},
			CustomResolvers: []map[string]interface{}{},
		},
		Security: map[string]interface{}{
			"allowed_request_origins": []interface{}{
				"http://localhost:8080",
			},
		},
	}}
}

// TODO(REALMC-7653): add round-trip test for new config version
// func appDataV2(app realm.App) local.AppDataV2 {
// 	return local.AppDataV2{local.AppStructureV2{
// 		ConfigVersion:         realm.AppConfigVersion20210101,
// 		ID:                    app.ClientAppID,
// 		Name:                  app.Name,
// 		Location:              app.Location,
// 		DeploymentModel:       app.DeploymentModel,
// 		AllowedRequestOrigins: []string{"http://localhost:8080"},
// 		Sync: local.SyncStructure{
// 			Config: map[string]interface{}{"development_mode_enabled": false},
// 		},
// 		Auth: local.AuthStructure{
// 			Config: map[string]interface{}{"enabled": false},
// 			Providers: []map[string]interface{}{
// 				{"name": realm.AuthProviderTypeAnonymous.String(), "type": realm.AuthProviderTypeAnonymous.String(), "disabled": false},
// 				{"name": realm.AuthProviderTypeAPIKey.String(), "type": realm.AuthProviderTypeAPIKey.String(), "disabled": false},
// 			},
// 		},
// 		Functions: local.FunctionsStructure{
// 			Config: map[string]interface{}{
// 				"test.js": map[string]interface{}{"private": true},
// 			},
// 			SrcMap: map[string]string{
// 				"test.js": `exports = function(){
//   console.log('got heem!');
// };`,
// 			},
// 		},
// 		Triggers: []map[string]interface{}{
// 			{
// 				"name":          "yell",
// 				"type":          "SCHEDULED",
// 				"config":        map[string]interface{}{"schedule": "0 0 * * 1"},
// 				"function_name": "test",
// 				"disabled":      false,
// 			},
// 		},
// 		GraphQL: local.GraphQLStructure{
// 			Config: map[string]interface{}{"use_natural_pluralization": true},
// 		},
// 	}}
// }
