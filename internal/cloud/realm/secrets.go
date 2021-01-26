package realm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	secretsPathPattern = appPathPattern + "/secrets"
)

// Secret is a secret stored in a Realm app
type Secret struct {
	ID   string `json:"_id"`
	Name string `json:"name"`
}

func (c *client) Secrets(groupID, appID string) ([]Secret, error) {
	res, resErr := c.do(
		http.MethodGet,
		fmt.Sprintf(secretsPathPattern, groupID, appID),
		api.RequestOptions{},
	)
	if resErr != nil {
		return nil, resErr
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, parseResponseError(res)
	}
	var secrets []Secret
	if err := json.NewDecoder(res.Body).Decode(&secrets); err != nil {
		return nil, err
	}
	return secrets, nil
}
