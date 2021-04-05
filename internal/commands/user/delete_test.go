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

var (
	testUsers = []realm.User{
		{
			ID:         "user-1",
			Type:       "type-1",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeAnonymous}},
		},
		{
			ID:         "user-2",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeUserPassword}},
			Data:       map[string]interface{}{"email": "user-2@test.com"},
		},
		{
			ID:         "user-3",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeAPIKey}},
			Data:       map[string]interface{}{"name": "name-3"},
		},
		{
			ID:         "user-4",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeCustomToken}},
		},
	}
)

func TestUserDeleteHandler(t *testing.T) {
	projectID := "projectID"
	appID := "appID"
	app := realm.App{
		ID:          appID,
		GroupID:     projectID,
		ClientAppID: "eggcorn-abcde",
		Name:        "eggcorn",
	}

	t.Run("should display empty state message no users are found to delete", func(t *testing.T) {
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

		cmd := &CommandDelete{deleteInputs{ProjectInputs: cli.ProjectInputs{
			Project: projectID,
			App:     appID,
		}}}

		assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
		assert.Equal(t, "No users to delete\n", out.String())
	})

	t.Run("should display users deleted by auth provider type", func(t *testing.T) {
		out, ui := mock.NewUI()

		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{app}, nil
		}
		realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
			return testUsers, nil
		}
		realmClient.DeleteUserFn = func(groupID, appID, userID string) error {
			return nil
		}

		cmd := &CommandDelete{deleteInputs{
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
			"  Email            ID      Type  Deleted  Details",
			"  ---------------  ------  ----  -------  -------",
			"  user-2@test.com  user-2        true            ",
			"Provider type: ApiKey",
			"  Name    ID      Type  Deleted  Details",
			"  ------  ------  ----  -------  -------",
			"  name-3  user-3        true            ",
			"Provider type: Anonymous",
			"  ID      Type    Deleted  Details",
			"  ------  ------  -------  -------",
			"  user-1  type-1  true            ",
			"Provider type: Custom JWT",
			"  ID      Type  Deleted  Details",
			"  ------  ----  -------  -------",
			"  user-4        true            ",
			"",
		}, "\n"), out.String())
	})

	for _, tc := range []struct {
		description    string
		deleteErr      error
		expectedOutput string
	}{
		{
			description: "should delete a user when a user id is provided",
			expectedOutput: strings.Join([]string{
				"Provider type: Anonymous",
				"  ID      Type    Deleted  Details",
				"  ------  ------  -------  -------",
				"  user-1  type-1  true            ",
				"",
			}, "\n"),
		},
		{
			description: "should save failed deletion errors",
			deleteErr:   errors.New("client error"),
			expectedOutput: strings.Join([]string{
				"Provider type: Anonymous",
				"  ID      Type    Deleted  Details     ",
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
				return testUsers[:1], nil
			}

			var capturedDeleteProjectID, capturedDeleteAppID string
			realmClient.DeleteUserFn = func(groupID, appID, userID string) error {
				capturedDeleteProjectID = groupID
				capturedDeleteAppID = appID
				return tc.deleteErr
			}

			cmd := &CommandDelete{deleteInputs{
				ProjectInputs: cli.ProjectInputs{
					Project: projectID,
					App:     appID,
				},
				multiUserInputs: multiUserInputs{
					Users: []string{testUsers[0].ID},
				},
			}}

			assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
			assert.Equal(t, tc.expectedOutput, out.String())

			assert.Equal(t, realm.AppFilter{App: appID, GroupID: projectID}, capturedAppFilter)
			assert.Equal(t, projectID, capturedFindProjectID)
			assert.Equal(t, appID, capturedFindAppID)
			assert.Equal(t, projectID, capturedDeleteProjectID)
			assert.Equal(t, appID, capturedDeleteAppID)
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
				cmd := &CommandDelete{}

				err := cmd.Handler(nil, nil, cli.Clients{Realm: tc.setupClient()})
				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})
}

func TestTableRowDelete(t *testing.T) {
	for _, tc := range []struct {
		description      string
		authProviderType realm.AuthProviderType
		err              error
		expectedRow      map[string]interface{}
	}{
		{
			description:      "should show successful delete user row",
			authProviderType: realm.AuthProviderTypeAPIKey,
			expectedRow: map[string]interface{}{
				"Deleted": true,
				"Details": "",
			},
		},
		{
			description:      "should show failed delete user row",
			authProviderType: realm.AuthProviderTypeUserPassword,
			err:              errors.New("client error"),

			expectedRow: map[string]interface{}{
				"Deleted": false,
				"Details": "client error",
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			row := map[string]interface{}{}
			output := userOutput{realm.User{}, tc.err}
			tableRowDelete(output, row)

			assert.Equal(t, tc.expectedRow, row)
		})
	}
}
