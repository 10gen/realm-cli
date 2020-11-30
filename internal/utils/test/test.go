package testutils

import (
	"os"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
)

// MustSkipf skips a test suite, but panics if REALM_NO_SKIP_TEST is set
func MustSkipf(t *testing.T, format string, args ...interface{}) {
	if len(os.Getenv("REALM_NO_SKIP_TEST")) > 0 {
		panic("test was skipped, but REALM_NO_SKIP_TEST is set")
	}
	t.Skipf(format, args...)
}

const (
	defaultServerURL = "http://localhost:9090"
)

var realmServerRunning = false
var realmServerNotRunning = false
var skipUnlessRealmServerCalled = false
var atlasServerRunning = false
var atlasServerNotRunning = false
var skipUnlessAtlasServerCalled = false

// CloudAdminGroupID returns the Cloud admin group id to use for testing
func CloudAdminGroupID() string {
	return os.Getenv("REALM_MONGODB_CLOUD_GROUP_ID")
}

// CloudAdminUsername returns the Cloud admin username to use for testing
func CloudAdminUsername() string {
	return os.Getenv("REALM_MONGODB_CLOUD_ADMIN_USERNAME")
}

// CloudAdminAPIKey returns the Cloud admin api key to use for testing
func CloudAdminAPIKey() string {
	return os.Getenv("REALM_MONGODB_CLOUD_ADMIN_API_KEY")
}

// AtlasServerURL returns the Atlas server url to use for testing
func AtlasServerURL() string {
	if !skipUnlessAtlasServerCalled {
		panic("testutils.SkipUnlessAtlasServerRunning(t) must be called before testutils.AtlasServerURL()")
	}
	if uri := os.Getenv("REALM_MONGODB_CLOUD_API_BASE_URL"); uri != "" {
		return uri
	}
	return defaultServerURL
}

// RealmServerURL returns the Realm server url to use for testing
func RealmServerURL() string {
	if !skipUnlessRealmServerCalled {
		panic("testutils.SkipUnlessRealmServerRunning(t) must be called before testutils.RealmServerURL()")
	}
	if uri := os.Getenv("REALM_SERVER_BASE_URL"); uri != "" {
		return uri
	}
	return defaultServerURL
}

// SkipUnlessAtlasServerRunning skips tests if there is no Atlas server running
// at the configured testing url (see: AtlasServerURL())
var SkipUnlessAtlasServerRunning = func() func(t *testing.T) {
	return func(t *testing.T) {
		if atlasServerRunning {
			return
		}
		skipUnlessAtlasServerCalled = true
		if atlasServerNotRunning {
			MustSkipf(t, "Atlas server not running at %s", AtlasServerURL())
			return
		}
		client := realm.NewClient(AtlasServerURL())
		if err := client.Status(); err != nil {
			atlasServerNotRunning = true
			MustSkipf(t, "Atlas server not running at %s", AtlasServerURL())
			return
		}
		atlasServerRunning = true
	}
}()

// SkipUnlessRealmServerRunning skips tests if there is no Realm server running
// at the configured testing url (see: RealmServerURL())
var SkipUnlessRealmServerRunning = func() func(t *testing.T) {
	return func(t *testing.T) {
		if realmServerRunning {
			return
		}
		skipUnlessRealmServerCalled = true
		if realmServerNotRunning {
			MustSkipf(t, "Realm server not running at %s", RealmServerURL())
			return
		}
		client := realm.NewClient(RealmServerURL())
		if err := client.Status(); err != nil {
			realmServerNotRunning = true
			MustSkipf(t, "Realm server not running at %s", RealmServerURL())
			return
		}
		realmServerRunning = true
	}
}()
