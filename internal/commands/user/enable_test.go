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

func TestUserEnableHandler(t *testing.T) {
	projectID := "projectID"
	appID := "appID"
	app := realm.App{
		ID:          appID,
		GroupID:     projectID,
		ClientAppID: "eggcorn-abcde",
		Name:        "eggcorn",
	}

	t.Run("should display empty state message no users are found to enable", func(t *testing.T) {
		out, ui := mock.NewUI()

		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{app}, nil
		}
		realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
			return nil, nil
		}
		realmClient.DeleteUserFn = func(groupID, appID, userID string) error {
			return nil
		}

		cmd := &CommandEnable{enableInputs{ProjectInputs: cli.ProjectInputs{
			Project: projectID,
			App:     appID,
		}}}

		assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
		assert.Equal(t, "No users to enable\n", out.String())
	})

	t.Run("should display users enabled by auth provider type", func(t *testing.T) {
		out, ui := mock.NewUI()

		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{app}, nil
		}
		realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
			return testUsers, nil
		}
		realmClient.EnableUserFn = func(groupID, appID, userID string) error {
			return nil
		}

		cmd := &CommandEnable{enableInputs{
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
			"  user-2@test.com  user-2        true            ",
			"Provider type: ApiKey",
			"  Name    ID      Type  Enabled  Details",
			"  ------  ------  ----  -------  -------",
			"  name-3  user-3        true            ",
			"Provider type: Anonymous",
			"  ID      Type    Enabled  Details",
			"  ------  ------  -------  -------",
			"  user-1  type-1  true            ",
			"Provider type: Custom JWT",
			"  ID      Type  Enabled  Details",
			"  ------  ----  -------  -------",
			"  user-4        true            ",
			"",
		}, "\n"), out.String())
	})

	for _, tc := range []struct {
		description    string
		enableErr      error
		expectedOutput string
	}{
		{
			description: "should enable a user when a user id is provided",
			expectedOutput: strings.Join([]string{
				"Provider type: Anonymous",
				"  ID      Type    Enabled  Details",
				"  ------  ------  -------  -------",
				"  user-1  type-1  true            ",
				"",
			}, "\n"),
		},
		{
			description: "should save failed enable errors",
			enableErr:   errors.New("client error"),
			expectedOutput: strings.Join([]string{
				"Provider type: Anonymous",
				"  ID      Type    Enabled  Details     ",
				"  ------  ------  -------  ------------",
				"  user-1  type-1  false    client error",
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

				return []realm.User{{
					ID:         "user-1",
					Type:       "type-1",
					Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeAnonymous}},
					Disabled:   true,
				}}, nil
			}

			var capturedEnableProjectID, capturedEnableAppID string
			realmClient.EnableUserFn = func(groupID, appID, userID string) error {
				capturedEnableProjectID = groupID
				capturedEnableAppID = appID
				return tc.enableErr
			}

			cmd := &CommandEnable{
				inputs: enableInputs{
					ProjectInputs: cli.ProjectInputs{
						Project: projectID,
						App:     appID,
					},
					multiUserInputs: multiUserInputs{
						Users: []string{testUsers[0].ID},
					},
				},
			}

			assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
			assert.Equal(t, tc.expectedOutput, out.String())

			assert.Equal(t, realm.AppFilter{App: appID, GroupID: projectID}, capturedAppFilter)
			assert.Equal(t, projectID, capturedFindProjectID)
			assert.Equal(t, appID, capturedFindAppID)
			assert.Equal(t, projectID, capturedEnableProjectID)
			assert.Equal(t, appID, capturedEnableAppID)
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
				cmd := &CommandEnable{}

				err := cmd.Handler(nil, nil, cli.Clients{Realm: tc.setupClient()})
				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})
}

func TestUserEnableRow(t *testing.T) {
	for _, tc := range []struct {
		description string
		err         error
		expectedRow map[string]interface{}
	}{
		{
			description: "should show successful enable user row",
			expectedRow: map[string]interface{}{
				"Enabled": true,
				"Details": "",
			},
		},
		{
			description: "should show failed enable user row",
			err:         errors.New("client error"),
			expectedRow: map[string]interface{}{
				"Enabled": false,
				"Details": "client error",
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			row := map[string]interface{}{}
			output := userOutput{realm.User{Disabled: true}, tc.err}
			tableRowEnable(output, row)

			assert.Equal(t, tc.expectedRow, row)
		})
	}
}
