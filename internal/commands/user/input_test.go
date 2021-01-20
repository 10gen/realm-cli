package user

import (
	"errors"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"github.com/Netflix/go-expect"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestResolveUsersInputs(t *testing.T) {
	testUsers := []realm.User{
		{
			ID: "user-1",
			Identities: []realm.UserIdentity{
				{ProviderType: providerTypeAnonymous},
			},
			Disabled: false,
		},
		{
			ID: "user-2",
			Identities: []realm.UserIdentity{
				{ProviderType: providerTypeLocalUserPass},
			},
			Disabled: true,
		},
		{
			ID: "user-3",
			Identities: []realm.UserIdentity{
				{ProviderType: providerTypeLocalUserPass},
			},
			Disabled: true,
		},
	}

	t.Run("setup should prompt for users", func(t *testing.T) {
		for _, tc := range []struct {
			description   string
			inputs        usersInputs
			procedure     func(c *expect.Console)
			users         []realm.User
			expectedUsers []realm.User
		}{
			{
				description: "with no input set",
				inputs:      usersInputs{},
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
				inputs:      usersInputs{ProviderTypes: []string{providerTypeLocalUserPass}},
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

				app := realm.App{
					ID:      primitive.NewObjectID().Hex(),
					GroupID: primitive.NewObjectID().Hex(),
				}

				users, err := tc.inputs.ResolveUsers(ui, realmClient, app)

				console.Tty().Close() // flush the writers
				<-doneCh              // wait for procedure to complete

				assert.Nil(t, err)
				assert.Equal(t, tc.expectedUsers, users)
			})
		}
	})
	t.Run("setup should error", func(t *testing.T) {
		for _, tc := range []struct {
			description string
			inputs      usersInputs
			procedure   func(c *expect.Console)
			expectedErr error
			mockClient  func() realm.Client
		}{
			{
				description: "from client",
				inputs:      usersInputs{},
				procedure:   func(c *expect.Console) {},
				expectedErr: errors.New("client error"),
				mockClient: func() realm.Client {
					realmClient := mock.RealmClient{}
					realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
						return nil, errors.New("client error")
					}
					return realmClient
				},
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

				app := realm.App{
					ID:      primitive.NewObjectID().Hex(),
					GroupID: primitive.NewObjectID().Hex(),
				}

				_, err := tc.inputs.ResolveUsers(ui, tc.mockClient(), app)

				console.Tty().Close() // flush the writers
				<-doneCh              // wait for procedure to complete

				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})
}
