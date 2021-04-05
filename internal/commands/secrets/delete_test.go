package secrets

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"github.com/10gen/realm-cli/internal/cloud/realm"
)

func TestSecretsDeleteHandler(t *testing.T) {
	projectID := "projectID"
	appID := "appID"
	app := realm.App{
		ID:          appID,
		GroupID:     projectID,
		ClientAppID: "eggcorn-abcde",
		Name:        "eggcorn",
	}

	testLen := 7
	secrets := make([]realm.Secret, testLen)
	for i := 0; i < testLen; i++ {
		secrets[i] = realm.Secret{
			ID:   fmt.Sprintf("secret_id_%d", i),
			Name: fmt.Sprintf("secret_name_%d", i),
		}
	}

	t.Run("should show empty state message if no secrets are found", func(t *testing.T) {
		out, ui := mock.NewUI()

		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{app}, nil
		}
		realmClient.SecretsFn = func(groupID, appID string) ([]realm.Secret, error) {
			return nil, nil
		}

		cmd := &CommandDelete{deleteInputs{
			ProjectInputs: cli.ProjectInputs{
				Project: projectID,
				App:     appID,
			},
		}}

		assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))

		assert.Equal(t, "No secrets to delete\n", out.String())
	})

	for _, tc := range []struct {
		description    string
		testInput      []string
		expectedOutput string
		deleteErr      error
	}{
		{
			description: "should return successful outputs for proper secret inputs",
			testInput:   []string{"secret_id_1"},
			expectedOutput: strings.Join([]string{
				"Deleted 1 secret(s)",
				"  ID           Name           Deleted  Details",
				"  -----------  -------------  -------  -------",
				"  secret_id_1  secret_name_1  true            ",
				"",
			}, "\n"),
		},
		{
			description: "should still output the errors for deletes on individual secrets",
			testInput:   []string{"secret_id_0", "secret_id_6"},
			deleteErr:   errors.New("something happened"),
			expectedOutput: strings.Join([]string{
				"Deleted 2 secret(s)",
				"  ID           Name           Deleted  Details           ",
				"  -----------  -------------  -------  ------------------",
				"  secret_id_0  secret_name_0  false    something happened",
				"  secret_id_6  secret_name_6  false    something happened",
				"",
			}, "\n"),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			out, ui := mock.NewUI()

			realmClient := mock.RealmClient{}

			var capturedAppFilter realm.AppFilter
			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				capturedAppFilter = filter
				return []realm.App{app}, nil
			}

			var capturedFindProjectID, capturedFindAppID string
			realmClient.SecretsFn = func(groupID, appID string) ([]realm.Secret, error) {
				capturedFindProjectID = groupID
				capturedFindAppID = appID
				return secrets, nil
			}

			var capturedDeleteProjectID, capturedDeleteAppID string

			realmClient.DeleteSecretFn = func(groupID, appID, secretId string) error {
				capturedDeleteProjectID = groupID
				capturedDeleteAppID = appID
				return tc.deleteErr
			}

			cmd := &CommandDelete{deleteInputs{
				ProjectInputs: cli.ProjectInputs{
					Project: projectID,
					App:     appID,
				},
				secrets: tc.testInput,
			}}

			assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))

			assert.Equal(t, tc.expectedOutput, out.String())

			assert.Equal(t, realm.AppFilter{App: appID, GroupID: projectID}, capturedAppFilter)
			assert.Equal(t, projectID, capturedFindProjectID)
			assert.Equal(t, appID, capturedFindAppID)
			assert.Equal(t, projectID, capturedDeleteProjectID)
			assert.Equal(t, appID, capturedDeleteAppID)
		})
	}

	t.Run("should return an error", func(t *testing.T) {
		for _, tc := range []struct {
			description string
			realmClient func() realm.Client
			expectedErr error
		}{
			{
				description: "if there is an issue with finding secrets",
				realmClient: func() realm.Client {
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
				description: "if there is no app",
				realmClient: func() realm.Client {
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
		} {
			t.Run(tc.description, func(t *testing.T) {
				cmd := &CommandDelete{}
				err := cmd.Handler(nil, nil, cli.Clients{Realm: tc.realmClient()})
				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})
}

func TestSecretsDeleteModifiers(t *testing.T) {
	for _, tc := range []struct {
		description    string
		outputInput    secretOutput
		rowInput       map[string]interface{}
		expectedOutput map[string]interface{}
	}{
		{
			description: "should set the details to nothing and deleted to true if successful",
			outputInput: secretOutput{
				realm.Secret{ID: "id1", Name: "name1"},
				nil,
			},
			rowInput: map[string]interface{}{
				headerID:   "id1",
				headerName: "name1",
			},
			expectedOutput: map[string]interface{}{
				headerID:      "id1",
				headerName:    "name1",
				headerDeleted: true,
			},
		},
		{
			description: "should set the details and deleted to false if not successful",
			outputInput: secretOutput{
				realm.Secret{ID: "id1", Name: "name1"},
				errors.New("new error"),
			},
			rowInput: map[string]interface{}{
				headerID:   "id1",
				headerName: "name1",
			},
			expectedOutput: map[string]interface{}{
				headerID:      "id1",
				headerName:    "name1",
				headerDetails: "new error",
				headerDeleted: false,
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			tableRowDelete(tc.outputInput, tc.rowInput)
			assert.Equal(t, tc.rowInput, tc.expectedOutput)
		})
	}
}
