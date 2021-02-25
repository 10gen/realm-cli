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

const (
	clustersByGroupIDPattern = "/api/atlas/v1.0/groups/%s/clusters"
)

func (c *client) ClustersByGroupID(groupID string) ([]Cluster, error) {
	res, err := c.do(
		http.MethodGet,
		fmt.Sprintf(clustersByGroupIDPattern, groupID),
		api.RequestOptions{},
	)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, api.ErrUnexpectedStatusCode{"get group clusters", res.StatusCode}
	}
	defer res.Body.Close()

	var clusters struct {
		Results []Cluster `json:"results"`
	}
	if err := json.NewDecoder(res.Body).Decode(&clusters); err != nil {
		return nil, err
	}

	return clusters.Results, nil
}
