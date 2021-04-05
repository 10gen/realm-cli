package user

import (
	"errors"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestUserDisableHandler(t *testing.T) {
	projectID := "projectID"
	appID := "appID"
	app := realm.App{
		ID:          appID,
		GroupID:     projectID,
		ClientAppID: "eggcorn-abcde",
		Name:        "eggcorn",
	}

	t.Run("should display empty state message no users are found to disable", func(t *testing.T) {
		out, ui := mock.NewUI()

		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{app}, nil
		}
		realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
			return nil, nil
		}
		realmClient.DisableUserFn = func(groupID, appID, userID string) error {
			return nil
		}

		cmd := &CommandDisable{disableInputs{ProjectInputs: cli.ProjectInputs{
			Project: projectID,
			App:     appID,
		}}}

		assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
		assert.Equal(t, "No users to disable\n", out.String())
	})

	t.Run("should display users disabled by auth provider type", func(t *testing.T) {
		out, ui := mock.NewUI()

		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{app}, nil
		}
		realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
			return testUsers, nil
		}
		realmClient.DisableUserFn = func(groupID, appID, userID string) error {
			return nil
		}

		cmd := &CommandDisable{disableInputs{
			ProjectInputs: cli.ProjectInputs{
				Project: projectID,
				App:     appID,
			},
			multiUserInputs: multiUserInputs{
				Users: []string{testUsers[0].ID},
			},
		}}

		assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
		assert.Equal(t, strings.Join([]string{
			"Provider type: User/Password",
			"  Email            ID      Type  Enabled  Details",
			"  ---------------  ------  ----  -------  -------",
			"  user-2@test.com  user-2        false           ",
			"Provider type: ApiKey",
			"  Name    ID      Type  Enabled  Details",
			"  ------  ------  ----  -------  -------",
			"  name-3  user-3        false           ",
			"Provider type: Anonymous",
			"  ID      Type    Enabled  Details",
			"  ------  ------  -------  -------",
			"  user-1  type-1  false           ",
			"Provider type: Custom JWT",
			"  ID      Type  Enabled  Details",
			"  ------  ----  -------  -------",
			"  user-4        false           ",
			"",
		}, "\n"), out.String())
	})

	for _, tc := range []struct {
		description    string
		expectedOutput string
		disableErr     error
	}{
		{
			description: "should disable a user when a user id is provided",
			expectedOutput: strings.Join([]string{
				"Provider type: Anonymous",
				"  ID      Type    Enabled  Details",
				"  ------  ------  -------  -------",
				"  user-1  type-1  false           ",
				"",
			}, "\n"),
		},
		{
			description: "should save failed disable errors",
			disableErr:  errors.New("client error"),
			expectedOutput: strings.Join([]string{
				"Provider type: Anonymous",
				"  ID      Type    Enabled  Details     ",
				"  ------  ------  -------  ------------",
				"  user-1  type-1  true     client error",
				"",
			}, "\n"),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			out, ui := mock.NewUI()

			realmClient := mock.RealmClient{}

			var capturedAppFilter realm.AppFilter
			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				capturedAppFilter = filter
				return []realm.App{app}, nil
			}

			var capturedFindProjectID, capturedFindAppID string
			realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
				capturedFindProjectID = groupID
				capturedFindAppID = appID
				return testUsers[:1], nil
			}

			var capturedDisableProjectID, capturedDisableAppID string
			realmClient.DisableUserFn = func(groupID, appID, userID string) error {
				capturedDisableProjectID = groupID
				capturedDisableAppID = appID
				return tc.disableErr
			}

			cmd := &CommandDisable{disableInputs{
				ProjectInputs: cli.ProjectInputs{
					Project: projectID,
					App:     appID,
				},
				multiUserInputs: multiUserInputs{
					Users: []string{testUsers[0].ID},
				},
			}}

			assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))

			assert.Equal(t, testUsers[0].ID, cmd.inputs.Users[0])
			assert.Equal(t, tc.expectedOutput, out.String())

			assert.Equal(t, realm.AppFilter{App: appID, GroupID: projectID}, capturedAppFilter)
			assert.Equal(t, projectID, capturedFindProjectID)
			assert.Equal(t, appID, capturedFindAppID)
			assert.Equal(t, projectID, capturedDisableProjectID)
			assert.Equal(t, appID, capturedDisableAppID)
		})
	}

	t.Run("should return an error", func(t *testing.T) {
		for _, tc := range []struct {
			description string
			setupClient func() realm.Client
			expectedErr error
		}{
			{
				description: "when resolving the app fails",
				setupClient: func() realm.Client {
					realmClient := mock.RealmClient{}
					realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
						return nil, errors.New("something bad happened")
					}
					return realmClient
				},
				expectedErr: errors.New("something bad happened"),
			},
			{
				description: "when finding the users fails",
				setupClient: func() realm.Client {
					realmClient := mock.RealmClient{}
					realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
						return []realm.App{app}, nil
					}
					realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
						return nil, errors.New("something bad happened")
					}
					return realmClient
				},
				expectedErr: errors.New("something bad happened"),
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				cmd := &CommandDisable{}

				err := cmd.Handler(nil, nil, cli.Clients{Realm: tc.setupClient()})
				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})
}

func TestTableRowDisable(t *testing.T) {
	for _, tc := range []struct {
		description string
		err         error
		expectedRow map[string]interface{}
	}{
		{
			description: "should show successful disable user row",
			expectedRow: map[string]interface{}{
				"Enabled": false,
				"Details": "",
			},
		},
		{
			description: "should show failed disable user row",
			err:         errors.New("client error"),
			expectedRow: map[string]interface{}{
				"Enabled": true,
				"Details": "client error",
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			row := map[string]interface{}{}
			output := userOutput{realm.User{}, tc.err}
			tableRowDisable(output, row)

			assert.Equal(t, tc.expectedRow, row)
		})
	}
}
