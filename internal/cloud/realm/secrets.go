package realm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	valuesPathPattern = appPathPattern + "/values"
)

// Value is a value or secret stored in a Realm application
type Value struct {
	ID           string `json:"_id"`
	Name         string `json:"name"`
	LastModified int64  `json:"last_modified"`
	Secret       bool   `json:"from_secret"`
}

func (c *client) FindValues(app App) ([]Value, error) {
	res, resErr := c.do(
		http.MethodGet,
		fmt.Sprintf(valuesPathPattern, app.GroupID, app.ID),
		api.RequestOptions{},
	)
	if resErr != nil {
		return nil, resErr
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, parseResponseError(res)
	}
	var values []Value
	if err := json.NewDecoder(res.Body).Decode(&values); err != nil {
		return nil, err
	}
	return values, nil
}
