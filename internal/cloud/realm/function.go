package realm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

// Routes for functions
const (
	FunctionsPattern               = appPathPattern + "/functions"
	AppDebugExecuteFunctionPattern = appPathPattern + "/debug/execute_function"
)

type stats struct {
	ExecutionTime string `json:"execution_time,omitempty"`
}

// ExecutionResults contains the details around a function execution
type ExecutionResults struct {
	Result    interface{} `json:"result,omitempty"`
	Logs      []string    `json:"logs,omitempty"`
	ErrorLogs []string    `json:"error_logs,omitempty"`
	Stats     stats       `json:"stats,omitempty"`
}

// Function is a realm Function
type Function struct {
	ID   string `json:"_id"`
	Name string `json:"name"`
}

func (c *client) AppDebugExecuteFunction(groupID, appID, userID, name string, args []interface{}) (ExecutionResults, error) {
	query := map[string]string{}
	if userID == "" {
		query["run_as_system"] = "true"
	} else {
		query["user_id"] = userID
	}
	res, err := c.doJSON(
		http.MethodPost,
		fmt.Sprintf(AppDebugExecuteFunctionPattern, groupID, appID),
		map[string]interface{}{
			"name":      name,
			"arguments": args,
		},
		api.RequestOptions{Query: query},
	)
	if err != nil {
		return ExecutionResults{}, err
	}
	if res.StatusCode != http.StatusOK {
		return ExecutionResults{}, api.ErrUnexpectedStatusCode{"debug execute function", res.StatusCode}
	}
	defer res.Body.Close()

	var response ExecutionResults
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return ExecutionResults{}, err
	}
	return response, nil
}

func (c *client) Functions(groupID, appID string) ([]Function, error) {
	res, err := c.do(
		http.MethodGet,
		fmt.Sprintf(FunctionsPattern, groupID, appID),
		api.RequestOptions{},
	)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, api.ErrUnexpectedStatusCode{"list functions", res.StatusCode}
	}
	defer res.Body.Close()

	var result []Function
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}
