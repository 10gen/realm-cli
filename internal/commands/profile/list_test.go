package profile

import (
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestProfileList(t *testing.T) {
	tmpDir, teardownTmpDir, tmpDirErr := u.NewTempDir("home")
	assert.Nil(t, tmpDirErr)
	defer teardownTmpDir()

	_, teardownHomeDir := u.SetupHomeDir(tmpDir)
	defer teardownHomeDir()

	profile1, profileErr := user.NewDefaultProfile()
	assert.Nil(t, profileErr)
	profile1.Save()

	profile2, profileErr := user.NewProfile("profile")
	assert.Nil(t, profileErr)

	profile2.SetCredentials(user.Credentials{PublicAPIKey: "abcde", PrivateAPIKey: "my-private-key"})
	profile2.Save()

	t.Run("should provide a list of existing profiles along with their api keys", func(t *testing.T) {
		out, ui := mock.NewUI()

		cmd := &CommandList{}
		assert.Nil(t, cmd.Handler(profile1, ui, cli.Clients{}))

		assert.Equal(t,
			fmt.Sprintf(`Found 2 profile(s)
  Profile  API Key               
  -------  ----------------------
  %s  (logged out)          
  %s  abcde (**-*******-key)
`, profile1.Name, profile2.Name),
			out.String(),
		)
	})
}
