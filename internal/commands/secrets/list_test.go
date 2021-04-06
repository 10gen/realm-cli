package secrets

import (
	"errors"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestSecretsListHandler(t *testing.T) {
	projectID := "projectID"
	appID := "appID"
	app := realm.App{
		ID:          appID,
		GroupID:     projectID,
		ClientAppID: "eggcorn-abcde",
		Name:        "eggcorn",
	}
	testSecrets := []realm.Secret{
		{ID: "secret1", Name: "test1"},
		{ID: "secret2", Name: "test2"},
		{ID: "secret3", Name: "dup"},
		{ID: "secret4", Name: "dup"},
	}

	for _, tc := range []struct {
		description    string
		secrets        []realm.Secret
		expectedOutput string
	}{
		{
			description:    "should list no secrets with no app secrets found",
			expectedOutput: "No available secrets to show\n",
		},
		{
			description: "should list the secrets found for the app",
			secrets:     testSecrets,
			expectedOutput: strings.Join(
				[]string{
					"Found 4 secrets",
					"  ID       Name ",
					"  -------  -----",
					"  secret1  test1",
					"  secret2  test2",
					"  secret3  dup  ",
					"  secret4  dup  ",
					"",
				},
				"\n",
			),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			out, ui := mock.NewUI()

			realmClient := mock.RealmClient{}
			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				return []realm.App{app}, nil
			}

			realmClient.SecretsFn = func(groupID, appID string) ([]realm.Secret, error) {
				return tc.secrets, nil
			}

			cmd := &CommandList{listInputs{cli.ProjectInputs{
				Project: projectID,
				App:     appID,
			}}}

			assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
			assert.Equal(t, tc.expectedOutput, out.String())
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
				description: "when finding the secrets fails",
				setupClient: func() realm.Client {
					realmClient := mock.RealmClient{}
					realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
						return []realm.App{app}, nil
					}
					realmClient.SecretsFn = func(groupID, appID string) ([]realm.Secret, error) {
						return nil, errors.New("something bad happened")
					}
					return realmClient
				},
				expectedErr: errors.New("something bad happened"),
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				realmClient := tc.setupClient()

				cmd := &CommandList{}

				err := cmd.Handler(nil, nil, cli.Clients{Realm: realmClient})
				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})
}
