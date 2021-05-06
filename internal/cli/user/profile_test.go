package user_test

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cli/user"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestProfile(t *testing.T) {
	tmpDir, teardownTmpDir, tmpDirErr := u.NewTempDir("home")
	assert.Nil(t, tmpDirErr)
	defer teardownTmpDir()

	_, teardownHomeDir := u.SetupHomeDir(tmpDir)
	defer teardownHomeDir()

	profile, profileErr := user.NewDefaultProfile()
	assert.Nil(t, profileErr)

	t.Run("Should initialize as an empty, default profile", func(t *testing.T) {
		assert.Equal(t, user.DefaultProfile, profile.Name)
		assert.Equal(t, tmpDir+"/.config/realm-cli", profile.Dir())
	})

	t.Run("Should load a config that does not exist without error", func(t *testing.T) {
		assert.Nil(t, profile.Load())
	})

	t.Run("Should set config values properly", func(t *testing.T) {
		profile.SetString("a", "ayyy")
		profile.SetString("b", "be")

		assert.Equal(t, profile.GetString("a"), "ayyy")
		assert.Equal(t, profile.GetString("b"), "be")
	})

	t.Run("Should save a config properly", func(t *testing.T) {
		assert.Nil(t, profile.Save())

		config, err := ioutil.ReadFile(profile.Path())
		assert.Nil(t, err)
		assert.True(t, strings.Contains(string(config), `default:
  a: ayyy
  b: be
`), "config must contain the expected contents")
	})

	t.Run("Should provide a path the the hosting asset cache file", func(t *testing.T) {
		cachePath := fmt.Sprintf("%s/%s/%s.json", profile.Dir(), user.HostingAssetCacheDir, profile.Name)
		assert.Equal(t, cachePath, profile.HostingAssetCachePath())
	})
}
