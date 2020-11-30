package realm_test

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestClientStatusSuccess(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)
	client := realm.NewClient(u.RealmServerURL())

	t.Run("Should return no error if the server is running", func(t *testing.T) {
		err := client.Status()
		assert.Nil(t, err)
	})
}

func TestClientStatusFailure(t *testing.T) {
	baseURL := "http://localhost:8081"
	client := realm.NewClient(baseURL)

	t.Run("Should return an error if the server is not running", func(t *testing.T) {
		err := client.Status()
		assert.Equal(t, realm.ErrServerNotRunning(baseURL), err)
	})
}
