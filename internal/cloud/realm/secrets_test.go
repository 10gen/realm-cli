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

		testApp, teardown := setupTestApp(t, client, groupID, "secrets-test")
		defer teardown()

		t.Run("should have no secrets upon app initialization", func(t *testing.T) {
			secrets, secretsErr := client.Secrets(groupID, testApp.ID)
			assert.Nil(t, secretsErr)
			assert.Equal(t, 0, len(secrets))
		})
		var secret realm.Secret
		t.Run("should create a secret", func(t *testing.T) {
			secretName := "secretName"
			secretValue := "secretValue"
			var secretErr error
			secret, secretErr = client.CreateSecret(groupID, testApp.ID, secretName, secretValue)
			assert.Nil(t, secretErr)

			t.Run("and list all app secrets", func(t *testing.T) {
				secrets, err := client.Secrets(groupID, testApp.ID)
				assert.Nil(t, err)
				assert.Equal(t, []realm.Secret{secret}, secrets)
			})
		})

		t.Run("should delete the app secret", func(t *testing.T) {
			err := client.DeleteSecret(groupID, testApp.ID, secret.ID)
			assert.Nil(t, err)
			t.Run("and list no more app secrets", func(t *testing.T) {
				secrets, err := client.Secrets(groupID, testApp.ID)
				assert.Nil(t, err)
				assert.Equal(t, []realm.Secret{}, secrets)
			})
			t.Run("and return an error if we can't find the secret", func(t *testing.T) {
				err := client.DeleteSecret(groupID, testApp.ID, secret.ID)
				assert.NotNil(t, err)
				assert.Equal(t, err.Error(), "secret not found: 'should not exist'")
			})
		})
	})
}
