package realm_test

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestRealmAuth(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	client := realm.NewClient(u.RealmServerURL())

	t.Run("Should fail with invalid credentials", func(t *testing.T) {
		_, err := client.Authenticate("username", "apiKey")
		assert.Equal(t,
			realm.ServerError{Message: "failed to authenticate with MongoDB Cloud API: You are not authorized for this resource."},
			err,
		)
	})

	t.Run("Should return session details with valid credentials", func(t *testing.T) {
		auth, err := client.Authenticate(u.CloudUsername(), u.CloudAPIKey())
		assert.Nil(t, err)
		assert.NotEqual(t, "", auth.AccessToken, "access token must not be blank")
		assert.NotEqual(t, "", auth.RefreshToken, "refresh token must not be blank")
	})
}
