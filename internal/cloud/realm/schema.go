package realm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	syncClientSchemasPathPattern = appPathPattern + "/sync/client_schemas/%s"
)

// set of supported Realm data model languages
const (
	DataModelLanguageCSharp     = "C_SHARP"
	DataModelLanguageJava       = "JAVA"
	DataModelLanguageJavascript = "JAVA_SCRIPT"
	DataModelLanguageKotlin     = "KOTLIN"
	DataModelLanguageObjectiveC = "OBJECTIVE_C"
	DataModelLanguageSwift      = "SWIFT"
	DataModelLanguageTypescript = "TYPE_SCRIPT"
)

// SchemaModel is a Realm app schema model
type SchemaModel struct {
	ServiceID string             `json:"service_id"`
	RuleID    string             `json:"rule_id"`
	Name      string             `json:"model_name"`
	Namespace string             `json:"collection_display_name"`
	Imports   []string           `json:"import_statements"`
	Code      string             `json:"schema"`
	Warnings  []SchemaModelAlert `json:"warnings"`
	Error     SchemaModelAlert   `json:"error"`
}

// SchemaModelAlert is a Realm app schema model alert
type SchemaModelAlert struct {
	Message string `json:"error"`
	Code    string `json:"error_code"`
}

func (c *client) SchemaModels(groupID, appID, language string) ([]SchemaModel, error) {
	res, err := c.do(
		http.MethodGet,
		fmt.Sprintf(syncClientSchemasPathPattern, groupID, appID, language),
		api.RequestOptions{},
	)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, api.ErrUnexpectedStatusCode{"get schema models", res.StatusCode}
	}
	defer res.Body.Close()

	var models []SchemaModel
	if err := json.NewDecoder(res.Body).Decode(&models); err != nil {
		return nil, err
	}
	return models, nil
}
