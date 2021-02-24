package app

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// CommandDelete is the `app delete` command
type CommandDelete struct {
	inputs deleteInputs
}

// Flags is the command flags
func (cmd *CommandDelete) Flags(fs *pflag.FlagSet) {
	fs.StringSliceVarP(&cmd.inputs.Apps, flagApps, "", []string{}, flagAppsUsage)
	fs.StringVar(&cmd.inputs.Project, flagProject, "", flagProjectUsage)
}

// Handler is the command handler
func (cmd *CommandDelete) Handler(profile *cli.Profile, ui terminal.UI, clients cli.Clients) error {
	apps, err := cmd.inputs.resolveApps(ui, clients.Realm)
	if err != nil {
		return err
	}

	outputs := make([]appOutput, 0, len(apps))
	deletedCount := 0
	for _, app := range apps {
		err := clients.Realm.DeleteApp(app.GroupID, app.ID)
		if err == nil {
			deletedCount++
		}
		outputs = append(outputs, appOutput{app, err})
	}

	if len(outputs) == 0 {
		ui.Print(terminal.NewTextLog("No apps to delete"))
		return nil
	}

	var logs []terminal.Log
	logs = append(logs, terminal.NewTableLog(
		fmt.Sprintf("Successfully deleted %d/%d app(s)", deletedCount, len(apps)),
		appDeleteTableHeaders,
		appDeleteRows(outputs)...,
	))

	ui.Print(logs...)
	return nil
}

var (
	appDeleteTableHeaders = []string{headerID, headerName, headerDeleted, headerDetails}
)

func appDeleteRows(ouputs []appOutput) []map[string]interface{} {
	rows := make([]map[string]interface{}, 0, len(ouputs))
	for _, output := range ouputs {
		var deleted bool
		var details string
		if output.err != nil {
			details = output.err.Error()
		} else {
			deleted = true
		}
		rows = append(rows, map[string]interface{}{
			headerID:      output.app.ID,
			headerName:    output.app.Name,
			headerDetails: details,
			headerDeleted: deleted,
		})
	}
	return rows
}
