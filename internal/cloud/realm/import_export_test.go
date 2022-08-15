package realm_test

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
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

			app, teardown := setupTestApp(t, client, groupID, "importexport-test")
			defer teardown()

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
			appLocal, appLocalErr := local.LoadApp(wd)
			assert.Nil(t, appLocalErr)
			assert.Equal(t, appData, appLocal.AppData)
		})
	}
}

func TestRealmImport20210101(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	client := newAuthClient(t)

	groupID := u.CloudGroupID()

	app, teardown := setupTestApp(t, client, groupID, "import20210101")
	defer teardown()

	t.Run("Should import a service with a secret successfully", func(t *testing.T) {
		assert.Nil(t, client.Import(groupID, app.ID, local.AppDataV2{local.AppStructureV2{
			ConfigVersion:   realm.AppConfigVersion20210101,
			ID:              app.ClientAppID,
			Name:            app.Name,
			Location:        app.Location,
			DeploymentModel: app.DeploymentModel,
			Services: []local.ServiceStructure{
				{
					Config: map[string]interface{}{"name": "twilio_svc", "type": "twilio", "config": map[string]interface{}{"sid": "my_sid"}},
					IncomingWebhooks: []map[string]interface{}{
						{
							"config": map[string]interface{}{
								"name":                         "twilioWebhook",
								"create_user_on_auth":          false,
								"fetch_custom_user_data":       false,
								"respond_result":               false,
								"run_as_authed_user":           false,
								"run_as_user_id_script_source": "",
							},
							"source": "exports = function() { return false }",
						},
					},
				},
			},
			Secrets: local.SecretsStructure{Services: map[string]map[string]string{"twilio_svc": map[string]string{"auth_token": "my-secret-auth-token"}}},
		}}))

		secrets, secretsErr := client.Secrets(groupID, app.ID)
		assert.Nil(t, secretsErr)
		assert.Equal(t, 1, len(secrets))
		assert.Equal(t, "__twilio_svc_auth_token", secrets[0].Name)

		_, zipPkg, exportErr := client.Export(groupID, app.ID, realm.ExportRequest{ConfigVersion: realm.AppConfigVersion20210101})
		assert.Nil(t, exportErr)

		exported := parseZipPkg(t, zipPkg)

		twilioConfig, twilioConfigOK := exported[filepath.Join(local.NameServices, "twilio_svc", local.FileConfig.String())]
		assert.True(t, twilioConfigOK, "expected to have twilio config file")
		assert.Equal(t, `{
    "name": "twilio_svc",
    "type": "twilio",
    "config": {
        "sid": "my_sid"
    },
    "secret_config": {
        "auth_token": "__twilio_svc_auth_token"
    }
}
`, twilioConfig)
	})
}

func TestRealmImportLegacy(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	client := newAuthClient(t)

	groupID := u.CloudGroupID()

	app, teardown := setupTestApp(t, client, groupID, "import20210101")
	defer teardown()

	for _, configVersion := range []realm.AppConfigVersion{realm.AppConfigVersion20180301, realm.AppConfigVersion20200603} {
		t.Run(fmt.Sprintf("Should import a service with a secret successfully for config version %d", configVersion), func(t *testing.T) {
			assert.Nil(t, client.Import(groupID, app.ID, local.AppDataV1{local.AppStructureV1{
				ConfigVersion:   configVersion,
				ID:              app.ClientAppID,
				Name:            app.Name,
				Location:        app.Location,
				DeploymentModel: app.DeploymentModel,
				Services: []local.ServiceStructure{
					{
						Config: map[string]interface{}{"name": "twilio_svc", "type": "twilio", "config": map[string]interface{}{"sid": "my_sid"}},
						IncomingWebhooks: []map[string]interface{}{
							{
								"config": map[string]interface{}{
									"name":                         "twilioWebhook",
									"create_user_on_auth":          false,
									"fetch_custom_user_data":       false,
									"respond_result":               false,
									"run_as_authed_user":           false,
									"run_as_user_id_script_source": "",
								},
								"source": "exports = function() { return false }",
							},
						},
					},
				},
				Secrets: local.SecretsStructure{Services: map[string]map[string]string{"twilio_svc": map[string]string{"auth_token": "my-secret-auth-token"}}},
			}}))

			secrets, secretsErr := client.Secrets(groupID, app.ID)
			assert.Nil(t, secretsErr)
			assert.Equal(t, 1, len(secrets))
			assert.Equal(t, "__twilio_svc_auth_token", secrets[0].Name)

			_, zipPkg, exportErr := client.Export(groupID, app.ID, realm.ExportRequest{ConfigVersion: configVersion})
			assert.Nil(t, exportErr)

			exported := parseZipPkg(t, zipPkg)

			twilioConfig, twilioConfigOK := exported[filepath.Join(local.NameServices, "twilio_svc", local.FileConfig.String())]
			assert.True(t, twilioConfigOK, "expected to have twilio config file")
			assert.Equal(t, `{
    "name": "twilio_svc",
    "type": "twilio",
    "config": {
        "sid": "my_sid"
    },
    "secret_config": {
        "auth_token": "__twilio_svc_auth_token"
    }
}
`, twilioConfig)
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
		Environments: map[string]map[string]interface{}{
			"no-environment.json": map[string]interface{}{"values": map[string]interface{}{"a": "0"}},
			"development.json":    map[string]interface{}{"values": map[string]interface{}{"a": "1"}},
			"testing.json":        map[string]interface{}{"values": map[string]interface{}{"a": "2"}},
			"qa.json":             map[string]interface{}{"values": map[string]interface{}{"a": "3"}},
			"production.json":     map[string]interface{}{"values": map[string]interface{}{"a": "4"}},
		},
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
				"name":   "yell",
				"type":   "SCHEDULED",
				"config": map[string]interface{}{"schedule": "0 0 * * 1", "skip_catchup_events": false},
				"event_processors": map[string]interface{}{
					"FUNCTION": map[string]interface{}{
						"config": map[string]interface{}{"function_name": "test"},
					},
				},
				"disabled": false,
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
		Values: []map[string]interface{}{},
		LogForwarders: []map[string]interface{}{
			{
				"name":         "forwarder",
				"log_types":    []interface{}{"auth"},
				"log_statuses": []interface{}{"error", "success"},
				"policy": map[string]interface{}{
					"type": "single",
				},
				"action": map[string]interface{}{
					"type": "function",
					"name": "test",
				},
				"disabled": false,
			},
		},
		DataAPIConfig: map[string]interface{}{
			"versions":                     []interface{}{"v1"},
			"run_as_user_id":               "",
			"run_as_user_id_script_source": "exports = function () { return 'goofygoof'; }",
			"disabled":                     false,
			"validation_method":            "NO_VALIDATION",
			"return_type":                  "EJSON",
			"fetch_custom_user_data":       true,
			"create_user_on_auth":          true,
			"secret_name":                  "",
			"log_function_arguments":       false,
		},
	}}
}

func appDataV2(app realm.App) local.AppDataV2 {
	return local.AppDataV2{local.AppStructureV2{
		ConfigVersion:         realm.AppConfigVersion20210101,
		ID:                    app.ClientAppID,
		Name:                  app.Name,
		Location:              app.Location,
		DeploymentModel:       app.DeploymentModel,
		AllowedRequestOrigins: []string{"http://localhost:8080"},
		Environments: map[string]map[string]interface{}{
			"no-environment.json": map[string]interface{}{"values": map[string]interface{}{"a": "0"}},
			"development.json":    map[string]interface{}{"values": map[string]interface{}{"a": "1"}},
			"testing.json":        map[string]interface{}{"values": map[string]interface{}{"a": "2"}},
			"qa.json":             map[string]interface{}{"values": map[string]interface{}{"a": "3"}},
			"production.json":     map[string]interface{}{"values": map[string]interface{}{"a": "4"}},
		},
		Auth: local.AuthStructure{
			CustomUserData: map[string]interface{}{"enabled": true, "mongo_service_name": "mdb", "database_name": "db", "collection_name": "coll", "user_id_field": "uid"},
			Providers: map[string]interface{}{
				realm.AuthProviderTypeAnonymous.String(): map[string]interface{}{
					"name":     realm.AuthProviderTypeAnonymous.String(),
					"type":     realm.AuthProviderTypeAnonymous.String(),
					"disabled": false,
				},
				realm.AuthProviderTypeAPIKey.String(): map[string]interface{}{
					"name":     realm.AuthProviderTypeAPIKey.String(),
					"type":     realm.AuthProviderTypeAPIKey.String(),
					"disabled": false,
				},
			},
		},
		DataSources: []local.DataSourceStructure{
			{
				Config: map[string]interface{}{"name": "mdb", "type": "mongodb-atlas", "config": map[string]interface{}{
					"clusterName":         "Cluster0",
					"readPreference":      "primary",
					"wireProtocolEnabled": false,
				}},
				Rules: []map[string]interface{}{{
					"database":   "db",
					"collection": "coll",
					"schema": map[string]interface{}{
						"title": "schemaTitle",
						"properties": map[string]interface{}{
							"name": map[string]interface{}{
								"bsonType": "string",
							},
							"country": map[string]interface{}{
								"bsonType": "string",
							},
						},
					},
					"relationships": map[string]interface{}{
						"name": map[string]interface{}{
							"ref":         "#/relationship/mongodb-atlas/db/coll",
							"source_key":  "name",
							"foreign_key": "country",
							"is_list":     false,
						},
						"country": map[string]interface{}{
							"ref":         "#/relationship/mongodb-atlas/db/coll",
							"source_key":  "country",
							"foreign_key": "name",
							"is_list":     false,
						},
					},
				}},
			},
		},
		HTTPServices: []local.HTTPServiceStructure{
			{
				Config: map[string]interface{}{"name": "api", "type": "http", "config": map[string]interface{}{}},
				IncomingWebhooks: []map[string]interface{}{{
					"config": map[string]interface{}{
						"name":                         "api_webhook",
						"options":                      map[string]interface{}{"validationMethod": "VERIFY_PAYLOAD", "secret": "the_secret"},
						"run_as_user_id":               "",
						"run_as_user_id_script_source": "",
						"run_as_authed_user":           false,
						"create_user_on_auth":          false,
						"fetch_custom_user_data":       false,
						"respond_result":               false,
					},
					"source": "exports = function() { return false }",
				}},
				Rules: []map[string]interface{}{{
					"name":    "rule",
					"actions": []interface{}{"get"},
					"when": map[string]interface{}{`%%args.url.host`: map[string]interface{}{
						`%in`: []interface{}{"google.com"},
					}},
				}},
			},
		},
		Sync: local.SyncStructure{
			Config: map[string]interface{}{"development_mode_enabled": true},
		},
		Functions: local.FunctionsStructure{
			Configs: []map[string]interface{}{
				{"name": "test", "private": true},
			},
			Sources: map[string]string{
				"test.js": `exports = function(){
		  console.log('got heem!');
		};`,
			},
		},
		Triggers: []map[string]interface{}{
			{
				"name": "onInsert",
				"type": "DATABASE",
				"config": map[string]interface{}{
					"service_name":                "mdb",
					"database":                    "db",
					"collection":                  "coll",
					"operation_types":             []interface{}{"INSERT"},
					"skip_catchup_events":         false,
					"unordered":                   false,
					"full_document":               false,
					"full_document_before_change": false,
					"match":                       map[string]interface{}{},
					"project":                     map[string]interface{}{},
				},
				"event_processors": map[string]interface{}{
					"FUNCTION": map[string]interface{}{
						"config": map[string]interface{}{"function_name": "test"},
					},
				},
				"disabled": false,
			},
			{
				"name":   "yell",
				"type":   "SCHEDULED",
				"config": map[string]interface{}{"schedule": "0 0 * * 1", "skip_catchup_events": false},
				"event_processors": map[string]interface{}{
					"FUNCTION": map[string]interface{}{
						"config": map[string]interface{}{"function_name": "test"},
					},
				},
				"disabled": false,
			},
		},
		GraphQL: local.GraphQLStructure{
			Config: map[string]interface{}{"use_natural_pluralization": true},
			CustomResolvers: []map[string]interface{}{
				{
					"function_name":       "test",
					"on_type":             "Query",
					"field_name":          "result",
					"input_type_format":   "scalar",
					"input_type":          "number",
					"payload_type_format": "scalar",
					"payload_type":        "number",
				},
			},
		},
		Values: []map[string]interface{}{},
		LogForwarders: []map[string]interface{}{
			{
				"name":         "forwarder",
				"log_types":    []interface{}{"auth"},
				"log_statuses": []interface{}{"error", "success"},
				"policy": map[string]interface{}{
					"type": "single",
				},
				"action": map[string]interface{}{
					"type": "function",
					"name": "test",
				},
				"disabled": false,
			},
		},
		Endpoints: local.EndpointStructure{
			Configs: []map[string]interface{}{
				{
					"create_user_on_auth":    true,
					"disabled":               true,
					"fetch_custom_user_data": true,
					"function_name":          "test",
					"http_method":            "GET",
					"respond_result":         true,
					"route":                  "/hello/world",
					"validation_method":      "NO_VALIDATION",
				},
				{
					"create_user_on_auth":    false,
					"disabled":               false,
					"fetch_custom_user_data": false,
					"function_name":          "test",
					"http_method":            "POST",
					"respond_result":         false,
					"route":                  "/hello/world",
					"validation_method":      "NO_VALIDATION",
				},
			},
		},
		DataAPIConfig: map[string]interface{}{
			"versions":                     []interface{}{"v1"},
			"run_as_user_id":               "",
			"run_as_user_id_script_source": "exports = function () { return 'goofygoof'; }",
			"disabled":                     false,
			"validation_method":            "NO_VALIDATION",
			"return_type":                  "EJSON",
			"fetch_custom_user_data":       true,
			"create_user_on_auth":          true,
			"secret_name":                  "",
			"log_function_arguments":       false,
		},
	}}
}

func parseZipPkg(t *testing.T, zipPkg *zip.Reader) map[string]string {
	t.Helper()

	out := make(map[string]string)
	for _, file := range zipPkg.File {
		if file.FileInfo().IsDir() {
			continue
		}
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
