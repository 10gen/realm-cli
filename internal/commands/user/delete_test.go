package user

import (
	"errors"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/app"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"github.com/Netflix/go-expect"
)

func TestResolveUsersInputs(t *testing.T) {
	testUsers := []realm.User{
		{
			ID:         "user-1",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeAnonymous}},
		},
		{
			ID:         "user-2",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeUserPassword}},
			Disabled:   true,
			Data:       map[string]interface{}{"email": "user-2@test.com"},
		},
		{
			ID:         "user-3",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeUserPassword}},
			Disabled:   true,
			Data:       map[string]interface{}{"email": "user-3@test.com"},
		},
	}

	t.Run("should prompt for users", func(t *testing.T) {
		for _, tc := range []struct {
			description   string
			inputs        deleteInputs
			procedure     func(c *expect.Console)
			users         []realm.User
			expectedUsers []realm.User
		}{
			{
				description: "with no input set",
				procedure: func(c *expect.Console) {
					c.ExpectString("Which user(s) would you like to delete?")
					c.Send("user-1")
					c.SendLine(" ")
					c.ExpectEOF()
				},
				users:         testUsers,
				expectedUsers: []realm.User{testUsers[0]},
			},
			{
				description: "with providers set",
				inputs:      deleteInputs{ProviderTypes: []string{realm.AuthProviderTypeUserPassword.String()}},
				procedure: func(c *expect.Console) {
					c.ExpectString("Which user(s) would you like to delete?")
					c.Send("user-2")
					c.SendLine(" ")
					c.ExpectEOF()
				},
				users:         testUsers[1:],
				expectedUsers: []realm.User{testUsers[1]},
			},
			{
				description: "with state set",
				inputs:      deleteInputs{State: realm.UserStateDisabled},
				procedure: func(c *expect.Console) {
					c.ExpectString("Which user(s) would you like to delete?")
					c.Send("user-2")
					c.SendLine(" ")
					c.ExpectEOF()
				},
				users:         testUsers[1:],
				expectedUsers: []realm.User{testUsers[1]},
			},
			{
				description: "with status set",
				inputs:      deleteInputs{State: realm.UserStateDisabled},
				procedure: func(c *expect.Console) {
					c.ExpectString("Which user(s) would you like to delete?")
					c.Send("user-3")
					c.SendLine(" ")
					c.ExpectEOF()
				},
				users:         testUsers[2:],
				expectedUsers: []realm.User{testUsers[2]},
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				realmClient := mock.RealmClient{}
				realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
					return tc.users, nil
				}

				_, console, _, ui, consoleErr := mock.NewVT10XConsole()
				assert.Nil(t, consoleErr)
				defer console.Close()

				doneCh := make(chan (struct{}))
				go func() {
					defer close(doneCh)
					tc.procedure(console)
				}()

				var app realm.App
				users, err := tc.inputs.resolveUsers(ui, realmClient, app)

				console.Tty().Close() // flush the writers
				<-doneCh              // wait for procedure to complete

				assert.Nil(t, err)
				assert.Equal(t, tc.expectedUsers, users)
			})
		}
	})

	for _, tc := range []struct {
		description   string
		users         []realm.User
		expectedUsers []realm.User
		expectedErr   error
	}{
		{
			description: "should error when a user cannot be found from provided ids",
			expectedErr: errors.New("no users found"),
		},
		{
			description:   "should find users from provided ids",
			users:         testUsers[:2],
			expectedUsers: testUsers[:2],
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			realmClient := mock.RealmClient{}
			realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
				return tc.users, nil
			}
			var app realm.App
			inputs := deleteInputs{Users: []string{testUsers[0].ID, testUsers[1].ID}}
			users, err := inputs.resolveUsers(nil, realmClient, app)

			assert.Equal(t, tc.expectedUsers, users)
			assert.Equal(t, tc.expectedErr, err)
		})
	}

	t.Run("should error from client", func(t *testing.T) {
		realmClient := mock.RealmClient{}
		realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
			return nil, errors.New("client error")
		}
		var app realm.App
		inputs := deleteInputs{}
		_, err := inputs.resolveUsers(nil, realmClient, app)

		assert.Equal(t, errors.New("client error"), err)
	})
}

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
		},
	}

	for _, tc := range []struct {
		description     string
		userDeleteErr   error
		expectedOutputs []userOutput
	}{
		{
			description: "should delete a user when a user id is provided",
			expectedOutputs: []userOutput{
				{
					user: testUsers[0],
				},
			},
		},
		{
			description:   "should save failed deletion errors",
			userDeleteErr: errors.New("client error"),
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
			var capturedDeleteProjectID, capturedDeleteAppID string
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
			realmClient.DeleteUserFn = func(groupID, appID, userID string) error {
				capturedDeleteProjectID = groupID
				capturedDeleteAppID = appID
				return tc.userDeleteErr
			}
			cmd := &CommandDelete{
				inputs: deleteInputs{
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
			assert.Equal(t, projectID, capturedDeleteProjectID)
			assert.Equal(t, appID, capturedDeleteAppID)
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
		outputs        []userOutput
		expectedOutput string
	}{
		{
			description:    "should show no users to delete",
			expectedOutput: "01:23:45 UTC INFO  No users to delete\n",
		},
		{
			description: "should show 1 failed user",
			outputs: []userOutput{
				{user: testUsers[0], err: errors.New("client error")},
			},
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
