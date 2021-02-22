package realm_test

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestRealmCommands(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	t.Run("should fail without an auth client", func(t *testing.T) {
		client := realm.NewClient(u.RealmServerURL())

		_, err := client.ListClusters(primitive.NewObjectID().Hex(), primitive.NewObjectID().Hex())
		assert.Equal(t, realm.ErrInvalidSession{}, err)
	})

	t.Run("with an active session", func(t *testing.T) {
		client := newAuthClient(t)
		groupID := u.CloudGroupID()

		t.Run("should list clusters", func(t *testing.T) {
			app, err := client.CreateApp(groupID, "commands-test", realm.AppMeta{})
			assert.Nil(t, err)

			_, err = client.ListClusters(groupID, app.ID)
			assert.Nil(t, err)
		})
	})
}
