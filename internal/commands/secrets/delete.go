package secrets

import (
	"sort"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// CommandDelete for the secrets delete command
type CommandDelete struct {
	inputs      deleteInputs
	realmClient realm.Client
	outputs     secretOutputs
}

// Inputs function for the secrets delete command
func (cmd *CommandDelete) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Flags function for the secrets delete command
func (cmd *CommandDelete) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)
	fs.StringSliceVarP(&cmd.inputs.secrets, flagSecret, flagSecretShort, []string{}, flagSecretUsage)
}

// Setup function for the secrets delete command
func (cmd *CommandDelete) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = profile.RealmAuthClient()
	return nil
}

// Handler function for the secrets delete command
func (cmd *CommandDelete) Handler(profile *cli.Profile, ui terminal.UI) error {
	app, appErr := cli.ResolveApp(ui, cmd.realmClient, cmd.inputs.Filter())
	if appErr != nil {
		return appErr
	}

	secretList, secretListErr := cmd.realmClient.Secrets(app.GroupID, app.ID)
	if secretListErr != nil {
		return secretListErr
	}

	toDelete, resolveErr := cmd.inputs.resolveSecrets(ui, secretList)
	if resolveErr != nil {
		return resolveErr
	}

	for _, secret := range toDelete {
		err := cmd.realmClient.DeleteSecret(app.GroupID, app.ID, secret.ID)
		cmd.outputs = append(cmd.outputs, secretOutput{secret: secret, err: err})
	}

	return nil
}

// Feedback function for the secrets delete command
func (cmd *CommandDelete) Feedback(profile *cli.Profile, ui terminal.UI) error {
	if len(cmd.inputs.secrets) == 0 {
		return ui.Print(terminal.NewTextLog("No secrets to delete"))
	}

	sort.SliceStable(cmd.outputs, secretOutputComparerBySuccess(cmd.outputs))
	logs := terminal.NewTableLog(
		"Deleted Secrets",
		secretHeaders(headerDeleted, headerDetails),
		secretTableRows(cmd.outputs, secretDeleteRow)...,
	)
	return ui.Print(logs)
}

func secretDeleteRow(output secretOutput, row map[string]interface{}) {
	deleted := false
	if output.err != nil {
		row[headerDetails] = output.err.Error()
	} else {
		deleted = true
	}
	row[headerDeleted] = deleted
}
