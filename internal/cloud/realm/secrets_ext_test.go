package realm_test

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestSecrets(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	t.Run("should fail without an auth client", func(t *testing.T) {
		client := realm.NewClient(u.RealmServerURL() + "")

		_, err := client.Secrets(primitive.NewObjectID().Hex(), primitive.NewObjectID().Hex())
		assert.Equal(t, realm.ErrInvalidSession{}, err)
	})

	t.Run("with an active session ", func(t *testing.T) {
		client := newAuthClient(t)
		groupID := u.CloudGroupID()

		secretName := "secretName"
		secretValue := "secretValue"

		testApp, teardown := setupTestApp(t, client, groupID, "secrets-test")
		defer teardown()

		t.Run("should have no secrets upon app initialization", func(t *testing.T) {
			secrets, secretsErr := client.Secrets(groupID, testApp.ID)
			assert.Nil(t, secretsErr)
			assert.Equal(t, 0, len(secrets))
		})
		var capturedSecretID string
		t.Run("should create a secret", func(t *testing.T) {
			secret, secretErr := client.CreateSecret(groupID, testApp.ID, secretName, secretValue)
			capturedSecretID = secret.ID
			assert.Nil(t, secretErr)
		})
		t.Run("should list a created secret", func(t *testing.T) {
			secrets, secretsErr := client.Secrets(groupID, testApp.ID)
			assert.Nil(t, secretsErr)
			assert.Equal(t, []realm.Secret{{capturedSecretID, secretName}}, secrets)
		})
	})
}
