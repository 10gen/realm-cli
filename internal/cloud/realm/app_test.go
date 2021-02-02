package realm_test

import (
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestRealmApps(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	t.Run("Should fail without an auth client", func(t *testing.T) {
		client := realm.NewClient(u.RealmServerURL())

		_, err := client.FindApps(realm.AppFilter{})
		assert.Equal(t, realm.ErrInvalidSession{}, err)
	})

	t.Run("With an active session ", func(t *testing.T) {
		client := newAuthClient(t)
		groupID := u.CloudGroupID()

		t.Run("Should create an app", func(t *testing.T) {
			name := "eggcorn"

			app, appErr := client.CreateApp(groupID, name, realm.AppMeta{})
			assert.Nil(t, appErr)

			assert.NotEqualf(t, "", app.ID, "expected app to have id")
			assert.Equal(t, groupID, app.GroupID)
			assert.Equal(t, name, app.Name)
			assert.True(t, strings.HasPrefix(app.ClientAppID, name), "expected client app id to be prefixed with name")

			t.Run("And find the app by id", func(t *testing.T) {
				apps, err := client.FindApps(realm.AppFilter{App: app.ClientAppID})
				assert.Nil(t, err)
				assert.Equal(t, []realm.App{app}, apps)
			})

			t.Run("And delete the app by id", func(t *testing.T) {
				assert.Nil(t, client.DeleteApp(groupID, app.ID))

				apps, err := client.FindApps(realm.AppFilter{App: app.ClientAppID})
				assert.Nil(t, err)
				assert.Equal(t, []realm.App{}, apps)
			})
		})
	})
}

func setupTestApp(t *testing.T, client realm.Client, groupID, name string) (realm.App, func()) {
	t.Helper()
	app, err := client.CreateApp(groupID, name, realm.AppMeta{})
	assert.Nil(t, err)
	teardown := func() {
		if deleteErr := client.DeleteApp(groupID, app.ID); deleteErr != nil {
			t.Logf("warning: failed to delete test app (id: %s): %s", app.ID, deleteErr)
		}
	}
	return app, teardown
}
