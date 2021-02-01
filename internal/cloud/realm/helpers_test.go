package realm_test

import (
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func setupTestApp(t *testing.T, client realm.Client, groupID, name string) (realm.App, func()) {
	t.Helper()
	app, err := client.CreateApp(groupID, name, realm.AppMeta{})
	assert.Nil(t, err)
	teardown := func() {
		deleteErr := client.DeleteApp(groupID, app.ID)
		if deleteErr != nil {
			t.Log(fmt.Sprintf("failed to delete test app with groupID: %v and ID: %v", groupID, app.ID))
		}
	}
	return app, teardown
}
