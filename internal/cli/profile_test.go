package cli

import (
	"fmt"
	"io/ioutil"
	"testing"

	u "github.com/10gen/realm-cli/internal/utils/test"

	"github.com/google/go-cmp/cmp"
)

func TestProfile(t *testing.T) {
	tmpDir, teardownTmpDir, tmpDirErr := u.NewTempDir("home")
	u.MustMatch(t, cmp.Diff(nil, tmpDirErr))
	defer teardownTmpDir()

	_, teardownHomeDir := u.SetupHomeDir(tmpDir)
	defer teardownHomeDir()

	profile, profileErr := NewDefaultProfile()
	u.MustMatch(t, cmp.Diff(nil, profileErr))

	t.Run("Should initialize as an empty, default profile", func(t *testing.T) {
		u.MustMatch(t, cmp.Diff(DefaultProfile, profile.Name))
		u.MustMatch(t, cmp.Diff(fmt.Sprintf("%s/%s", tmpDir, profileDir), profile.dir))
		u.MustNotBeNil(t, profile.fs)
	})

	t.Run("Should load a config that does not exist without error", func(t *testing.T) {
		u.MustMatch(t, cmp.Diff(nil, profile.Load()))
	})

	t.Run("Should set config values properly", func(t *testing.T) {
		profile.SetString("a", "ayyy")
		profile.SetString("b", "be")

		u.MustMatch(t, cmp.Diff(profile.GetString("a"), "ayyy"))
		u.MustMatch(t, cmp.Diff(profile.GetString("b"), "be"))
	})

	t.Run("Should save a config properly", func(t *testing.T) {
		u.MustMatch(t, cmp.Diff(nil, profile.Save()))

		config, err := ioutil.ReadFile(profile.path())
		u.MustMatch(t, cmp.Diff(nil, err))
		u.MustContainSubstring(t, string(config), `default:
  a: ayyy
  b: be
`)
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
				u.MustMatch(t, cmp.Diff(tc.display, user.RedactedPrivateAPIKey()))
			})
		}
	})
}
