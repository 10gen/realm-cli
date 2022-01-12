package atlas_test

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestClusters(t *testing.T) {
	u.SkipUnlessAtlasServerRunning(t)

	for _, tc := range []struct {
		description string
		client      atlas.Client
		expectedErr error
	}{
		{
			description: "without an auth client",
			client:      atlas.NewClient(u.AtlasServerURL()),
			expectedErr: atlas.ErrMissingAuth,
		},
		{
			description: "with a client with bad credentials",
			client:      atlas.NewAuthClient(u.AtlasServerURL(), user.Credentials{PublicAPIKey: "username", PrivateAPIKey: "password"}),
			expectedErr: atlas.ErrUnauthorized{"You are not authorized for this resource."},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			_, err := tc.client.Clusters(u.CloudGroupID())
			assert.Equal(t, tc.expectedErr, err)
		})
	}

	t.Run("with an authenticated client should return the list of atlas clusters", func(t *testing.T) {
		client := newAuthClient(t)

		clusters, err := client.Clusters(u.CloudGroupID())
		assert.Nil(t, err)
		assert.Equal(t, u.CloudAtlasClusterCount(), len(clusters))
	})
}
