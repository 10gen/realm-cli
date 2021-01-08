package list

import (
	"errors"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestUserListSetup(t *testing.T) {
	t.Run("Should construct a Realm client with the configured base url", func(t *testing.T) {
		profile := mock.NewProfile(t)
		profile.SetRealmBaseURL("http://localhost:8080")

		cmd := &command{inputs: inputs{}}
		assert.Nil(t, cmd.realmClient)

		assert.Nil(t, cmd.Setup(profile, nil))
		assert.NotNil(t, cmd.realmClient)
	})
}

func TestUserListHandler(t *testing.T) {

	const projectID = "projectID"
	const appID = "appID"
	testApp := realm.App{
		ID:          appID,
		GroupID:     projectID,
		ClientAppID: "eggcorn-abcde",
		Name:        "eggcorn",
	}

	t.Run("Should find app users", func(t *testing.T) {
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

		cmd := &command{
			inputs: inputs{
				ProjectAppInputs: cli.ProjectAppInputs{
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
		assert.Equal(t, testUsers, cmd.users)
	})

	t.Run("Should return an error", func(t *testing.T) {
		for _, tc := range []struct {
			description string
			setupClient func() realm.Client
			expectedErr error
		}{
			{
				description: "When resolving the app fails",
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
				description: "When finding the users fails",
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

				cmd := &command{
					realmClient: realmClient,
				}

				err := cmd.Handler(nil, nil)
				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})
}

func TestUserTableHeaders(t *testing.T) {
	for _, tc := range []struct {
		description     string
		providerType    string
		expectedHeaders []string
	}{
		{
			description:     "Should show name for apikey",
			providerType:    "api-key",
			expectedHeaders: []string{"Name", "ID", "Enabled", "Type", "Last Authentication"},
		},
		{
			description:     "Should show email for local-userpass",
			providerType:    "local-userpass",
			expectedHeaders: []string{"Email", "ID", "Enabled", "Type", "Last Authentication"},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expectedHeaders, userTableHeaders(tc.providerType))
		})
	}
}

func TestUserTableRow(t *testing.T) {
	for _, tc := range []struct {
		description  string
		providerType string
		user         realm.User
		expectedRow  map[string]interface{}
	}{
		{
			description:  "Should show name for apikey type user",
			providerType: "api-key",
			user: realm.User{
				ID:                     "id1",
				Identities:             []realm.UserIdentity{{ProviderType: "api-key"}},
				Type:                   "type1",
				Disabled:               false,
				Data:                   map[string]interface{}{"name": "myName"},
				CreationDate:           1111111111,
				LastAuthenticationDate: 1111111111,
			},
			expectedRow: map[string]interface{}{
				"Enabled":             true,
				"ID":                  "id1",
				"Last Authentication": "2005-03-18 01:58:31 +0000 UTC",
				"Name":                "myName",
				"Type":                "type1",
			},
		},
		{
			description:  "Should show email for local-userpass type user",
			providerType: "local-userpass",
			user: realm.User{
				ID:                     "id1",
				Identities:             []realm.UserIdentity{{ProviderType: "local-userpass"}},
				Type:                   "type1",
				Disabled:               false,
				Data:                   map[string]interface{}{"email": "myEmail"},
				CreationDate:           1111111111,
				LastAuthenticationDate: 1111111111,
			},
			expectedRow: map[string]interface{}{
				"Enabled":             true,
				"ID":                  "id1",
				"Last Authentication": "2005-03-18 01:58:31 +0000 UTC",
				"Email":               "myEmail",
				"Type":                "type1",
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expectedRow, userTableRow(tc.providerType, tc.user))
		})
	}
}

func TestUserListFeedback(t *testing.T) {
	for _, tc := range []struct {
		description    string
		users          []realm.User
		expectedOutput string
	}{
		{
			description:    "Should indicate no users found when none are found",
			users:          []realm.User{},
			expectedOutput: "01:23:45 UTC INFO  No available users to show\n",
		},
		{
			description: "Should group the users by provider type and sort by LastAuthenticationDate",
			users: []realm.User{
				{
					ID:                     "id1",
					Identities:             []realm.UserIdentity{{ProviderType: "local-userpass"}},
					Type:                   "type1",
					Disabled:               false,
					Data:                   map[string]interface{}{"email": "myEmail1"},
					CreationDate:           1111111111,
					LastAuthenticationDate: 1111111111,
				},
				{
					ID:                     "id2",
					Identities:             []realm.UserIdentity{{ProviderType: "local-userpass"}},
					Type:                   "type2",
					Disabled:               false,
					Data:                   map[string]interface{}{"email": "myEmail2"},
					CreationDate:           1111333333,
					LastAuthenticationDate: 1111333333,
				},
				{
					ID:                     "id3",
					Identities:             []realm.UserIdentity{{ProviderType: "local-userpass"}},
					Type:                   "type1",
					Disabled:               false,
					Data:                   map[string]interface{}{"email": "myEmail3"},
					CreationDate:           1111222222,
					LastAuthenticationDate: 1111222222,
				},
				{
					ID:                     "id4",
					Identities:             []realm.UserIdentity{{ProviderType: "api-key"}},
					Type:                   "type1",
					Disabled:               false,
					Data:                   map[string]interface{}{"name": "myName"},
					CreationDate:           1111111111,
					LastAuthenticationDate: 1111111111,
				},
			},
			expectedOutput: strings.Join(
				[]string{
					"01:23:45 UTC INFO  Provider type: local-userpass",
					"  Email     ID   Enabled  Type   Last Authentication          ",
					"  --------  ---  -------  -----  -----------------------------",
					"  myEmail2  id2  true     type2  2005-03-20 15:42:13 +0000 UTC",
					"  myEmail3  id3  true     type1  2005-03-19 08:50:22 +0000 UTC",
					"  myEmail1  id1  true     type1  2005-03-18 01:58:31 +0000 UTC",
					"01:23:45 UTC INFO  Provider type: api-key",
					"  Name    ID   Enabled  Type   Last Authentication          ",
					"  ------  ---  -------  -----  -----------------------------",
					"  myName  id4  true     type1  2005-03-18 01:58:31 +0000 UTC",
					"",
				},
				"\n",
			),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			out, ui := mock.NewUI()

			cmd := &command{
				users: tc.users,
			}

			assert.Nil(t, cmd.Feedback(nil, ui))

			assert.Equal(t, tc.expectedOutput, out.String())
		})
	}
}
