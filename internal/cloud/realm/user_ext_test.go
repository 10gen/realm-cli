package realm_test

import (
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestRealmUsers(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	t.Run("Should fail without an auth client", func(t *testing.T) {
		client := realm.NewClient(u.RealmServerURL())

		_, err := client.FindUsers(primitive.NewObjectID().Hex(), primitive.NewObjectID().Hex(), realm.UserFilter{})
		assert.Equal(t, realm.ErrInvalidSession(user.DefaultProfile), err)
	})

	t.Run("With an active session", func(t *testing.T) {
		client := newAuthClient(t)
		groupID := u.CloudGroupID()

		app, teardown := setupTestApp(t, client, groupID, "users-test")
		defer teardown()

		assert.Nil(t, client.Import(groupID, app.ID, local.AppDataV2{local.AppStructureV2{
			ConfigVersion:   realm.AppConfigVersion20210101,
			ID:              app.ClientAppID,
			Name:            app.Name,
			Location:        app.Location,
			DeploymentModel: app.DeploymentModel,
			Auth: local.AuthStructure{
				Providers: map[string]interface{}{
					"api-key": map[string]interface{}{"name": "api-key", "type": "api-key"},
					"local-userpass": map[string]interface{}{"name": "local-userpass", "type": "local-userpass", "config": map[string]interface{}{
						"resetPasswordUrl":     "http://localhost:8080/reset_password",
						"emailConfirmationUrl": "http://localhost:8080/confirm_email",
					}},
				},
			},
		}}))

		t.Run("Should create users", func(t *testing.T) {
			email1, createErr := client.CreateUser(groupID, app.ID, "one@domain.com", "password1")
			assert.Nil(t, createErr)
			email2, createErr := client.CreateUser(groupID, app.ID, "two@domain.com", "password2")
			assert.Nil(t, createErr)
			email3, createErr := client.CreateUser(groupID, app.ID, "three@domain.com", "password3")
			assert.Nil(t, createErr)

			apiKey1, createErr := client.CreateAPIKey(groupID, app.ID, "one")
			assert.Nil(t, createErr)
			apiKey2, createErr := client.CreateAPIKey(groupID, app.ID, "two")
			assert.Nil(t, createErr)

			apiKeyIDs := map[string]string{
				apiKey1.ID: "",
				apiKey2.ID: "",
			}

			t.Run("And find all types of users", func(t *testing.T) {
				users, err := client.FindUsers(groupID, app.ID, realm.UserFilter{})
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
					assert.True(t, ok, "expected %s to match a previously created API Key id", identity.UID)
					apiKeyIDs[identity.UID] = user.ID
				}
			})

			t.Run("And find a certain type of user", func(t *testing.T) {
				users, err := client.FindUsers(groupID, app.ID, realm.UserFilter{Providers: []realm.AuthProviderType{realm.AuthProviderTypeUserPassword}})
				assert.Nil(t, err)
				assert.Equal(t, []realm.User{email1, email2, email3}, users)
			})

			t.Run("And find specific user ids", func(t *testing.T) {
				users, err := client.FindUsers(groupID, app.ID, realm.UserFilter{IDs: []string{email2.ID, email3.ID}})
				assert.Nil(t, err)
				assert.Equal(t, []realm.User{email2, email3}, users)
			})

			t.Run("And disable users", func(t *testing.T) {
				assert.Nil(t, client.DisableUser(groupID, app.ID, email1.ID))
				users1, users1Err := client.FindUsers(groupID, app.ID, realm.UserFilter{IDs: []string{email1.ID}})
				assert.Nil(t, users1Err)
				assert.Equal(t, 1, len(users1))
				email1Disabled := users1[0]
				assert.True(t, email1Disabled.Disabled, fmt.Sprintf("expected %s to be disabled", email1Disabled.Data["email"]))

				assert.Nil(t, client.DisableUser(groupID, app.ID, email3.ID))
				users3, users3Err := client.FindUsers(groupID, app.ID, realm.UserFilter{IDs: []string{email3.ID}})
				assert.Nil(t, users3Err)
				assert.Equal(t, 1, len(users3))
				email3Disabled := users3[0]
				assert.True(t, email3Disabled.Disabled, fmt.Sprintf("expected %s to be disabled", email3Disabled.Data["email"]))

				t.Run("And find all disabled users", func(t *testing.T) {
					users, err := client.FindUsers(groupID, app.ID, realm.UserFilter{State: realm.UserStateDisabled})
					assert.Nil(t, err)
					assert.Equal(t, []realm.User{email1Disabled, email3Disabled}, users)
				})

				t.Run("And find specific user using all filter options", func(t *testing.T) {
					filter := realm.UserFilter{
						IDs:       []string{email2.ID, email3.ID, apiKeyIDs[apiKey1.ID]},
						State:     realm.UserStateDisabled,
						Providers: []realm.AuthProviderType{realm.AuthProviderTypeUserPassword},
					}
					users, err := client.FindUsers(groupID, app.ID, filter)
					assert.Nil(t, err)
					assert.Equal(t, []realm.User{email3Disabled}, users)
				})
			})

			t.Run("And enable users", func(t *testing.T) {
				assert.Nil(t, client.EnableUser(groupID, app.ID, email1.ID))
				users1, users1Err := client.FindUsers(groupID, app.ID, realm.UserFilter{IDs: []string{email1.ID}})
				assert.Nil(t, users1Err)
				email1Enabled := users1[0]
				assert.False(t, email1Enabled.Disabled, fmt.Sprintf("expected %s to be enabled", email1Enabled.Data["email"]))

				assert.Nil(t, client.EnableUser(groupID, app.ID, email3.ID))
				users3, users3Err := client.FindUsers(groupID, app.ID, realm.UserFilter{IDs: []string{email3.ID}})
				assert.Nil(t, users3Err)
				email3Enabled := users3[0]
				assert.False(t, email3Enabled.Disabled, fmt.Sprintf("expected %s to be enabled", email3Enabled.Data["email"]))

				t.Run("And find all enabled users", func(t *testing.T) {
					users, err := client.FindUsers(groupID, app.ID, realm.UserFilter{State: realm.UserStateEnabled})
					assert.Nil(t, err)
					assert.Equal(t, 5, len(users))

					emailUsers := make([]realm.User, 0, 3)
					apiKeyIDs := make([]string, 0, 2)
					for _, user := range users {
						assert.Equalf(t, 1, len(user.Identities), "expected user to have only one identity")
						switch user.Identities[0].ProviderType {
						case realm.AuthProviderTypeUserPassword:
							emailUsers = append(emailUsers, user)
						case realm.AuthProviderTypeAPIKey:
							apiKeyIDs = append(apiKeyIDs, user.Identities[0].UID)
						}
					}
					assert.Equal(t, []realm.User{email1, email2, email3}, emailUsers)
					assert.Equal(t, []string{apiKey1.ID, apiKey2.ID}, apiKeyIDs)
				})
			})

			t.Run("And find enabled user/password users", func(t *testing.T) {
				filter := realm.UserFilter{State: realm.UserStateEnabled, Providers: []realm.AuthProviderType{realm.AuthProviderTypeUserPassword}, IDs: []string{email1.ID, email2.ID, email3.ID}}
				users, err := client.FindUsers(groupID, app.ID, filter)
				assert.Nil(t, err)
				for _, user := range users {
					assert.False(t, user.Disabled, fmt.Sprintf("expected %s to be enabled", user.Data["email"]))
				}
			})

			t.Run("And revoking a user session should succeed", func(t *testing.T) {
				assert.Nil(t, client.RevokeUserSessions(groupID, app.ID, email1.ID))
			})

			t.Run("And delete users", func(t *testing.T) {
				for _, userID := range []string{email1.ID, email2.ID, email3.ID, apiKeyIDs[apiKey1.ID], apiKeyIDs[apiKey2.ID]} {
					assert.Nilf(t, client.DeleteUser(groupID, app.ID, userID), "failed to successfully delete user: %s", userID)
				}
			})
		})

		t.Run("And finding pending users should return an empty list", func(t *testing.T) {
			users, err := client.FindUsers(groupID, app.ID, realm.UserFilter{Pending: true})
			assert.Nil(t, err)
			assert.Equal(t, []realm.User{}, users)
		})
	})
}
