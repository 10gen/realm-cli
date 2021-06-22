package app

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

// CommandMetaDelete is the command meta for the `app delete` command
var CommandMetaDelete = cli.CommandMeta{
	Use:         "delete",
	Display:     "app delete",
	Description: "Delete a Realm app",
	HelpText: `If you have more than one Realm app, you will be prompted to select one or
multiple app(s) that you would like to delete from a list of all your Realm apps.
The list includes Realm apps from all projects associated with your user profile.`,
}

// CommandDelete is the `app delete` command
type CommandDelete struct {
	inputs deleteInputs
}

// Flags is the command flags
func (cmd *CommandDelete) Flags() []flags.Flag {
	return []flags.Flag{
		flags.StringSliceFlag{
			Value: &cmd.inputs.Apps,
			Meta: flags.Meta{
				Name:      "app",
				Shorthand: "a",
				Usage: flags.Usage{
					Description: "Specify the name(s) or ID(s) of Realm apps to delete",
				},
			},
		},
		cli.ProjectFlag(&cmd.inputs.Project),
	}
}

// Handler is the command handler
func (cmd *CommandDelete) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	apps, err := cmd.inputs.resolveApps(ui, clients.Realm)
	if err != nil {
		return err
	}

	if len(apps) == 0 {
		ui.Print(terminal.NewTextLog("No apps to delete"))
		return nil
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

	ui.Print(terminal.NewTableLog(
		fmt.Sprintf("Successfully deleted %d/%d app(s)", deletedCount, len(apps)),
		tableHeadersDelete,
		tableRowsDelete(outputs)...,
	))

	return nil
}

var (
	tableHeadersDelete = []string{headerID, headerName, headerDeleted, headerDetails}
)

func tableRowsDelete(ouputs []appOutput) []map[string]interface{} {
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
