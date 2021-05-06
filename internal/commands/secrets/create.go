package secrets

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// CommandCreate is the `secrets create` command
type CommandCreate struct {
	inputs createInputs
}

// Flags is the command flags
func (cmd *CommandCreate) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)

	fs.StringVarP(&cmd.inputs.Name, flagName, flagNameShort, "", flagNameUsageCreate)
	fs.StringVarP(&cmd.inputs.Value, flagValue, flagValueShort, "", flagValueUsageCreate)
}

// Inputs is the command inputs
func (cmd *CommandCreate) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *CommandCreate) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	app, err := cli.ResolveApp(ui, clients.Realm, cmd.inputs.Filter())
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
