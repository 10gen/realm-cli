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
	clustersByGroupIDPattern = groupPath + "/clusters"
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

	var clusters []Cluster
	if err := json.NewDecoder(res.Body).Decode(&clusters); err != nil {
		return nil, err
	}
	return clusters, nil
}
