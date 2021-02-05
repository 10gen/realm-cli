package secrets

import (
	"fmt"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
	"testing"
)

func TestResolveDeleteInputs(t *testing.T) {
	secrets := make([]realm.Secret, 4)
	for i := 0; i < 3; i++ {
		secrets[i] = realm.Secret{
			ID:   fmt.Sprintf("secret_id_%d", i),
			Name: fmt.Sprintf("secret_name_%d", i),
		}
	}

	t.Run("should prompt for secrets with no input", func(t *testing.T) {
		inputs := deleteInputs{}

		_, console, _, ui, consoleErr := mock.NewVT10XConsole()
		assert.Nil(t, consoleErr)
		defer console.Close()

		selectedSecrets := []string{}
		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)
			console.ExpectString("Which secret(s) would you like to delete?")
			console.Send("secret_id_1")
			console.SendLine(" ")
			console.ExpectEOF()
		}()

		expectedSecrets := []realm.Secret{secrets[1]}

		secretsResult, err := inputs.resolveDelete(selectedSecrets, secrets, ui)

		assert.Nil(t, err)
		assert.Equal(t, expectedSecrets, secretsResult)
	})

	t.Run("should not prompt for secrets with an input and return appropriate secrets", func(t *testing.T) {

	})

}
