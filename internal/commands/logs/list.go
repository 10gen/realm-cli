package logs

import (
	"fmt"
	"sort"
	"time"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

const (
	dateFormat = "2006-01-02T15:04:05.000-0700"
)

var (
	tailLookBehind = 5
)

// CommandMetaList is the command meta for the `logs list` command
var CommandMetaList = cli.CommandMeta{
	Use:         "list",
	Aliases:     []string{"ls"},
	Display:     "logs list",
	Description: "Lists the Logs in your Realm app",
	HelpText: `Displays a list of your Realm app’s Logs sorted by recentness, with most recent
Logs appearing towards the bottom. You can specify a "--tail" flag to monitor
your Logs and follow any newly created Logs in real-time.`,
}

// CommandList is the `logs list` command
type CommandList struct {
	inputs listInputs
}

// Flags is the command flags
func (cmd *CommandList) Flags() []flags.Flag {
	return []flags.Flag{
		cli.AppFlagWithContext(&cmd.inputs.App, "to list its logs"),
		cli.ProjectFlag(&cmd.inputs.Project),
		cli.ProductFlag(&cmd.inputs.Products),
		flags.NewStringSetFlag(
			&cmd.inputs.Types,
			flags.StringSetOptions{
				Meta: flags.Meta{
					Name: "type",
					Usage: flags.Usage{
						Description: "Specify the type(s) of logs to list",
					},
				},
				ValidValues: []string{
					logTypeAuth,
					logTypeFunction,
					logTypePush,
					logTypeService,
					logTypeTrigger,
					logTypeGraphQL,
					logTypeSync,
					logTypeSchema,
				},
			},
		),
		flags.BoolFlag{
			Value: &cmd.inputs.Errors,
			Meta: flags.Meta{
				Name: "errors",
				Usage: flags.Usage{
					Description: "View your Realm app's error logs",
				},
			},
		},
		flags.CustomFlag{
			Value: &cmd.inputs.Start,
			Meta: flags.Meta{
				Name: "start",
				Usage: flags.Usage{
					Description:   "Specify when to begin listing logs",
					AllowedFormat: "2006-01-02[T15:04:05.000-0700]",
					DocsLink:      "https://docs.mongodb.com/realm/logs/cli/#view-logs-for-a-date-range",
				},
			},
		},
		flags.CustomFlag{
			Value: &cmd.inputs.End,
			Meta: flags.Meta{
				Name: "end",
				Usage: flags.Usage{
					Description:   "Specify when to finish listing logs",
					AllowedFormat: "2006-01-02[T15:04:05.000-0700]",
					DocsLink:      "https://docs.mongodb.com/realm/logs/cli/#view-logs-for-a-date-range",
				},
			},
		},
		flags.BoolFlag{
			Value: &cmd.inputs.Tail,
			Meta: flags.Meta{
				Name: "tail",
				Usage: flags.Usage{
					Description: "View your Realm app's logs in real-time",
					Note:        `"--start" and "--end" flags do not apply here`,
				},
			},
		},
	}
}

// Inputs is the command inputs
func (cmd *CommandList) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *CommandList) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	cmdStart := time.Now() // for use with tail later

	app, err := cli.ResolveApp(ui, clients.Realm, cli.AppOptions{
		AppMeta: cmd.inputs.AppMeta,
		Filter:  cmd.inputs.Filter(),
	})
	if err != nil {
		return err
	}

	opts := realm.LogsOptions{
		Types:      cmd.inputs.logTypes(),
		ErrorsOnly: cmd.inputs.Errors,
	}
	if !cmd.inputs.Tail {
		opts.Start = cmd.inputs.Start.Time
		opts.End = cmd.inputs.End.Time
	}

	logs, err := clients.Realm.Logs(app.GroupID, app.ID, opts)
	if err != nil {
		return err
	}

	if cmd.inputs.Tail && len(logs) > tailLookBehind {
		logs = logs[0:tailLookBehind]
	}

	printLogs(ui, logs)
	if !cmd.inputs.Tail {
		return nil // if not tailing, command stops here
	}

	logsCh, errCh, closeCh := make(chan realm.Logs), make(chan error), make(chan struct{})
	defer close(closeCh)

	opts.Start = cmdStart
	go pollForLogs(clients.Realm, app.GroupID, app.ID, opts, logsCh, errCh, closeCh)

	for {
		select {
		case logs := <-logsCh:
			printLogs(ui, logs)
		case err := <-errCh:
			return err
		case <-cmd.inputs.sigShutdown:
			return nil
		}
	}
}

func pollForLogs(realmClient realm.Client, groupID, appID string, opts realm.LogsOptions, logsCh chan<- realm.Logs, errCh chan<- error, closeCh <-chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-closeCh:
			return
		case <-ticker.C:
			startNext := time.Now()

			logs, err := realmClient.Logs(groupID, appID, opts)
			if err != nil {
				select {
				case <-closeCh: // if closed during API call
				case errCh <- err:
				}
				return
			}
			select {
			case <-closeCh:
				return // if closed during API call
			case logsCh <- logs:
			}

			opts.Start = startNext
		}
	}
}

func printLogs(ui terminal.UI, logs realm.Logs) {
	sort.Sort(logs)
	for _, log := range logs {
		ui.Print(terminal.NewListLog(
			fmt.Sprintf(
				"%s %9s %26s%s: %s",
				log.Started.Format(dateFormat),
				fmt.Sprintf("[%s]", log.Completed.Sub(log.Started)), // 9 provides spacing for runtime
				logTypeDisplay(log),                                 // 26 provides spacing for type (see test)
				logNameDisplay(log),
				logStatusDisplay(log),
			),
			log.Messages...,
		))
	}
}

func logNameDisplay(log realm.Log) string {
	var name, prefix string

	switch log.Type {
	case realm.LogTypeAuthTrigger, realm.LogTypeDBTrigger, realm.LogTypeScheduledTrigger:
		if log.EventSubscriptionName != "" {
			name = log.EventSubscriptionName
		} else {
			name = log.EventSubscriptionID
		}
	case realm.LogTypeFunction:
		if log.FunctionName != "" {
			name = log.FunctionName
		} else {
			name = log.FunctionID
		}
	case realm.LogTypeServiceStreamFunction, realm.LogTypeServiceFunction:
		if log.FunctionName != "" {
			name = log.FunctionName
		}
	case realm.LogTypeAuth:
		if log.AuthEvent.Provider != "" {
			name = log.AuthEvent.Provider
		}
	case realm.LogTypeWebhook:
		if log.IncomingWebhookName != "" {
			name = log.IncomingWebhookName
		} else {
			name = log.IncomingWebhookID
		}
	}

	if name != "" {
		prefix = " "
	}
	return prefix + name
}

func logStatusDisplay(log realm.Log) string {
	if log.Error == "" {
		return "OK"
	}
	return fmt.Sprintf("%sError - %s", log.ErrorCode, log.Error)
}

const (
	logDisplayAPIKey           = "API Key"
	logDisplayAuthentication   = "Authentication"
	logDisplayFunction         = "Function"
	logDisplayGraphQL          = "GraphQL"
	logDisplayOther            = "Other"
	logDisplayPushNotification = "Push Notification"
	logDisplaySchema           = "Schema"
	logDisplayStreamFunction   = "Stream Function"
	logDisplaySync             = "Sync"
	logDisplayTrigger          = "Trigger"
	logDisplayWebhook          = "Webhook"

	logSubTypeAdditiveChange  = "Additive Change"
	logSubTypeAuth            = "Auth"
	logSubTypeConnectionStart = "Connection Start"
	logSubTypeConnectionEnd   = "Connection End"
	logSubTypeDatabase        = "Database"
	logSubTypeError           = "Error"
	logSubTypeGeneration      = "Generation"
	logSubTypeOther           = "Other"
	logSubTypeScheduled       = "Scheduled"
	logSubTypeService         = "Service"
	logSubTypeSessionStart    = "Session Start"
	logSubTypeSessionEnd      = "Session End"
	logSubTypeValidation      = "Validation"
	logSubTypeWrite           = "Write"
)

func logTypeDisplay(log realm.Log) string {
	var display, subType string

	switch log.Type {
	case realm.LogTypeAuthTrigger:
		display, subType = logDisplayTrigger, logSubTypeAuth
	case realm.LogTypeDBTrigger:
		display, subType = logDisplayTrigger, logSubTypeDatabase
	case realm.LogTypeScheduledTrigger:
		display, subType = logDisplayTrigger, logSubTypeScheduled
	case realm.LogTypeFunction:
		display = logDisplayFunction
	case realm.LogTypeServiceFunction:
		display, subType = logDisplayFunction, logSubTypeService
	case realm.LogTypeStreamFunction:
		display = logDisplayStreamFunction
	case realm.LogTypeServiceStreamFunction:
		display, subType = logDisplayStreamFunction, logSubTypeService
	case realm.LogTypeAuth:
		display = logDisplayAuthentication
	case realm.LogTypeWebhook:
		display = logDisplayWebhook
	case realm.LogTypePush:
		display = logDisplayPushNotification
	case realm.LogTypeAPI:
		display = logDisplayOther
	case realm.LogTypeAPIKey:
		display = logDisplayAPIKey
	case realm.LogTypeGraphQL:
		display = logDisplayGraphQL
	case realm.LogTypeSyncConnectionStart:
		display, subType = logDisplaySync, logSubTypeConnectionStart
	case realm.LogTypeSyncConnectionEnd:
		display, subType = logDisplaySync, logSubTypeConnectionEnd
	case realm.LogTypeSyncSessionStart:
		display, subType = logDisplaySync, logSubTypeSessionStart
	case realm.LogTypeSyncSessionEnd:
		display, subType = logDisplaySync, logSubTypeSessionEnd
	case realm.LogTypeSyncClientWrite:
		display, subType = logDisplaySync, logSubTypeWrite
	case realm.LogTypeSyncError:
		display, subType = logDisplaySync, logSubTypeError
	case realm.LogTypeSyncOther:
		display, subType = logDisplaySync, logSubTypeOther
	case realm.LogTypeSchemaAdditiveChange:
		display, subType = logDisplaySchema, logSubTypeAdditiveChange
	case realm.LogTypeSchemaGeneration:
		display, subType = logDisplaySchema, logSubTypeGeneration
	case realm.LogTypeSchemaValidation:
		display, subType = logDisplaySchema, logSubTypeValidation
	}

	if subType == "" {
		return display
	}
	return fmt.Sprintf("%s -> %s", display, subType)
}
