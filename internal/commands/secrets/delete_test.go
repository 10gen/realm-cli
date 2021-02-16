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

func TestSecretHandler(t *testing.T) {
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
	for _, tc := range []struct {
		description    string
		testInput      []string
		expectedOutput []secretOutput
		expectedErr    error
	}{
		{
			description:    "should return successful outputs for proper secret inputs",
			testInput:      []string{"secret_id_1"},
			expectedOutput: []secretOutput{{secrets[1], nil}},
		},
		{
			description: "should still output the errors for deletes on individual secrets",
			testInput:   []string{"secret_id_0", "secret_id_6"},
			expectedOutput: []secretOutput{
				{
					secrets[0],
					errors.New("something happened"),
				},
				{
					secrets[6],
					errors.New("something happened"),
				},
			},
			expectedErr: errors.New("something happened"),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
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
				return tc.expectedErr
			}

			cmd := &CommandDelete{
				inputs: deleteInputs{
					ProjectInputs: cli.ProjectInputs{
						Project: projectID,
						App:     appID,
					},
					secrets: tc.testInput,
				},
				realmClient: realmClient,
			}

			assert.Nil(t, cmd.Handler(nil, nil))

			for i, actual := range cmd.outputs {
				assert.Equal(t, actual.secret, tc.expectedOutput[i].secret)
				assert.Equal(t, actual.err, tc.expectedOutput[i].err)
			}

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
			clientSetup func() realm.Client
			expectedErr error
		}{
			{
				description: "if there is an issue with finding secrets",
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
		} {
			t.Run(tc.description, func(t *testing.T) {
				realmClient := tc.clientSetup()
				cmd := &CommandDelete{realmClient: realmClient}
				err := cmd.Handler(nil, nil)
				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})
}

func TestSecretFeedback(t *testing.T) {
	testLen := 7
	secrets := make([]realm.Secret, testLen)
	for i := 0; i < testLen; i++ {
		secrets[i] = realm.Secret{
			ID:   fmt.Sprintf("secret_id_%d", i),
			Name: fmt.Sprintf("secret_name_%d", i),
		}
	}

	for _, tc := range []struct {
		description    string
		outputs        secretOutputs
		expectedOutput string
	}{
		{
			description:    "should show no secrets to delete",
			expectedOutput: "01:23:45 UTC INFO  No secrets to delete\n",
		},
		{
			description: "should show a successfully deleted secret",
			outputs:     secretOutputs{{secrets[1], nil}},
			expectedOutput: strings.Join([]string{
				"01:23:45 UTC INFO  Deleted Secrets",
				"  ID           Name           Deleted  Details",
				"  -----------  -------------  -------  -------",
				"  secret_id_1  secret_name_1  true            \n",
			}, "\n"),
		},
		{
			description: "should show many successfully deleted secrets",
			outputs:     secretOutputs{{secrets[0], nil}, {secrets[5], nil}},
			expectedOutput: strings.Join([]string{
				"01:23:45 UTC INFO  Deleted Secrets",
				"  ID           Name           Deleted  Details",
				"  -----------  -------------  -------  -------",
				"  secret_id_0  secret_name_0  true            ",
				"  secret_id_5  secret_name_5  true            \n",
			}, "\n"),
		},
		{
			description: "should show an unsuccessfully deleted secret",
			outputs:     secretOutputs{{secrets[3], errors.New("something happened")}},
			expectedOutput: strings.Join([]string{
				"01:23:45 UTC INFO  Deleted Secrets",
				"  ID           Name           Deleted  Details           ",
				"  -----------  -------------  -------  ------------------",
				"  secret_id_3  secret_name_3  false    something happened\n",
			}, "\n"),
		},
		{
			description: "should show many unsuccessfully deleted secrets",
			outputs:     secretOutputs{{secrets[4], errors.New("something happened")}, {secrets[5], errors.New("something else happened")}},
			expectedOutput: strings.Join([]string{
				"01:23:45 UTC INFO  Deleted Secrets",
				"  ID           Name           Deleted  Details                ",
				"  -----------  -------------  -------  -----------------------",
				"  secret_id_4  secret_name_4  false    something happened     ",
				"  secret_id_5  secret_name_5  false    something else happened\n",
			}, "\n"),
		},
		{
			description: "should show a mix of successfully and unsuccessfully deleted secrets",
			outputs: secretOutputs{
				{secrets[4], errors.New("something happened")},
				{secrets[5], nil},
				{secrets[0], errors.New("something else happened")},
			},
			expectedOutput: strings.Join([]string{
				"01:23:45 UTC INFO  Deleted Secrets",
				"  ID           Name           Deleted  Details                ",
				"  -----------  -------------  -------  -----------------------",
				"  secret_id_4  secret_name_4  false    something happened     ",
				"  secret_id_0  secret_name_0  false    something else happened",
				"  secret_id_5  secret_name_5  true                            \n",
			}, "\n"),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			out, ui := mock.NewUI()
			inputs := make([]string, len(tc.outputs))
			for i, o := range tc.outputs {
				inputs[i] = o.secret.ID
			}
			cmd := &CommandDelete{
				inputs: deleteInputs{
					secrets: inputs,
				},
				outputs: tc.outputs,
			}

			assert.Nil(t, cmd.Feedback(nil, ui))
			assert.Equal(t, tc.expectedOutput, out.String())
		})
	}
}

func TestSecretDeleteModifiers(t *testing.T) {
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
			secretDeleteRow(tc.outputInput, tc.rowInput)
			assert.Equal(t, tc.rowInput, tc.expectedOutput)
		})
	}
}
