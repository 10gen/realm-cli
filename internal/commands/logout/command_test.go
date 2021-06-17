package logout

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestLogoutHandler(t *testing.T) {
	t.Run("should clear user and session and save the config", func(t *testing.T) {
		tmpDir, teardownTmpDir, tmpDirErr := u.NewTempDir("home")
		assert.Nil(t, tmpDirErr)
		defer teardownTmpDir()

		_, teardownHomeDir := u.SetupHomeDir(tmpDir)
		defer teardownHomeDir()

		profile := mock.NewProfile(t)

		profile.SetCredentials(user.Credentials{"username", "password"})
		profile.SetSession(user.Session{"accessToken", "refreshToken"})
		assert.Nil(t, profile.Save())

		creds := profile.Credentials()
		session := profile.Session()
		assert.Equal(t, user.Credentials{"username", "password"}, creds)
		assert.Equal(t, user.Session{"accessToken", "refreshToken"}, session)

		out, err := ioutil.ReadFile(profile.Path())
		assert.Nil(t, err)
		assert.True(t, strings.Contains(string(out), fmt.Sprintf(`%s:
  access_token: accessToken
  private_api_key: password
  public_api_key: username
  refresh_token: refreshToken
`, profile.Name)), "profile must contain the expected contents")

		_, ui := mock.NewUI()

		cmd := &Command{}

		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{}))

		assert.Equal(t, user.Credentials{}, profile.Credentials())
		assert.Equal(t, user.Session{}, profile.Session())

		out, err = ioutil.ReadFile(profile.Path())
		assert.Nil(t, err)
		assert.True(t, strings.Contains(string(out), fmt.Sprintf(`%s:
  access_token: ""
  private_api_key: ""
  public_api_key: ""
  refresh_token: ""
`, profile.Name)), "profile must contain the expected contents")
	})
}

func TestLogoutFeedback(t *testing.T) {
	t.Run("should print a message that logout was successful", func(t *testing.T) {
		profile := mock.NewProfile(t)

		out, ui := mock.NewUI()

		cmd := &Command{}

		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{}))

		assert.Equal(t, "Successfully logged out\n", out.String())
	})
}
