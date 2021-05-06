package secrets

import (
	"errors"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
)

type deleteInputs struct {
	cli.ProjectInputs
	secrets []string
}

func (i *deleteInputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory, false); err != nil {
		return err
	}
	return nil
}

// If there are inputs then use , then
func (i *deleteInputs) resolveSecrets(ui terminal.UI, appSecrets []realm.Secret) ([]realm.Secret, error) {
	if len(appSecrets) == 0 {
		return nil, nil
	}

	if len(i.secrets) > 0 {
		secretsByID := make(map[string]realm.Secret, len(appSecrets))
		secretsByName := make(map[string]realm.Secret, len(appSecrets))
		for _, secret := range appSecrets {
			secretsByID[secret.ID] = secret
			secretsByName[secret.Name] = secret
		}

		secrets := make([]realm.Secret, 0, len(i.secrets))
		for _, identifier := range i.secrets {
			if secret, ok := secretsByName[identifier]; ok {
				secrets = append(secrets, secret)
			} else if secret, ok := secretsByID[identifier]; ok {
				secrets = append(secrets, secret)
			}
		}

		if len(secrets) == 0 {
			return nil, errors.New("unable to find secrets")
		}
		return secrets, nil
	}

	options := make([]string, 0, len(appSecrets))
	secretsByOption := map[string]realm.Secret{}
	for _, secret := range appSecrets {
		option := displaySecretOption(secret)

		options = append(options, option)
		secretsByOption[option] = secret
	}

	var selections []string
	if err := ui.AskOne(
		&selections,
		&survey.MultiSelect{
			Message: "Which secret(s) would you like to delete?",
			Options: options,
		},
	); err != nil {
		return nil, err
	}

	secrets := make([]realm.Secret, 0, len(selections))
	for _, selection := range selections {
		secrets = append(secrets, secretsByOption[selection])
	}
	return secrets, nil
}
