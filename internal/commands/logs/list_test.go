package logs

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestLogsList(t *testing.T) {
	t.Run("should return an error when client fails to find app", func(t *testing.T) {
		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return nil, errors.New("something bad happened")
		}

		cmd := &CommandList{listInputs{}}

		err := cmd.Handler(nil, nil, cli.Clients{Realm: realmClient})
		assert.Equal(t, errors.New("something bad happened"), err)
	})

	t.Run("should return an error when client fails to get logs", func(t *testing.T) {
		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{}}, nil
		}
		realmClient.LogsFn = func(groupID, appID string, opts realm.LogsOptions) ([]realm.Log, error) {
			return nil, errors.New("something bad happened")
		}

		cmd := &CommandList{}

		err := cmd.Handler(nil, nil, cli.Clients{Realm: realmClient})
		assert.Equal(t, errors.New("something bad happened"), err)
	})

	t.Run("should print logs returned by the client", func(t *testing.T) {
		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{}}, nil
		}
		realmClient.LogsFn = func(groupID, appID string, opts realm.LogsOptions) ([]realm.Log, error) {
			return []realm.Log{
				{
					Type:         realm.LogTypeServiceStreamFunction,
					Messages:     []interface{}{"a test log message"},
					Started:      time.Date(2021, time.June, 22, 7, 54, 42, 0, time.UTC),
					Completed:    time.Date(2021, time.June, 22, 7, 54, 43, 234_000_000, time.UTC),
					FunctionName: "func0",
				},
				{
					Type:                  realm.LogTypeScheduledTrigger,
					Messages:              []interface{}{"one message", "two message", "red message", "blue message"},
					Started:               time.Date(2020, time.June, 22, 7, 54, 42, 0, time.UTC),
					Completed:             time.Date(2020, time.June, 22, 7, 54, 42, 123_000_000, time.UTC),
					EventSubscriptionName: "suessTrigger",
				},
				{
					Type:      realm.LogTypeAuth,
					Error:     "something bad happened",
					ErrorCode: "Test",
					Started:   time.Date(2019, time.June, 22, 7, 54, 42, 0, time.UTC),
					Completed: time.Date(2019, time.June, 22, 7, 54, 42, 5_000_000, time.UTC),
				},
			}, nil
		}

		out, ui := mock.NewUI()

		cmd := &CommandList{listInputs{ProjectInputs: cli.ProjectInputs{Project: "project", App: "test-app"}}}

		err := cmd.Handler(nil, ui, cli.Clients{Realm: realmClient})
		assert.Nil(t, err)

		assert.Equal(t, `2019-06-22T07:54:42.000+0000     [5ms]             Authentication: TestError - something bad happened
2020-06-22T07:54:42.000+0000   [123ms]       Trigger -> Scheduled suessTrigger: OK
  one message
  two message
  red message
  blue message
2021-06-22T07:54:42.000+0000  [1.234s] Stream Function -> Service func0: OK
  a test log message
`, out.String())
	})
}

func TestLogNameDisplay(t *testing.T) {
	for _, tc := range []struct {
		description string
		log         realm.Log
		name        string
	}{
		{
			description: "nothing for a log type that has no name",
		},
		{
			description: "name for an auth trigger log",
			log:         realm.Log{Type: realm.LogTypeAuthTrigger, EventSubscriptionID: "id", EventSubscriptionName: "name"},
			name:        " name",
		},
		{
			description: "id for an auth trigger log without a name",
			log:         realm.Log{Type: realm.LogTypeAuthTrigger, EventSubscriptionID: "id"},
			name:        " id",
		},
		{
			description: "name for a database trigger log",
			log:         realm.Log{Type: realm.LogTypeDBTrigger, EventSubscriptionID: "id", EventSubscriptionName: "name"},
			name:        " name",
		},
		{
			description: "id for a database trigger log without a name",
			log:         realm.Log{Type: realm.LogTypeDBTrigger, EventSubscriptionID: "id"},
			name:        " id",
		},
		{
			description: "name for a scheduled trigger log",
			log:         realm.Log{Type: realm.LogTypeScheduledTrigger, EventSubscriptionID: "id", EventSubscriptionName: "name"},
			name:        " name",
		},
		{
			description: "id for a scheduled trigger log without a name",
			log:         realm.Log{Type: realm.LogTypeScheduledTrigger, EventSubscriptionID: "id"},
			name:        " id",
		},
		{
			description: "name for a function log",
			log:         realm.Log{Type: realm.LogTypeFunction, FunctionID: "id", FunctionName: "name"},
			name:        " name",
		},
		{
			description: "id for a function log without a name",
			log:         realm.Log{Type: realm.LogTypeFunction, FunctionID: "id"},
			name:        " id",
		},
		{
			description: "name for a service stream function log",
			log:         realm.Log{Type: realm.LogTypeServiceStreamFunction, FunctionName: "name"},
			name:        " name",
		},
		{
			description: "name for a service function log",
			log:         realm.Log{Type: realm.LogTypeServiceFunction, FunctionName: "name"},
			name:        " name",
		},
		{
			description: "name for an auth log",
			log:         realm.Log{Type: realm.LogTypeAuth, AuthEvent: realm.LogAuthEvent{Provider: "provider"}},
			name:        " provider",
		},
		{
			description: "name for a webhook log",
			log:         realm.Log{Type: realm.LogTypeWebhook, IncomingWebhookID: "id", IncomingWebhookName: "name"},
			name:        " name",
		},
		{
			description: "id for a webhook log without a name",
			log:         realm.Log{Type: realm.LogTypeWebhook, IncomingWebhookID: "id"},
			name:        " id",
		},
	} {
		t.Run("should display name for "+tc.description, func(t *testing.T) {
			assert.Equal(t, tc.name, logNameDisplay(tc.log))
		})
	}
}

func TestLogStatusDisplay(t *testing.T) {
	for _, tc := range []struct {
		description string
		log         realm.Log
		status      string
	}{
		{
			description: "a log without error",
			status:      "OK",
		},
		{
			description: "a log with a generic error",
			log:         realm.Log{Error: "something bad happened"},
			status:      "Error - something bad happened",
		},
		{
			description: "a log with a custom error",
			log:         realm.Log{Error: "something bad happened", ErrorCode: "Custom"},
			status:      "CustomError - something bad happened",
		},
	} {
		t.Run("should display status for "+tc.description, func(t *testing.T) {
			assert.Equal(t, tc.status, logStatusDisplay(tc.log))
		})
	}
}

func TestLogTypeDisplay(t *testing.T) {
	var maxWidth int
	for _, tc := range []struct {
		logType string
		display string
	}{
		{realm.LogTypeAPI, "Other"},
		{realm.LogTypeAPIKey, "API Key"},
		{realm.LogTypeAuth, "Authentication"},
		{realm.LogTypeAuthTrigger, "Trigger -> Auth"},
		{realm.LogTypeDBTrigger, "Trigger -> Database"},
		{realm.LogTypeFunction, "Function"},
		{realm.LogTypeGraphQL, "GraphQL"},
		{realm.LogTypePush, "Push Notification"},
		{realm.LogTypeScheduledTrigger, "Trigger -> Scheduled"},
		{realm.LogTypeSchemaAdditiveChange, "Schema -> Additive Change"},
		{realm.LogTypeSchemaGeneration, "Schema -> Generation"},
		{realm.LogTypeSchemaValidation, "Schema -> Validation"},
		{realm.LogTypeServiceFunction, "Function -> Service"},
		{realm.LogTypeServiceStreamFunction, "Stream Function -> Service"},
		{realm.LogTypeStreamFunction, "Stream Function"},
		{realm.LogTypeSyncClientWrite, "Sync -> Write"},
		{realm.LogTypeSyncConnectionEnd, "Sync -> Connection End"},
		{realm.LogTypeSyncConnectionStart, "Sync -> Connection Start"},
		{realm.LogTypeSyncError, "Sync -> Error"},
		{realm.LogTypeSyncOther, "Sync -> Other"},
		{realm.LogTypeSyncSessionEnd, "Sync -> Session End"},
		{realm.LogTypeSyncSessionStart, "Sync -> Session Start"},
		{realm.LogTypeWebhook, "Webhook"},
	} {
		t.Run(fmt.Sprintf("should show proper display for log type %s", strings.ToLower(tc.logType)), func(t *testing.T) {
			display := logTypeDisplay(realm.Log{Type: tc.logType})
			assert.Equal(t, tc.display, display)

			if maxWidth < len(display) {
				maxWidth = len(display)
			}
		})
	}
	t.Logf("the max width for log type display is: %d", maxWidth)
}
