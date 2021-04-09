package realm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	logsPathPattern = appPathPattern + "/logs"

	logsQueryEndDate    = "end_date"
	logsQueryErrorsOnly = "errors_only"
	logsQueryStartDate  = "start_date"
	logsQueryType       = "type"

	logsDateFormat = "2006-01-02T15:04:05.999Z07:00"
)

// set of supported Realm app log types
const (
	LogTypeAPI                   = "API"
	LogTypeAPIKey                = "API_KEY"
	LogTypeAuth                  = "AUTH"
	LogTypeAuthTrigger           = "AUTH_TRIGGER"
	LogTypeDBTrigger             = "DB_TRIGGER"
	LogTypeFunction              = "FUNCTION"
	LogTypeGraphQL               = "GRAPHQL"
	LogTypePush                  = "PUSH"
	LogTypeScheduledTrigger      = "SCHEDULED_TRIGGER"
	LogTypeSchemaAdditiveChange  = "SCHEMA_ADDITIVE_CHANGE"
	LogTypeSchemaGeneration      = "SCHEMA_GENERATION"
	LogTypeSchemaValidation      = "SCHEMA_VALIDATION"
	LogTypeServiceFunction       = "SERVICE_FUNCTION"
	LogTypeServiceStreamFunction = "SERVICE_STREAM_FUNCTION"
	LogTypeStreamFunction        = "STREAM_FUNCTION"
	LogTypeSyncClientWrite       = "SYNC_CLIENT_WRITE"
	LogTypeSyncConnectionEnd     = "SYNC_CONNECTION_END"
	LogTypeSyncConnectionStart   = "SYNC_CONNECTION_START"
	LogTypeSyncError             = "SYNC_ERROR"
	LogTypeSyncOther             = "SYNC_OTHER"
	LogTypeSyncSessionEnd        = "SYNC_SESSION_END"
	LogTypeSyncSessionStart      = "SYNC_SESSION_START"
	LogTypeWebhook               = "WEBHOOK"
)

// LogsOptions are options to query for a Realm app's logs
type LogsOptions struct {
	ErrorsOnly bool
	Types      []string
	Start      time.Time
	End        time.Time
}

// Logs is an array of Realm app logs
type Logs []Log

func (l Logs) Len() int           { return len(l) }
func (l Logs) Less(i, j int) bool { return l[i].Started.Before(l[j].Started) }
func (l Logs) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }

// Log is a Realm app log
type Log struct {
	Messages              []interface{} `json:"messages"`
	Type                  string        `json:"type"`
	Started               time.Time     `json:"started"`
	Completed             time.Time     `json:"completed"`
	MemTimeUsage          int64         `json:"mem_time_usage"`
	Error                 string        `json:"error"`
	ErrorCode             string        `json:"error_code"`
	AuthEvent             LogAuthEvent  `json:"auth_event"`
	EventSubscriptionID   string        `json:"event_subscription_id"`
	EventSubscriptionName string        `json:"event_subscription_name"`
	FunctionID            string        `json:"function_id"`
	FunctionName          string        `json:"function_name"`
	IncomingWebhookID     string        `json:"incoming_webhook_id"`
	IncomingWebhookName   string        `json:"incoming_webhook_name"`
}

// LogAuthEvent is a Realm app log auth event
type LogAuthEvent struct {
	Failed   bool   `json:"failed"`
	Type     string `json:"type"`
	Provider string `json:"provider"`
}

type logsResponse struct {
	Logs []Log `json:"logs"`
}

func (c *client) Logs(groupID, appID string, opts LogsOptions) (Logs, error) {
	query := map[string]string{}
	if len(opts.Types) > 0 {
		query[logsQueryType] = strings.Join(opts.Types, ",")
	}
	if opts.ErrorsOnly {
		query[logsQueryErrorsOnly] = trueVal
	}
	if !opts.Start.IsZero() {
		query[logsQueryStartDate] = opts.Start.Format(logsDateFormat)
	}
	if !opts.End.IsZero() {
		query[logsQueryEndDate] = opts.End.Format(logsDateFormat)
	}

	res, err := c.do(
		http.MethodGet,
		fmt.Sprintf(logsPathPattern, groupID, appID),
		api.RequestOptions{Query: query},
	)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, api.ErrUnexpectedStatusCode{"get logs", res.StatusCode}
	}
	defer res.Body.Close()

	var out logsResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Logs, nil
}
