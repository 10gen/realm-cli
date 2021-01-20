package secrets

import (
	"errors"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/app"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestSecretsListSetup(t *testing.T) {
	t.Run("Should construct a Realm client with the configured base url", func(t *testing.T) {
		profile := mock.NewProfile(t)
		profile.SetRealmBaseURL("http://localhost:8080")

		cmd := &CommandList{inputs: listInputs{}}
		assert.Nil(t, cmd.realmClient)

		assert.Nil(t, cmd.Setup(profile, nil))
		assert.NotNil(t, cmd.realmClient)
	})
}

func TestSecretsListHandler(t *testing.T) {
	projectID := "projectID"
	appID := "appID"
	testApp := realm.App{
		ID:          appID,
		GroupID:     projectID,
		ClientAppID: "eggcorn-abcde",
		Name:        "eggcorn",
	}

	t.Run("Should find app values and secrets", func(t *testing.T) {
		testValues := []realm.Value{
			{
				ID:           "user1",
				Name:         "test1",
				LastModified: 1111111111,
			},
			{
				ID:           "user2",
				Name:         "test2",
				LastModified: 11111122,
				Secret:       false,
			},
			{
				ID:           "user3",
				Name:         "duplicate",
				LastModified: 11111122,
				Secret:       false,
			},
			{
				ID:           "user4",
				Name:         "duplicate",
				LastModified: 11111122,
				Secret:       true,
			},
		}

		var capturedAppFilter realm.AppFilter
		var capturedProjectID, capturedAppID string

		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			capturedAppFilter = filter
			return []realm.App{testApp}, nil
		}

		realmClient.FindValuesFn = func(app realm.App) ([]realm.Value, error) {
			capturedProjectID = app.GroupID
			capturedAppID = app.ID
			return testValues, nil
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
		assert.Equal(t, testValues, cmd.values)
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
				description: "When finding the secrets fails",
				setupClient: func() realm.Client {
					realmClient := mock.RealmClient{}
					realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
						return []realm.App{testApp}, nil
					}
					realmClient.FindValuesFn = func(app realm.App) ([]realm.Value, error) {
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

func TestSecretsListFeedback(t *testing.T) {
	for _, tc := range []struct {
		description    string
		values         []realm.Value
		expectedOutput string
	}{
		{
			description:    "Should indicate no secrets found when none are found",
			values:         []realm.Value{},
			expectedOutput: "01:23:45 UTC INFO  No available secrets to show\n",
		},
		{
			description: "Should display all found secrets",
			values: []realm.Value{
				{
					ID:           "id1",
					Name:         "test1",
					LastModified: 1111111111,
					Secret:       true,
				},
				{
					ID:           "id2",
					Name:         "test2",
					LastModified: 1111333333,
				},
				{
					ID:           "id3",
					Name:         "dup",
					LastModified: 1111222222,
					Secret:       true,
				},
				{
					ID:           "id4",
					Name:         "dup",
					LastModified: 1111111111,
				},
			},
			expectedOutput: strings.Join(
				[]string{
					"01:23:45 UTC INFO  Found 4 secrets",
					"  Name   ID   Is Secret  Last Modified                ",
					"  -----  ---  ---------  -----------------------------",
					"  test1  id1  true       2005-03-18 01:58:31 +0000 UTC",
					"  test2  id2  false      2005-03-20 15:42:13 +0000 UTC",
					"  dup    id3  true       2005-03-19 08:50:22 +0000 UTC",
					"  dup    id4  false      2005-03-18 01:58:31 +0000 UTC",
					"",
				},
				"\n",
			),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			out, ui := mock.NewUI()
			cmd := &CommandList{
				values: tc.values,
			}
			assert.Nil(t, cmd.Feedback(nil, ui))
			assert.Equal(t, tc.expectedOutput, out.String())
		})
	}
}
