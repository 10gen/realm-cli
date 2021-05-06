package secrets

import (
	"errors"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

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
		out, ui := mock.NewUI()

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

		cmd := &CommandCreate{createInputs{
			ProjectInputs: cli.ProjectInputs{
				Project: projectID,
				App:     appID,
			},
			Name:  secretName,
			Value: secretValue,
		}}

		assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
		assert.Equal(t, "Successfully created secret, id: secretID\n", out.String())

		t.Log("and should properly pass through the expected inputs")
		assert.Equal(t, realm.AppFilter{projectID, appID, nil}, capturedFilter)
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

				cmd := &CommandCreate{}

				err := cmd.Handler(nil, nil, cli.Clients{Realm: realmClient})
				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})
}
