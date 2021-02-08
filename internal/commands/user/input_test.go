package user

import (
	"errors"
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"github.com/Netflix/go-expect"
)

func TestProviderTypeDisplayUser(t *testing.T) {
	testUsers := []realm.User{
		{
			ID:         "user-1",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeAnonymous}},
		},
		{
			ID:         "user-2",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeUserPassword}},
			Data:       map[string]interface{}{"email": "user-2@test.com"},
		},
		{
			ID:         "user-3",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeAPIKey}},
			Data:       map[string]interface{}{"name": "name-3"},
		},
		{
			ID:         "user-4",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeApple}},
		},
		{
			ID:         "user-5",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeGoogle}},
		},
		{
			ID:         "user-6",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeFacebook}},
		},
		{
			ID:         "user-7",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeCustomToken}},
		},
		{
			ID:         "user-8",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeCustomFunction}},
		},
	}
	for _, tc := range []struct {
		apt            realm.AuthProviderType
		user           realm.User
		expectedOutput string
	}{
		{
			apt:            realm.AuthProviderTypeAnonymous,
			user:           testUsers[0],
			expectedOutput: "Anonymous - user-1",
		},
		{
			apt:            realm.AuthProviderTypeUserPassword,
			user:           testUsers[1],
			expectedOutput: "User/Password - user-2@test.com - user-2",
		},
		{
			apt:            realm.AuthProviderTypeAPIKey,
			user:           testUsers[2],
			expectedOutput: "ApiKey - name-3 - user-3",
		},
		{
			apt:            realm.AuthProviderTypeApple,
			user:           testUsers[3],
			expectedOutput: "Apple - user-4",
		},
		{
			apt:            realm.AuthProviderTypeGoogle,
			user:           testUsers[4],
			expectedOutput: "Google - user-5",
		},
		{
			apt:            realm.AuthProviderTypeFacebook,
			user:           testUsers[5],
			expectedOutput: "Facebook - user-6",
		},
		{
			apt:            realm.AuthProviderTypeCustomToken,
			user:           testUsers[6],
			expectedOutput: "Custom JWT - user-7",
		},
		{
			apt:            realm.AuthProviderTypeCustomFunction,
			user:           testUsers[7],
			expectedOutput: "Custom Function - user-8",
		},
	} {
		t.Run(fmt.Sprintf("should return %s", tc.expectedOutput), func(t *testing.T) {
			assert.Equal(t, displayUser(tc.apt, tc.user), tc.expectedOutput)
		})
	}
}

func TestUserFilter(t *testing.T) {
	t.Run("should create a user filter from inputs", func(t *testing.T) {
		state := realm.UserStateDisabled
		providerTypes := []string{"testProvider"}
		users := []string{"testUser"}
		inputs := multiUserInputs{
			State:         state,
			ProviderTypes: providerTypes,
			Pending:       true,
			Users:         users,
		}
		assert.Equal(t, realm.UserFilter{
			IDs:       users,
			State:     state,
			Pending:   true,
			Providers: realm.NewAuthProviderTypes(providerTypes...),
		}, inputs.filter())
	})
}

func TestResolveMultiUsersInputs(t *testing.T) {
	testUsers := []realm.User{
		{
			ID:         "user-1",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeAnonymous}},
		},
		{
			ID:         "user-2",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeUserPassword}},
			Disabled:   true,
			Data:       map[string]interface{}{"email": "user-2@test.com"},
		},
		{
			ID:         "user-3",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeUserPassword}},
			Disabled:   true,
			Data:       map[string]interface{}{"email": "user-3@test.com"},
		},
	}

	t.Run("should resolve users", func(t *testing.T) {
		for _, tc := range []struct {
			description   string
			inputs        multiUserInputs
			users         []realm.User
			expectedUsers []realm.User
		}{
			{
				description:   "with no input set",
				users:         testUsers,
				expectedUsers: testUsers,
			},
			{
				description:   "with no users",
				inputs:        multiUserInputs{ProviderTypes: []string{realm.AuthProviderTypeUserPassword.String()}, State: realm.UserStateDisabled, Pending: false},
				expectedUsers: nil,
			},
			{
				description:   "with no users and no input",
				expectedUsers: nil,
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				realmClient := mock.RealmClient{}
				realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
					return tc.users, nil
				}

				users, err := tc.inputs.resolveUsers(realmClient, "groupID", "appID")
				assert.Nil(t, err)
				assert.Equal(t, tc.expectedUsers, users)
			})
		}
	})

	for _, tc := range []struct {
		description   string
		users         []realm.User
		expectedUsers []realm.User
		expectedErr   error
	}{
		{
			description: "should error when a user cannot be found from provided ids",
			expectedErr: errors.New("no users found"),
		},
		{
			description:   "should find users from provided ids",
			users:         testUsers[:2],
			expectedUsers: testUsers[:2],
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			realmClient := mock.RealmClient{}
			realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
				return tc.users, nil
			}

			inputs := multiUserInputs{Users: []string{testUsers[0].ID, testUsers[1].ID}}
			users, err := inputs.resolveUsers(realmClient, "groupID", "appID")

			assert.Equal(t, tc.expectedUsers, users)
			assert.Equal(t, tc.expectedErr, err)
		})
	}

	t.Run("should error from client", func(t *testing.T) {
		realmClient := mock.RealmClient{}
		realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
			return nil, errors.New("client error")
		}
		inputs := multiUserInputs{}
		_, err := inputs.resolveUsers(realmClient, "groupID", "appID")

		assert.Equal(t, errors.New("client error"), err)
	})
}

func TestSelectMultiUsersInputs(t *testing.T) {
	testUsers := []realm.User{
		{
			ID:         "user-1",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeAnonymous}},
		},
		{
			ID:         "user-2",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeUserPassword}},
			Disabled:   true,
			Data:       map[string]interface{}{"email": "user-2@test.com"},
		},
		{
			ID:         "user-3",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeUserPassword}},
			Disabled:   true,
			Data:       map[string]interface{}{"email": "user-3@test.com"},
		},
	}

	t.Run("should prompt for users", func(t *testing.T) {
		for _, tc := range []struct {
			description   string
			inputs        multiUserInputs
			procedure     func(c *expect.Console)
			users         []realm.User
			expectedUsers []realm.User
		}{
			{
				description: "with no input set",
				procedure: func(c *expect.Console) {
					c.ExpectString("Which user(s) would you like to delete?")
					c.Send("user-1")
					c.SendLine(" ")
					c.ExpectEOF()
				},
				users:         testUsers,
				expectedUsers: []realm.User{testUsers[0]},
			},
			{
				description: "unless no users are found",
				procedure: func(c *expect.Console) {
					console, _ := c.ExpectEOF()
					assert.Equal(t, "", console)
				},
				expectedUsers: nil,
			},
			{
				description: "unless users inputs are resolved",
				procedure: func(c *expect.Console) {
					console, _ := c.ExpectEOF()
					assert.Equal(t, "", console)
				},
				inputs:        multiUserInputs{Users: []string{"user-1"}},
				users:         testUsers[:1],
				expectedUsers: testUsers[:1],
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				_, console, _, ui, consoleErr := mock.NewVT10XConsole()
				assert.Nil(t, consoleErr)
				defer console.Close()

				doneCh := make(chan (struct{}))
				go func() {
					defer close(doneCh)
					tc.procedure(console)
				}()
				users, selectErr := tc.inputs.selectUsers(ui, tc.users, "delete")

				console.Tty().Close() // flush the writers
				<-doneCh              // wait for procedure to complete

				assert.Nil(t, selectErr)
				assert.Equal(t, tc.expectedUsers, users)
			})
		}
	})

}
