package realm_test

import (
	"fmt"
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

		t.Run("should create a secret", func(t *testing.T) {
			secretName := "secretName"
			secretValue := "secretValue"

			secret, secretErr := client.CreateSecret(groupID, testApp.ID, secretName, secretValue)
			assert.Nil(t, secretErr)

			t.Run("and list all app secrets", func(t *testing.T) {
				secrets, err := client.Secrets(groupID, testApp.ID)
				assert.Nil(t, err)
				assert.Equal(t, []realm.Secret{secret}, secrets)
			})
		})

		t.Run("should create secrets for deletion", func(t *testing.T) {
			testLen := 3
			testSecrets := make([]realm.Secret, testLen)
			for i := 0; i < testLen; i++ {
				secret, err := client.CreateSecret(groupID, testApp.ID, fmt.Sprintf("deleteSecret%d", i), fmt.Sprintf("deleteName%d", i))
				assert.Nil(t, err)
				testSecrets[i] = secret
			}

			t.Run("and delete them successfully", func(t *testing.T) {
				for _, secret := range testSecrets {
					err := client.DeleteSecret(groupID, testApp.ID, secret.ID)
					assert.Nil(t, err)
				}
			})
		})
		t.Run("should return an error if we can't delete the secret", func(t *testing.T) {
			secretID := "should not exist"
			err := client.DeleteSecret(groupID, testApp.ID, secretID)
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), "secret not found: 'should not exist'")
		})
	})
}
