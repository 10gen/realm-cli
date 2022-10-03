package realm_test

import (
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestRealmSecrets(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	t.Run("should fail without an auth client", func(t *testing.T) {
		client := realm.NewClient(u.RealmServerURL())

		_, err := client.Secrets(primitive.NewObjectID().Hex(), primitive.NewObjectID().Hex())
		assert.Equal(t, realm.ErrInvalidSession(user.DefaultProfile), err)
	})

	t.Run("with an active session ", func(t *testing.T) {
		client := newAuthClient(t)
		groupID := u.CloudGroupID()

		testApp, teardown := setupTestApp(t, client, groupID, "secrets-test")
		defer teardown()

		t.Run("should have no secrets upon app initialization", func(t *testing.T) {
			secrets, err := client.Secrets(groupID, testApp.ID)
			assert.Nil(t, err)
			assert.Equal(t, 0, len(secrets))
		})

		t.Run("should create a secret", func(t *testing.T) {
			secretName := "secretName"
			secretValue := "secretValue"

			secret, err := client.CreateSecret(groupID, testApp.ID, secretName, secretValue)
			assert.Nil(t, err)

			t.Run("and list all app secrets", func(t *testing.T) {
				secrets, err := client.Secrets(groupID, testApp.ID)
				assert.Nil(t, err)
				assert.Equal(t, []realm.Secret{secret}, secrets)
			})

			t.Run("and should update the app secret name", func(t *testing.T) {
				assert.Nil(t, client.UpdateSecret(groupID, testApp.ID, secret.ID, "newName", ""))

				t.Run("and list the new app secret value", func(t *testing.T) {
					secrets, err := client.Secrets(groupID, testApp.ID)
					assert.Nil(t, err)
					assert.Equal(t, "newName", secrets[0].Name)
				})

				t.Run("and return an error if we can't find the secret", func(t *testing.T) {
					err := client.UpdateSecret(groupID, testApp.ID, "notFound", "notUsed", "notUsed")
					assert.Equal(t, realm.ServerError{Message: `secret not found: "notFound"`}, err)
				})
			})

			t.Run("and should delete the app secret", func(t *testing.T) {
				assert.Nil(t, client.DeleteSecret(groupID, testApp.ID, secret.ID))

				t.Run("and list no more app secrets", func(t *testing.T) {
					secrets, err := client.Secrets(groupID, testApp.ID)
					assert.Nil(t, err)
					assert.Equal(t, []realm.Secret{}, secrets)
				})

				t.Run("and return an error if we can't find the secret", func(t *testing.T) {
					err := client.DeleteSecret(groupID, testApp.ID, secret.ID)
					assert.Equal(t, realm.ServerError{Message: fmt.Sprintf("secret not found: %q", secret.ID)}, err)
				})
			})
		})
	})
}
