package shared

import (
	"errors"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
	"github.com/Netflix/go-expect"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TODO: Test on ui error
func TestResolveUsersInputs(t *testing.T) {
	testUsers := []realm.User{
		{
			ID: "salad-fingers",
			Identities: []realm.UserIdentity{
				{ProviderType: ProviderTypeAnonymous},
			},
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
					c.Send("salad-fingers")
					c.SendLine(" ")
					c.ExpectEOF()
				},
				users:         testUsers,
				expectedUsers: []string{testUsers[0].ID},
			},
			// TODO: Test on valid inputs
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
			users       []realm.User
			expectedErr error
		}{
			{
				description: "With no input set, from client",
				inputs:      UsersInputs{},
				procedure:   func(c *expect.Console) {},
				expectedErr: errors.New("client error"),
			},
			// TODO: Test on invalid inputs
		} {
			t.Run(tc.description, func(t *testing.T) {
				realmClient := mock.RealmClient{}
				realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
					return nil, tc.expectedErr
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

				_, err := tc.inputs.ResolveUsers(ui, realmClient, app)

				console.Tty().Close() // flush the writers
				<-doneCh              // wait for procedure to complete

				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})
}
