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

func TestUserListSetup(t *testing.T) {
	t.Run("Should construct a Realm client with the configured base url", func(t *testing.T) {
		profile := mock.NewProfile(t)
		profile.SetRealmBaseURL("http://localhost:8080")
		cmd := &CommandList{inputs: listInputs{}}

		assert.Nil(t, cmd.realmClient)
		assert.Nil(t, cmd.Setup(profile, nil))
		assert.NotNil(t, cmd.realmClient)
	})
}

func TestUserListHandler(t *testing.T) {
	projectID := "projectID"
	appID := "appID"
	testApp := realm.App{
		ID:          appID,
		GroupID:     projectID,
		ClientAppID: "eggcorn-abcde",
		Name:        "eggcorn",
	}

	t.Run("should find app users", func(t *testing.T) {
		testUsers := []realm.User{{ID: "user1"}, {ID: "user2"}}
		var capturedAppFilter realm.AppFilter
		var capturedProjectID, capturedAppID string
		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			capturedAppFilter = filter
			return []realm.App{testApp}, nil
		}
		realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
			capturedProjectID = groupID
			capturedAppID = appID
			return testUsers, nil
		}
		cmd := &CommandList{
			inputs: listInputs{
				ProjectInputs: app.ProjectInputs{
					Project: projectID,
					App:     appID,
				},
			},
			realmClient: realmClient,
		}

		assert.Nil(t, cmd.Handler(nil, nil))
		assert.Equal(t, realm.AppFilter{App: appID, GroupID: projectID}, capturedAppFilter)
		assert.Equal(t, projectID, capturedProjectID)
		assert.Equal(t, appID, capturedAppID)
		assert.Equal(t, testUsers[0], cmd.outputs[0].user)
		assert.Equal(t, testUsers[1], cmd.outputs[1].user)
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
				cmd := &CommandList{
					realmClient: realmClient,
				}
				err := cmd.Handler(nil, nil)
				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})
}

func TestUserListFeedback(t *testing.T) {
	for _, tc := range []struct {
		description     string
		outputs         []userOutput
		expectedContent string
	}{
		{
			description:     "whould indicate no users found when none are found",
			expectedContent: "01:23:45 UTC INFO  No available users to show\n",
		},
		{
			description: "whould group the users by provider type and sort by LastAuthenticationDate",
			outputs: []userOutput{
				{
					user: realm.User{
						ID:                     "id1",
						Identities:             []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeUserPassword}},
						Type:                   "type1",
						Data:                   map[string]interface{}{"email": "myEmail1"},
						CreationDate:           1111111111,
						LastAuthenticationDate: 1111111111,
					},
				},
				{
					user: realm.User{
						ID:                     "id2",
						Identities:             []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeUserPassword}},
						Type:                   "type2",
						Data:                   map[string]interface{}{"email": "myEmail2"},
						CreationDate:           1111333333,
						LastAuthenticationDate: 1111333333,
					},
				},
				{
					user: realm.User{
						ID:                     "id3",
						Identities:             []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeUserPassword}},
						Type:                   "type1",
						Data:                   map[string]interface{}{"email": "myEmail3"},
						CreationDate:           1111222222,
						LastAuthenticationDate: 1111222222,
					},
				},
				{
					user: realm.User{
						ID:                     "id4",
						Identities:             []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeAPIKey}},
						Type:                   "type1",
						Data:                   map[string]interface{}{"name": "myName"},
						CreationDate:           1111111111,
						LastAuthenticationDate: 1111111111,
					},
				},
			},
			expectedContent: strings.Join(
				[]string{
					"01:23:45 UTC INFO  Provider type: User/Password",
					"  Email     ID   Type   Enabled  Last Authenticated           ",
					"  --------  ---  -----  -------  -----------------------------",
					"  myEmail2  id2  type2  true     2005-03-20 15:42:13 +0000 UTC",
					"  myEmail3  id3  type1  true     2005-03-19 08:50:22 +0000 UTC",
					"  myEmail1  id1  type1  true     2005-03-18 01:58:31 +0000 UTC",
					"01:23:45 UTC INFO  Provider type: ApiKey",
					"  Name    ID   Type   Enabled  Last Authenticated           ",
					"  ------  ---  -----  -------  -----------------------------",
					"  myName  id4  type1  true     2005-03-18 01:58:31 +0000 UTC",
					"",
				},
				"\n",
			),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			out, ui := mock.NewUI()
			cmd := &CommandList{
				outputs: tc.outputs,
			}

			assert.Nil(t, cmd.Feedback(nil, ui))
			assert.Equal(t, tc.expectedContent, out.String())
		})
	}
}

func TestUserListRow(t *testing.T) {
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
		userListRow(output, row)

		assert.Equal(t, expectedRow, row)
	})
}
