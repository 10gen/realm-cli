package local

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestLoadProfiles(t *testing.T) {
	t.Run("should return a list of profile names along with file paths and api keys", func(t *testing.T) {
		profile, _ := mock.NewProfileFromTmpDir(t, "temp")

		profile.SetCredentials(user.Credentials{PublicAPIKey: "abcde", PrivateAPIKey: "my-private-key"})
		profile.Save()

		environments, err := LoadEnvironments()
		assert.Nil(t, err)

		assert.Equal(t,
			[]Environment{{
				Name:     profile.Name,
				Filepath: profile.Path(),
				Credentials: Credentials{
					PublicAPIKey:  "abcde",
					PrivateAPIKey: "my-private-key",
				},
			}},
			environments,
		)
	})
}
