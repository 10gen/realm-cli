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

func (c *client) Diff(groupID, appID string, appData interface{}) ([]string, error) {
	res, resErr := c.doImport(groupID, appID, appData, true)
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

func (c *client) Import(groupID, appID string, appData interface{}) error {
	res, resErr := c.doImport(groupID, appID, appData, false)
	if resErr != nil {
		return resErr
	}
	if res.StatusCode != http.StatusNoContent {
		return api.ErrUnexpectedStatusCode{"import", res.StatusCode}
	}
	return nil
}

func (c *client) doImport(groupID, appID string, appData interface{}, diff bool) (*http.Response, error) {
	query := map[string]string{importQueryStrategy: importStrategyReplaceByName}
	if diff {
		query[importQueryDiff] = trueVal
	}

	return c.doJSON(
		http.MethodPost,
		fmt.Sprintf(importPathPattern, groupID, appID),
		appData,
		api.RequestOptions{Query: query},
	)
}
