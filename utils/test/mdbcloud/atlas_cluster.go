package mdbcloud

// AtlasCluster represents an Atlas cluster
type AtlasCluster struct {
	ID               string           `json:"id,omitempty"`
	Name             string           `json:"name"`
	MongoURI         string           `json:"mongoURI,omitempty"`
	StateName        string           `json:"stateName,omitempty"`
	ProviderSettings ProviderSettings `json:"providerSettings"`
}

// ProviderSettings represents the providerSettings in an atlas cluster
type ProviderSettings struct {
	InstanceSize string `json:"instanceSizeName"`
	ProviderName string `json:"providerName"`
	RegionName   string `json:"regionName"`
}

// CreateAtlasCluster represents the body of a Create Cluster request
type CreateAtlasCluster struct {
	AtlasCluster
	BackupEnabled string `json:"backupEnabled"`
}
