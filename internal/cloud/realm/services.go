package realm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	appServicesPattern = "/groups/%s/apps/%s/services"
	appServicePattern  = "/groups/%s/apps/%s/services/%s"
)

// ServiceDescData is the underlying service desc data
type ServiceDescData struct {
	ID     primitive.ObjectID     `json:"_id"`
	Name   string                 `json:"name"`
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config,omitempty"`
}

func (c *client) CreateAppService(groupID, appID string, service ServiceDescData) (ServiceDescData, error) {
	res, resErr := c.doJSON(
		http.MethodPost,
		fmt.Sprintf(runCommandPattern, groupID, appID, listAtlasClusterCommand),
		service,
		api.RequestOptions{},
	)
	if resErr != nil {
		return ServiceDescData{}, resErr
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return ServiceDescData{}, parseResponseError(res)
	}

	var newService ServiceDescData
	if err := json.NewDecoder(res.Body).Decode(&newService); err != nil {
		return ServiceDescData{}, err
	}
	return newService, nil
}
