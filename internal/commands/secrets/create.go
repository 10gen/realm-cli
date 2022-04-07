package secrets

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

// CommandMetaCreate is the command meta for the `secrets create` command
var CommandMetaCreate = cli.CommandMeta{
	Use:         "create",
	Display:     "secrets create",
	Description: "Create a Secret for your Realm app",
	HelpText:    `You will be prompted to name your Secret and define the value of your Secret.`,
}

// CommandCreate is the `secrets create` command
type CommandCreate struct {
	inputs createInputs
}

// Flags is the command flags
func (cmd *CommandCreate) Flags() []flags.Flag {
	return []flags.Flag{
		cli.AppFlagWithContext(&cmd.inputs.App, "to create its secrets"),
		cli.ProjectFlag(&cmd.inputs.Project),
		cli.ProductFlag(&cmd.inputs.Products),
		nameFlag(&cmd.inputs.Name, "Name the secret"),
		valueFlag(&cmd.inputs.Value, "Specify the secret value"),
	}
}

// Inputs is the command inputs
func (cmd *CommandCreate) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *CommandCreate) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	app, err := cli.ResolveApp(ui, clients.Realm, cli.AppOptions{
		AppMeta: cmd.inputs.AppMeta,
		Filter:  cmd.inputs.Filter(),
	})
	if err != nil {
		return err
	}

	secret, err := clients.Realm.CreateSecret(app.GroupID, app.ID, cmd.inputs.Name, cmd.inputs.Value)
	if err != nil {
		return err
	}

	ui.Print(terminal.NewTextLog("Successfully created secret, id: %s", secret.ID))
	return nil
}
