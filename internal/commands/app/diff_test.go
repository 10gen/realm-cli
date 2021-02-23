package app

import (
	"errors"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAppDiffHandler(t *testing.T) {
	groupID1 := primitive.NewObjectID().Hex()
	app1 := realm.App{
		ID:          "app1",
		GroupID:     groupID1,
		ClientAppID: "app1-abcde",
		Name:        "app1",
	}

	apps := []realm.App{app1}

	for _, tc := range []struct {
		description        string
		inputs             diffInputs
		expectedAppFilter  realm.AppFilter
		expectedDiff       []string
		expectedDiffOutput string
		expectedErr        error
	}{
		{
			description:        "with no project nor app flag set should diff based on input",
			expectedAppFilter:  realm.AppFilter{},
			expectedDiff:       []string{"diff1"},
			expectedDiffOutput: "01:23:45 UTC INFO  The following reflects the proposed changes to your Realm app\ndiff1\n",
		},
		{
			description:        "with no project flag set and an app flag set should show the diff for the app",
			inputs:             diffInputs{AppDirectory: "testdata/project", ProjectInputs: cli.ProjectInputs{App: "app1"}},
			expectedAppFilter:  realm.AppFilter{App: "app1"},
			expectedDiff:       []string{"diff1"},
			expectedDiffOutput: "01:23:45 UTC INFO  The following reflects the proposed changes to your Realm app\ndiff1\n",
		},
		{
			description:        "with no diffs between local and remote app",
			inputs:             diffInputs{AppDirectory: "testdata/project", ProjectInputs: cli.ProjectInputs{App: "app1"}},
			expectedAppFilter:  realm.AppFilter{App: "app1"},
			expectedDiff:       nil,
			expectedDiffOutput: "01:23:45 UTC INFO  Deployed app is identical to proposed version\n",
		},
		{
			description:        "with a project flag set and no app flag set should diff based on input",
			inputs:             diffInputs{AppDirectory: "testdata/project", ProjectInputs: cli.ProjectInputs{Project: groupID1}},
			expectedAppFilter:  realm.AppFilter{GroupID: groupID1},
			expectedDiff:       []string{"diff1"},
			expectedDiffOutput: "01:23:45 UTC INFO  The following reflects the proposed changes to your Realm app\ndiff1\n",
		},
		{
			description:       "error on the diff",
			inputs:            diffInputs{AppDirectory: "testdata/project", ProjectInputs: cli.ProjectInputs{Project: groupID1, App: "app1"}},
			expectedAppFilter: realm.AppFilter{GroupID: groupID1, App: "app1"},
			expectedErr:       errors.New("something went wrong"),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			out, ui := mock.NewUI()

			realmClient := mock.RealmClient{}

			var appFilter realm.AppFilter
			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				appFilter = filter
				return apps, nil
			}
			realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
				return tc.expectedDiff, tc.expectedErr
			}

			cmd := &CommandDiff{inputs: tc.inputs}
			assert.Equal(t, tc.expectedErr, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
			assert.Equal(t, tc.expectedAppFilter, appFilter)
			assert.Equal(t, tc.expectedDiffOutput, out.String())
		})
	}
}
