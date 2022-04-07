package secrets

import (
	"errors"
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestSecretUpdateHandler(t *testing.T) {
	projectID := "projectID"
	appID := "appID"
	app := realm.App{
		ID:          appID,
		GroupID:     projectID,
		ClientAppID: "test-abcde",
		Name:        "test",
	}

	testLen := 7
	secrets := make([]realm.Secret, testLen)
	for i := 0; i < testLen; i++ {
		secrets[i] = realm.Secret{
			ID:   fmt.Sprintf("secretID%d", i),
			Name: fmt.Sprintf("secretName%d", i),
		}
	}

	for _, tc := range []struct {
		description string
		testSecret  string
		testName    string
		testValue   string
	}{
		{
			description: "should return a successful message for a successful update for name and value",
			testSecret:  "secretID5",
			testName:    "newName5",
			testValue:   "newValue5",
		},
		{
			description: "should return a successful message for an update with only a name",
			testSecret:  "secretName3",
			testName:    "newName3",
		},
		{
			description: "should return a successful message for an update with only a value",
			testSecret:  "secretID2",
			testValue:   "newValue2",
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			realmClient := mock.RealmClient{}

			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				return []realm.App{app}, nil
			}
			realmClient.SecretsFn = func(groupID, appID string) ([]realm.Secret, error) {
				return secrets, nil
			}
			realmClient.UpdateSecretFn = func(groupID, appID, secretId, name, value string) error {
				return nil
			}

			cmd := &CommandUpdate{updateInputs{
				cli.ProjectInputs{Project: projectID, App: appID},
				tc.testSecret,
				tc.testName,
				tc.testValue,
			}}

			out, ui := mock.NewUI()

			assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))

			assert.Equal(t, "Successfully updated secret\n", out.String())
		})
	}

	t.Run("should return an error", func(t *testing.T) {
		for _, tc := range []struct {
			description string
			inputs      updateInputs
			clientSetup func() realm.Client
			expectedErr error
		}{
			{
				description: "if there is no app",
				clientSetup: func() realm.Client {
					return mock.RealmClient{
						FindAppsFn: func(filter realm.AppFilter) ([]realm.App, error) {
							return nil, errors.New("Something went wrong with the app")
						},
						SecretsFn: func(groupID, appID string) ([]realm.Secret, error) {
							return secrets, nil
						},
					}
				},
				expectedErr: errors.New("Something went wrong with the app"),
			},
			{
				description: "if there is an issue with finding secrets for the app",
				clientSetup: func() realm.Client {
					return mock.RealmClient{
						FindAppsFn: func(filter realm.AppFilter) ([]realm.App, error) {
							return []realm.App{app}, nil
						},
						SecretsFn: func(groupID, appID string) ([]realm.Secret, error) {
							return nil, errors.New("Something happened with secrets")
						},
					}
				},
				expectedErr: errors.New("Something happened with secrets"),
			},
			{
				description: "if there is an issue with finding the secret specified in the list of app secrets",
				inputs:      updateInputs{secret: "illegal"},
				clientSetup: func() realm.Client {
					return mock.RealmClient{
						FindAppsFn: func(filter realm.AppFilter) ([]realm.App, error) {
							return []realm.App{app}, nil
						},
						SecretsFn: func(groupID, appID string) ([]realm.Secret, error) {
							return secrets, nil
						},
					}
				},
				expectedErr: errors.New("unable to find secret: illegal"),
			},
			{
				description: "if there is an issue with updating the secret",
				inputs:      updateInputs{secret: secrets[0].Name},
				clientSetup: func() realm.Client {
					return mock.RealmClient{
						FindAppsFn: func(filter realm.AppFilter) ([]realm.App, error) {
							return []realm.App{app}, nil
						},
						SecretsFn: func(groupID, appID string) ([]realm.Secret, error) {
							return secrets, nil
						},
						UpdateSecretFn: func(groupID, appID, secretID, name, value string) error {
							return errors.New("something bad happened")
						},
					}
				},
				expectedErr: errors.New("something bad happened"),
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				_, ui := mock.NewUI()

				realmClient := tc.clientSetup()
				cmd := &CommandUpdate{tc.inputs}
				assert.Equal(t, tc.expectedErr, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
			})
		}
	})
}
