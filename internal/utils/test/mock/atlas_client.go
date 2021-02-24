package mock

import "github.com/10gen/realm-cli/internal/cloud/atlas"

// AtlasClient is a mocked Atlas client
type AtlasClient struct {
	atlas.Client
	GroupsFn            func() ([]atlas.Group, error)
	ClustersByGroupIDFn func(groupID string) ([]atlas.Cluster, error)
}

// Groups calls the mocked Groups implementation if provided,
// otherwise the call falls back to the underlying atlas.Client implementation.
// NOTE: this may panic if the underlying atlas.Client is left undefined
func (ac AtlasClient) Groups() ([]atlas.Group, error) {
	if ac.GroupsFn != nil {
		return ac.GroupsFn()
	}
	return ac.Client.Groups()
}

// ClustersByGroupID calls the mocked Groups implementation if provided,
// otherwise the call falls back to the underlying atlas.Client implementation.
// NOTE: this may panic if the underlying atlas.Client is left undefined
func (ac AtlasClient) ClustersByGroupID(groupID string) ([]atlas.Cluster, error) {
	if ac.ClustersByGroupIDFn != nil {
		return ac.ClustersByGroupIDFn(groupID)
	}
	return ac.Client.ClustersByGroupID(groupID)
}
