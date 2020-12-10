package cli

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestProfile(t *testing.T) {
	tmpDir, teardownTmpDir, tmpDirErr := u.NewTempDir("home")
	assert.Nil(t, tmpDirErr)
	defer teardownTmpDir()

	_, teardownHomeDir := u.SetupHomeDir(tmpDir)
	defer teardownHomeDir()

	profile, profileErr := NewDefaultProfile()
	assert.Nil(t, profileErr)

	t.Run("Should initialize as an empty, default profile", func(t *testing.T) {
		assert.Equal(t, DefaultProfile, profile.Name)
		assert.Equal(t, fmt.Sprintf("%s/%s", tmpDir, profileDir), profile.dir)
		assert.NotNil(t, profile.fs)
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
}

func TestUser(t *testing.T) {
	t.Run("User should redact its private API key by displaying only the last portion", func(t *testing.T) {
		for _, tc := range []struct {
			description   string
			privateAPIKey string
			display       string
		}{
			{
				description:   "With an empty key",
				privateAPIKey: "",
				display:       "",
			},
			{
				description:   "With a key that has no dashes",
				privateAPIKey: "password",
				display:       "********",
			},
			{
				description:   "With a key that has one dash",
				privateAPIKey: "api-key",
				display:       "***-key",
			},
			{
				description:   "With a key that has many dashes",
				privateAPIKey: "some-super-secret-key",
				display:       "****-*****-******-key",
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				user := User{PrivateAPIKey: tc.privateAPIKey}
				assert.Equal(t, tc.display, user.RedactedPrivateAPIKey())
			})
		}
	})
}
