package list

import (
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAppListSetup(t *testing.T) {
	t.Run("Setup creates a realm client with a session", func(t *testing.T) {
		profile := mock.NewProfile(t)
		profile.SetRealmBaseURL("http://localhost:8080")

		cmd := &command{}
		assert.Nil(t, cmd.realmClient)

		assert.Nil(t, cmd.Setup(profile, nil))
		assert.NotNil(t, cmd.realmClient)
	})
}

func TestAppListHandler(t *testing.T) {
	groupID1, groupID2 := primitive.NewObjectID().Hex(), primitive.NewObjectID().Hex()
	app1 := realm.App{
		ID:          primitive.NewObjectID().Hex(),
		GroupID:     groupID1,
		ClientAppID: "app1-abcde",
		Name:        "app1",
	}
	app2 := realm.App{
		ID:          primitive.NewObjectID().Hex(),
		GroupID:     groupID1,
		ClientAppID: "app2-fghij",
		Name:        "app2",
	}
	app3 := realm.App{
		ID:          primitive.NewObjectID().Hex(),
		GroupID:     groupID2,
		ClientAppID: "app1-abcde",
		Name:        "app1",
	}

	apps := []realm.App{app1, app2, app3}

	for _, tc := range []struct {
		description       string
		inputs            cli.ProjectAppInputs
		expectedAppFilter realm.AppFilter
	}{
		{
			description: "With no project nor app flag set should return all apps",
		},
		{
			description:       "With no project flag set and an app flag set should return all apps that match the app flag",
			inputs:            cli.ProjectAppInputs{App: "app1"},
			expectedAppFilter: realm.AppFilter{App: "app1"},
		},
		{
			description:       "With a project flag set and no app flag set should return all project apps",
			inputs:            cli.ProjectAppInputs{Project: groupID1},
			expectedAppFilter: realm.AppFilter{GroupID: groupID1},
		},
		{
			description:       "With no project flag set and an app flag set should return all apps that match the app flag",
			inputs:            cli.ProjectAppInputs{Project: groupID1, App: "app1"},
			expectedAppFilter: realm.AppFilter{GroupID: groupID1, App: "app1"},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			realmClient := mock.RealmClient{}

			var appFilter realm.AppFilter
			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				appFilter = filter
				return apps, nil
			}

			cmd := &command{inputs: tc.inputs, realmClient: realmClient}
			assert.Nil(t, cmd.Handler(nil, nil))

			assert.Equal(t, tc.expectedAppFilter, appFilter)
			assert.Equal(t, apps, cmd.apps)
		})
	}
}

func TestAppListFeedback(t *testing.T) {
	groupID1, groupID2 := primitive.NewObjectID().Hex(), primitive.NewObjectID().Hex()
	for _, tc := range []struct {
		description    string
		apps           []realm.App
		expectedOutput string
	}{
		{
			description:    "Should print an empty state message when no apps were found",
			expectedOutput: "01:23:45 UTC INFO  No available apps to show\n",
		},
		{
			description: "Should print a list of apps that were found",
			apps: []realm.App{
				{
					ID:          primitive.NewObjectID().Hex(),
					GroupID:     groupID1,
					ClientAppID: "app1-abcde",
					Name:        "app1",
				},
				{
					ID:          primitive.NewObjectID().Hex(),
					GroupID:     groupID1,
					ClientAppID: "app2-fghij",
					Name:        "app2",
				},
				{
					ID:          primitive.NewObjectID().Hex(),
					GroupID:     groupID2,
					ClientAppID: "app1-abcde",
					Name:        "app1",
				},
			},
			expectedOutput: fmt.Sprintf(`01:23:45 UTC INFO  Found 3 apps
  app1-abcde (%s)
  app2-fghij (%s)
  app1-abcde (%s)
`, groupID1, groupID1, groupID2),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			out, ui := mock.NewUI()

			cmd := &command{apps: tc.apps}

			assert.Nil(t, cmd.Feedback(nil, ui))

			assert.Equal(t, tc.expectedOutput, out.String())
		})
	}
}
