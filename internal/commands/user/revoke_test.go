package user

import (
	"errors"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/app"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestUserRevokeSetup(t *testing.T) {
	t.Run("Should construct a Realm client with the configured base url", func(t *testing.T) {
		profile := mock.NewProfile(t)
		profile.SetRealmBaseURL("http://localhost:8080")

		cmd := &CommandRevoke{inputs: revokeInputs{}}
		assert.Nil(t, cmd.realmClient)

		assert.Nil(t, cmd.Setup(profile, nil))
		assert.NotNil(t, cmd.realmClient)
	})
}

func TestUserRevokeHandler(t *testing.T) {
	projectID := "projectID"
	appID := "appID"
	testApp := realm.App{
		ID:          appID,
		GroupID:     projectID,
		ClientAppID: "eggcorn-abcde",
		Name:        "eggcorn",
	}
	testUsers := []realm.User{
		{
			ID:         "user-1",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeAnonymous}},
			Disabled:   false,
		},
	}

	for _, tc := range []struct {
		description     string
		userRevokeErr   error
		expectedOutputs []userOutput
	}{
		{
			description: "should revoke a user's sesions when a user id is provided",
			expectedOutputs: []userOutput{
				{
					user: testUsers[0],
				},
			},
		},
		{
			description:   "should save failed revoke errors",
			userRevokeErr: errors.New("client error"),
			expectedOutputs: []userOutput{
				{
					user: testUsers[0],
					err:  errors.New("client error"),
				},
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			var capturedAppFilter realm.AppFilter
			var capturedFindProjectID, capturedFindAppID string
			var capturedRevokeProjectID, capturedRevokeAppID string

			realmClient := mock.RealmClient{}
			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				capturedAppFilter = filter
				return []realm.App{testApp}, nil
			}

			realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
				capturedFindProjectID = groupID
				capturedFindAppID = appID
				return testUsers, nil
			}

			realmClient.RevokeUserSessionsFn = func(groupID, appID, userID string) error {
				capturedRevokeProjectID = groupID
				capturedRevokeAppID = appID
				return tc.userRevokeErr
			}

			cmd := &CommandRevoke{
				inputs: revokeInputs{
					ProjectInputs: app.ProjectInputs{
						Project: projectID,
						App:     appID,
					},
					usersInputs: usersInputs{
						Users: []string{testUsers[0].ID},
					},
				},
				realmClient: realmClient,
			}

			assert.Nil(t, cmd.Handler(nil, nil))
			assert.Equal(t, realm.AppFilter{App: appID, GroupID: projectID}, capturedAppFilter)
			assert.Equal(t, projectID, capturedFindProjectID)
			assert.Equal(t, appID, capturedFindAppID)
			assert.Equal(t, projectID, capturedRevokeProjectID)
			assert.Equal(t, appID, capturedRevokeAppID)
			assert.Equal(t, cmd.outputs[0].err, tc.expectedOutputs[0].err)
			assert.Equal(t, cmd.outputs[0].user, tc.expectedOutputs[0].user)
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
						return []realm.App{testApp}, nil
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

				cmd := &CommandRevoke{
					realmClient: realmClient,
				}

				err := cmd.Handler(nil, nil)
				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})
}

func TestUserRevokeFeedback(t *testing.T) {
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
		outputs        []userOutput
		expectedOutput string
	}{
		{
			description:    "should show no users to revoke",
			expectedOutput: "01:23:45 UTC INFO  No users to revoke, try changing the --user input\n",
		},
		{
			description: "should show 1 failed user",
			outputs: []userOutput{
				{user: testUsers[0], err: errors.New("client error")},
			},
			expectedOutput: strings.Join(
				[]string{
					"01:23:45 UTC INFO  Provider type: User/Password",
					"  Email            ID      Type    Revoked  Details     ",
					"  ---------------  ------  ------  -------  ------------",
					"  user-1@test.com  user-1  type-1  no       client error",
					"",
				},
				"\n",
			),
		},
		{
			description: "should show 2 failed users",
			outputs: []userOutput{
				{user: testUsers[0], err: nil},
				{user: testUsers[1], err: errors.New("client error")},
				{user: testUsers[2], err: nil},
				{user: testUsers[3], err: errors.New("client error")},
				{user: testUsers[4], err: nil},
			},
			expectedOutput: strings.Join(
				[]string{
					"01:23:45 UTC INFO  Provider type: User/Password",
					"  Email            ID      Type    Revoked  Details     ",
					"  ---------------  ------  ------  -------  ------------",
					"  user-2@test.com  user-2  type-2  no       client error",
					"  user-1@test.com  user-1  type-1  yes                  ",
					"  user-3@test.com  user-3  type-1  yes                  ",
					"01:23:45 UTC INFO  Provider type: ApiKey",
					"  Name    ID      Type    Revoked  Details     ",
					"  ------  ------  ------  -------  ------------",
					"  name-4  user-4  type-1  no       client error",
					"01:23:45 UTC INFO  Provider type: Custom JWT",
					"  ID      Type    Revoked  Details",
					"  ------  ------  -------  -------",
					"  user-5  type-3  yes             ",
					"",
				},
				"\n",
			),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			out, ui := mock.NewUI()

			cmd := &CommandRevoke{
				outputs: tc.outputs,
			}

			assert.Nil(t, cmd.Feedback(nil, ui))
			assert.Equal(t, tc.expectedOutput, out.String())
		})
	}
}

func TestUserRevokeTableHeaders(t *testing.T) {
	for _, tc := range []struct {
		description      string
		authProviderType realm.AuthProviderType
		expectedHeaders  []string
	}{
		{
			description:      "should show name for apikey",
			authProviderType: realm.AuthProviderTypeAPIKey,
			expectedHeaders:  []string{"Name", "ID", "Type", "Revoked", "Details"},
		},
		{
			description:      "should show email for local-userpass",
			authProviderType: realm.AuthProviderTypeUserPassword,
			expectedHeaders:  []string{"Email", "ID", "Type", "Revoked", "Details"},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expectedHeaders, userRevokeTableHeaders(tc.authProviderType))
		})
	}
}

func TestUserRevokeTableRow(t *testing.T) {
	for _, tc := range []struct {
		description      string
		authProviderType realm.AuthProviderType
		output           userOutput
		expectedRow      map[string]interface{}
	}{
		{
			description:      "should show name for apikey type user",
			authProviderType: realm.AuthProviderTypeAPIKey,
			output: userOutput{
				user: realm.User{
					ID:         "user-1",
					Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeAPIKey}},
					Type:       "type-1",
					Data:       map[string]interface{}{"name": "name-1"},
				},
				err: nil,
			},
			expectedRow: map[string]interface{}{
				"ID":      "user-1",
				"Name":    "name-1",
				"Type":    "type-1",
				"Revoked": true,
				"Details": "",
			},
		},
		{
			description:      "should show email for local-userpass type user",
			authProviderType: realm.AuthProviderTypeUserPassword,
			output: userOutput{
				user: realm.User{
					ID:         "user-1",
					Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeUserPassword}},
					Type:       "type-1",
					Data:       map[string]interface{}{"email": "user-1@test.com"},
				},
				err: nil,
			},
			expectedRow: map[string]interface{}{
				"ID":      "user-1",
				"Email":   "user-1@test.com",
				"Type":    "type-1",
				"Revoked": true,
				"Details": "",
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expectedRow, userRevokeTableRow(tc.authProviderType, tc.output))
		})
	}
}
