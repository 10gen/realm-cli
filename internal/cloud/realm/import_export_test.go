package realm_test

import (
	"fmt"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestRealmImportExportRoundTrip(t *testing.T) {
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
		{
			configVersion: realm.AppConfigVersion20210101,
			importData: func(app realm.App) local.AppData {
				return &local.AppRealmConfigJSON{appDataV2(app)}
			},
		},
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

func appDataV2(app realm.App) local.AppDataV2 {
	return local.AppDataV2{local.AppStructureV2{
		ConfigVersion:   realm.AppConfigVersion20210101,
		ID:              app.ClientAppID,
		Name:            app.Name,
		Location:        app.Location,
		DeploymentModel: app.DeploymentModel,
		// TODO(REALMC-7989): include auth, functions, triggers, allowed request origina, and graphql
		// AllowedRequestOrigins: []string{"http://localhost:8080"},
		Sync: &local.SyncStructure{
			Config: map[string]interface{}{"development_mode_enabled": false},
		},
		// in 20210101 round-trip test once its supported in export on the backend
		// Auth: &local.AuthStructure{
		// 	CustomUserData: map[string]interface{}{"enabled": false},
		// 	Providers: map[string]map[string]interface{}{
		// 		realm.AuthProviderTypeAnonymous.String(): map[string]interface{}{
		// 			"name":     realm.AuthProviderTypeAnonymous.String(),
		// 			"type":     realm.AuthProviderTypeAnonymous.String(),
		// 			"disabled": false,
		// 		},
		// 		realm.AuthProviderTypeAPIKey.String(): map[string]interface{}{
		// 			"name":     realm.AuthProviderTypeAPIKey.String(),
		// 			"type":     realm.AuthProviderTypeAPIKey.String(),
		// 			"disabled": false,
		// 		},
		// 	},
		// },
		// 		Functions: &local.FunctionsStructure{
		// 			Config: map[string]interface{}{
		// 				"test.js": map[string]interface{}{"private": true},
		// 			},
		// 			SrcMap: map[string]string{
		// 				"test.js": `exports = function(){
		//   console.log('got heem!');
		// };`,
		// 			},
		// 		},
		// Triggers: []map[string]interface{}{
		// 	{
		// 		"name":          "yell",
		// 		"type":          "SCHEDULED",
		// 		"config":        map[string]interface{}{"schedule": "0 0 * * 1"},
		// 		"function_name": "test",
		// 		"disabled":      false,
		// 	},
		// },
		// GraphQL: &local.GraphQLStructure{
		// 	Config: map[string]interface{}{"use_natural_pluralization": true},
		// },
	}}
}
