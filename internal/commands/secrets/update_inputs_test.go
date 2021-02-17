package secrets

import (
	"errors"
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestSecretInputResolve(t *testing.T) {
	testLen := 7
	secrets := make([]realm.Secret, testLen)

	for i := 0; i < testLen; i++ {
		secrets[i] = realm.Secret{
			fmt.Sprintf("secretID%d", i),
			fmt.Sprintf("secretName%d", i),
		}
	}

	for _, tc := range []struct {
		description    string
		testInput      string
		expectedOutput realm.Secret
	}{
		{
			description:    "should return the secret if searching by name",
			testInput:      "secretName4",
			expectedOutput: secrets[4],
		},
		{
			description:    "should return the secret if searching by ID",
			testInput:      "secretID3",
			expectedOutput: secrets[3],
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			inputs := updateInputs{
				secret: tc.testInput,
			}
			result, err := inputs.resolveSecret(nil, secrets)
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedOutput, result)
		})
	}

	t.Run("should return an error if we cannot find the secret to update", func(t *testing.T) {
		inputs := updateInputs{
			secret: "cannotFind",
		}

		_, err := inputs.resolveSecret(nil, secrets)
		assert.Equal(t, errors.New("unable to find secret: cannotFind"), err)
	})

	t.Run("should show the prompt for a user if the input is empty", func(t *testing.T) {
		inputs := updateInputs{}

		_, console, _, ui, consoleErr := mock.NewVT10XConsole()
		assert.Nil(t, consoleErr)
		defer console.Close()

		doneCh := make(chan struct{})
		go func() {
			defer close(doneCh)
			console.ExpectString("Which secret would you like to update?")
			console.SendLine("secretID2")
			console.ExpectEOF()
		}()

		secretsResult, err := inputs.resolveSecret(ui, secrets)

		console.Tty().Close()
		<-doneCh

		assert.Nil(t, err)
		assert.Equal(t, secrets[2], secretsResult)
	})
}
