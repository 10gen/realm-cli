package atlas

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

// ServerlessInstance contains non sensitive data about an Atlas serverless instance
type ServerlessInstance struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	State string `json:"stateName"`
}

type serverlessInstancesResponse struct {
	Results []ServerlessInstance `json:"results"`
}

const (
	serverlessInstancesPattern = atlasAPI + "/groups/%s/serverless"
)

func (c *client) ServerlessInstances(groupID string) ([]ServerlessInstance, error) {
	res, err := c.doWithBaseURL(
		http.MethodGet,
		fmt.Sprintf(serverlessInstancesPattern, groupID),
		api.RequestOptions{},
	)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, api.ErrUnexpectedStatusCode{"get serverless instances", res.StatusCode}
	}
	defer res.Body.Close()

	var serverlessInstances serverlessInstancesResponse
	if err := json.NewDecoder(res.Body).Decode(&serverlessInstances); err != nil {
		return nil, err
	}

	return serverlessInstances.Results, nil
}
