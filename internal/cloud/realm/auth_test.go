package realm_test

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestRealmAuthenticateCloud(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	client := realm.NewClient(u.RealmServerURL())

	t.Run("Should fail with invalid cloud credentials", func(t *testing.T) {
		_, err := client.Authenticate(realm.AuthTypeCloud, user.Credentials{PublicAPIKey: "username", PrivateAPIKey: "apiKey"})
		assert.Equal(t,
			realm.ServerError{Message: "failed to authenticate with MongoDB Cloud API: You are not authorized for this resource."},
			err,
		)
	})

	t.Run("Should return session details with valid cloud credentials", func(t *testing.T) {
		session, err := client.Authenticate(realm.AuthTypeCloud, user.Credentials{
			PublicAPIKey:  u.CloudUsername(),
			PrivateAPIKey: u.CloudAPIKey(),
		})
		assert.Nil(t, err)
		assert.NotEqual(t, "", session.AccessToken, "access token must not be blank")
		assert.NotEqual(t, "", session.RefreshToken, "refresh token must not be blank")
	})
}

func TestRealmAuthenticateLocal(t *testing.T) {
	t.Skip("this test needs to run with a local Realm server configured to support the local-userpass auth provider")

	client := realm.NewClient(u.RealmServerURL())

	t.Run("Should return session details with valid local credentials", func(t *testing.T) {
		session, err := client.Authenticate(realm.AuthTypeLocal, user.Credentials{
			Username: "unique_user@domain.com",
			Password: "password",
		})
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
		assert.Equal(t, realm.ErrInvalidSession(user.DefaultProfile), err)
	})

	t.Run("With an active session should return session details with valid credentials", func(t *testing.T) {
		client := newAuthClient(t)

		profile, err := client.AuthProfile()
		assert.Nil(t, err)
		assert.NotEqualf(t, 0, len(profile.Roles), "expected profile to have role(s)")

		groupIDs := profile.AllGroupIDs()
		assert.Equal(t, u.CloudGroupCount(), len(groupIDs))

		groupIDSet := map[string]bool{}
		for _, groupID := range groupIDs {
			groupIDSet[groupID] = true
		}
		assert.True(t, groupIDSet[u.CloudGroupID()], "expected groupID, %s, to be returned in auth profile", u.CloudGroupID())
	})
}

func TestRealmAuthRefresh(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	t.Run("Does not refresh auth if request does not return invalid session code", func(t *testing.T) {
		client := realm.NewClient(u.RealmServerURL())

		session, err := client.Authenticate(realm.AuthTypeCloud, user.Credentials{
			PublicAPIKey:  u.CloudUsername(),
			PrivateAPIKey: u.CloudAPIKey(),
		})
		assert.Equal(t, nil, err)

		// invalidate the session's access token
		session.AccessToken = session.RefreshToken

		profile := mock.NewProfileWithSession(t, session)

		client = realm.NewAuthClient(profile.RealmBaseURL(), profile)
		_, err = client.AuthProfile()
		serverError, ok := err.(realm.ServerError)
		assert.True(t, ok, "expected %T to be server error", err)
		assert.Equal(t, realm.ServerError{Message: "invalid session: valid Issuer required"}, serverError)
	})

	t.Run("Should return the invalid session error when credentials are invalid", func(t *testing.T) {
		client := realm.NewClient(u.RealmServerURL())

		session, err := client.Authenticate(realm.AuthTypeCloud, user.Credentials{
			PublicAPIKey:  u.CloudUsername(),
			PrivateAPIKey: u.CloudAPIKey(),
		})
		assert.Equal(t, nil, err)

		// invalidate the session's tokens
		session.RefreshToken = session.AccessToken
		session.AccessToken = ""

		profile := mock.NewProfileWithSession(t, session)

		client = realm.NewAuthClient(profile.RealmBaseURL(), profile)
		_, err = client.AuthProfile()
		assert.Equal(t, realm.ErrInvalidSession(profile.Name), err)
	})

	t.Run("with an expired access token", func(t *testing.T) {
		u.SkipUnlessExpiredAccessTokenPresent(t)

		t.Run("should use the refresh token to generate a new access token", func(t *testing.T) {
			client := realm.NewClient(u.RealmServerURL())

			session, err := client.Authenticate(realm.AuthTypeCloud, user.Credentials{
				PublicAPIKey:  u.CloudUsername(),
				PrivateAPIKey: u.CloudAPIKey(),
			})
			assert.Nil(t, err)

			profile, teardown := mock.NewProfileFromTmpDir(t, "auth_refresh_test")
			defer teardown()
			profile.SetRealmBaseURL(u.RealmServerURL())
			// set the access token value to an expired token
			profile.SetSession(user.Session{u.ExpiredAccessToken(), session.RefreshToken})

			client = realm.NewAuthClient(profile.RealmBaseURL(), profile)
			_, err = client.AuthProfile()
			assert.Nil(t, err)

			updatedSession := profile.Session()

			t.Log("and update the access token")
			assert.NotEqual(t, u.ExpiredAccessToken(), updatedSession.AccessToken, "access token was not updated")
			assert.Equalf(t, session.RefreshToken, updatedSession.RefreshToken, "refresh token was incorrectly updated")
		})

		t.Run("should error if both the access and refresh tokens are expired", func(t *testing.T) {
			profile, teardown := mock.NewProfileFromTmpDir(t, "auth_refresh_test")
			defer teardown()
			profile.SetRealmBaseURL(u.RealmServerURL())
			// set the access and refresh token values to expired tokens
			profile.SetSession(user.Session{u.ExpiredAccessToken(), u.ExpiredAccessToken()})

			client := realm.NewAuthClient(profile.RealmBaseURL(), profile)
			_, err := client.AuthProfile()
			assert.Equal(t, realm.ErrInvalidSession(profile.Name), err)

			session := profile.Session()

			t.Log("and clear the session tokens")
			assert.Equalf(t, "", session.AccessToken, "access token was not cleared")
			assert.Equalf(t, "", session.RefreshToken, "refresh token was not cleared")
		})
	})
}

func newAuthClient(t *testing.T) realm.Client {
	t.Helper()

	client := realm.NewClient(u.RealmServerURL())

	session, err := client.Authenticate(realm.AuthTypeCloud, user.Credentials{
		PublicAPIKey:  u.CloudUsername(),
		PrivateAPIKey: u.CloudAPIKey(),
	})
	assert.Nil(t, err)

	profile := mock.NewProfileWithSession(t, session)

	return realm.NewAuthClient(profile.RealmBaseURL(), profile)
}
