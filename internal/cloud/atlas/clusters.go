package atlas

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

// Cluster contains non sensitive data about an Atlas cluster
type Cluster struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	State string `json:"stateName"`
}

type clustersResponse struct {
	Results []Cluster `json:"results"`
}

const (
	clustersPattern = atlasAPI + "/groups/%s/clusters"
)

func (c *client) Clusters(groupID string) ([]Cluster, error) {
	res, err := c.doWithBaseURL(
		http.MethodGet,
		fmt.Sprintf(clustersPattern, groupID),
		api.RequestOptions{},
	)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, api.ErrUnexpectedStatusCode{"get clusters", res.StatusCode}
	}
	defer res.Body.Close()

	var clusters clustersResponse
	if err := json.NewDecoder(res.Body).Decode(&clusters); err != nil {
		return nil, err
	}

	return clusters.Results, nil
}
