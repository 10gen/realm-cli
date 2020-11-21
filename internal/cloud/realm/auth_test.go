package realm_test

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"

	"github.com/google/go-cmp/cmp"
)

func TestClientAuthenticate(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	client := realm.NewClient(u.RealmServerURL())

	t.Run("Should fail with invalid credentials", func(t *testing.T) {
		_, err := client.Authenticate("username", "apiKey")

		u.MustNotBeNil(t, err)
		u.MustMatch(t, cmp.Diff(
			"failed to authenticate with MongoDB Cloud API: You are not authorized for this resource.",
			err.Error(),
		))
	})

	t.Run("Should return session details with valid credentials", func(t *testing.T) {
		auth, err := client.Authenticate(u.CloudAdminUsername(), u.CloudAdminAPIKey())
		u.MustMatch(t, cmp.Diff(nil, err))
		u.MustNotBeBlank(t, auth.AccessToken)
		u.MustNotBeBlank(t, auth.RefreshToken)
	})
}
