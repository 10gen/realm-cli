package atlas_test

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestAtlasStatus(t *testing.T) {
	u.SkipUnlessAtlasServerRunning(t)
	client := atlas.NewAuthClient(u.AtlasServerURL(), user.Credentials{u.CloudUsername(), u.CloudAPIKey()})

	t.Run("Should return no error if the server is running", func(t *testing.T) {
		err := client.Status()
		assert.Nil(t, err)
	})
}

func TestAtlasStatusFailure(t *testing.T) {
	baseURL := "http://localhost:8081"
	client := atlas.NewClient(baseURL)

	t.Run("Should return an error if the server is not running", func(t *testing.T) {
		err := client.Status()
		assert.Equal(t, atlas.ErrServerUnavailable, err)
	})
}
