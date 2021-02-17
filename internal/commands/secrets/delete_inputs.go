package secrets

import (
	"errors"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
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

func (i *deleteInputs) resolveSecrets(ui terminal.UI, secrets []realm.Secret) ([]realm.Secret, error) {
	if len(i.secrets) > 0 {
		ids := make(map[string]realm.Secret, len(secrets))
		names := make(map[string]realm.Secret, len(secrets))
		for _, secret := range secrets {
			ids[secret.ID] = secret
			names[secret.Name] = secret
		}

		filtered := make([]realm.Secret, 0, len(i.secrets))
		for _, identifier := range i.secrets {
			if secret, ok := names[identifier]; ok {
				filtered = append(filtered, secret)
			} else if secret, ok := ids[identifier]; ok {
				filtered = append(filtered, secret)
			}
		}

		if len(filtered) == 0 {
			return nil, errors.New("unable to find any of the secrets")
		}
		return filtered, nil
	}

	selectableSecrets := map[string]realm.Secret{}
	selectableOptions := make([]string, len(secrets))
	for i, secret := range secrets {
		option := displaySecretOption(secret)
		selectableOptions[i] = option
		selectableSecrets[option] = secret
	}

	var selectedSecrets []string
	if err := ui.AskOne(
		&selectedSecrets,
		&survey.MultiSelect{
			Message: "Which secret(s) would you like to delete?",
			Options: selectableOptions,
		},
	); err != nil {
		return nil, err
	}

	resolvedSecrets := make([]realm.Secret, len(selectedSecrets))
	for i, secret := range selectedSecrets {
		resolvedSecrets[i] = selectableSecrets[secret]
	}

	return resolvedSecrets, nil
}
