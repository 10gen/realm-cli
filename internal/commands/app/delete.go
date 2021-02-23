package app

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// CommandDelete is the `app delete` command
type CommandDelete struct {
	inputs      cli.ProjectInputs
	realmClient realm.Client
	outputs     []appOutput
}

// Flags is the command flags
func (cmd *CommandDelete) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)
}

// Setup is the command setup
func (cmd *CommandDelete) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = profile.RealmAuthClient()
	return nil
}

// Handler is the command handler
func (cmd *CommandDelete) Handler(profile *cli.Profile, ui terminal.UI) error {
	apps, appErr := cli.ResolveApps(ui, cmd.realmClient, cmd.inputs.Filter())
	if appErr != nil {
		return appErr
	}

	for _, app := range apps {
		err := cmd.realmClient.DeleteApp(app.GroupID, app.ID)
		cmd.outputs = append(cmd.outputs, appOutput{app: app, err: err})
	}
	return nil
}

// Feedback is the command feedback
func (cmd *CommandDelete) Feedback(profile *cli.Profile, ui terminal.UI) error {
	if len(cmd.outputs) == 0 {
		return ui.Print(terminal.NewTextLog("No apps to delete"))
	}

	var logs []terminal.Log
	logs = append(logs, terminal.NewTableLog(
		"Deleted app(s)",
		appDeleteTableHeaders,
		appDeleteRows(cmd.outputs)...,
	))
	return ui.Print(logs...)
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
