package realm_test

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"

	"github.com/google/go-cmp/cmp"
)

func TestClientStatusSuccess(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)
	client := realm.NewClient(u.RealmServerURL())

	t.Run("Should return no error if the server is running", func(t *testing.T) {
		err := client.Status()
		u.MustMatch(t, cmp.Diff(nil, err))
	})
}

func TestClientStatusFailure(t *testing.T) {
	baseURL := "http://localhost:8081"
	client := realm.NewClient(baseURL)

	t.Run("Should return an error if the server is not running", func(t *testing.T) {
		err := client.Status()
		u.MustNotBeNil(t, err)
		u.MustMatch(t, cmp.Diff(realm.ErrServerNotRunning(baseURL).Error(), err.Error()))
	})
}
