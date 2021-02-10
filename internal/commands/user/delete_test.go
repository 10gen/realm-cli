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

func TestUserDeleteSetup(t *testing.T) {
	t.Run("should construct a realm client with the configured base url", func(t *testing.T) {
		profile := mock.NewProfile(t)
		profile.SetRealmBaseURL("http://localhost:8080")
		cmd := &CommandDelete{inputs: deleteInputs{}}

		assert.Nil(t, cmd.realmClient)
		assert.Nil(t, cmd.Setup(profile, nil))
		assert.NotNil(t, cmd.realmClient)
	})
}

func TestUserDeleteHandler(t *testing.T) {
	projectID := "projectID"
	appID := "appID"
	app := realm.App{
		ID:          appID,
		GroupID:     projectID,
		ClientAppID: "eggcorn-abcde",
		Name:        "eggcorn",
	}
	testUsers := []realm.User{
		{
			ID:         "user-1",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeAnonymous}},
		},
	}

	for _, tc := range []struct {
		description     string
		userDeleteErr   error
		expectedOutputs userOutputs
	}{
		{
			description:     "should delete a user when a user id is provided",
			expectedOutputs: userOutputs{{testUsers[0], nil}},
		},
		{
			description:     "should save failed deletion errors",
			userDeleteErr:   errors.New("client error"),
			expectedOutputs: userOutputs{{testUsers[0], errors.New("client error")}},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
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
				return testUsers, nil
			}

			var capturedDeleteProjectID, capturedDeleteAppID string
			realmClient.DeleteUserFn = func(groupID, appID, userID string) error {
				capturedDeleteProjectID = groupID
				capturedDeleteAppID = appID
				return tc.userDeleteErr
			}

			cmd := &CommandDelete{
				inputs: deleteInputs{
					ProjectInputs: cli.ProjectInputs{
						Project: projectID,
						App:     appID,
					},
					multiUserInputs: multiUserInputs{
						Users: []string{testUsers[0].ID},
					},
				},
				realmClient: realmClient,
			}

			assert.Nil(t, cmd.Handler(nil, nil))

			assert.Equal(t, cmd.outputs[0].user, tc.expectedOutputs[0].user)
			assert.Equal(t, cmd.outputs[0].err, tc.expectedOutputs[0].err)

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
				realmClient := tc.setupClient()

				cmd := &CommandDelete{
					realmClient: realmClient,
				}

				err := cmd.Handler(nil, nil)
				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})
}

func TestUserDeleteFeedback(t *testing.T) {
	testUsers := []realm.User{
		{
			ID:         "user-1",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeUserPassword}},
			Type:       "type-1",
			Data:       map[string]interface{}{"email": "user-1@test.com"},
		},
		{
			ID:         "user-2",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeUserPassword}},
			Type:       "type-2",
			Data:       map[string]interface{}{"email": "user-2@test.com"},
		},
		{
			ID:         "user-3",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeUserPassword}},
			Type:       "type-1",
			Data:       map[string]interface{}{"email": "user-3@test.com"},
		},
		{
			ID:         "user-4",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeAPIKey}},
			Type:       "type-1",
			Data:       map[string]interface{}{"name": "name-4"},
		},
		{
			ID:         "user-5",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeCustomToken}},
			Type:       "type-3",
		},
	}
	for _, tc := range []struct {
		description    string
		outputs        userOutputs
		expectedOutput string
	}{
		{
			description:    "should show no users to delete",
			expectedOutput: "01:23:45 UTC INFO  No users to delete\n",
		},
		{
			description: "should show 1 failed user",
			outputs:     userOutputs{{testUsers[0], errors.New("client error")}},
			expectedOutput: strings.Join(
				[]string{
					"01:23:45 UTC INFO  Provider type: User/Password",
					"  Email            ID      Type    Deleted  Details     ",
					"  ---------------  ------  ------  -------  ------------",
					"  user-1@test.com  user-1  type-1  false    client error",
					"",
				},
				"\n",
			),
		},
		{
			description: "should show 2 failed users",
			outputs: userOutputs{
				{testUsers[0], nil},
				{testUsers[1], errors.New("client error")},
				{testUsers[2], nil},
				{testUsers[3], errors.New("client error")},
				{testUsers[4], nil},
			},
			expectedOutput: strings.Join(
				[]string{
					"01:23:45 UTC INFO  Provider type: User/Password",
					"  Email            ID      Type    Deleted  Details     ",
					"  ---------------  ------  ------  -------  ------------",
					"  user-2@test.com  user-2  type-2  false    client error",
					"  user-1@test.com  user-1  type-1  true                 ",
					"  user-3@test.com  user-3  type-1  true                 ",
					"01:23:45 UTC INFO  Provider type: ApiKey",
					"  Name    ID      Type    Deleted  Details     ",
					"  ------  ------  ------  -------  ------------",
					"  name-4  user-4  type-1  false    client error",
					"01:23:45 UTC INFO  Provider type: Custom JWT",
					"  ID      Type    Deleted  Details",
					"  ------  ------  -------  -------",
					"  user-5  type-3  true            ",
					"",
				},
				"\n",
			),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			out, ui := mock.NewUI()
			cmd := &CommandDelete{
				outputs: tc.outputs,
			}

			assert.Nil(t, cmd.Feedback(nil, ui))
			assert.Equal(t, tc.expectedOutput, out.String())
		})
	}
}

func TestUserDeleteRow(t *testing.T) {
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
			userDeleteRow(output, row)

			assert.Equal(t, tc.expectedRow, row)
		})
	}
}
