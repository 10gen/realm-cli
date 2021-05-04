package mock

import (
	"os"
	"testing"

	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NewProfile returns a new CLI profile with a random name
func NewProfile(t *testing.T) *user.Profile {
	t.Helper()
	profile, err := user.NewProfile(primitive.NewObjectID().Hex())
	assert.Nil(t, err)
	return profile
}

// NewProfileFromTmpDir returns a new CLI profile with a random name
// and a current working directory based on a temporary directory
// along with the associated cleanup function
func NewProfileFromTmpDir(t *testing.T, name string) (*user.Profile, func()) {
	t.Helper()

	tmpDir, teardown, err := u.NewTempDir(name)
	assert.Nil(t, err)

	_, resetHomeDir := u.SetupHomeDir(tmpDir)

	profile := NewProfile(t)
	profile.WorkingDirectory = tmpDir

	return profile,
		func() {
			resetHomeDir()
			teardown()
		}
}

// NewProfileFromWd returns a new CLI profile with a random name
// and the current working directory
func NewProfileFromWd(t *testing.T) *user.Profile {
	t.Helper()

	wd, err := os.Getwd()
	assert.Nil(t, err)

	profile := NewProfile(t)
	profile.WorkingDirectory = wd

	return profile
}

// NewProfileWithSession returns a new CLI profile with a session
func NewProfileWithSession(t *testing.T, session realm.Session) *user.Profile {
	profile := NewProfile(t)
	profile.SetRealmBaseURL(u.RealmServerURL())
	profile.SetSession(user.Session{session.AccessToken, session.RefreshToken})
	return profile
}
