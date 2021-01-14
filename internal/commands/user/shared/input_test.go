package shared

import (
	"errors"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
	"github.com/AlecAivazis/survey/v2"
	"github.com/Netflix/go-expect"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestResolveUsersInputs(t *testing.T) {
	testUsers := []realm.User{
		{
			ID: "user-1",
			Identities: []realm.UserIdentity{
				{ProviderType: ProviderTypeAnonymous},
			},
			Disabled: false,
		},
		{
			ID: "user-2",
			Identities: []realm.UserIdentity{
				{ProviderType: ProviderTypeLocalUserPass},
			},
			Disabled: true,
		},
		{
			ID: "user-3",
			Identities: []realm.UserIdentity{
				{ProviderType: ProviderTypeLocalUserPass},
			},
			Disabled: true,
		},
	}

	t.Run("Setup should prompt for Users", func(t *testing.T) {
		for _, tc := range []struct {
			description   string
			inputs        UsersInputs
			procedure     func(c *expect.Console)
			users         []realm.User
			expectedUsers []string
		}{
			{
				description: "With no input set",
				inputs:      UsersInputs{},
				procedure: func(c *expect.Console) {
					c.ExpectString("Which user(s) would you like to delete?")
					c.Send("user-1")
					c.SendLine(" ")
					c.ExpectEOF()
				},
				users:         testUsers,
				expectedUsers: []string{testUsers[0].ID},
			},
			{
				description: "With providers set",
				inputs:      UsersInputs{ProviderTypes: []string{ProviderTypeLocalUserPass}},
				procedure: func(c *expect.Console) {
					c.ExpectString("Which user(s) would you like to delete?")
					c.Send("user-2")
					c.SendLine(" ")
					c.ExpectEOF()
				},
				users:         testUsers[1:],
				expectedUsers: []string{testUsers[1].ID},
			},
			{
				description: "With state set",
				inputs:      UsersInputs{State: UserStateTypeDisabled},
				procedure: func(c *expect.Console) {
					c.ExpectString("Which user(s) would you like to delete?")
					c.Send("user-2")
					c.SendLine(" ")
					c.ExpectEOF()
				},
				users:         testUsers[1:],
				expectedUsers: []string{testUsers[1].ID},
			},
			{
				description: "With status set",
				inputs:      UsersInputs{State: UserStateTypeDisabled},
				procedure: func(c *expect.Console) {
					c.ExpectString("Which user(s) would you like to delete?")
					c.Send("user-3")
					c.SendLine(" ")
					c.ExpectEOF()
				},
				users:         testUsers[2:],
				expectedUsers: []string{testUsers[2].ID},
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

				users, _ := tc.inputs.ResolveUsers(ui, realmClient, app)

				console.Tty().Close() // flush the writers
				<-doneCh              // wait for procedure to complete

				assert.Equal(t, tc.expectedUsers, users)
			})
		}
	})
	t.Run("Setup should Error", func(t *testing.T) {
		for _, tc := range []struct {
			description string
			inputs      UsersInputs
			procedure   func(c *expect.Console)
			expectedErr error
			mockClient  func() realm.Client
			mockUI      func(ui mock.UI) terminal.UI
		}{
			{
				description: "From client",
				inputs:      UsersInputs{},
				procedure:   func(c *expect.Console) {},
				expectedErr: errors.New("client error"),
				mockClient: func() realm.Client {
					realmClient := mock.RealmClient{}
					realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
						return nil, errors.New("client error")
					}
					return realmClient
				},
				mockUI: func(ui mock.UI) terminal.UI {
					return ui
				},
			},
			{
				description: "From UI",
				inputs:      UsersInputs{},
				procedure:   func(c *expect.Console) {},
				expectedErr: errors.New("ui error"),
				mockClient: func() realm.Client {
					realmClient := mock.RealmClient{}
					realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
						return nil, nil
					}
					return realmClient
				},
				mockUI: func(ui mock.UI) terminal.UI {
					ui.AskOneFn = func(answer interface{}, prompt survey.Prompt) error {
						return errors.New("ui error")
					}
					return ui
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

				mui := tc.mockUI(ui)

				app := realm.App{
					ID:      primitive.NewObjectID().Hex(),
					GroupID: primitive.NewObjectID().Hex(),
				}

				_, err := tc.inputs.ResolveUsers(mui, tc.mockClient(), app)

				console.Tty().Close() // flush the writers
				<-doneCh              // wait for procedure to complete

				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})
}
