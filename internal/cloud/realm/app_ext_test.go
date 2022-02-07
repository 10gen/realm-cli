package realm_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestRealmApps(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	t.Run("should fail without an auth client", func(t *testing.T) {
		client := realm.NewClient(u.RealmServerURL())

		_, err := client.FindApps(realm.AppFilter{})
		assert.Equal(t, realm.ErrInvalidSession(user.DefaultProfile), err)
	})

	t.Run("with an active session", func(t *testing.T) {
		client := newAuthClient(t)
		groupID := u.CloudGroupID()

		t.Run("should create an app", func(t *testing.T) {
			name := "eggcorn"

			app, appErr := client.CreateApp(groupID, name, realm.AppMeta{})
			assert.Nil(t, appErr)

			assert.NotEqualf(t, "", app.ID, "expected app to have id")
			assert.Equal(t, groupID, app.GroupID)
			assert.Equal(t, name, app.Name)
			assert.True(t, strings.HasPrefix(app.ClientAppID, name), "expected client app id to be prefixed with name")

			t.Run("and find the app by client app id", func(t *testing.T) {
				apps, err := client.FindApps(realm.AppFilter{App: app.ClientAppID})
				assert.Nil(t, err)
				assert.Equal(t, []realm.App{app}, apps)
			})

			// TODO(REALMC-9462): remove this once /apps has "template_id" in the payload
			t.Run("and find the app by group and app id", func(t *testing.T) {
				found, err := client.FindApp(app.GroupID, app.ID)
				assert.Nil(t, err)
				assert.Equal(t, app.ID, found.ID)
			})

			t.Run("and get the app description by id", func(t *testing.T) {
				appDesc, err := client.AppDescription(groupID, app.ID)
				assert.Nil(t, err)
				assert.Match(t, realm.AppDescription{
					ClientAppID:    app.ClientAppID,
					Name:           app.Name,
					RealmURL:       fmt.Sprintf("%s/groups/%s/apps/%s/dashboard", u.RealmServerURL(), groupID, app.ID),
					DataSources:    []realm.DataSourceSummary{},
					HTTPEndpoints:  realm.HTTPEndpoints{Summaries: []interface{}{}},
					ServiceDescs:   []realm.ServiceSummary{},
					AuthProviders:  []realm.AuthProviderSummary{{"api-key", "api-key", false}},
					CustomUserData: realm.CustomUserDataSummary{},
					Values:         []string{},
					Functions:      []realm.FunctionSummary{},
					Sync:           realm.SyncSummary{},
					GraphQL: realm.GraphQLSummary{
						URL:             fmt.Sprintf("%s/api/client/v2.0/app/%s/graphql", u.RealmServerURL(), app.ClientAppID),
						CustomResolvers: []string{},
					},
					Environment:       "",
					EventSubscription: []realm.EventSubscriptionSummary{},
					LogForwarders:     []realm.LogForwarderSummary{},
				}, appDesc)
			})

			t.Run("and delete the app by id", func(t *testing.T) {
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
