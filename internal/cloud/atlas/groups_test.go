package atlas_test

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestAtlasGroups(t *testing.T) {
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
			client:      atlas.NewAuthClient(u.AtlasServerURL(), user.Credentials{PublicAPIKey: "username", PrivateAPIKey: "password"}),
			expectedErr: atlas.ErrUnauthorized{"You are not authorized for this resource."},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			_, err := atlas.AllGroups(tc.client)
			assert.Equal(t, tc.expectedErr, err)
		})
	}

	t.Run("With an authenticated client should return the list of groups", func(t *testing.T) {
		client := newAuthClient(t)

		groups, err := atlas.AllGroups(client)
		assert.Nil(t, err)
		assert.Equal(t, u.CloudGroupCount(), len(groups))

		groupsM := map[string]string{}
		for _, group := range groups {
			groupsM[group.ID] = group.Name
		}
		assert.Equal(t, groupsM[u.CloudGroupID()], u.CloudGroupName())
	})
}

func TestFetchGroups(t *testing.T) {
	var groupsIdx int
	groupsCalls := []atlas.Groups{
		{
			Results: []atlas.Group{
				{Name: "one"},
				{Name: "two"},
			},
			Links: []atlas.Link{{Rel: "next"}},
		},
		{
			Results: []atlas.Group{
				{Name: "three"},
				{Name: "four"},
			},
			Links: []atlas.Link{{Rel: "next"}},
		},
		{
			Results: []atlas.Group{
				{Name: "five"},
			},
		},
	}

	client := mock.AtlasClient{
		Client: atlas.NewClient(""),
		GroupsFn: func(url string, useBaseURL bool) (atlas.Groups, error) {
			defer func() { groupsIdx++ }()
			return groupsCalls[groupsIdx], nil
		},
	}

	t.Run("should fetch all groups", func(t *testing.T) {
		groups, err := atlas.AllGroups(client)
		assert.Nil(t, err)
		assert.Equal(t, []atlas.Group{
			{Name: "one"},
			{Name: "two"},
			{Name: "three"},
			{Name: "four"},
			{Name: "five"},
		}, groups)
	})
}

func newAuthClient(t *testing.T) atlas.Client {
	return atlas.NewAuthClient(u.AtlasServerURL(), user.Credentials{
		PublicAPIKey:  u.CloudUsername(),
		PrivateAPIKey: u.CloudAPIKey(),
	})
}
