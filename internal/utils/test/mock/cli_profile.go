package mock

import (
	"testing"

	"github.com/10gen/realm-cli/internal/auth"
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NewProfile returns a new CLI profile with a random name
func NewProfile(t *testing.T) *cli.Profile {
	t.Helper()
	profile, err := cli.NewProfile(primitive.NewObjectID().Hex())
	assert.Nil(t, err)
	return profile
}

// NewProfileFromTmpDir returns a new CLI profile with a random name
// and a current working directory based on a temporary directory
// along with the associated cleanup function
func NewProfileFromTmpDir(t *testing.T, name string) (*cli.Profile, func()) {
	t.Helper()

	tmpDir, teardown, err := u.NewTempDir(name)
	assert.Nil(t, err)

	profile := NewProfile(t)
	profile.WorkingDirectory = tmpDir

	return profile, teardown
}

// NewProfileWithSession returns a new CLI profile with a session
func NewProfileWithSession(t *testing.T, session realm.Session) *cli.Profile {
	profile := NewProfile(t)
	profile.SetRealmBaseURL(u.RealmServerURL())
	profile.SetSession(auth.Session{session.AccessToken, session.RefreshToken})
	return profile
}
