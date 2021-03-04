package atlas

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

// DataLake contains non sensitive data about an Atlas data lake
type DataLake struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

const (
	dataLakesPattern = atlasAPI + "/groups/%s/dataLakes"
)

func (c *client) DataLakes(groupID string) ([]DataLake, error) {
	res, err := c.do(
		http.MethodGet,
		fmt.Sprintf(dataLakesPattern, groupID),
		api.RequestOptions{},
	)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, api.ErrUnexpectedStatusCode{"get data lakes", res.StatusCode}
	}
	defer res.Body.Close()

	var dataLakes []DataLake
	if err := json.NewDecoder(res.Body).Decode(&dataLakes); err != nil {
		return nil, err
	}

	return dataLakes, nil
}
