package realm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	runCommandPattern = "/groups/%s/apps/%s/commands/%s"
)

// Commands that are currently supported
const (
	listAtlasClusterCommand = "list_clusters"
	listDataSourcesCommand  = "list_data_sources"
)

// PartialDataSource contains non sensitive data about a data source, which can be an Atlas Cluster or Data Lake
type PartialDataSource struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
	Type  string `json:"type"`
}

// PartialAtlasCluster contains non sensitive data about an Atlas cluster
type PartialAtlasCluster struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
}

func (c *client) ListClusters(groupID, appID string) ([]PartialAtlasCluster, error) {
	res, resErr := c.doJSON(
		http.MethodPost,
		fmt.Sprintf(runCommandPattern, groupID, appID, listAtlasClusterCommand),
		nil,
		api.RequestOptions{},
	)
	if resErr != nil {
		return nil, resErr
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, parseResponseError(res)
	}

	var clusters []PartialAtlasCluster
	if err := json.NewDecoder(res.Body).Decode(&clusters); err != nil {
		return nil, err
	}
	return clusters, nil
}

func (c *client) ListDataSources(groupID, appID string) ([]PartialDataSource, error) {
	res, resErr := c.doJSON(
		http.MethodPost,
		fmt.Sprintf(runCommandPattern, groupID, appID, listDataSourcesCommand),
		nil,
		api.RequestOptions{},
	)
	if resErr != nil {
		return nil, resErr
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, parseResponseError(res)
	}

	var dataSources []PartialDataSource
	if err := json.NewDecoder(res.Body).Decode(&dataSources); err != nil {
		return nil, err
	}
	return dataSources, nil
}
