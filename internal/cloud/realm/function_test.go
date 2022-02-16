package realm_test

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestFunctions(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	t.Run("should fail without an auth client", func(t *testing.T) {
		client := realm.NewClient(u.RealmServerURL())

		_, err := client.Functions(u.CloudGroupID(), "test-app-1234")
		assert.Equal(t, realm.ErrInvalidSession(user.DefaultProfile), err)
	})

	t.Run("should return list of functions associated with app", func(t *testing.T) {
		client := newAuthClient(t)

		groupID := u.CloudGroupID()

		app, teardown := setupTestApp(t, client, groupID, "functions-test")
		defer teardown()

		t.Run("should find 0 functions", func(t *testing.T) {
			functions, err := client.Functions(u.CloudGroupID(), app.ID)
			assert.Nil(t, err)

			assert.Equal(t, 0, len(functions))
		})

		t.Run("should find 1 function", func(t *testing.T) {
			appData := local.AppDataV2{local.AppStructureV2{
				ConfigVersion:   realm.AppConfigVersion20210101,
				ID:              app.ClientAppID,
				Name:            app.Name,
				Location:        app.Location,
				DeploymentModel: app.DeploymentModel,
				Functions: local.FunctionsStructure{
					Configs: []map[string]interface{}{
						{"name": "test", "private": true},
					},
					Sources: map[string]string{
						"test.js": "exports = function(){\n  return \"successful test\";\n};",
					},
				},
			}}

			err := client.Import(groupID, app.ID, appData)
			assert.Nil(t, err)

			functions, err := client.Functions(u.CloudGroupID(), app.ID)
			assert.Nil(t, err)

			assert.Equal(t, 1, len(functions))
			assert.Equal(t, "test", functions[0].Name)
		})
	})
}

func TestAppDebugExecuteFunction(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	t.Run("should fail without an auth client", func(t *testing.T) {
		client := realm.NewClient(u.RealmServerURL())

		_, err := client.AppDebugExecuteFunction(u.CloudGroupID(), "test-app-1234", "", "test-function", nil)
		assert.Equal(t, realm.ErrInvalidSession(user.DefaultProfile), err)
	})

	t.Run("should execute function", func(t *testing.T) {
		client := newAuthClient(t)

		groupID := u.CloudGroupID()

		app, teardown := setupTestApp(t, client, groupID, "app-debug-execute-function-test")
		defer teardown()

		appData := local.AppDataV2{local.AppStructureV2{
			ConfigVersion:   realm.AppConfigVersion20210101,
			ID:              app.ClientAppID,
			Name:            app.Name,
			Location:        app.Location,
			DeploymentModel: app.DeploymentModel,
			Functions: local.FunctionsStructure{
				Configs: []map[string]interface{}{
					{"name": "simple_test", "private": true},
					{"name": "passed_args_test", "private": true},
				},
				Sources: map[string]string{
					"simple_test.js":      "exports = function(){\n  return \"successful test\";\n};",
					"passed_args_test.js": "exports = function(arg1, arg2){\n  return {arg1: arg1, arg2: arg2};\n};",
				},
			},
		}}

		err := client.Import(groupID, app.ID, appData)
		assert.Nil(t, err)

		t.Run("should return string", func(t *testing.T) {
			response, err := client.AppDebugExecuteFunction(u.CloudGroupID(), app.ID, "", "simple_test", nil)
			assert.Nil(t, err)

			assert.Equal(t, "successful test", response.Result)
		})

		t.Run("should return passed args", func(t *testing.T) {
			args := []interface{}{
				map[string]interface{}{
					"value1": 1,
					"abcs":   []string{"x", "y", "z"},
				},
				[]int{1, 2},
			}

			response, err := client.AppDebugExecuteFunction(u.CloudGroupID(), app.ID, "", "passed_args_test", args)
			assert.Nil(t, err)

			assert.Equal(t, map[string]interface{}{
				"arg1": map[string]interface{}{
					"value1": map[string]interface{}{"$numberInt": "1"},
					"abcs":   []interface{}{"x", "y", "z"},
				},
				"arg2": []interface{}{
					map[string]interface{}{
						"$numberInt": "1",
					},
					map[string]interface{}{
						"$numberInt": "2",
					},
				},
			}, response.Result)
		})
	})
}
