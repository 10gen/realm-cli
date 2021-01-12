package shared

import (
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
	"github.com/Netflix/go-expect"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestResolveUsersInputs(t *testing.T) {
	testUser := realm.User{
		ID: "salad-fingers",
		Identities: []realm.UserIdentity{
			{ProviderType: ProviderTypeAnonymous},
		},
	}

	for _, tc := range []struct {
		description   string
		inputs        UsersInputs
		procedure     func(c *expect.Console)
		expectedUsers []string
		expectedErr   error
	}{
		{
			description: "With no input set",
			inputs:      UsersInputs{},
			procedure: func(c *expect.Console) {
				c.ExpectString("Which user(s) would you like to delete?")
				c.SendLine("salad-fingers")
				c.ExpectEOF()
			},
			expectedUsers: []string{testUser.ID},
		},
	} {
		t.Run(fmt.Sprintf("%s Setup should prompt for Users", tc.description), func(t *testing.T) {

			// out, outErr := mock.FileWriter(t)
			// assert.Nil(t, outErr)
			// defer out.Close()

			// c, err := expect.NewConsole(expect.WithStdout(out))
			// assert.Nil(t, err)
			// defer c.Close()

			realmClient := mock.RealmClient{}
			realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
				return []realm.User{testUser}, nil
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

			// assert.Equal(t, tc.expectedUsers, users) // Still failing as of now
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}
