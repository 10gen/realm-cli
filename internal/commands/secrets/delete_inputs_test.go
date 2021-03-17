package secrets

import (
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestSecretsDeleteInputsResolve(t *testing.T) {
	testLen := 7
	secrets := make([]realm.Secret, testLen)
	for i := 0; i < testLen; i++ {
		secrets[i] = realm.Secret{
			ID:   fmt.Sprintf("secret_id_%d", i),
			Name: fmt.Sprintf("secret_name_%d", i),
		}
	}

	for _, tc := range []struct {
		description     string
		selectedSecrets []string
		expectedOutput  []realm.Secret
	}{
		{
			description:     "select by ID",
			selectedSecrets: []string{"secret_id_0"},
			expectedOutput:  []realm.Secret{secrets[0]},
		},
		{
			description:     "select by name",
			selectedSecrets: []string{"secret_name_2"},
			expectedOutput:  []realm.Secret{secrets[2]},
		},
		{
			description:     "select by name and ID",
			selectedSecrets: []string{"secret_name_2", "secret_id_4", "secret_name_1"},
			expectedOutput:  []realm.Secret{secrets[2], secrets[4], secrets[1]},
		},
	} {
		t.Run("should return the secrets when specified by "+tc.description, func(t *testing.T) {
			inputs := deleteInputs{
				secrets: tc.selectedSecrets,
			}

			secretsResult, err := inputs.resolveSecrets(nil, secrets)

			assert.Nil(t, err)
			assert.Equal(t, tc.expectedOutput, secretsResult)
		})
	}

	for _, tc := range []struct {
		description     string
		selectedSecrets []string
		expectedOutput  []realm.Secret
	}{
		{
			selectedSecrets: []string{"secret_id_0"},
			expectedOutput:  []realm.Secret{secrets[0]},
		},
		{
			description:     "allow multiple selections",
			selectedSecrets: []string{"secret_id_6", "secret_id_1", "secret_id_2"},
			expectedOutput:  []realm.Secret{secrets[1], secrets[2], secrets[6]},
		},
	} {
		t.Run("should prompt for secrets with no input: "+tc.description, func(t *testing.T) {
			inputs := deleteInputs{}

			_, console, _, ui, consoleErr := mock.NewVT10XConsole()
			assert.Nil(t, consoleErr)
			defer console.Close()

			doneCh := make(chan struct{})
			go func() {
				defer close(doneCh)
				console.ExpectString("Which secret(s) would you like to delete?")
				for _, selected := range tc.selectedSecrets {
					console.Send(selected)
					console.Send(" ")
				}
				console.SendLine("")
				console.ExpectEOF()
			}()

			secretsResult, err := inputs.resolveSecrets(ui, secrets)

			console.Tty().Close()
			<-doneCh

			assert.Nil(t, err)
			assert.Equal(t, tc.expectedOutput, secretsResult)
		})
	}
}
