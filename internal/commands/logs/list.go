package logs

import (
	"fmt"
	"sort"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/spf13/pflag"
)

const (
	dateFormat = "2006-01-02T15:04:05.000-0700"
)

// CommandList is the `logs list` command
type CommandList struct {
	inputs listInputs
}

// Flags is the command flags
func (cmd *CommandList) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)

	fs.Var(flags.NewEnumSet(&cmd.inputs.Types, allLogTypes), flagType, flagTypeUsage)
	fs.BoolVar(&cmd.inputs.Errors, flagErrors, false, flagErrorsUsage)
	fs.Var(&cmd.inputs.Start, flagStartDate, flagStartDateUsage)
	fs.Var(&cmd.inputs.End, flagEndDate, flagEndDateUsage)
}

// Inputs is the command inputs
func (cmd *CommandList) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *CommandList) Handler(profile *cli.Profile, ui terminal.UI, clients cli.Clients) error {
	app, err := cli.ResolveApp(ui, clients.Realm, cmd.inputs.Filter())
	if err != nil {
		return err
	}

	logs, err := clients.Realm.Logs(app.GroupID, app.ID, realm.LogsOptions{
		Types:      cmd.inputs.Types,
		ErrorsOnly: cmd.inputs.Errors,
		Start:      cmd.inputs.Start.Time,
		End:        cmd.inputs.End.Time,
	})
	if err != nil {
		return err
	}

	sort.SliceStable(logs, func(i, j int) bool {
		return logs[i].Started.Before(logs[j].Started)
	})

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
	return nil
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
