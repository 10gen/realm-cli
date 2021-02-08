package secrets

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
)

const (
	flagSecret      = "secret"
	flagSecretShort = "s"
	flagSecretUsage = "set the list of secrets to delete by ID or Name"
)

type deleteInputs struct {
	cli.ProjectInputs
	secrets []string
}

func (i *deleteInputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory); err != nil {
		return err
	}
	return nil
}

func (i *deleteInputs) resolveDelete(allSecrets []realm.Secret, ui terminal.UI) ([]realm.Secret, error) {
	var toDelete []realm.Secret

	if len(i.secrets) == 0 {

		selectableSecrets := map[string]realm.Secret{}
		selectableSecretOptions := make([]string, len(allSecrets))
		for i, secret := range allSecrets {
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

		toDelete = make([]realm.Secret, len(selectedSecrets))

		for i, secret := range selectedSecrets {
			s := selectableSecrets[secret]
			toDelete[i] = s
		}

	} else {

		toDelete = make([]realm.Secret, len(i.secrets))

		ids := make(map[string]realm.Secret, len(allSecrets))
		names := make(map[string]realm.Secret, len(allSecrets))
		for _, secret := range allSecrets {
			ids[secret.ID] = secret
			names[secret.Name] = secret
		}

		for i, arg := range i.secrets {
			if _, ok := ids[arg]; ok {
				toDelete[i] = ids[arg]
			} else if _, ok := names[arg]; ok {
				toDelete[i] = names[arg]
			}
		}

	}

	return toDelete, nil
}