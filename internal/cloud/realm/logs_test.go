package realm_test

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestRealmLogs(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	t.Run("with an active session", func(t *testing.T) {
		client := newAuthClient(t)
		groupID := u.CloudGroupID()

		app, teardown := setupTestApp(t, client, groupID, "logs-test")
		defer teardown()

		t.Run("getting logs should return an empty list if there are none", func(t *testing.T) {
			logs, err := client.Logs(groupID, app.ID, realm.LogsOptions{})
			assert.Nil(t, err)
			assert.Equal(t, 0, len(logs))
		})
	})
}
