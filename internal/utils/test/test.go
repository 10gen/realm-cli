package testutils

import (
	"os"
	"strconv"
	"testing"

	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
)

// MustSkipf skips a test suite, but panics if BAAS_NO_SKIP_TEST is set
func MustSkipf(t *testing.T, format string, args ...interface{}) {
	if len(os.Getenv("BAAS_NO_SKIP_TEST")) > 0 {
		panic("test was skipped, but BAAS_NO_SKIP_TEST is set")
	}
	t.Skipf(format, args...)
}

const (
	defaultGroupCount                   = 1
	defaultGroupID                      = "5fd45718cface356de9d104d"
	defaultGroupName                    = "Project 0"
	defaultAtlasServerURL               = "https://cloud-dev.mongodb.com"
	defaultRealmServerURL               = "http://localhost:8080"
	defaultAtlasClusterCount            = 3
	defaultAtlasServerlessInstanceCount = 1
	defaultAtlasDatalakeCount           = 1
)

var realmServerRunning = false
var realmServerNotRunning = false
var skipUnlessRealmServerCalled = false
var skipUnlessExpiredAccessTokenCalled = false
var atlasServerRunning = false
var atlasServerNotRunning = false
var skipUnlessAtlasServerCalled = false

// CloudGroupCount returns the Cloud groups count to use for testing
func CloudGroupCount() int {
	if count := os.Getenv("BAAS_MONGODB_CLOUD_GROUP_COUNT"); count != "" {
		c, err := strconv.Atoi(count)
		if err != nil {
			panic("BAAS_MONGODB_CLOUD_GROUP_COUNT must be set with an integer")
		}
		return c
	}
	return defaultGroupCount
}

// CloudGroupID returns the Cloud group id to use for testing
func CloudGroupID() string {
	if groupID := os.Getenv("BAAS_MONGODB_CLOUD_GROUP_ID"); groupID != "" {
		return groupID
	}
	return defaultGroupID
}

// CloudGroupName returns the Cloud group id to use for testing
func CloudGroupName() string {
	if groupName := os.Getenv("BAAS_MONGODB_CLOUD_GROUP_NAME"); groupName != "" {
		return groupName
	}
	return defaultGroupName
}

// CloudUsername returns the Cloud username to use for testing
func CloudUsername() string {
	return os.Getenv("BAAS_MONGODB_CLOUD_USERNAME")
}

// CloudAPIKey returns the Cloud api key to use for testing
func CloudAPIKey() string {
	return os.Getenv("BAAS_MONGODB_CLOUD_API_KEY")
}

// CloudAdminUsername returns the Cloud admin username
func CloudAdminUsername() string {
	return os.Getenv("BAAS_MONGODB_CLOUD_ADMIN_USERNAME")
}

// CloudAdminAPIKey returns the Cloud admin api key
func CloudAdminAPIKey() string {
	return os.Getenv("BAAS_MONGODB_CLOUD_ADMIN_API_KEY")
}

// CloudAtlasClusterCount returns the count of clusters to use for testing
func CloudAtlasClusterCount() int {
	if count := os.Getenv("BAAS_MONGODB_CLOUD_ATLAS_CLUSTER_COUNT"); count != "" {
		c, err := strconv.Atoi(count)
		if err != nil {
			panic("BAAS_MONGODB_CLOUD_ATLAS_CLUSTER_COUNT must be set with an integer")
		}
		return c
	}
	return defaultAtlasClusterCount
}

// CloudAtlasServerlessInstanceCount returns the count of serverless instances to use for testing
func CloudAtlasServerlessInstanceCount() int {
	if count := os.Getenv("BAAS_MONGODB_CLOUD_ATLAS_SERVERLESS_INSTANCE_COUNT"); count != "" {
		c, err := strconv.Atoi(count)
		if err != nil {
			panic("BAAS_MONGODB_CLOUD_ATLAS_SERVERLESS_INSTANCE_COUNT must be set with an integer")
		}
		return c
	}
	return defaultAtlasServerlessInstanceCount
}

// CloudAtlasDatalakeCount returns the count of clusters to use for testing
func CloudAtlasDatalakeCount() int {
	if count := os.Getenv("BAAS_MONGODB_CLOUD_ATLAS_DATA_LAKE_COUNT"); count != "" {
		c, err := strconv.Atoi(count)
		if err != nil {
			panic("BAAS_MONGODB_CLOUD_ATLAS_DATA_LAKE_COUNT must be set with an integer")
		}
		return c
	}
	return defaultAtlasDatalakeCount
}

// AtlasServerURL returns the Atlas server url to use for testing
func AtlasServerURL() string {
	if !skipUnlessAtlasServerCalled {
		panic("testutils.SkipUnlessAtlasServerRunning(t) must be called before testutils.AtlasServerURL()")
	}
	if uri := os.Getenv("BAAS_MONGODB_CLOUD_API_BASE_URL"); uri != "" {
		return uri
	}
	return defaultAtlasServerURL
}

// RealmServerURL returns the Realm server url to use for testing
func RealmServerURL() string {
	if !skipUnlessRealmServerCalled {
		panic("testutils.SkipUnlessRealmServerRunning(t) must be called before testutils.RealmServerURL()")
	}
	if uri := os.Getenv("BAAS_SERVER_BASE_URL"); uri != "" {
		return uri
	}
	return defaultRealmServerURL
}

// ExpiredAccessToken returns an expired access token to use for testing
func ExpiredAccessToken() string {
	if !skipUnlessExpiredAccessTokenCalled {
		panic("testutils.SkipUnlessExpiredAccessTokenPresent(t) must be called before testutils.ExpiredAccessToken()")
	}
	return os.Getenv("BAAS_MONGODB_EXPIRED_ACCESS_TOKEN")
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
		client := atlas.NewAuthClient(AtlasServerURL(), user.Credentials{
			PublicAPIKey:  CloudUsername(),
			PrivateAPIKey: CloudAPIKey(),
		})
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

// SkipUnlessExpiredAccessTokenPresent skips tests if there is no expired access token
// set as an environment URL
var SkipUnlessExpiredAccessTokenPresent = func() func(t *testing.T) {
	return func(t *testing.T) {
		skipUnlessExpiredAccessTokenCalled = true
		if ExpiredAccessToken() == "" {
			MustSkipf(t, "expired access token not detected in environment")
		}
		return
	}
}()
