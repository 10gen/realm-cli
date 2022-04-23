package app

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

// CommandMetaDescribe is the command meta for the `app describe` command
var CommandMetaDescribe = cli.CommandMeta{
	Use:         "describe",
	Display:     "app describe",
	Description: "Displays information about your Realm app",
	HelpText: `View all of the aspects of your Realm app to see what is configured and enabled
(e.g. services, functions, etc.). If you have more than one Realm app, you will
be prompted to select a Realm app to view.`,
}

// CommandDescribe is the `app describe` command
type CommandDescribe struct {
	inputs describeInputs
}

type describeInputs struct {
	cli.ProjectInputs
}

// Flags is the command flags
func (cmd *CommandDescribe) Flags() []flags.Flag {
	return []flags.Flag{
		cli.AppFlagWithContext(&cmd.inputs.App, "to describe"),
		cli.ProjectFlag(&cmd.inputs.Project),
		cli.ProductFlag(&cmd.inputs.Products),
	}
}

// Inputs is the command inputs
func (cmd *CommandDescribe) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *CommandDescribe) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	app, err := cli.ResolveApp(ui, clients.Realm, cli.AppOptions{
		Filter:  cmd.inputs.Filter(),
		AppMeta: cmd.inputs.AppMeta,
	})
	if err != nil {
		return err
	}

	appDesc, err := clients.Realm.AppDescription(app.GroupID, app.ID)
	if err != nil {
		return err
	}

	ui.Print(terminal.NewJSONLog("App description", appDesc))
	return nil
}

func (i *describeInputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	return i.ProjectInputs.Resolve(ui, profile.WorkingDirectory, true)
}
