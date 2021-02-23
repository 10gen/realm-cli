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

func TestAppDeleteSetup(t *testing.T) {
	t.Run("setup creates a realm client with a session", func(t *testing.T) {
		profile := mock.NewProfile(t)
		profile.SetRealmBaseURL("http://localhost:8080")

		cmd := &CommandDelete{}
		assert.Nil(t, cmd.realmClient)

		assert.Nil(t, cmd.Setup(profile, nil))
		assert.NotNil(t, cmd.realmClient)
	})
}

func TestAppDeleteHandler(t *testing.T) {
	groupID1 := primitive.NewObjectID().Hex()
	appID := primitive.NewObjectID().Hex()
	app1 := realm.App{
		ID:          appID,
		GroupID:     groupID1,
		ClientAppID: "app1-abcde",
		Name:        "app1",
	}

	apps := []realm.App{app1}

	for _, tc := range []struct {
		description  string
		inputs       cli.ProjectInputs
		expectedApps []string
	}{
		{
			description:  "with no project flag set and an app flag set should return all apps that match the app flag",
			inputs:       cli.ProjectInputs{App: "app1"},
			expectedApps: []string{appID},
		},
		{
			description:  "with no project flag set and an app flag set should return all apps that match the app flag",
			inputs:       cli.ProjectInputs{Project: groupID1, App: "app1"},
			expectedApps: []string{appID},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			realmClient := mock.RealmClient{}

			var capturedApps = make([]string, 0)
			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				return apps, nil
			}

			realmClient.DeleteAppFn = func(groupID, appID string) error {
				capturedApps = append(capturedApps, appID)
				return nil
			}

			cmd := &CommandDelete{inputs: tc.inputs, realmClient: realmClient}
			assert.Nil(t, cmd.Handler(nil, nil))
			assert.Equal(t, tc.expectedApps, capturedApps)
		})
	}
}

func TestAppDeleteFeedback(t *testing.T) {
	groupID1, groupID2 := primitive.NewObjectID().Hex(), primitive.NewObjectID().Hex()
	for _, tc := range []struct {
		description    string
		apps           []appOutput
		expectedOutput string
	}{
		{
			description:    "should print an empty state message when no apps were found",
			expectedOutput: "01:23:45 UTC INFO  No apps to delete\n",
		},
		{
			description: "should print a list of apps that were found",
			apps: []appOutput{
				{
					app: realm.App{
						ID:          "60344735b37e3733de2adf40",
						GroupID:     groupID1,
						ClientAppID: "app1-abcde",
						Name:        "app1",
					},
				},
				{
					app: realm.App{
						ID:          "60344735b37e3733de2adf41",
						GroupID:     groupID1,
						ClientAppID: "app2-fghij",
						Name:        "app2",
					},
					err: errors.New("client error")},
				{
					app: realm.App{
						ID:          "60344735b37e3733de2adf42",
						GroupID:     groupID2,
						ClientAppID: "app3-wxyz",
						Name:        "app3",
					},
				},
			},
			expectedOutput: strings.Join(
				[]string{
					"01:23:45 UTC INFO  Deleted app(s)",
					"  ID                        Name  Deleted  Details     ",
					"  ------------------------  ----  -------  ------------",
					"  60344735b37e3733de2adf40  app1  true                 ",
					"  60344735b37e3733de2adf41  app2  false    client error",
					"  60344735b37e3733de2adf42  app3  true                 ",
					"",
				},
				"\n",
			),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			out, ui := mock.NewUI()

			cmd := &CommandDelete{outputs: tc.apps}

			assert.Nil(t, cmd.Feedback(nil, ui))

			assert.Equal(t, tc.expectedOutput, out.String())
		})
	}
}
