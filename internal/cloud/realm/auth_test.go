package realm_test

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestRealmAuthenticate(t *testing.T) {
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
		session, err := client.Authenticate(u.CloudUsername(), u.CloudAPIKey())
		assert.Nil(t, err)
		assert.NotEqual(t, "", session.AccessToken, "access token must not be blank")
		assert.NotEqual(t, "", session.RefreshToken, "refresh token must not be blank")
	})
}

func TestRealmAuthProfile(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	t.Run("Should fail without an auth client", func(t *testing.T) {
		client := realm.NewClient(u.RealmServerURL())

		_, err := client.AuthProfile()
		assert.Equal(t, realm.ErrInvalidSession, err)
	})

	t.Run("With an active session should return session details with valid credentials", func(t *testing.T) {
		client := newAuthClient(t)

		profile, err := client.AuthProfile()
		assert.Nil(t, err)
		assert.NotEqualf(t, 0, len(profile.Roles), "expected profile to have role(s)")
		assert.Equal(t, []string{u.CloudGroupID()}, profile.AllGroupIDs())
	})
}

func newAuthClient(t *testing.T) realm.Client {
	t.Helper()

	client := realm.NewClient(u.RealmServerURL())

	session, err := client.Authenticate(u.CloudUsername(), u.CloudAPIKey())
	assert.Nil(t, err)

	return realm.NewAuthClient(u.RealmServerURL(), session)
}
