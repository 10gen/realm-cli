package secrets

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

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
func (cmd *CommandUpdate) Handler(profile *cli.Profile, ui terminal.UI, clients cli.Clients) error {
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

	if err := clients.Realm.UpdateSecret(
		app.GroupID,
		app.ID,
		secret.ID,
		cmd.inputs.name,
		cmd.inputs.value,
	); err != nil {
		return err
	}

	ui.Print(terminal.NewTextLog("Successfully updated secret"))
	return nil
}
