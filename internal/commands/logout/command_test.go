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

		existingCreds := user.Credentials{
			PublicAPIKey:  "publicAPIKey",
			PrivateAPIKey: "privateAPIKey",
			Username:      "username",
			Password:      "password",
		}

		existingSess := user.Session{"accessToken", "refreshToken"}

		profile.SetCredentials(existingCreds)
		profile.SetSession(existingSess)
		assert.Nil(t, profile.Save())

		creds := profile.Credentials()
		session := profile.Session()
		assert.Equal(t, existingCreds, creds)
		assert.Equal(t, existingSess, session)

		out, err := ioutil.ReadFile(profile.Path())
		assert.Nil(t, err)
		assert.True(t, strings.Contains(string(out), fmt.Sprintf(`%s:
  access_token: accessToken
  password: password
  private_api_key: privateAPIKey
  public_api_key: publicAPIKey
  refresh_token: refreshToken
  username: username
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
  password: ""
  private_api_key: ""
  public_api_key: ""
  refresh_token: ""
  username: ""
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
