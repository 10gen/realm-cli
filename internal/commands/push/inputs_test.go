package push

import (
	"errors"
	"fmt"
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

func TestPushInputsResolve(t *testing.T) {
	t.Run("should return an error if run from outside a project directory and no local flag is set", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_init_input_test")
		defer teardown()

		var i inputs
		assert.Equal(t, errProjectInvalid(profile.WorkingDirectory, true), i.Resolve(profile, nil))
	})

	t.Run("should return an error when more than one dependencies flag is set", func(t *testing.T) {
		t.Run("when include node modules and include package json are both set", func(t *testing.T) {
			i := inputs{IncludeNodeModules: true, IncludePackageJSON: true}
			assert.Equal(t, errors.New(`cannot use both "include-node-modules" and "include-package-json" at the same time`), i.Resolve(nil, nil))
		})

		t.Run("when include dependencies and include package json are both set", func(t *testing.T) {
			i := inputs{IncludeDependencies: true, IncludePackageJSON: true}
			assert.Equal(t, errors.New(`cannot use both "include-dependencies" and "include-package-json" at the same time`), i.Resolve(nil, nil))
		})
	})

	t.Run("should return an error when specified local path does not exist", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_init_input_test")
		defer teardown()

		localPath := "fakePath"

		i := inputs{LocalPath: localPath}
		assert.Equal(t, errProjectInvalid(localPath, false), i.Resolve(profile, nil))
	})

	t.Run("should return an error when specified local path is an absolute path and it does not exist", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_init_input_test")
		defer teardown()

		localPath := "fakePath"
		searchPathAbs, _ := filepath.Abs(localPath)

		i := inputs{LocalPath: searchPathAbs}
		assert.Equal(t, errProjectInvalid(searchPathAbs, false), i.Resolve(profile, nil))
	})

	t.Run("should return an error when specified local path is not a realm app project", func(t *testing.T) {
		localPath := "testdata"

		profile, teardown := mock.NewProfileFromTmpDir(t, localPath)
		defer teardown()

		i := inputs{LocalPath: localPath}
		assert.Equal(t, errProjectInvalid(localPath, true), i.Resolve(profile, nil))
	})

	t.Run("should set the app data if no flags are set but is run from inside a project directory", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_init_input_test")
		defer teardown()

		assert.Nil(t, ioutil.WriteFile(
			filepath.Join(profile.WorkingDirectory, local.FileRealmConfig.String()),
			[]byte(fmt.Sprintf(`{"config_version": %d, "app_id": "eggcorn-abcde", "name":"eggcorn"}`, realm.DefaultAppConfigVersion)),
			0666,
		))

		var i inputs
		assert.Nil(t, i.Resolve(profile, nil))

		assert.Equal(t, profile.WorkingDirectory, i.LocalPath)
		assert.Equal(t, "eggcorn-abcde", i.RemoteApp)
	})

	t.Run("should not set remote app if app meta is present", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_init_input_test")
		defer teardown()

		assert.Nil(t, ioutil.WriteFile(
			filepath.Join(profile.WorkingDirectory, local.FileRealmConfig.String()),
			[]byte(fmt.Sprintf(`{"config_version": %d, "app_id": "eggcorn-abcde", "name":"eggcorn"}`, realm.DefaultAppConfigVersion)),
			0666,
		))

		assert.Nil(t, os.Mkdir(filepath.Join(profile.WorkingDirectory, local.NameDotMDB), os.ModePerm))
		assert.Nil(t, ioutil.WriteFile(
			filepath.Join(profile.WorkingDirectory, local.NameDotMDB, local.FileAppMeta.String()),
			[]byte(fmt.Sprintf(`{"group_id":"groupID","app_id":"appID","config_version":%d}`, realm.DefaultAppConfigVersion)),
			0666,
		))

		var i inputs
		assert.Nil(t, i.Resolve(profile, nil))

		assert.Equal(t, profile.WorkingDirectory, i.LocalPath)
		assert.Equal(t, "", i.RemoteApp)
	})
}

func TestPushInputsResolveTo(t *testing.T) {
	t.Run("should return the app id and group id of specified app if to is set to app", func(t *testing.T) {
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

		f, err := i.resolveRemoteApp(nil, client, local.AppMeta{})
		assert.Nil(t, err)

		assert.Equal(t, appRemote{GroupID: app.GroupID, AppID: app.ID, ClientAppID: app.ClientAppID}, f)
		assert.Equal(t, realm.AppFilter{GroupID: app.GroupID, App: app.ClientAppID}, appFilter)
	})

	t.Run("should return the app id and group id from app meta when provided", func(t *testing.T) {
		app := realm.App{
			ID:          primitive.NewObjectID().Hex(),
			GroupID:     primitive.NewObjectID().Hex(),
			ClientAppID: "test-app-abcde",
			Name:        "test-app",
		}

		var client mock.RealmClient
		var findAppsCalled bool
		client.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			findAppsCalled = true
			return []realm.App{app}, nil
		}

		var i inputs
		remoteApp, err := i.resolveRemoteApp(nil, client, local.AppMeta{AppID: app.ID, GroupID: app.GroupID, ConfigVersion: realm.AppConfigVersion20210101})
		assert.Nil(t, err)

		assert.Equal(t, appRemote{AppID: app.ID, GroupID: app.GroupID}, remoteApp)
		assert.False(t, findAppsCalled, "expected app to skip resolve")
	})

	t.Run("should return minimal app remote if cannot find app", func(t *testing.T) {
		var client mock.RealmClient
		client.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{}, cli.ErrAppNotFound{App: "app"}
		}

		i := inputs{Project: "groupID", RemoteApp: "appID"}
		tt, err := i.resolveRemoteApp(nil, client, local.AppMeta{})
		assert.Nil(t, err)
		assert.Equal(t, appRemote{GroupID: "groupID"}, tt)
	})

	t.Run("should error if find apps cannot find group", func(t *testing.T) {
		var client mock.RealmClient
		client.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{}, cli.ErrGroupNotFound
		}

		var i inputs
		_, err := i.resolveRemoteApp(nil, client, local.AppMeta{})
		assert.Equal(t, cli.ErrGroupNotFound, err)
	})
}
