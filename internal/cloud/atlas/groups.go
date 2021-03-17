package atlas

import (
	"encoding/json"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	groupsPath = publicAPI + "/groups"
)

// Group is an Atlas group
type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type groupResponse struct {
	Results []Group `json:"results"`
}

func (c *client) Groups() ([]Group, error) {
	res, resErr := c.do(
		http.MethodGet,
		groupsPath,
		api.RequestOptions{},
	)
	if resErr != nil {
		return nil, resErr
	}
	if res.StatusCode != http.StatusOK {
		return nil, api.ErrUnexpectedStatusCode{"get groups", res.StatusCode}
	}
	defer res.Body.Close()

	var groupRes groupResponse
	if err := json.NewDecoder(res.Body).Decode(&groupRes); err != nil {
		return nil, err
	}
	return groupRes.Results, nil
}
