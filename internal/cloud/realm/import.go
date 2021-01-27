package realm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	importPathPattern = appPathPattern + "/import"

	importQueryDiff     = "diff"
	importQueryStrategy = "strategy"

	importStrategyReplaceByName = "replace-by-name"
)

// ImportRequest is a Realm application import request
type ImportRequest struct {
	ConfigVersion AppConfigVersion `json:"config_version"`
	AppPackage    map[string]interface{}
}

func (c *client) Diff(groupID, appID string, pkg map[string]interface{}) ([]string, error) {
	res, resErr := c.doImport(groupID, appID, pkg, true)
	if resErr != nil {
		return nil, resErr
	}
	if res.StatusCode != http.StatusOK {
		return nil, api.ErrUnexpectedStatusCode{"diff", res.StatusCode}
	}
	defer res.Body.Close()

	var diffs []string
	if err := json.NewDecoder(res.Body).Decode(&diffs); err != nil {
		return nil, err
	}
	return diffs, nil
}

func (c *client) Import(groupID, appID string, pkg map[string]interface{}) error {
	res, resErr := c.doImport(groupID, appID, pkg, false)
	if resErr != nil {
		return resErr
	}
	if res.StatusCode != http.StatusNoContent {
		return api.ErrUnexpectedStatusCode{"import", res.StatusCode}
	}
	return nil
}

func (c *client) doImport(groupID, appID string, pkg map[string]interface{}, diff bool) (*http.Response, error) {
	query := map[string]string{importQueryStrategy: importStrategyReplaceByName}
	if diff {
		query[importQueryDiff] = trueVal
	}

	return c.doJSON(
		http.MethodPost,
		fmt.Sprintf(importPathPattern, groupID, appID),
		pkg,
		api.RequestOptions{Query: query},
	)
}
