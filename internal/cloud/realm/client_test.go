package realm_test

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestRealmClient(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	t.Run("Should work correctly execute request with valid credentials", func(t *testing.T) {
		client := newAuthClient(t)
		authProfile, err := client.AuthProfile()
		assert.Equal(t, nil, err)
		assert.NotEqualf(t, realm.AuthProfile{}, authProfile, "expected auth profile to not be empty")

	})

	t.Run("Should return the correct error when credentials are invalid", func(t *testing.T) {
		client := realm.NewClient(u.RealmServerURL())

		session, err := client.Authenticate(u.CloudUsername(), u.CloudAPIKey())
		assert.Equal(t, nil, err)

		session.AccessToken = "ssgsfg"

		client = realm.NewAuthClient(u.RealmServerURL(), session)
		_, err = client.AuthProfile()

		serverError := err.(realm.ServerError)
		assert.Equal(t, "401 Unauthorized", serverError.Message)

	})

	// TODO: REALMC-7719 add test for expired credentials
}
