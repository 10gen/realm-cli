package user

import (
	"errors"
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
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

func TestMultiUsersInputsFind(t *testing.T) {
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

	t.Run("when finding users", func(t *testing.T) {
		for _, tc := range []struct {
			description   string
			inputs        multiUserInputs
			users         []realm.User
			expectedUsers []realm.User
			expectedErr   error
		}{
			{
				description:   "should return found all users with no input users ",
				users:         testUsers,
				expectedUsers: testUsers,
			},
			{
				description: "should error with input users and no found users",
				inputs:      multiUserInputs{Users: []string{"user-1"}},
				expectedErr: errors.New("no users found"),
			},

			{
				description: "should return an empty users slice with no found users or input users",
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				realmClient := mock.RealmClient{}
				realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
					return tc.users, nil
				}

				users, err := tc.inputs.findUsers(realmClient, "groupID", "appID")
				assert.Equal(t, tc.expectedErr, err)
				assert.Equal(t, tc.expectedUsers, users)
			})
		}
	})

	t.Run("should error from client", func(t *testing.T) {
		realmClient := mock.RealmClient{}
		realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
			return nil, errors.New("client error")
		}
		inputs := multiUserInputs{}
		_, err := inputs.findUsers(realmClient, "groupID", "appID")

		assert.Equal(t, errors.New("client error"), err)
	})
}

func TestMultiUsersInputsSelect(t *testing.T) {
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

	t.Run("should prompt with no input set", func(t *testing.T) {
		_, console, _, ui, consoleErr := mock.NewVT10XConsole()
		assert.Nil(t, consoleErr)
		defer console.Close()

		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)

			console.ExpectString("Which user(s) would you like to delete?")
			console.Send("user-1")
			console.SendLine(" ")
			console.ExpectEOF()
		}()

		var i multiUserInputs
		users, err := i.selectUsers(ui, testUsers, "delete")

		console.Tty().Close() // flush the writers
		<-doneCh              // wait for procedure to complete

		assert.Nil(t, err)
		assert.Equal(t, testUsers[:1], users)
	})

	t.Run("should not prompt the user", func(t *testing.T) {
		for _, tc := range []struct {
			description string
			inputs      multiUserInputs
			users       []realm.User
		}{
			{description: "with no inputs set and no users found"},
			{
				"with user inputs set and some users found",
				multiUserInputs{Users: []string{"user-1"}},
				testUsers,
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				users, err := tc.inputs.selectUsers(nil, tc.users, "")
				assert.Nil(t, err)
				assert.Equal(t, tc.users, users)
			})
		}
	})
}
