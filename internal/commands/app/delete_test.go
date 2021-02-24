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
	appID := "60344735b37e3733de2adf40"
	app1 := realm.App{
		ID:          appID,
		GroupID:     groupID1,
		ClientAppID: "app1-abcde",
		Name:        "app1",
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
			inputs:         deleteInputs{},
			expectedOutput: "01:23:45 UTC INFO  No apps to delete\n",
		},
		{
			description:  "with no project flag set and an apps flag set should delete all apps that match the apps flag",
			inputs:       deleteInputs{Apps: []string{"app1"}},
			apps:         []realm.App{app1},
			expectedApps: []string{appID},
			expectedOutput: strings.Join(
				[]string{
					"01:23:45 UTC INFO  Successfully deleted 1/1 app(s)",
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
			inputs:       deleteInputs{Apps: []string{"app1"}, Project: groupID1},
			apps:         []realm.App{app1},
			expectedApps: []string{appID},
			expectedOutput: strings.Join(
				[]string{
					"01:23:45 UTC INFO  Successfully deleted 1/1 app(s)",
					"  ID                        Name  Deleted  Details",
					"  ------------------------  ----  -------  -------",
					"  60344735b37e3733de2adf40  app1  true            ",
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
					"01:23:45 UTC INFO  Successfully deleted 0/1 app(s)",
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
