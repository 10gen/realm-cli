package app

import (
	"errors"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAppDeleteHandler(t *testing.T) {
	groupID1 := primitive.NewObjectID().Hex()
	groupID2 := primitive.NewObjectID().Hex()
	appID1 := "60344735b37e3733de2adf40"
	appID2 := "60344735b37e3733de2adf41"
	appID3 := "60344735b37e3733de2adf42"
	app1 := realm.App{
		ID:          appID1,
		GroupID:     groupID1,
		ClientAppID: "app1-abcde",
		Name:        "app1",
	}
	app2 := realm.App{
		ID:          appID2,
		GroupID:     groupID2,
		ClientAppID: "app2-defgh",
		Name:        "app2",
	}
	app3 := realm.App{
		ID:          appID3,
		GroupID:     groupID1,
		ClientAppID: "app2-hijkl",
		Name:        "app2",
	}

	for _, tc := range []struct {
		description    string
		inputs         deleteInputs
		apps           []realm.App
		deleteErr      error
		expectedApps   []string
		expectedOutput string
	}{
		{
			description:    "should delete no apps if none are found",
			expectedOutput: "No apps to delete\n",
		},
		{
			description:  "with no project flag set and an apps flag set should delete all apps that match the apps flag",
			inputs:       deleteInputs{Apps: []string{"app1"}},
			apps:         []realm.App{app1, app2, app3},
			expectedApps: []string{appID1},
			expectedOutput: strings.Join(
				[]string{
					"Successfully deleted 1/1 app(s)",
					"  ID                        Name  Deleted  Details",
					"  ------------------------  ----  -------  -------",
					"  60344735b37e3733de2adf40  app1  true            ",
					"",
				},
				"\n",
			),
		},
		{
			description:  "with a project flag set and an apps flag set should delete all apps that match the apps flag",
			inputs:       deleteInputs{Apps: []string{"app1", "app2"}, Project: groupID1},
			apps:         []realm.App{app1, app2, app3},
			expectedApps: []string{appID1, appID3},
			expectedOutput: strings.Join(
				[]string{
					"Successfully deleted 2/2 app(s)",
					"  ID                        Name  Deleted  Details",
					"  ------------------------  ----  -------  -------",
					"  60344735b37e3733de2adf40  app1  true            ",
					"  60344735b37e3733de2adf42  app2  true            ",
					"",
				},
				"\n",
			),
		},
		{
			description:  "should indicate an error if deleting an app fails",
			inputs:       deleteInputs{Apps: []string{"app1"}, Project: groupID1},
			apps:         []realm.App{app1},
			expectedApps: []string{},
			deleteErr:    errors.New("client error"),
			expectedOutput: strings.Join(
				[]string{
					"Successfully deleted 0/1 app(s)",
					"  ID                        Name  Deleted  Details     ",
					"  ------------------------  ----  -------  ------------",
					"  60344735b37e3733de2adf40  app1  false    client error",
					"",
				},
				"\n",
			),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			out, ui := mock.NewUI()
			realmClient := mock.RealmClient{}

			var capturedFindGroupID string
			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				capturedFindGroupID = filter.GroupID
				return tc.apps, nil
			}

			var capturedApps = make([]string, 0)
			realmClient.DeleteAppFn = func(groupID, appID string) error {
				capturedApps = append(capturedApps, appID)
				return tc.deleteErr
			}

			cmd := &CommandDelete{inputs: tc.inputs}
			assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
			assert.Equal(t, tc.expectedOutput, out.String())

			assert.Equal(t, tc.inputs.Project, capturedFindGroupID)
		})
	}
}
