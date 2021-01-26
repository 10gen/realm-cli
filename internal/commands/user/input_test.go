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

func TestResolveUsersInputs(t *testing.T) {
	testUsers := []realm.User{
		{
			ID:         "user-1",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeAnonymous}},
			Disabled:   false,
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
			inputs        usersInputs
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
				description: "with providers set",
				inputs:      usersInputs{ProviderTypes: []string{realm.AuthProviderTypeUserPassword.String()}},
				procedure: func(c *expect.Console) {
					c.ExpectString("Which user(s) would you like to delete?")
					c.Send("user-2")
					c.SendLine(" ")
					c.ExpectEOF()
				},
				users:         testUsers[1:],
				expectedUsers: []realm.User{testUsers[1]},
			},
			{
				description: "with state set",
				inputs:      usersInputs{State: realm.UserStateDisabled},
				procedure: func(c *expect.Console) {
					c.ExpectString("Which user(s) would you like to delete?")
					c.Send("user-2")
					c.SendLine(" ")
					c.ExpectEOF()
				},
				users:         testUsers[1:],
				expectedUsers: []realm.User{testUsers[1]},
			},
			{
				description: "with status set",
				inputs:      usersInputs{State: realm.UserStateDisabled},
				procedure: func(c *expect.Console) {
					c.ExpectString("Which user(s) would you like to delete?")
					c.Send("user-3")
					c.SendLine(" ")
					c.ExpectEOF()
				},
				users:         testUsers[2:],
				expectedUsers: []realm.User{testUsers[2]},
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				realmClient := mock.RealmClient{}
				realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
					return tc.users, nil
				}

				_, console, _, ui, consoleErr := mock.NewVT10XConsole()
				assert.Nil(t, consoleErr)
				defer console.Close()

				doneCh := make(chan (struct{}))
				go func() {
					defer close(doneCh)
					tc.procedure(console)
				}()

				var app realm.App

				users, err := tc.inputs.ResolveUsers(ui, realmClient, app)

				console.Tty().Close() // flush the writers
				<-doneCh              // wait for procedure to complete

				assert.Nil(t, err)
				assert.Equal(t, tc.expectedUsers, users)
			})
		}
	})

	t.Run("should error from client", func(t *testing.T) {
		_, console, _, ui, consoleErr := mock.NewVT10XConsole()
		assert.Nil(t, consoleErr)
		defer console.Close()

		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)
		}()

		realmClient := mock.RealmClient{}
		realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
			return nil, errors.New("client error")
		}
		var app realm.App

		inputs := usersInputs{}
		_, err := inputs.ResolveUsers(ui, realmClient, app)

		console.Tty().Close() // flush the writers
		<-doneCh              // wait for procedure to complete

		assert.Equal(t, errors.New("client error"), err)
	})
}

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
