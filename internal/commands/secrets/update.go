package secrets

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// CommandMetaUpdate is the command meta for the `secrets update` command
var CommandMetaUpdate = cli.CommandMeta{
	Use:         "update",
	Display:     "secret update",
	Description: "Update a Secret in your Realm app",
	HelpText: `NOTE: The Name of the Secret cannot be modified. In order to do so, you will
need to delete and re-create the Secret.`,
}

// CommandUpdate is the `secret update` command
type CommandUpdate struct {
	inputs updateInputs
}

// Inputs function for the secrets update command
func (cmd *CommandUpdate) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Flags function for the secrets update command
func (cmd *CommandUpdate) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)
	fs.StringVarP(&cmd.inputs.secret, flagSecret, flagSecretShort, "", flagSecretUsageUpdate)
	fs.StringVarP(&cmd.inputs.name, flagName, flagNameShort, "", flagNameUsageUpdate)
	fs.StringVarP(&cmd.inputs.value, flagValue, flagValueShort, "", flagValueUsageUpdate)
}

// Handler function for the secrets update command
func (cmd *CommandUpdate) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	app, err := cli.ResolveApp(ui, clients.Realm, cmd.inputs.Filter())
	if err != nil {
		return err
	}

	secrets, err := clients.Realm.Secrets(app.GroupID, app.ID)
	if err != nil {
		return err
	}

	secret, err := cmd.inputs.resolveSecret(ui, secrets)
	if err != nil {
		return err
	}

	name := cmd.inputs.name
	if name == "" {
		name = secret.Name // when admin api _says_ patch, but never means it...
	}

	if err := clients.Realm.UpdateSecret(
		app.GroupID,
		app.ID,
		secret.ID,
		name,
		cmd.inputs.value,
	); err != nil {
		return err
	}

	ui.Print(terminal.NewTextLog("Successfully updated secret"))
	return nil
}
