package logout

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestHandler(t *testing.T) {
	t.Run("Handler should clear user and session and save the config", func(t *testing.T) {
		tmpDir, teardownTmpDir, tmpDirErr := u.NewTempDir("home")
		assert.Nil(t, tmpDirErr)
		defer teardownTmpDir()

		_, teardownHomeDir := u.SetupHomeDir(tmpDir)
		defer teardownHomeDir()

		profile := mock.NewProfile(t)

		profile.SetUser("username", "password")
		profile.SetSession("accessToken", "refreshToken")
		assert.Nil(t, profile.Save())

		user := profile.GetUser()
		session := profile.GetSession()
		assert.Equal(t, cli.User{"username", "password"}, user)
		assert.Equal(t, realm.Session{"accessToken", "refreshToken"}, session)

		out, err := ioutil.ReadFile(profile.Path())
		assert.Nil(t, err)
		assert.True(t, strings.Contains(string(out), fmt.Sprintf(`%s:
  access_token: accessToken
  private_api_key: password
  public_api_key: username
  refresh_token: refreshToken
`, profile.Name)), "profile must contain the expected contents")

		buf := new(bytes.Buffer)
		ui := mock.NewUI(mock.UIOptions{}, buf)

		cmd := &command{}

		assert.Nil(t, cmd.Handler(profile, ui, nil))

		assert.Equal(t, cli.User{PublicAPIKey: "username"}, profile.GetUser())
		assert.Equal(t, realm.Session{}, profile.GetSession())

		out, err = ioutil.ReadFile(profile.Path())
		assert.Nil(t, err)
		assert.True(t, strings.Contains(string(out), fmt.Sprintf(`%s:
  access_token: ""
  private_api_key: ""
  public_api_key: username
  refresh_token: ""
`, profile.Name)), "profile must contain the expected contents")
	})
}

func TestFeedback(t *testing.T) {
	t.Run("Feedback should print a message that logout was successful", func(t *testing.T) {
		out := new(bytes.Buffer)
		ui := mock.NewUI(mock.UIOptions{}, out)

		cmd := &command{}

		err := cmd.Feedback(nil, ui)
		assert.Nil(t, err)

		assert.Equal(t, "01:23:45 UTC INFO  Successfully logged out\n", out.String())
	})
}
