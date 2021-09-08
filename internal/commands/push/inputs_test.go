package push

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPushInputsResolve(t *testing.T) {
	t.Run("Should return an error if run from outside a project directory and no local flag is set", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_init_input_test")
		defer teardown()

		var i inputs
		assert.Equal(t, errProjectNotFound{}, i.Resolve(profile, nil))
	})

	t.Run("Should return an error if both include node modules and package json flags are set", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_init_input_test")
		defer teardown()

		i := inputs{IncludePackageJSON: true, IncludeNodeModules: true}
		assert.Equal(t, errDependencyFlagConflict, i.Resolve(profile, nil))
	})

	t.Run("Should set the app data if no flags are set but is run from inside a project directory", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_init_input_test")
		defer teardown()

		assert.Nil(t, ioutil.WriteFile(
			filepath.Join(profile.WorkingDirectory, local.FileConfig.String()),
			[]byte(`{"app_id": "eggcorn-abcde", "name":"eggcorn"}`),
			0666,
		))

		var i inputs
		assert.Nil(t, i.Resolve(profile, nil))

		assert.Equal(t, profile.WorkingDirectory, i.LocalPath)
		assert.Equal(t, "eggcorn-abcde", i.RemoteApp)
	})
}

func TestPushInputsResolveTo(t *testing.T) {
	t.Run("Should do nothing if to is not set", func(t *testing.T) {
		var i inputs
		tt, err := i.resolveRemoteApp(nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, appRemote{}, tt)
	})

	t.Run("Should return the app id and group id of specified app if to is set to app", func(t *testing.T) {
		var appFilter realm.AppFilter
		app := realm.App{
			ID:          primitive.NewObjectID().Hex(),
			GroupID:     primitive.NewObjectID().Hex(),
			ClientAppID: "test-app-abcde",
			Name:        "test-app",
		}

		client := mock.RealmClient{}
		client.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			appFilter = filter
			return []realm.App{app}, nil
		}

		i := inputs{Project: app.GroupID, RemoteApp: app.ClientAppID}

		f, err := i.resolveRemoteApp(nil, client)
		assert.Nil(t, err)

		assert.Equal(t, appRemote{GroupID: app.GroupID, AppID: app.ID}, f)
		assert.Equal(t, realm.AppFilter{GroupID: app.GroupID, App: app.ClientAppID}, appFilter)
	})
}
