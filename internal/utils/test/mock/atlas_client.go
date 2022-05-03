package mock

import "github.com/10gen/realm-cli/internal/cloud/atlas"

// AtlasClient is a mocked Atlas client
type AtlasClient struct {
	atlas.Client
	GroupsFn              func(url string, useBaseURL bool) (atlas.Groups, error)
	ClustersFn            func(groupID string) ([]atlas.Cluster, error)
	ServerlessInstancesFn func(groupID string) ([]atlas.ServerlessInstance, error)
	DatalakesFn           func(groupID string) ([]atlas.Datalake, error)
}

// Groups calls the mocked Groups implementation if provided,
// otherwise the call falls back to the underlying atlas.Client implementation.
// NOTE: this may panic if the underlying atlas.Client is left undefined
func (ac AtlasClient) Groups(url string, useBaseURL bool) (atlas.Groups, error) {
	if ac.GroupsFn != nil {
		return ac.GroupsFn(url, useBaseURL)
	}
	return ac.Client.Groups(url, useBaseURL)
}

// Clusters calls the mocked Clusters implementation if provided,
// otherwise the call falls back to the underlying atlas.Client implementation.
// NOTE: this may panic if the underlying atlas.Client is left undefined
func (ac AtlasClient) Clusters(groupID string) ([]atlas.Cluster, error) {
	if ac.ClustersFn != nil {
		return ac.ClustersFn(groupID)
	}
	return ac.Client.Clusters(groupID)
}

// ServerlessInstances calls the mocked ServerlessInstances implementation if provided,
// otherwise the call falls back to the underlying atlas.Client implementation.
// NOTE: this may panic if the underlying atlas.Client is left undefined
func (ac AtlasClient) ServerlessInstances(groupID string) ([]atlas.ServerlessInstance, error) {
	if ac.ServerlessInstancesFn != nil {
		return ac.ServerlessInstancesFn(groupID)
	}
	return ac.Client.ServerlessInstances(groupID)
}

// Datalakes calls the mocked Datalakes implementation if provided,
// otherwise the call falls back to the underlying atlas.Client implementation.
// NOTE: this may panic if the underlying atlas.Client is left undefined
func (ac AtlasClient) Datalakes(groupID string) ([]atlas.Datalake, error) {
	if ac.DatalakesFn != nil {
		return ac.DatalakesFn(groupID)
	}
	return ac.Client.Datalakes(groupID)
}
