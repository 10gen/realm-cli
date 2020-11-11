package cli

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"github.com/google/go-cmp/cmp"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestLogoutHandler(t *testing.T) {
	t.Run("Handler should clear session tokens and save the config", func(t *testing.T) {
		tmpDir, teardownTmpDir, tmpDirErr := u.NewTempDir("home")
		u.MustMatch(t, cmp.Diff(nil, tmpDirErr))
		defer teardownTmpDir()

		_, teardownHomeDir := u.SetupHomeDir(tmpDir)
		defer teardownHomeDir()

		profile, profileErr := NewProfile(primitive.NewObjectID().Hex())
		u.MustMatch(t, cmp.Diff(nil, profileErr))

		profile.SetSession("accessToken", "refreshToken")
		u.MustMatch(t, cmp.Diff(nil, profile.Save()))

		u.MustMatch(t, cmp.Diff(Session{"accessToken", "refreshToken"}, profile.GetSession()))

		t.Log("ensure profile has the expected contents")
		out, err := ioutil.ReadFile(profile.path())
		u.MustMatch(t, cmp.Diff(nil, err))
		u.MustContainSubstring(t, string(out), fmt.Sprintf(`%s:
  access_token: accessToken
  refresh_token: refreshToken
`, profile.Name))

		buf := new(bytes.Buffer)
		ui := mock.NewUI(mock.UIOptions{}, buf)

		cmd := &logoutCommand{}

		u.MustMatch(t, cmp.Diff(nil, cmd.Handler(profile, ui, nil)))

		u.MustMatch(t, cmp.Diff(Session{}, profile.GetSession()))

		t.Log("ensure profile has been cleared")
		out, err = ioutil.ReadFile(profile.path())
		u.MustMatch(t, cmp.Diff(nil, err))
		u.MustContainSubstring(t, string(out), fmt.Sprintf(`%s:
  access_token: ""
  refresh_token: ""
`, profile.Name))
	})
}

func TestLogoutFeedback(t *testing.T) {
	t.Run("Feedback should print a message that logout was successful", func(t *testing.T) {
		buf := new(bytes.Buffer)
		ui := mock.NewUI(mock.UIOptions{}, buf)

		cmd := &logoutCommand{}

		err := cmd.Feedback(nil, ui)
		u.MustMatch(t, cmp.Diff(nil, err))

		u.MustMatch(t, cmp.Diff("Successfully logged out.\n", buf.String()))
	})
}
