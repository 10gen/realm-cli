package secrets

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/pflag"
	"sort"
)

type CommandDelete struct {
	inputs      deleteInputs
	realmClient realm.Client
	idToSecret  map[string]realm.Secret
	outputs     secretOutputs
}

func (cmd *CommandDelete) Inputs() cli.InputResolver {
	return &cmd.inputs
}

func (cmd *CommandDelete) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)
	fs.StringSliceVarP(&cmd.inputs.secrets, flagSecret, flagSecretShort, []string{}, flagSecretUsage)
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

func (cmd *CommandDelete) Handler(profile *cli.Profile, ui terminal.UI) error {
	app, appErr := cli.ResolveApp(ui, cmd.realmClient, cmd.inputs.Filter())
	if appErr != nil {
		return appErr
	}

	secretList, secretListErr := cmd.realmClient.Secrets(app.GroupID, app.ID)
	if secretListErr != nil {
		return secretListErr
	}

	resolve, resolveErr := cmd.resolveDelete(cmd.inputs.secrets, secretList, ui)
	if resolveErr != nil {
		return resolveErr
	}

	for _, deleteId := range resolve {
		deleteErr := cmd.realmClient.DeleteSecret(app.GroupID, app.ID, deleteId)
		cmd.outputs = append(cmd.outputs, secretOutput{
			secret: cmd.idToSecret[deleteId],
			err:    deleteErr,
		})
	}

	return nil
}

func (cmd *CommandDelete) Feedback(profile *cli.Profile, ui terminal.UI) error {
	//ui.Print(terminal.NewTextLog("Successfully deleted secret: %s", cmd.inputs.secretID))
	if len(cmd.inputs.secrets) == 0 {
		return ui.Print(terminal.NewTextLog("No secrets to delete"))
	}

	sort.SliceStable(cmd.outputs, secretOutputComparerBySuccess(cmd.outputs))
	logs := terminal.NewTableLog(
		"",
		secretHeaders(),
		secretTableRows(cmd.outputs, secretDeleteRow)...,
	)
	return ui.Print(logs)
}

func (cmd *CommandDelete) resolveDelete(args []string, secrets []realm.Secret, ui terminal.UI) ([]string, error) {
	var toDelete []string

	if len(args) != 0 {
		toDelete = make([]string, len(args))

		ids := make(map[string]realm.Secret, len(secrets))
		names := make(map[string]realm.Secret, len(secrets))
		for _, secret := range secrets {
			ids[secret.ID] = secret
			names[secret.Name] = secret
		}

		for i, arg := range args {
			if _, ok := ids[arg]; ok {
				cmd.idToSecret[arg] = ids[arg]
				toDelete[i] = arg
			} else if _, ok := names[arg]; ok {
				cmd.idToSecret[names[arg].ID] = names[arg]
				toDelete[i] = names[arg].ID
			}
		}
	} else {
		selectableSecrets := map[string]realm.Secret{}
		selectableSecretOptions := make([]string, len(secrets))
		for i, secret := range secrets {
			option := displaySecretOption(secret)
			selectableSecretOptions[i] = option
			selectableSecrets[option] = secret
		}
		var selectedSecrets []string
		askErr := ui.AskOne(
			&selectedSecrets,
			&survey.MultiSelect{
				Message: "Which secret(s) would you like to delete?",
				Options: selectableSecretOptions,
			},
		)
		if askErr != nil {
			return nil, askErr
		}

		toDelete = make([]string, len(selectedSecrets))
		for i, secret := range selectedSecrets {
			s := selectableSecrets[secret]
			cmd.idToSecret[s.ID] = s
			toDelete[i] = s.ID
		}
	}

	return toDelete, nil
}

func secretDeleteRow(output secretOutput, row map[string]interface{}) {
	if output.err != nil {
		row[headerDetails] = output.err.Error()
	} else {
		row[headerDeleted] = true
	}
}

