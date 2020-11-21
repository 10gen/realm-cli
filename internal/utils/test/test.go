package testutils

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"

	"github.com/mitchellh/go-homedir"
)

// MustContainSubstring asserts the second provided arg is a substring of the first
// TODO(REALMC-7315): make this a go-cmp Option instead
func MustContainSubstring(t *testing.T, str, substr string) {
	t.Helper()
	if !strings.Contains(str, substr) {
		t.Errorf("expected %q to be a substring of %q, but it was not", substr, str)
	}
}

// MustNotBeBlank asserts the provided arg is not blank
// TODO(REALMC-7315): make this a go-cmp Option instead
func MustNotBeBlank(t *testing.T, s string) {
	t.Helper()
	if s == "" {
		t.Error("expected value to not be <blank>, but it was")
	}
}

// MustNotBeNil asserts the provided arg is not nil
// TODO(REALMC-7315): make this a go-cmp Option instead
func MustNotBeNil(t *testing.T, o interface{}) {
	t.Helper()
	if o == nil {
		t.Error("expected value to not be <nil>, but it was")
	}
}

// MustMatch asserts the provided diff matches and fails the test if not
func MustMatch(t *testing.T, o string) {
	t.Helper()
	if o != "" {
		t.Error(o)
	}
}

// MustSkipf skips a test suite, but panics if REALM_NO_SKIP_TEST is set
func MustSkipf(t *testing.T, format string, args ...interface{}) {
	if len(os.Getenv("REALM_NO_SKIP_TEST")) > 0 {
		panic("test was skipped, but REALM_NO_SKIP_TEST is set")
	}
	t.Skipf(format, args...)
}

// NewTempDir constructs a new temporary directory
// and returns the directory name along with a cleanup function
// or any error that occurred during the process
func NewTempDir(name string) (string, func(), error) {
	dir, err := ioutil.TempDir("", name)
	if err != nil {
		return "", nil, err
	}
	return dir, func() { os.RemoveAll(dir) }, nil
}

// SetupHomeDir sets up the $HOME directory for a test
// and returns the directory name along with a reset function
func SetupHomeDir(newHome string) (string, func()) {
	origHome := os.Getenv("HOME")
	if newHome == "" {
		newHome = "."
	}

	homedir.DisableCache = true
	_ = os.Setenv("HOME", newHome)

	return newHome, func() {
		homedir.DisableCache = false
		_ = os.Setenv("HOME", origHome)
	}
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
