package atlas_test

import (
	"testing"

	"github.com/10gen/realm-cli/internal/auth"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestClustersByGroupID(t *testing.T) {
	u.SkipUnlessAtlasServerRunning(t)

	for _, tc := range []struct {
		description string
		client      atlas.Client
		expectedErr error
	}{
		{
			description: "Without an auth client",
			client:      atlas.NewClient(u.AtlasServerURL()),
			expectedErr: atlas.ErrMissingAuth,
		},
		{
			description: "With a client with bad credentials",
			client:      atlas.NewAuthClient(u.AtlasServerURL(), auth.User{"username", "password"}),
			expectedErr: atlas.ErrUnauthorized{"You are not authorized for this resource."},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			_, err := tc.client.ClustersByGroupID(u.CloudGroupID())
			assert.Equal(t, tc.expectedErr, err)
		})
	}

	t.Run("With an authenticated client should return the list of atlas clusters", func(t *testing.T) {
		client := newAuthClient(t)

		clusters, err := client.ClustersByGroupID(u.CloudGroupID())
		assert.Nil(t, err)
		assert.Equal(t, u.CloudAtlasClusterCount(), len(clusters))
	})
}
