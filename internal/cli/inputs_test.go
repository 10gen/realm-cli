package cli_test

import (
	"errors"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Netflix/go-expect"
)

func TestProjectAppInputsResolve(t *testing.T) {
	for _, tc := range []struct {
		description string
		inputs      cli.ProjectAppInputs
		appData     cli.AppData
		procedure   func(c *expect.Console)
		test        func(t *testing.T, i cli.ProjectAppInputs)
	}{
		{
			description: "Should not prompt for app when set by flag already",
			inputs:      cli.ProjectAppInputs{App: "some-app"},
			procedure:   func(c *expect.Console) {},
			test: func(t *testing.T, i cli.ProjectAppInputs) {
				assert.Equal(t, "some-app", i.App)
			},
		},
		{
			description: "When outside a project directory should prompt for app when not flagged",
			procedure: func(c *expect.Console) {
				c.ExpectString("App Filter")
				c.SendLine("some-app")
			},
			test: func(t *testing.T, i cli.ProjectAppInputs) {
				assert.Equal(t, "some-app", i.App)
			},
		},
		{
			description: "When inside a project directory should prompt for app when not flagged and provide client app id as a default",
			appData:     cli.AppData{ID: "eggcorn-abcde"},
			procedure: func(c *expect.Console) {
				c.ExpectString("App Filter")
				c.SendLine("") // accept default
			},
			test: func(t *testing.T, i cli.ProjectAppInputs) {
				assert.Equal(t, "eggcorn-abcde", i.App)
			},
		},
		{
			description: "When inside a project directory should prompt for app when not flagged and provide name as a default when client app id is not available",
			appData:     cli.AppData{Name: "eggcorn"},
			procedure: func(c *expect.Console) {
				c.ExpectString("App Filter")
				c.SendLine("") // accept default
			},
			test: func(t *testing.T, i cli.ProjectAppInputs) {
				assert.Equal(t, "eggcorn", i.App)
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

			err := tc.inputs.Resolve(ui, tc.appData)
			assert.Nil(t, err)

			console.Tty().Close() // flush the writers
			<-doneCh              // wait for procedure to complete

			tc.test(t, tc.inputs)
		})
	}
}

func TestProjectAppInputsResolveApp(t *testing.T) {
	testApp := realm.App{
		ID:          primitive.NewObjectID(),
		GroupID:     primitive.NewObjectID().Hex(),
		ClientAppID: "eggcorn-abcde",
		Name:        "eggcorn",
	}

	for _, tc := range []struct {
		description string
		apps        []realm.App
		procedure   func(c *expect.Console)
		expectedApp realm.App
		expectedErr error
	}{
		{
			description: "Should return the single app found from the client call",
			apps:        []realm.App{testApp},
			procedure:   func(c *expect.Console) {},
			expectedApp: testApp,
		},
		{
			description: "Should return an error when no apps are returned from the client call",
			procedure:   func(c *expect.Console) {},
			expectedErr: errors.New("failed to find app 'app'"),
		},
		{
			description: "Should prompt user to select an app when more than one is returned from the client call",
			apps:        []realm.App{testApp, testApp},
			procedure: func(c *expect.Console) {
				c.ExpectString("Select App")
				c.SendLine("egg")
			},
			expectedApp: testApp,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			var appFilter realm.AppFilter

			realmClient := mock.RealmClient{}
			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				appFilter = filter
				return tc.apps, nil
			}

			_, console, _, ui, consoleErr := mock.NewVT10XConsole()
			assert.Nil(t, consoleErr)
			defer console.Close()

			doneCh := make(chan (struct{}))
			go func() {
				defer close(doneCh)
				tc.procedure(console)
			}()

			inputs := cli.ProjectAppInputs{Project: "groupID", App: "app"}

			app, err := inputs.ResolveApp(ui, realmClient)

			console.Tty().Close() // flush the writers
			<-doneCh              // wait for procedure to complete

			assert.Equal(t, tc.expectedApp, app)
			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, realm.AppFilter{GroupID: "groupID", App: "app"}, appFilter)
		})
	}

	t.Run("Should return the client error if one occurs", func(t *testing.T) {
		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return nil, errors.New("something bad happened")
		}

		var inputs cli.ProjectAppInputs

		_, err := inputs.ResolveApp(nil, realmClient)
		assert.Equal(t, errors.New("something bad happened"), err)
	})
}
