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
	testSecrets := []realm.Secret{
		{
			ID:   "secret1",
			Name: "test1",
		},
		{
			ID:   "secret2",
			Name: "test2",
		},
		{
			ID:   "secret3",
			Name: "duplicate",
		},
		{
			ID:   "secret4",
			Name: "duplicate",
		},
	}

	t.Run("Should find app secrets", func(t *testing.T) {
		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{testApp}, nil
		}

		realmClient.FindSecretsFn = func(groupID, appID string) ([]realm.Secret, error) {
			return testSecrets, nil
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
		assert.Equal(t, testSecrets, cmd.secrets)
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
					realmClient.FindSecretsFn = func(groupID, appID string) ([]realm.Secret, error) {
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
	testSecrets := []realm.Secret{
		{
			ID:   "60066e14734d0b6c336ffc23",
			Name: "test1",
		},
		{
			ID:   "234566e14734d0b6c336ffc2",
			Name: "test2",
		},
		{
			ID:   "60066e14564d0b6c336ffc23",
			Name: "dup",
		},
		{
			ID:   "60066e14734d0b6c886ffc23",
			Name: "dup",
		},
	}

	for _, tc := range []struct {
		description    string
		secrets        []realm.Secret
		expectedOutput string
	}{
		{
			description:    "Should indicate no secrets found when none are found",
			secrets:        []realm.Secret{},
			expectedOutput: "01:23:45 UTC INFO  No available secrets to show\n",
		},
		{
			description: "Should display all found secrets",
			secrets:     testSecrets,
			expectedOutput: strings.Join(
				[]string{
					"01:23:45 UTC INFO  Found 4 secrets",
					"  ID                        Name ",
					"  ------------------------  -----",
					"  60066e14734d0b6c336ffc23  test1",
					"  234566e14734d0b6c336ffc2  test2",
					"  60066e14564d0b6c336ffc23  dup  ",
					"  60066e14734d0b6c886ffc23  dup  ",
					"",
				},
				"\n",
			),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			out, ui := mock.NewUI()
			cmd := &CommandList{
				secrets: tc.secrets,
			}
			assert.Nil(t, cmd.Feedback(nil, ui))
			assert.Equal(t, tc.expectedOutput, out.String())
		})
	}
}
