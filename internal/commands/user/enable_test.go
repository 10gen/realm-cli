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

func TestUserEnableSetup(t *testing.T) {
	t.Run("should construct a realm client with the configured base url", func(t *testing.T) {
		profile := mock.NewProfile(t)
		profile.SetRealmBaseURL("http://localhost:8080")
		cmd := &CommandEnable{inputs: enableInputs{}}

		assert.Nil(t, cmd.realmClient)
		assert.Nil(t, cmd.Setup(profile, nil))
		assert.NotNil(t, cmd.realmClient)
	})
}

func TestUserEnableHandler(t *testing.T) {
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
			Disabled:   true,
		},
		{
			ID:         "user-2",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeAnonymous}},
		},
	}

	for _, tc := range []struct {
		description   string
		enableUserErr error
	}{
		{description: "should enable a user when a user id is provided"},
		{description: "should save failed enable errors", enableUserErr: errors.New("client error")},
	} {
		t.Run(tc.description, func(t *testing.T) {
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
				return testUsers[:1], nil
			}
			realmClient.EnableUserFn = func(groupID, appID, userID string) error {
				capturedEnableProjectID = groupID
				capturedEnableAppID = appID
				return tc.enableUserErr
			}
			cmd := &CommandEnable{
				inputs: enableInputs{
					ProjectInputs: app.ProjectInputs{
						Project: projectID,
						App:     appID,
					},
					Users: []string{testUsers[0].ID},
				},
				realmClient: realmClient,
			}

			assert.Nil(t, cmd.Handler(nil, nil))
			assert.Equal(t, realm.AppFilter{App: appID, GroupID: projectID}, capturedAppFilter)
			assert.Equal(t, projectID, capturedFindProjectID)
			assert.Equal(t, appID, capturedFindAppID)
			assert.Equal(t, projectID, capturedEnableProjectID)
			assert.Equal(t, appID, capturedEnableAppID)
			assert.Equal(t, testUsers[0].ID, cmd.inputs.Users[0])
			assert.Equal(t, tc.enableUserErr, cmd.outputs[0].err)
			assert.Equal(t, testUsers[0], cmd.outputs[0].user)
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
				cmd := &CommandEnable{
					realmClient: realmClient,
				}
				err := cmd.Handler(nil, nil)

				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})
}

func TestUserEnableFeedback(t *testing.T) {
	testUsers := []realm.User{
		{
			ID:         "user-1",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeUserPassword}},
			Type:       "type-1",
			Disabled:   true,
			Data:       map[string]interface{}{"email": "user-1@test.com"},
		},
		{
			ID:         "user-2",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeUserPassword}},
			Type:       "type-2",
			Disabled:   true,
			Data:       map[string]interface{}{"email": "user-2@test.com"},
		},
		{
			ID:         "user-3",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeUserPassword}},
			Type:       "type-1",
			Disabled:   true,
			Data:       map[string]interface{}{"email": "user-3@test.com"},
		},
		{
			ID:         "user-4",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeAPIKey}},
			Type:       "type-1",
			Disabled:   true,
			Data:       map[string]interface{}{"name": "name-4"},
		},
		{
			ID:         "user-5",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeCustomToken}},
			Type:       "type-3",
			Disabled:   true,
		},
	}
	for _, tc := range []struct {
		description     string
		outputs         []userOutput
		expectedContent string
	}{
		{
			description:     "should show no users to enable",
			expectedContent: "01:23:45 UTC INFO  No users to enable\n",
		},
		{
			description: "should show 1 failed user",
			outputs: []userOutput{
				{user: testUsers[0], err: errors.New("client error")},
			},
			expectedContent: strings.Join(
				[]string{
					"01:23:45 UTC INFO  Provider type: User/Password",
					"  Email            ID      Type    Enabled  Details     ",
					"  ---------------  ------  ------  -------  ------------",
					"  user-1@test.com  user-1  type-1  false    client error",
					"",
				},
				"\n",
			),
		},
		{
			description: "should show failures to enable 2 users amongst successful results across different auth provider types",
			outputs: []userOutput{
				{user: testUsers[0], err: nil},
				{user: testUsers[1], err: errors.New("client error")},
				{user: testUsers[2], err: nil},
				{user: testUsers[3], err: errors.New("client error")},
				{user: testUsers[4], err: nil},
			},
			expectedContent: strings.Join(
				[]string{
					"01:23:45 UTC INFO  Provider type: User/Password",
					"  Email            ID      Type    Enabled  Details     ",
					"  ---------------  ------  ------  -------  ------------",
					"  user-2@test.com  user-2  type-2  false    client error",
					"  user-1@test.com  user-1  type-1  true                 ",
					"  user-3@test.com  user-3  type-1  true                 ",
					"01:23:45 UTC INFO  Provider type: ApiKey",
					"  Name    ID      Type    Enabled  Details     ",
					"  ------  ------  ------  -------  ------------",
					"  name-4  user-4  type-1  false    client error",
					"01:23:45 UTC INFO  Provider type: Custom JWT",
					"  ID      Type    Enabled  Details",
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
			cmd := &CommandEnable{outputs: tc.outputs}

			assert.Nil(t, cmd.Feedback(nil, ui))
			assert.Equal(t, tc.expectedContent, out.String())
		})
	}
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
			userEnableRow(output, row)

			assert.Equal(t, tc.expectedRow, row)
		})
	}
}
