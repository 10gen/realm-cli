package realm_test

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TODO(REALMC-9207): Add tests once backend is fully supported
func TestRealmIPAccess(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	t.Run("should fail without an auth client", func(t *testing.T) {
		client := realm.NewClient(u.RealmServerURL() + "")

		_, err := client.AllowedIPCreate(primitive.NewObjectID().Hex(), primitive.NewObjectID().Hex(), "0.0.0.0/0", "comment", false)
		assert.Equal(t, realm.ErrInvalidSession{}, err)
	})
}
