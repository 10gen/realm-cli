package realm

import (
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	importPathPattern = appPathPattern + "/import"

	importQueryStrategy = "strategy"

	importStrategyReplaceByName = "replace-by-name"
)

// ImportRequest is a Realm application import request
type ImportRequest struct {
	ConfigVersion AppConfigVersion `json:"config_version"`
	AuthProviders []AuthProvider   `json:"auth_providers"`
}

func (c *client) Import(groupID, appID string, req ImportRequest) error {
	if req.ConfigVersion == AppConfigVersionZero {
		req.ConfigVersion = DefaultAppConfigVersion
	}

	res, resErr := c.doJSON(
		http.MethodPost,
		fmt.Sprintf(importPathPattern, groupID, appID),
		req,
		api.RequestOptions{Query: map[string]string{
			importQueryStrategy: importStrategyReplaceByName,
		}},
	)
	if resErr != nil {
		return resErr
	}
	if res.StatusCode != http.StatusNoContent {
		defer res.Body.Close()
		return parseResponseError(res)
	}
	return nil
}
