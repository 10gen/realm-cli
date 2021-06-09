package logs

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

const (
	testDateFormat = "2006-01-02T15:04:05-0700" // avoid millisecond precision
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
		realmClient.LogsFn = func(groupID, appID string, opts realm.LogsOptions) (realm.Logs, error) {
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
		realmClient.LogsFn = func(groupID, appID string, opts realm.LogsOptions) (realm.Logs, error) {
			return realm.Logs{
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

func TestLogsListTail(t *testing.T) {
	t.Run("should poll for logs until a shutdown signal is received", func(t *testing.T) {
		var logIdx int
		testLogs := []realm.Logs{
			{{
				Type:      realm.LogTypeAuth,
				Started:   time.Date(2019, time.June, 22, 7, 54, 42, 0, time.UTC),
				Completed: time.Date(2019, time.June, 22, 7, 54, 42, 5_000_000, time.UTC),
				Messages:  []interface{}{"initial log"},
			}},
			{{
				Type:      realm.LogTypeAuth,
				Started:   time.Date(2020, time.June, 22, 7, 54, 42, 0, time.UTC),
				Completed: time.Date(2020, time.June, 22, 7, 54, 42, 5_000_000, time.UTC),
				Messages:  []interface{}{"second log"},
			}},
			{{
				Type:      realm.LogTypeAuth,
				Started:   time.Date(2021, time.June, 22, 7, 54, 42, 0, time.UTC),
				Completed: time.Date(2021, time.June, 22, 7, 54, 42, 5_000_000, time.UTC),
				Messages:  []interface{}{"third log"},
			}},
		}
		startDates := make([]time.Time, len(testLogs))

		var wg sync.WaitGroup
		wg.Add(len(testLogs))

		realmClient := mock.RealmClient{}

		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{}}, nil
		}

		realmClient.LogsFn = func(groupID, appID string, opts realm.LogsOptions) (realm.Logs, error) {
			logs := testLogs[logIdx]
			startDates[logIdx] = opts.Start

			wg.Done()
			logIdx++

			return logs, nil
		}

		out, ui := mock.NewUI()

		sigShutdown := make(chan os.Signal, 1)
		go func() {
			wg.Wait()
			sigShutdown <- os.Interrupt
		}()

		cmd := &CommandList{listInputs{sigShutdown: sigShutdown, Tail: true}}

		cmdStart := time.Now()
		assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))

		assert.Equal(t, `2019-06-22T07:54:42.000+0000     [5ms]             Authentication: OK
  initial log
2020-06-22T07:54:42.000+0000     [5ms]             Authentication: OK
  second log
2021-06-22T07:54:42.000+0000     [5ms]             Authentication: OK
  third log
`, out.String())

		assert.Equal(t, time.Time{}, startDates[0])
		for i := 0; i < len(startDates)-1; i++ {
			expectedStart := cmdStart.Add(time.Duration(5*i) * time.Second).Format(testDateFormat)
			actualStart := startDates[i+1].Format(testDateFormat)

			assert.Equal(t, expectedStart, actualStart)
		}
	})

	t.Run("should poll for logs until an api call returns an error", func(t *testing.T) {
		origTailLookBehind := tailLookBehind
		defer func() { tailLookBehind = origTailLookBehind }()
		tailLookBehind = 2

		testLogs := []realm.Logs{
			{
				{
					Type:      realm.LogTypeAuth,
					Started:   time.Date(2021, time.June, 22, 7, 54, 42, 0, time.UTC),
					Completed: time.Date(2021, time.June, 22, 7, 54, 42, 5_000_000, time.UTC),
					Messages:  []interface{}{"lower log"},
				},
				{
					Type:      realm.LogTypeAuth,
					Started:   time.Date(2020, time.June, 22, 7, 54, 42, 0, time.UTC),
					Completed: time.Date(2020, time.June, 22, 7, 54, 42, 5_000_000, time.UTC),
					Messages:  []interface{}{"upper log"},
				},
				{
					Type:      realm.LogTypeAuth,
					Started:   time.Date(2019, time.June, 22, 7, 54, 42, 0, time.UTC),
					Completed: time.Date(2019, time.June, 22, 7, 54, 42, 5_000_000, time.UTC),
					Messages:  []interface{}{"skipped log"},
				},
			},
			{{
				Type:      realm.LogTypeAuth,
				Started:   time.Date(2022, time.June, 22, 7, 54, 42, 0, time.UTC),
				Completed: time.Date(2022, time.June, 22, 7, 54, 42, 5_000_000, time.UTC),
				Messages:  []interface{}{"tailed log"},
			}},
		}

		realmClient := mock.RealmClient{}

		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{}}, nil
		}

		var counter int
		realmClient.LogsFn = func(groupID, appID string, opts realm.LogsOptions) (realm.Logs, error) {
			defer func() { counter++ }()
			if counter < len(testLogs) {
				return testLogs[counter], nil
			}
			return nil, errors.New("something bad happened")
		}

		out, ui := mock.NewUI()

		cmd := &CommandList{listInputs{Tail: true}}

		err := cmd.Handler(nil, ui, cli.Clients{Realm: realmClient})
		assert.Equal(t, errors.New("something bad happened"), err)

		assert.Equal(t, `2020-06-22T07:54:42.000+0000     [5ms]             Authentication: OK
  upper log
2021-06-22T07:54:42.000+0000     [5ms]             Authentication: OK
  lower log
2022-06-22T07:54:42.000+0000     [5ms]             Authentication: OK
  tailed log
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

func TestLogTypeFormat(t *testing.T) {
	t.Run("should query for all types of schema logs for log type schema", func(t *testing.T) {
		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{}}, nil
		}
		var logsOpts realm.LogsOptions
		realmClient.LogsFn = func(groupID, appID string, opts realm.LogsOptions) (realm.Logs, error) {
			logsOpts = opts
			return realm.Logs{}, nil
		}

		typeInputs := listInputs{Types: []string{logTypeSchema}}
		logTypes := []string{realm.LogTypeSchemaAdditiveChange, realm.LogTypeSchemaGeneration, realm.LogTypeSchemaValidation}

		cmd := &CommandList{typeInputs}
		assert.Nil(t, cmd.Handler(nil, nil, cli.Clients{Realm: realmClient}))

		assert.Equal(t, logTypes, logsOpts.Types)
	})
}
