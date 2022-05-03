package atlas

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

// Datalake contains non sensitive data about an Atlas data lake
type Datalake struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

const (
	datalakesPattern = atlasAPI + "/groups/%s/dataLakes"
)

func (c *client) Datalakes(groupID string) ([]Datalake, error) {
	res, err := c.doWithBaseURL(
		http.MethodGet,
		fmt.Sprintf(datalakesPattern, groupID),
		api.RequestOptions{},
	)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, api.ErrUnexpectedStatusCode{"get data lakes", res.StatusCode}
	}
	defer res.Body.Close()

	var datalakes []Datalake
	if err := json.NewDecoder(res.Body).Decode(&datalakes); err != nil {
		return nil, err
	}

	return datalakes, nil
}
