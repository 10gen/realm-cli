package realm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	runCommandPattern = appPathPattern + "/commands/%s"
)

// Commands that are currently supported
const (
	listAtlasClusterCommand = "list_clusters"
)

// PartialAtlasCluster contains non sensitive data about an Atlas cluster
type PartialAtlasCluster struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
}

func (c *client) ListClusters(groupID, appID string) ([]PartialAtlasCluster, error) {
	res, resErr := c.do(
		http.MethodPost,
		fmt.Sprintf(runCommandPattern, groupID, appID, listAtlasClusterCommand),
		api.RequestOptions{},
	)
	if resErr != nil {
		return nil, resErr
	}
	if res.StatusCode != http.StatusOK {
		return nil, api.ErrUnexpectedStatusCode{"list clusters", res.StatusCode}
	}
	defer res.Body.Close()

	var clusters []PartialAtlasCluster
	if err := json.NewDecoder(res.Body).Decode(&clusters); err != nil {
		return nil, err
	}
	return clusters, nil
}
