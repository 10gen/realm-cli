package realm_test

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestRealmDeployments(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	t.Run("Should fail without an auth client", func(t *testing.T) {
		client := realm.NewClient(u.RealmServerURL())

		_, err := client.Deployment(primitive.NewObjectID().Hex(), primitive.NewObjectID().Hex(), primitive.NewObjectID().Hex())
		assert.Equal(t, realm.ErrInvalidSession{}, err)
	})

	t.Run("With an active session", func(t *testing.T) {
		client := newAuthClient(t)
		groupID := u.CloudGroupID()

		testApp, appErr := client.CreateApp(groupID, "users-test", realm.AppMeta{})
		assert.Nil(t, appErr)

		draft, draftErr := client.CreateDraft(groupID, testApp.ID)
		assert.Nil(t, draftErr)

		t.Run("Should be able to deploy an existing draft", func(t *testing.T) {
			deployment, deploymentErr := client.DeployDraft(groupID, testApp.ID, draft.ID)
			assert.Nil(t, deploymentErr)

			assert.True(t, deployment.ID != "", "deployment id should not be empty")
			assert.Equal(t, realm.DeploymentStatusCreated, deployment.Status)

			t.Run("And be able to retrieve the deployment", func(t *testing.T) {
				found, err := client.Deployment(groupID, testApp.ID, deployment.ID)
				assert.Nil(t, err)
				assert.Equal(t, deployment.ID, found.ID)
			})
		})
	})
}
