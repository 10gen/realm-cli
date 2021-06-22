package logs

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

// set of supported log type flags
const (
	logTypeAuth     = "auth"
	logTypeFunction = "function"
	logTypePush     = "push"
	logTypeService  = "service"
	logTypeTrigger  = "trigger"
	logTypeGraphQL  = "graphql"
	logTypeSync     = "sync"
	logTypeSchema   = "schema"
)

type listInputs struct {
	cli.ProjectInputs
	Types       []string
	Errors      bool
	Start       flags.Date
	End         flags.Date
	Tail        bool
	sigShutdown chan os.Signal
}

func (i *listInputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	i.sigShutdown = make(chan os.Signal, 1)
	signal.Notify(i.sigShutdown, syscall.SIGTERM, syscall.SIGINT)

	return i.ProjectInputs.Resolve(ui, profile.WorkingDirectory, true)
}

func (i *listInputs) logTypes() []string {
	var types []string
	for _, lt := range i.Types {
		switch lt {
		case logTypeAuth:
			types = append(types, realm.LogTypeAuth, realm.LogTypeAPIKey)
		case logTypeFunction:
			types = append(types, realm.LogTypeFunction)
		case logTypePush:
			types = append(types, realm.LogTypePush)
		case logTypeService:
			types = append(types, realm.LogTypeServiceFunction, realm.LogTypeWebhook, realm.LogTypeServiceStreamFunction, realm.LogTypeStreamFunction)
		case logTypeTrigger:
			types = append(types, realm.LogTypeAuthTrigger, realm.LogTypeDBTrigger, realm.LogTypeScheduledTrigger)
		case logTypeGraphQL:
			types = append(types, realm.LogTypeGraphQL)
		case logTypeSync:
			types = append(types, realm.LogTypeSyncConnectionStart, realm.LogTypeSyncConnectionEnd, realm.LogTypeSyncSessionStart, realm.LogTypeSyncSessionEnd, realm.LogTypeSyncClientWrite, realm.LogTypeSyncError, realm.LogTypeSyncOther)
		case logTypeSchema:
			types = append(types, realm.LogTypeSchemaAdditiveChange, realm.LogTypeSchemaGeneration, realm.LogTypeSchemaValidation)
		}
	}
	return types
}
