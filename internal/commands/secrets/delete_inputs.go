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

// If there are inputs then use , then
func (i *deleteInputs) resolveSecrets(ui terminal.UI, allSecrets []realm.Secret) ([]realm.Secret, error) {
	if len(i.secrets) > 0 {
		resolvedSecrets := make([]realm.Secret, len(i.secrets))
		ids := make(map[string]realm.Secret, len(allSecrets))
		names := make(map[string]realm.Secret, len(allSecrets))
		for _, secret := range allSecrets {
			ids[secret.ID] = secret
			names[secret.Name] = secret
		}

		for i, arg := range i.secrets {
			if secret, ok := names[arg]; ok {
				resolvedSecrets[i] = secret
			} else if secret, ok := ids[arg]; ok {
				resolvedSecrets[i] = secret
			}
		}
		return resolvedSecrets, nil
	}

	selectableSecrets := map[string]realm.Secret{}
	selectableSecretOptions := make([]string, len(allSecrets))
	for i, secret := range allSecrets {
		option := displaySecretOption(secret)
		selectableSecretOptions[i] = option
		selectableSecrets[option] = secret
	}
	var selectedSecrets []string
	if err := ui.AskOne(
		&selectedSecrets,
		&survey.MultiSelect{
			Message: "Which secret(s) would you like to delete?",
			Options: selectableSecretOptions,
		},
	); err != nil {
		return nil, err
	}

	resolvedSecrets := make([]realm.Secret, len(selectedSecrets))
	for i, secret := range selectedSecrets {
		s := selectableSecrets[secret]
		resolvedSecrets[i] = s
	}
	return resolvedSecrets, nil
}
