package app

import (
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAppListHandler(t *testing.T) {
	groupID1, groupID2 := primitive.NewObjectID().Hex(), primitive.NewObjectID().Hex()
	app1 := realm.App{
		ID:          "app1",
		GroupID:     groupID1,
		ClientAppID: "app1-abcde",
		Name:        "app1",
	}
	app2 := realm.App{
		ID:          "app2",
		GroupID:     groupID1,
		ClientAppID: "app2-abcde",
		Name:        "app2",
	}
	app3 := realm.App{
		ID:          "app3",
		GroupID:     groupID2,
		ClientAppID: "app1-fghij",
		Name:        "app1",
	}

	apps := []realm.App{app1, app2, app3}

	for _, tc := range []struct {
		description       string
		inputs            cli.ProjectInputs
		expectedAppFilter realm.AppFilter
	}{
		{
			description: "with no project nor app flag set should return all apps",
		},
		{
			description:       "with no project flag set and an app flag set should return all apps that match the app flag",
			inputs:            cli.ProjectInputs{App: "app1"},
			expectedAppFilter: realm.AppFilter{App: "app1"},
		},
		{
			description:       "with a project flag set and no app flag set should return all project apps",
			inputs:            cli.ProjectInputs{Project: groupID1},
			expectedAppFilter: realm.AppFilter{GroupID: groupID1},
		},
		{
			description:       "with no project flag set and an app flag set should return all apps that match the app flag",
			inputs:            cli.ProjectInputs{Project: groupID1, App: "app1"},
			expectedAppFilter: realm.AppFilter{GroupID: groupID1, App: "app1"},
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

			cmd := &CommandList{tc.inputs}
			assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))

			assert.Equal(t, tc.expectedAppFilter, appFilter)
			assert.Equal(t, fmt.Sprintf(`Found 3 apps
  app1-abcde (%s)
  app2-abcde (%s)
  app1-fghij (%s)
`, groupID1, groupID1, groupID2), out.String())
		})
	}
}
