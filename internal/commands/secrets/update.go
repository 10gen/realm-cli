package secrets

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// CommandUpdate is the `secret update` command
type CommandUpdate struct {
	inputs      updateInputs
	realmClient realm.Client
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

// Setup function for the secrets update command
func (cmd *CommandUpdate) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = profile.RealmAuthClient()
	return nil
}

// Handler function for the secrets update command
func (cmd *CommandUpdate) Handler(profile *cli.Profile, ui terminal.UI) error {
	app, err := cli.ResolveApp(ui, cmd.realmClient, cmd.inputs.Filter())
	if err != nil {
		return err
	}

	secrets, err := cmd.realmClient.Secrets(app.GroupID, app.ID)
	if err != nil {
		return err
	}

	secret, err := cmd.inputs.resolveSecret(ui, secrets)
	if err != nil {
		return err
	}

	if err := cmd.realmClient.UpdateSecret(
		app.GroupID,
		app.ID,
		secret.ID,
		cmd.inputs.name,
		cmd.inputs.value,
	); err != nil {
		return err
	}
	return ui.Print(terminal.NewTextLog("Successfully updated secret"))
}
