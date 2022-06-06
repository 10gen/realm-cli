package user_test

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestProfiles(t *testing.T) {
	t.Run("should return a list of profile names and file paths", func(t *testing.T) {
		profile, _ := mock.NewProfileFromTmpDir(t, "temp")
		profile.Save()

		profileMetas, err := user.Profiles()
		assert.Nil(t, err)

		assert.Equal(t,
			[]user.ProfileMeta{{Name: profile.Name, Filepath: profile.Path()}},
			profileMetas,
		)
	})
}
