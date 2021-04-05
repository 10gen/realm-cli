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

func TestUserListHandler(t *testing.T) {
	projectID := "projectID"
	appID := "appID"
	app := realm.App{
		ID:          appID,
		GroupID:     projectID,
		ClientAppID: "eggcorn-abcde",
		Name:        "eggcorn",
	}

	t.Run("should display empty state message no users are found", func(t *testing.T) {
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

		cmd := &CommandList{listInputs{ProjectInputs: cli.ProjectInputs{
			Project: projectID,
			App:     appID,
		}}}

		assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
		assert.Equal(t, "No available users to show\n", out.String())
	})

	t.Run("should display users by auth provider type", func(t *testing.T) {
		out, ui := mock.NewUI()

		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{app}, nil
		}
		realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
			return testUsers, nil
		}

		cmd := &CommandList{listInputs{
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
			"  Email            ID      Type  Enabled  Last Authenticated",
			"  ---------------  ------  ----  -------  ------------------",
			"  user-2@test.com  user-2        true     n/a               ",
			"Provider type: ApiKey",
			"  Name    ID      Type  Enabled  Last Authenticated",
			"  ------  ------  ----  -------  ------------------",
			"  name-3  user-3        true     n/a               ",
			"Provider type: Anonymous",
			"  ID      Type    Enabled  Last Authenticated",
			"  ------  ------  -------  ------------------",
			"  user-1  type-1  true     n/a               ",
			"Provider type: Custom JWT",
			"  ID      Type  Enabled  Last Authenticated",
			"  ------  ----  -------  ------------------",
			"  user-4        true     n/a               ",
			"",
		}, "\n"), out.String())
	})

	t.Run("should find app users", func(t *testing.T) {
		out, ui := mock.NewUI()

		var capturedAppFilter realm.AppFilter
		var capturedProjectID, capturedAppID string

		realmClient := mock.RealmClient{}

		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			capturedAppFilter = filter
			return []realm.App{app}, nil
		}

		realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
			capturedProjectID = groupID
			capturedAppID = appID
			return testUsers[:1], nil
		}

		cmd := &CommandList{listInputs{
			ProjectInputs: cli.ProjectInputs{
				Project: projectID,
				App:     appID,
			},
		}}

		assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
		assert.Equal(t, strings.Join([]string{
			"Provider type: Anonymous",
			"  ID      Type    Enabled  Last Authenticated",
			"  ------  ------  -------  ------------------",
			"  user-1  type-1  true     n/a               ",
			"",
		}, "\n"), out.String())

		assert.Equal(t, realm.AppFilter{App: appID, GroupID: projectID}, capturedAppFilter)
		assert.Equal(t, projectID, capturedProjectID)
		assert.Equal(t, appID, capturedAppID)
	})

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
				cmd := &CommandList{}

				err := cmd.Handler(nil, nil, cli.Clients{Realm: tc.setupClient()})
				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})
}

func TestTableRowList(t *testing.T) {
	t.Run("should show successful list user row", func(t *testing.T) {
		output := userOutput{
			user: realm.User{
				LastAuthenticationDate: 1111111111,
			},
		}
		expectedRow := map[string]interface{}{
			"Enabled":            true,
			"Last Authenticated": "2005-03-18 01:58:31 +0000 UTC",
		}
		row := map[string]interface{}{}
		tableRowList(output, row)

		assert.Equal(t, expectedRow, row)
	})
}
