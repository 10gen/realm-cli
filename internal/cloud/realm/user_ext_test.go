package realm_test

import (
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/app"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestRealmUsers(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	t.Run("Should fail without an auth client", func(t *testing.T) {
		client := realm.NewClient(u.RealmServerURL())

		_, err := client.FindUsers(primitive.NewObjectID().Hex(), primitive.NewObjectID().Hex(), realm.UserFilter{})
		assert.Equal(t, realm.ErrInvalidSession{}, err)
	})

	t.Run("With an active session ", func(t *testing.T) {
		client := newAuthClient(t)
		groupID := u.CloudGroupID()

		testApp, appErr := client.CreateApp(groupID, "users-test", realm.AppMeta{})
		assert.Nil(t, appErr)

		assert.Nil(t, client.Import(groupID, testApp.ID, map[string]interface{}{
			app.NameAuthProviders: []map[string]interface{}{
				{"name": "api-key", "type": "api-key"},
				{"name": "local-userpass", "type": "local-userpass", "config": map[string]interface{}{
					"resetPasswordUrl":     "http://localhost:8080/reset_password",
					"emailConfirmationUrl": "http://localhost:8080/confirm_email",
				}},
			},
		}))

		t.Run("Should create users", func(t *testing.T) {
			email1, createErr := client.CreateUser(groupID, testApp.ID, "one@domain.com", "password1")
			assert.Nil(t, createErr)
			email2, createErr := client.CreateUser(groupID, testApp.ID, "two@domain.com", "password2")
			assert.Nil(t, createErr)
			email3, createErr := client.CreateUser(groupID, testApp.ID, "three@domain.com", "password3")
			assert.Nil(t, createErr)

			apiKey1, createErr := client.CreateAPIKey(groupID, testApp.ID, "one")
			assert.Nil(t, createErr)
			apiKey2, createErr := client.CreateAPIKey(groupID, testApp.ID, "two")
			assert.Nil(t, createErr)

			apiKeyIDs := map[string]string{
				apiKey1.ID: "",
				apiKey2.ID: "",
			}

			t.Run("And find all types of users", func(t *testing.T) {
				users, err := client.FindUsers(groupID, testApp.ID, realm.UserFilter{})
				assert.Nil(t, err)

				emailUsers := make([]realm.User, 0, 3)
				apiKeyUsers := make([]realm.User, 0, 2)

				for _, user := range users {
					assert.Equalf(t, 1, len(user.Identities), "expected user to have only one identity")
					switch user.Identities[0].ProviderType {
					case realm.AuthProviderTypeUserPassword:
						emailUsers = append(emailUsers, user)
					case realm.AuthProviderTypeAPIKey:
						apiKeyUsers = append(apiKeyUsers, user)
					}
				}

				assert.Equal(t, 3, len(emailUsers))
				assert.Equal(t, []realm.User{email1, email2, email3}, emailUsers)

				assert.Equal(t, 2, len(apiKeyUsers))
				for _, user := range apiKeyUsers {
					identity := user.Identities[0]
					_, ok := apiKeyIDs[identity.UID]
					assert.True(t, ok, "expected %s to match a previously created api key id", identity.UID)
					apiKeyIDs[identity.UID] = user.ID
				}
			})

			t.Run("And find a certain type of user", func(t *testing.T) {
				users, err := client.FindUsers(groupID, testApp.ID, realm.UserFilter{Providers: []realm.AuthProviderType{realm.AuthProviderTypeUserPassword}})
				assert.Nil(t, err)
				assert.Equal(t, []realm.User{email1, email2, email3}, users)
			})

			t.Run("And find specific user ids", func(t *testing.T) {
				users, err := client.FindUsers(groupID, testApp.ID, realm.UserFilter{IDs: []string{email2.ID, email3.ID}})
				assert.Nil(t, err)
				assert.Equal(t, []realm.User{email2, email3}, users)
			})

			t.Run("And disable users", func(t *testing.T) {
				assert.Nil(t, client.DisableUser(groupID, testApp.ID, email1.ID))
				assert.Nil(t, client.DisableUser(groupID, testApp.ID, email3.ID))
			})

			t.Run("And find all disabled users", func(t *testing.T) {
				filter := realm.UserFilter{State: realm.UserStateEnabled, Providers: []realm.AuthProviderType{realm.AuthProviderTypeUserPassword}, IDs: []string{email1.ID, email3.ID}}
				users, err := client.FindUsers(groupID, testApp.ID, filter)
				assert.Nil(t, err)
				for _, user := range users {
					assert.True(t, user.Disabled, fmt.Sprintf("expected %s to be disabled", user.Data["email"]))
				}
			})

			t.Run("And find specific user using all filter options", func(t *testing.T) {
				filter := realm.UserFilter{
					IDs:       []string{email2.ID, email3.ID, apiKeyIDs[apiKey1.ID]},
					State:     realm.UserStateDisabled,
					Providers: []realm.AuthProviderType{realm.AuthProviderTypeUserPassword},
				}
				users, err := client.FindUsers(groupID, testApp.ID, filter)
				assert.Nil(t, err)
				assert.Equal(t, 1, len(users))
				assert.Equal(t, email3.ID, users[0].ID)
				assert.True(t, users[0].Disabled, fmt.Sprintf("expected %s to be disabled", users[0].Data["email"]))
				assert.Equal(t, email3.Identities[0].ProviderType, users[0].Identities[0].ProviderType)
			})

			t.Run("And enable users", func(t *testing.T) {
				assert.Nil(t, client.EnableUser(groupID, testApp.ID, email1.ID))
				assert.Nil(t, client.EnableUser(groupID, testApp.ID, email3.ID))
			})

			t.Run("And find enabled user/password users", func(t *testing.T) {
				filter := realm.UserFilter{State: realm.UserStateEnabled, Providers: []realm.AuthProviderType{realm.AuthProviderTypeUserPassword}, IDs: []string{email1.ID, email2.ID, email3.ID}}
				users, err := client.FindUsers(groupID, testApp.ID, filter)
				assert.Nil(t, err)
				for _, user := range users {
					assert.False(t, user.Disabled, fmt.Sprintf("expected %s to be enabled", user.Data["email"]))
				}
			})

			t.Run("And revoking a user session should succeed", func(t *testing.T) {
				assert.Nil(t, client.RevokeUserSessions(groupID, testApp.ID, email1.ID))
			})

			t.Run("And delete users", func(t *testing.T) {
				for _, userID := range []string{email1.ID, email2.ID, email3.ID, apiKeyIDs[apiKey1.ID], apiKeyIDs[apiKey2.ID]} {
					assert.Nilf(t, client.DeleteUser(groupID, testApp.ID, userID), "failed to successfully delete user: %s", userID)
				}
			})
		})

		t.Run("And finding pending users should return an empty list", func(t *testing.T) {
			users, err := client.FindUsers(groupID, testApp.ID, realm.UserFilter{Pending: true})
			assert.Nil(t, err)
			assert.Equal(t, []realm.User{}, users)
		})
	})
}
