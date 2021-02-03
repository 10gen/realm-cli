package secrets

import (
	"errors"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestSecretsCreateSetup(t *testing.T) {
	t.Run("should construct a realm client with the configured base url", func(t *testing.T) {
		profile := mock.NewProfile(t)
		profile.SetRealmBaseURL("http://localhost:8080")

		cmd := &CommandCreate{inputs: createInputs{}}
		assert.Nil(t, cmd.realmClient)

		assert.Nil(t, cmd.Setup(profile, nil))
		assert.NotNil(t, cmd.realmClient)
	})
}

func TestSecretsCreateHandler(t *testing.T) {
	projectID := "projectID"
	appID := "appID"
	secretID := "secretID"
	secretName := "secretname"
	secretValue := "secretvalue"
	app := realm.App{
		ID:          appID,
		GroupID:     projectID,
		ClientAppID: "eggcorn-abcde",
		Name:        "eggcorn",
	}

	t.Run("should create app secrets", func(t *testing.T) {
		realmClient := mock.RealmClient{}
		var capturedFilter realm.AppFilter
		var capturedGroupID, capturedAppID, capturedName, capturedValue string
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			capturedFilter = filter
			return []realm.App{app}, nil
		}

		realmClient.CreateSecretFn = func(groupID, appID, name, value string) (realm.Secret, error) {
			capturedGroupID = groupID
			capturedAppID = appID
			capturedName = name
			capturedValue = value
			return realm.Secret{secretID, secretName}, nil
		}

		cmd := &CommandCreate{
			inputs: createInputs{
				ProjectInputs: cli.ProjectInputs{
					Project: projectID,
					App:     appID,
				},
				Name:  secretName,
				Value: secretValue,
			},
			realmClient: realmClient,
		}

		assert.Nil(t, cmd.Handler(nil, nil))
		assert.Equal(t, realm.Secret{secretID, secretName}, cmd.secret)

		t.Log("and should properly pass through the expected inputs")
		assert.Equal(t, realm.AppFilter{projectID, appID}, capturedFilter)
		assert.Equal(t, projectID, capturedGroupID)
		assert.Equal(t, appID, capturedAppID)
		assert.Equal(t, secretName, capturedName)
		assert.Equal(t, secretValue, capturedValue)

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
				description: "when creating a secret fails",
				setupClient: func() realm.Client {
					realmClient := mock.RealmClient{}
					realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
						return []realm.App{app}, nil
					}
					realmClient.CreateSecretFn = func(groupID, appID, name, value string) (realm.Secret, error) {
						return realm.Secret{}, errors.New("something bad happened")
					}
					return realmClient
				},
				expectedErr: errors.New("something bad happened"),
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				realmClient := tc.setupClient()

				cmd := &CommandCreate{
					realmClient: realmClient,
				}

				err := cmd.Handler(nil, nil)
				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})
}

func TestSecretsCreateFeedback(t *testing.T) {
	t.Run("should print a message that secret creation was successful", func(t *testing.T) {
		out, ui := mock.NewUI()

		cmd := &CommandCreate{secret: realm.Secret{ID: "testID"}}

		err := cmd.Feedback(nil, ui)
		assert.Nil(t, err)

		assert.Equal(t, "01:23:45 UTC INFO  Successfully created secret, id: testID\n", out.String())
	})
}
