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

func TestUserStateSetup(t *testing.T) {
	t.Run("should construct a realm client with the configured base url", func(t *testing.T) {
		profile := mock.NewProfile(t)
		profile.SetRealmBaseURL("http://localhost:8080")

		cmd := &CommandUserState{inputs: userStateInputs{}}
		assert.Nil(t, cmd.realmClient)

		assert.Nil(t, cmd.Setup(profile, nil))
		assert.NotNil(t, cmd.realmClient)
	})
}

func TestUserStateHandler(t *testing.T) {
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
		{
			ID:         "user-2",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeAnonymous}},
			Disabled:   true,
		},
	}

	t.Run("should disable a user when a user id is provided", func(t *testing.T) {
		var capturedAppFilter realm.AppFilter
		var capturedFindProjectID, capturedFindAppID string
		var capturedDisableProjectID, capturedDisableAppID string

		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			capturedAppFilter = filter
			return []realm.App{testApp}, nil
		}

		realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
			capturedFindProjectID = groupID
			capturedFindAppID = appID
			return testUsers[:1], nil
		}

		realmClient.DisableUserFn = func(groupID, appID, userID string) error {
			capturedDisableProjectID = groupID
			capturedDisableAppID = appID
			return nil
		}

		cmd := &CommandUserState{
			userEnable: false,
			inputs: userStateInputs{
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
		assert.Equal(t, projectID, capturedDisableProjectID)
		assert.Equal(t, appID, capturedDisableAppID)
		assert.Equal(t, testUsers[0].ID, cmd.inputs.Users[0])
		assert.Nil(t, cmd.outputs[0].err)
		assert.Equal(t, cmd.outputs[0].user, testUsers[0])
	})

	t.Run("should save failed disable errors", func(t *testing.T) {
		var capturedAppFilter realm.AppFilter
		var capturedFindProjectID, capturedFindAppID string
		var capturedDisableProjectID, capturedDisableAppID string

		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			capturedAppFilter = filter
			return []realm.App{testApp}, nil
		}

		realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
			capturedFindProjectID = groupID
			capturedFindAppID = appID
			return testUsers[:1], nil
		}

		realmClient.DisableUserFn = func(groupID, appID, userID string) error {
			capturedDisableProjectID = groupID
			capturedDisableAppID = appID
			return errors.New("client error")
		}

		cmd := &CommandUserState{
			userEnable: false,
			inputs: userStateInputs{
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
		assert.Equal(t, projectID, capturedDisableProjectID)
		assert.Equal(t, appID, capturedDisableAppID)
		assert.Equal(t, testUsers[0].ID, cmd.inputs.Users[0])
		assert.Equal(t, cmd.outputs[0].err, errors.New("client error"))
		assert.Equal(t, cmd.outputs[0].user, testUsers[0])
	})

	t.Run("should enable a user when a user id is provided", func(t *testing.T) {
		var capturedAppFilter realm.AppFilter
		var capturedFindProjectID, capturedFindAppID string
		var capturedEnableProjectID, capturedEnableAppID string

		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			capturedAppFilter = filter
			return []realm.App{testApp}, nil
		}

		realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
			capturedFindProjectID = groupID
			capturedFindAppID = appID
			return testUsers[1:], nil
		}

		realmClient.EnableUserFn = func(groupID, appID, userID string) error {
			capturedEnableProjectID = groupID
			capturedEnableAppID = appID
			return nil
		}

		cmd := &CommandUserState{
			userEnable: true,
			inputs: userStateInputs{
				ProjectInputs: app.ProjectInputs{
					Project: projectID,
					App:     appID,
				},
				usersInputs: usersInputs{
					Users: []string{testUsers[1].ID},
				},
			},
			realmClient: realmClient,
		}

		assert.Nil(t, cmd.Handler(nil, nil))
		assert.Equal(t, realm.AppFilter{App: appID, GroupID: projectID}, capturedAppFilter)
		assert.Equal(t, projectID, capturedFindProjectID)
		assert.Equal(t, appID, capturedFindAppID)
		assert.Equal(t, projectID, capturedEnableProjectID)
		assert.Equal(t, appID, capturedEnableAppID)
		assert.Equal(t, testUsers[1].ID, cmd.inputs.Users[0])
		assert.Nil(t, cmd.outputs[0].err)
		assert.Equal(t, cmd.outputs[0].user, testUsers[1])
	})

	t.Run("should save failed disable errors", func(t *testing.T) {
		var capturedAppFilter realm.AppFilter
		var capturedFindProjectID, capturedFindAppID string
		var capturedEnableProjectID, capturedEnableAppID string

		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			capturedAppFilter = filter
			return []realm.App{testApp}, nil
		}

		realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
			capturedFindProjectID = groupID
			capturedFindAppID = appID
			return testUsers[1:], nil
		}

		realmClient.EnableUserFn = func(groupID, appID, userID string) error {
			capturedEnableProjectID = groupID
			capturedEnableAppID = appID
			return errors.New("client error")
		}

		cmd := &CommandUserState{
			userEnable: true,
			inputs: userStateInputs{
				ProjectInputs: app.ProjectInputs{
					Project: projectID,
					App:     appID,
				},
				usersInputs: usersInputs{
					Users: []string{testUsers[1].ID},
				},
			},
			realmClient: realmClient,
		}

		assert.Nil(t, cmd.Handler(nil, nil))
		assert.Equal(t, realm.AppFilter{App: appID, GroupID: projectID}, capturedAppFilter)
		assert.Equal(t, projectID, capturedFindProjectID)
		assert.Equal(t, appID, capturedFindAppID)
		assert.Equal(t, projectID, capturedEnableProjectID)
		assert.Equal(t, appID, capturedEnableAppID)
		assert.Equal(t, testUsers[1].ID, cmd.inputs.Users[0])
		assert.Equal(t, cmd.outputs[0].err, errors.New("client error"))
		assert.Equal(t, cmd.outputs[0].user, testUsers[1])
	})

	t.Run("should return an error", func(t *testing.T) {
		for _, tc := range []struct {
			description string
			userEnable  bool
			setupClient func() realm.Client
			expectedErr error
		}{
			{
				description: "when resolving the app fails",
				userEnable:  false,
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
				userEnable:  false,
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

				cmd := &CommandUserState{
					userEnable:  tc.userEnable,
					realmClient: realmClient,
				}

				err := cmd.Handler(nil, nil)
				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})
}

func TestUserStateFeedback(t *testing.T) {
	testUsers := []realm.User{
		{
			ID:         "user-1",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeUserPassword}},
			Data:       map[string]interface{}{"email": "user-1@test.com"},
			Disabled:   false,
			Type:       "type-1",
		},
	}
	for _, tc := range []struct {
		description    string
		userEnable     bool
		outputs        []userOutput
		expectedOutput string
	}{
		{
			description:    "should show indicate no users to disable",
			userEnable:     false,
			outputs:        []userOutput{},
			expectedOutput: "01:23:45 UTC INFO  No users to disable\n",
		},
		{
			description: "should show 1 failed user, command: user disable",
			userEnable:  false,
			outputs: []userOutput{
				{user: testUsers[0], err: errors.New("client error")},
			},
			expectedOutput: strings.Join(
				[]string{
					"01:23:45 UTC INFO  Provider type: local-userpass",
					"  Email            ID      Type    Enabled  Details     ",
					"  ---------------  ------  ------  -------  ------------",
					"  user-1@test.com  user-1  type-1  true     client error",
					"",
				},
				"\n",
			),
		},
		{
			description:    "should show indicate no users to enable",
			userEnable:     true,
			outputs:        []userOutput{},
			expectedOutput: "01:23:45 UTC INFO  No users to enable\n",
		},
		{
			description: "should show 1 failed user, command: user enable",
			userEnable:  true,
			outputs: []userOutput{
				{user: testUsers[0], err: errors.New("client error")},
			},
			expectedOutput: strings.Join(
				[]string{
					"01:23:45 UTC INFO  Provider type: local-userpass",
					"  Email            ID      Type    Enabled  Details     ",
					"  ---------------  ------  ------  -------  ------------",
					"  user-1@test.com  user-1  type-1  false    client error",
					"",
				},
				"\n",
			),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			out, ui := mock.NewUI()

			cmd := &CommandUserState{
				userEnable: tc.userEnable,
				outputs:    tc.outputs,
			}

			assert.Nil(t, cmd.Feedback(nil, ui))
			assert.Equal(t, tc.expectedOutput, out.String())
		})
	}
}

func TestUserStateTableHeaders(t *testing.T) {
	for _, tc := range []struct {
		description      string
		authProviderType realm.AuthProviderType
		expectedHeaders  []string
	}{
		{
			description:      "should show name for apikey",
			authProviderType: realm.AuthProviderTypeAPIKey,
			expectedHeaders:  []string{"Name", "ID", "Type", "Enabled", "Details"},
		},
		{
			description:      "should show email for local-userpass",
			authProviderType: realm.AuthProviderTypeUserPassword,
			expectedHeaders:  []string{"Email", "ID", "Type", "Enabled", "Details"},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expectedHeaders, userStateTableHeaders(tc.authProviderType))
		})
	}
}

func TestUserStateTableRow(t *testing.T) {
	for _, tc := range []struct {
		description      string
		userEnable       bool
		authProviderType realm.AuthProviderType
		output           userOutput
		expectedRow      map[string]interface{}
	}{
		{
			description:      "should show name for apikey type user",
			userEnable:       false,
			authProviderType: "api-key",
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
				"Enabled": false,
				"Details": "n/a",
			},
		},
		{
			description:      "should show email for local-userpass type user",
			userEnable:       false,
			authProviderType: "local-userpass",
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
				"Enabled": false,
				"Details": "n/a",
			},
		},
		{
			description:      "should show name for apikey type user",
			userEnable:       true,
			authProviderType: "api-key",
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
				"Enabled": true,
				"Details": "n/a",
			},
		},
		{
			description:      "should show email for local-userpass type user",
			userEnable:       true,
			authProviderType: "local-userpass",
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
				"Enabled": true,
				"Details": "n/a",
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expectedRow, userStateTableRow(tc.authProviderType, tc.output, tc.userEnable))
		})
	}
}
