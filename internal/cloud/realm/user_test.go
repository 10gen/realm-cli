package realm_test

import (
	"errors"
	"fmt"
	"testing"

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

		app, appErr := client.CreateApp(groupID, "users-test", realm.AppMeta{})
		assert.Nil(t, appErr)

		assert.Nil(t, client.Import(groupID, app.ID, realm.ImportRequest{
			AuthProviders: []realm.AuthProvider{
				{Name: "api-key", Type: "api-key"},
				{Name: "local-userpass", Type: "local-userpass", Config: map[string]interface{}{
					"resetPasswordUrl":     "http://localhost:8080/reset_password",
					"emailConfirmationUrl": "http://localhost:8080/confirm_email",
				}},
			},
		}))

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
					case "local-userpass":
						emailUsers = append(emailUsers, user)
					case "api-key":
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
				users, err := client.FindUsers(groupID, app.ID, realm.UserFilter{Providers: []realm.ProviderType{"local-userpass"}})
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
				assert.Nil(t, client.DisableUser(groupID, app.ID, email3.ID))
			})

			t.Run("And find all disabled users", func(t *testing.T) {
				users, err := client.FindUsers(groupID, app.ID, realm.UserFilter{State: realm.UserStateDisabled})
				assert.Nil(t, err)

				email1.Disabled = true
				email3.Disabled = true

				assert.Equal(t, []realm.User{email1, email3}, users)
			})

			t.Run("And find specific user using all filter options", func(t *testing.T) {
				filter := realm.UserFilter{
					IDs:       []string{email2.ID, email3.ID, apiKeyIDs[apiKey1.ID]},
					State:     realm.UserStateDisabled,
					Providers: []realm.ProviderType{"local-userpass"},
				}
				users, err := client.FindUsers(groupID, app.ID, filter)
				assert.Nil(t, err)
				assert.Equal(t, []realm.User{email3}, users)
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

func TestProviderTypeIsValid(t *testing.T) {
	for _, tc := range realm.ValidProviderTypes {
		t.Run(fmt.Sprintf("%s should be valid", tc), func(t *testing.T) {
			assert.Nil(t, realm.ProviderType(tc).IsValid())
		})
	}
	t.Run(fmt.Sprintf("%s should be invalid", "invalid type"), func(t *testing.T) {
		assert.Equal(t, realm.ProviderType("invalid type").IsValid(), errors.New("Invalid ProviderType"))
	})
}

func TestProviderTypeDisplay(t *testing.T) {
	for _, tc := range []struct {
		pt             realm.ProviderType
		expectedOutput string
	}{
		{
			pt:             realm.ProviderTypeAnonymous,
			expectedOutput: "Anonymous",
		},
		{
			pt:             realm.ProviderTypeUserPassord,
			expectedOutput: "User/Password",
		},
		{
			pt:             realm.ProviderTypeAPIKey,
			expectedOutput: "ApiKey",
		},
		{
			pt:             realm.ProviderTypeApple,
			expectedOutput: "Apple",
		},
		{
			pt:             realm.ProviderTypeGoogle,
			expectedOutput: "Google",
		},
		{
			pt:             realm.ProviderTypeFacebook,
			expectedOutput: "Facebook",
		},
		{
			pt:             realm.ProviderTypeCustomToken,
			expectedOutput: "Custom JWT",
		},
		{
			pt:             realm.ProviderTypeCustomFunction,
			expectedOutput: "Custom Function",
		},
		{
			pt:             realm.ProviderType("invalid_provider_type"),
			expectedOutput: "Unknown",
		},
	} {
		t.Run(fmt.Sprintf("should return %s", tc.expectedOutput), func(t *testing.T) {
			assert.Equal(t, tc.pt.Display(), tc.expectedOutput)
		})
	}
}

func TestProviderTypeDisplayUser(t *testing.T) {
	testUsers := []realm.User{
		{
			ID:         "user-1",
			Identities: []realm.UserIdentity{{ProviderType: realm.ProviderTypeAnonymous}},
		},
		{
			ID:         "user-2",
			Identities: []realm.UserIdentity{{ProviderType: realm.ProviderTypeUserPassord}},
			Data:       map[string]interface{}{"email": "user-2@test.com"},
		},
		{
			ID:         "user-3",
			Identities: []realm.UserIdentity{{ProviderType: realm.ProviderTypeAPIKey}},
			Data:       map[string]interface{}{"name": "name-3"},
		},
		{
			ID:         "user-4",
			Identities: []realm.UserIdentity{{ProviderType: realm.ProviderTypeApple}},
		},
		{
			ID:         "user-5",
			Identities: []realm.UserIdentity{{ProviderType: realm.ProviderTypeGoogle}},
		},
		{
			ID:         "user-6",
			Identities: []realm.UserIdentity{{ProviderType: realm.ProviderTypeFacebook}},
		},
		{
			ID:         "user-7",
			Identities: []realm.UserIdentity{{ProviderType: realm.ProviderTypeCustomToken}},
		},
		{
			ID:         "user-8",
			Identities: []realm.UserIdentity{{ProviderType: realm.ProviderTypeCustomFunction}},
		},
	}
	for _, tc := range []struct {
		pt             realm.ProviderType
		user           realm.User
		expectedOutput string
	}{
		{
			pt:             realm.ProviderTypeAnonymous,
			user:           testUsers[0],
			expectedOutput: "Anonymous - user-1",
		},
		{
			pt:             realm.ProviderTypeUserPassord,
			user:           testUsers[1],
			expectedOutput: "User/Password - user-2@test.com - user-2",
		},
		{
			pt:             realm.ProviderTypeAPIKey,
			user:           testUsers[2],
			expectedOutput: "ApiKey - name-3 - user-3",
		},
		{
			pt:             realm.ProviderTypeApple,
			user:           testUsers[3],
			expectedOutput: "Apple - user-4",
		},
		{
			pt:             realm.ProviderTypeGoogle,
			user:           testUsers[4],
			expectedOutput: "Google - user-5",
		},
		{
			pt:             realm.ProviderTypeFacebook,
			user:           testUsers[5],
			expectedOutput: "Facebook - user-6",
		},
		{
			pt:             realm.ProviderTypeCustomToken,
			user:           testUsers[6],
			expectedOutput: "Custom JWT - user-7",
		},
		{
			pt:             realm.ProviderTypeCustomFunction,
			user:           testUsers[7],
			expectedOutput: "Custom Function - user-8",
		},
	} {
		t.Run(fmt.Sprintf("should return %s", tc.expectedOutput), func(t *testing.T) {
			assert.Equal(t, tc.pt.DisplayUser(tc.user), tc.expectedOutput)
		})
	}
}

func TestStringSliceToProviderTypes(t *testing.T) {
	for _, tc := range []struct {
		inSlice  []string
		outSlice []realm.ProviderType
	}{
		{
			inSlice: []string{"anon-user", "local-userpass", "api-key"},
			outSlice: []realm.ProviderType{
				realm.ProviderTypeAnonymous,
				realm.ProviderTypeUserPassord,
				realm.ProviderTypeAPIKey,
			},
		},
		{
			inSlice: []string{"anon-user", "local-userpass", "api-key", "oauth2-facebook", "oauth2-google", "oauth2-apple"},
			outSlice: []realm.ProviderType{
				realm.ProviderTypeAnonymous,
				realm.ProviderTypeUserPassord,
				realm.ProviderTypeAPIKey,
				realm.ProviderTypeFacebook,
				realm.ProviderTypeGoogle,
				realm.ProviderTypeApple,
			},
		},
	} {
		t.Run("should return provider type slice", func(t *testing.T) {
			assert.Equal(t, realm.StringSliceToProviderTypes(tc.inSlice...), tc.outSlice)
		})
	}
}

func TestJoinProviderTypes(t *testing.T) {
	for _, tc := range []struct {
		pts            []realm.ProviderType
		sep            string
		expectedOutput string
	}{
		{
			pts: []realm.ProviderType{
				realm.ProviderTypeAnonymous,
				realm.ProviderTypeUserPassord,
				realm.ProviderTypeAPIKey,
			},
			sep:            ",",
			expectedOutput: "anon-user,local-userpass,api-key",
		},
		{
			pts: []realm.ProviderType{
				realm.ProviderTypeAnonymous,
				realm.ProviderTypeUserPassord,
				realm.ProviderTypeAPIKey,
				realm.ProviderTypeFacebook,
				realm.ProviderTypeGoogle,
				realm.ProviderTypeApple,
			},
			sep:            ", ",
			expectedOutput: "anon-user, local-userpass, api-key, oauth2-facebook, oauth2-google, oauth2-apple",
		},
	} {
		t.Run(fmt.Sprintf("should return %s", tc.expectedOutput), func(t *testing.T) {
			assert.Equal(t, realm.JoinProviderTypes(tc.sep, tc.pts...), tc.expectedOutput)
		})
	}
}
