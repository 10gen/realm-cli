package secrets

import (
	"github.com/10gen/realm-cli/internal/app"
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/spf13/pflag"
)

type CommandDelete struct {
	inputs      deleteInputs
	realmClient realm.Client
	err  error
}

type deleteInputs struct {
	app.ProjectInputs
	secretIDs []string
	secretNames []string
}

const (
	flagSecretID = "id"
	flagSecretIDShort = "i"
	flagSecretIDUsage = "set the list of secret by ID to delete"

	flagSecretName = "name"
	flagSecretNameShort = "n"
	flagSecretNameUsage = "set the list of secrets by name to delete"
)

func (cmd *CommandDelete) Inputs() cli.InputResolver {
	return &cmd.inputs
}

func (cmd *CommandDelete) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)
	fs.StringSliceVarP(&cmd.inputs.secretIDs, flagSecretID, flagSecretIDShort, []string{}, flagSecretIDUsage)
	fs.StringSliceVarP(&cmd.inputs.secretNames, flagSecretName, flagSecretNameShort, []string{}, flagSecretNameUsage)
}

func (i *deleteInputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory); err != nil {
		return err
	}
	return nil
}

func (cmd *CommandDelete) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = profile.RealmAuthClient()
	return nil
}

// TODO: validations
// 	resolve the app, user, and secret
// 	check if the user can delete the secret for this app?
func (cmd *CommandDelete) Handler(profile *cli.Profile, ui terminal.UI) error {
	app, appErr := app.Resolve(ui, cmd.realmClient, cmd.inputs.Filter())
	if appErr != nil {
		cmd.err = appErr
		return appErr
	}

	// TODO: get the list of secrets by name and ID; make a combined list of IDs to delete then delete

	deleteErr := cmd.realmClient.DeleteSecret(app.GroupID, app.ID, cmd.inputs.secretID)
	if deleteErr != nil {
		cmd.err = deleteErr
		return deleteErr
	}
	return nil
}

func (cmd *CommandDelete) Feedback(profile *cli.Profile, ui terminal.UI) error {
	if cmd.err != nil {
		return ui.Print(terminal.NewTextLog("Could not delete the secret: %s", cmd.inputs.secretID))
	}
	return ui.Print(terminal.NewTextLog("Successfully deleted secret: %s", cmd.inputs.secretID))
}
