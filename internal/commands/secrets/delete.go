package secrets

import (
	"github.com/10gen/realm-cli/internal/app"
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/spf13/pflag"
	"math"
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

func intersection(a []string, b[]string) []string {
	hash := make(map[string]interface{}, len(a))
	for _, a1 := range a {
		hash[a1] = nil
	}
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}

	result := make([]string, maxLen)
	for _, b1 := range b {
		if _, ok := hash[b1]; ok {
			result = append(result, b1)
		}
	}
	return result
}

func getIDsFromNames(secrets []realm.Secret, names []string) []string{
	nameMap := make(map[string]interface{}, len(names ))
	for _, name := range names {
		nameMap[name] = nil
	}
	result := make([]string, len(names))
	for _, secret := range secrets {
		if _, ok := nameMap[secret.Name]; ok {
			result = append(result, secret.ID)
		}
	}
	return result
}

func getNames(secrets []realm.Secret) []string {
	result := make([]string, len(secrets))
	for _, secret:= range secrets {
		result = append(result, secret.Name)
	}
	return result
}

func getIDs(secrets []realm.Secret) []string {
	result := make([]string, len(secrets))
	for _, secret:= range secrets {
		result = append(result, secret.ID)
	}
	return result
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
	
	secretList, secretListErr := cmd.realmClient.Secrets(app.GroupID, app.ID)
	if secretListErr != nil {
		cmd.err = secretListErr
		return secretListErr
	}

	toDelete := make([]string, len(cmd.inputs.secretNames) + len(cmd.inputs.secretIDs))
	// TODO: do i only delete the IDs that are valid or reject the entire request if there's an invalid ID/name
	// 	is this the resolve step  ??
	if len(cmd.inputs.secretIDs) != 0 {
		toDelete = intersection(cmd.inputs.secretIDs,  getIDs(secretList))
	}

	if len(cmd.inputs.secretNames) != 0 {
		toDelete = intersection(toDelete, getIDsFromNames(secretList, cmd.inputs.secretNames))
	}

	for _, secret := range secretList {
		secret.ID
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
