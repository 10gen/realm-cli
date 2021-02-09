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

		app, appErr := client.CreateApp(groupID, "users-test", realm.AppMeta{})
		assert.Nil(t, appErr)

		draft, draftErr := client.CreateDraft(groupID, app.ID)
		assert.Nil(t, draftErr)

		t.Run("Should initially find no deployments", func(t *testing.T) {
			deployments, err := client.Deployments(groupID, app.ID)
			assert.Nil(t, err)
			assert.Equal(t, 0, len(deployments))
		})

		t.Run("Should be able to deploy an existing draft", func(t *testing.T) {
			deployment, deploymentErr := client.DeployDraft(groupID, app.ID, draft.ID)
			assert.Nil(t, deploymentErr)

			assert.True(t, deployment.ID != "", "deployment id should not be empty")
			assert.Equal(t, realm.DeploymentStatusCreated, deployment.Status)

			t.Run("And be able to retrieve the deployment", func(t *testing.T) {
				found, err := client.Deployment(groupID, app.ID, deployment.ID)
				assert.Nil(t, err)
				assert.Equal(t, deployment.ID, found.ID)

				all, err := client.Deployments(groupID, app.ID)
				assert.Nil(t, err)
				assert.Equal(t, []realm.AppDeployment{found}, all)
			})
		})
	})
}
