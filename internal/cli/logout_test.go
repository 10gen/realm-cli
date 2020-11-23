package cli

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestLogoutHandler(t *testing.T) {
	t.Run("Handler should clear session tokens and save the config", func(t *testing.T) {
		tmpDir, teardownTmpDir, tmpDirErr := u.NewTempDir("home")
		assert.Nil(t, tmpDirErr)
		defer teardownTmpDir()

		_, teardownHomeDir := u.SetupHomeDir(tmpDir)
		defer teardownHomeDir()

		profile, profileErr := NewProfile(primitive.NewObjectID().Hex())
		assert.Nil(t, profileErr)

		profile.SetSession("accessToken", "refreshToken")
		assert.Nil(t, profile.Save())

		assert.Match(t, Session{"accessToken", "refreshToken"}, profile.GetSession())

		out, err := ioutil.ReadFile(profile.path())
		assert.Nil(t, err)
		assert.True(t, strings.Contains(string(out), fmt.Sprintf(`%s:
  access_token: accessToken
  refresh_token: refreshToken
`, profile.Name)), "profile must contain the expected contents")

		buf := new(bytes.Buffer)
		ui := mock.NewUI(mock.UIOptions{}, buf)

		cmd := &logoutCommand{}

		assert.Nil(t, cmd.Handler(profile, ui, nil))

		assert.Match(t, Session{}, profile.GetSession())

		out, err = ioutil.ReadFile(profile.path())
		assert.Nil(t, err)
		assert.True(t, strings.Contains(string(out), fmt.Sprintf(`%s:
  access_token: ""
  refresh_token: ""
`, profile.Name)), "profile must contain the expected contents")
	})
}

func TestLogoutFeedback(t *testing.T) {
	t.Run("Feedback should print a message that logout was successful", func(t *testing.T) {
		buf := new(bytes.Buffer)
		ui := mock.NewUI(mock.UIOptions{}, buf)

		cmd := &logoutCommand{}

		err := cmd.Feedback(nil, ui)
		assert.Nil(t, err)

		assert.Equal(t, "Successfully logged out.\n", buf.String())
	})
}
