package realm_test

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestRealmSchemaModels(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	t.Run("should respond with 401 when not authenticated", func(t *testing.T) {
		client := realm.NewClient(u.RealmServerURL())

		_, err := client.SchemaModels("", "", realm.DataModelLanguageJava)
		assert.Equal(t, realm.ErrInvalidSession(user.DefaultProfile), err)
	})

	t.Run("with an active session", func(t *testing.T) {
		client := newAuthClient(t)
		groupID := u.CloudGroupID()

		app, teardown := setupTestApp(t, client, groupID, "schema-test")
		defer teardown()

		t.Run("should respond with an empty list of data models with no app schema", func(t *testing.T) {
			models, err := client.SchemaModels(groupID, app.ID, realm.DataModelLanguageTypescript)
			assert.Nil(t, err)
			assert.Equal(t, 0, len(models))
		})

		// TODO(REALMC-7175): once `schema generate` is a supported command
		// write mkore of this test to actually retrieve data models generated from a schema
	})
}
