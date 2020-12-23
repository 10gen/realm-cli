package realm_test

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestRealmImportExport(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	groupID := u.CloudGroupID()

	t.Run("Should fail without an auth client", func(t *testing.T) {
		client := realm.NewClient(u.RealmServerURL())

		err := client.Import(groupID, primitive.NewObjectID().Hex(), realm.ImportRequest{})
		assert.Equal(t, realm.ErrInvalidSession, err)
	})

	t.Run("With an active session", func(t *testing.T) {
		client := newAuthClient(t)

		app, appErr := client.CreateApp(groupID, "importexport-test", realm.AppMeta{})
		assert.Nil(t, appErr)

		resetPasswordURL := "http://localhost:8080/reset_password"
		emailConfirmationURL := "http://localhost:8080/confirm_email"

		t.Run("Should import an app with auth providers", func(t *testing.T) {
			err := client.Import(groupID, app.ID, realm.ImportRequest{
				AuthProviders: []realm.AuthProvider{
					{Name: "api-key", Type: "api-key"},
					{Name: "local-userpass", Type: "local-userpass", Config: map[string]interface{}{
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
				appConfig, appConfigOK := exported[realm.FileAppConfig]
				assert.True(t, appConfigOK, "expected exported app to have file: %s", realm.FileAppConfig)
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
    },
    "environment": "none"
}
`, app.ClientAppID, realm.DefaultAppConfigVersion), appConfig)
			})

			t.Run("And the auth provider contents should be as expected", func(t *testing.T) {
				apiKeyConfigFilepath := realm.FileAuthProvider("api-key")
				apiKeyConfigPayload, apiKeyConfigOK := exported[apiKeyConfigFilepath]
				assert.True(t, apiKeyConfigOK, "expected exported app to have file: %s", apiKeyConfigFilepath)

				var apiKeyConfig map[string]interface{}
				assert.Nil(t, json.Unmarshal([]byte(apiKeyConfigPayload), &apiKeyConfig))
				assert.Nilf(t, apiKeyConfig["id"], "expected api-key.json to not have an 'id' field")
				assert.Equal(t, "api-key", apiKeyConfig["name"])
				assert.Equal(t, "api-key", apiKeyConfig["type"])
				assert.False(t, apiKeyConfig["disabled"], "expected api-key.json to have 'disabled' field set to false")

				localUserpassConfigFilepath := realm.FileAuthProvider("local-userpass")
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
				graphQLConfig, graphQLConfigOK := exported[realm.FileGraphQLConfig]
				assert.True(t, graphQLConfigOK, "expected exported app to have file: %s", realm.FileGraphQLConfig)
				assert.Equal(t, `{
    "use_natural_pluralization": true
}
`, graphQLConfig)
			})
		})
	})
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
