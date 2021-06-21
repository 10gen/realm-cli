package cli_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Netflix/go-expect"
)

func TestProjectAppInputsResolve(t *testing.T) {
	wd, wdErr := os.Getwd()
	assert.Nil(t, wdErr)

	testRoot := wd
	projectRoot := filepath.Join(testRoot, "testdata", "project")
	localProjectRoot := filepath.Join(testRoot, "testdata", "local_project")

	for _, tc := range []struct {
		description string
		inputs      cli.ProjectInputs
		wd          string
		procedure   func(c *expect.Console)
		test        func(t *testing.T, i cli.ProjectInputs)
	}{
		{
			description: "should not prompt for app when set by flag already",
			inputs:      cli.ProjectInputs{App: "some-app"},
			wd:          testRoot,
			procedure:   func(c *expect.Console) {},
			test: func(t *testing.T, i cli.ProjectInputs) {
				assert.Equal(t, "some-app", i.App)
			},
		},
		{
			description: "when outside a project directory should prompt for app when not flagged",
			wd:          testRoot,
			procedure: func(c *expect.Console) {
				c.ExpectString("App ID or Name")
				c.SendLine("some-app")
			},
			test: func(t *testing.T, i cli.ProjectInputs) {
				assert.Equal(t, "some-app", i.App)
			},
		},
		{
			description: "when inside a project directory should prompt for app when not flagged and provide client app id as a default",
			wd:          projectRoot,
			procedure: func(c *expect.Console) {
				c.ExpectString("App ID or Name")
				c.SendLine("") // accept default
			},
			test: func(t *testing.T, i cli.ProjectInputs) {
				assert.Equal(t, "eggcorn-abcde", i.App)
			},
		},
		{
			description: "when inside a project directory should prompt for app when not flagged and provide name as a default when client app id is not available",
			wd:          localProjectRoot,
			procedure: func(c *expect.Console) {
				c.ExpectString("App ID or Name")
				c.SendLine("") // accept default
			},
			test: func(t *testing.T, i cli.ProjectInputs) {
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

			err := tc.inputs.Resolve(ui, tc.wd, false)
			assert.Nil(t, err)

			console.Tty().Close() // flush the writers
			<-doneCh              // wait for procedure to complete

			tc.test(t, tc.inputs)
		})
	}
}

func TestResolveApp(t *testing.T) {
	groupID := "groupID"
	app := realm.App{
		ID:          primitive.NewObjectID().Hex(),
		GroupID:     primitive.NewObjectID().Hex(),
		ClientAppID: "eggcorn-abcde",
		Name:        "eggcorn",
	}

	for _, tc := range []struct {
		description string
		appID       string
		apps        []realm.App
		procedure   func(c *expect.Console)
		expectedApp realm.App
		expectedErr error
	}{
		{
			description: "should return the single app found from the client call",
			appID:       "app",
			apps:        []realm.App{app},
			procedure:   func(c *expect.Console) {},
			expectedApp: app,
		},
		{
			description: "should return an error when no apps are returned from the client call with no app id specified",
			procedure:   func(c *expect.Console) {},
			expectedErr: cli.ErrAppNotFound{},
		},
		{
			description: "should return an error when no apps are returned from the client call with an app id specified",
			appID:       "app",
			procedure:   func(c *expect.Console) {},
			expectedErr: cli.ErrAppNotFound{"app"},
		},
		{
			description: "should prompt user to select an app when more than one is returned from the client call",
			appID:       "app",
			apps:        []realm.App{app, app},
			procedure: func(c *expect.Console) {
				c.ExpectString("Select App")
				c.SendLine("egg")
			},
			expectedApp: app,
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

			inputs := cli.ProjectInputs{Project: groupID, App: tc.appID}

			app, err := cli.ResolveApp(ui, realmClient, inputs.Filter())

			console.Tty().Close() // flush the writers
			<-doneCh              // wait for procedure to complete

			assert.Equal(t, tc.expectedApp, app)
			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, realm.AppFilter{GroupID: "groupID", App: tc.appID}, appFilter)
		})
	}

	t.Run("should return the client error if one occurs", func(t *testing.T) {
		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return nil, errors.New("something bad happened")
		}

		_, err := cli.ResolveApp(nil, realmClient, realm.AppFilter{})
		assert.Equal(t, errors.New("something bad happened"), err)
	})
}

func TestResolveGroupID(t *testing.T) {
	testGroup := atlas.Group{
		ID:   "some-id",
		Name: "eggcorn",
	}

	for _, tc := range []struct {
		description     string
		groups          []atlas.Group
		procedure       func(c *expect.Console)
		expectedGroupID string
		expectedErr     error
	}{
		{
			description:     "should return the single group found from the client call",
			groups:          []atlas.Group{testGroup},
			procedure:       func(c *expect.Console) {},
			expectedGroupID: testGroup.ID,
		},
		{
			description: "should return an error when no groups are returned from the client call",
			procedure:   func(c *expect.Console) {},
			expectedErr: cli.ErrGroupNotFound,
		},
		{
			description: "should prompt user to select a group when more than one is returned from the client call",
			groups:      []atlas.Group{testGroup, testGroup},
			procedure: func(c *expect.Console) {
				c.ExpectString("Atlas Project")
				c.SendLine("egg")
			},
			expectedGroupID: testGroup.ID,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			atlasClient := mock.AtlasClient{}
			atlasClient.GroupsFn = func() ([]atlas.Group, error) {
				return tc.groups, nil
			}

			_, console, _, ui, consoleErr := mock.NewVT10XConsole()
			assert.Nil(t, consoleErr)
			defer console.Close()

			doneCh := make(chan (struct{}))
			go func() {
				defer close(doneCh)
				tc.procedure(console)
			}()
			groupID, err := cli.ResolveGroupID(ui, atlasClient)

			console.Tty().Close() // flush the writers
			<-doneCh              // wait for procedure to complete

			assert.Equal(t, tc.expectedGroupID, groupID)
			assert.Equal(t, tc.expectedErr, err)
		})
	}

	t.Run("should return the client error if one occurs", func(t *testing.T) {
		atlasClient := mock.AtlasClient{}
		atlasClient.GroupsFn = func() ([]atlas.Group, error) {
			return nil, errors.New("something bad happened")
		}

		_, err := cli.ResolveGroupID(nil, atlasClient)
		assert.Equal(t, errors.New("something bad happened"), err)
	})
}
