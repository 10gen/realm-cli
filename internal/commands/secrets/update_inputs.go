package secrets

import (
	"errors"
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
)

type updateInputs struct {
	cli.ProjectInputs
	secret string
	name   string
	value  string
}

func (i *updateInputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory, false); err != nil {
		return err
	}

	if i.name == "" && i.value == "" {
		return errors.New("must set either --name or --value when updating a secret")
	}

	return nil
}

func (i *updateInputs) resolveSecret(ui terminal.UI, secrets []realm.Secret) (realm.Secret, error) {

	if len(i.secret) > 0 {
		for _, secret := range secrets {
			if secret.ID == i.secret || secret.Name == i.secret {
				return secret, nil
			}
		}
		return realm.Secret{}, fmt.Errorf("unable to find secret: %s", i.secret)
	}

	selectableSecrets := map[string]realm.Secret{}
	selectableOptions := make([]string, len(secrets))
	for i, secret := range secrets {
		option := displaySecretOption(secret)
		selectableOptions[i] = option
		selectableSecrets[option] = secret
	}

	var selected string
	if err := ui.AskOne(
		&selected,
		&survey.Select{
			Message: "Which secret would you like to update?",
			Options: selectableOptions,
		},
	); err != nil {
		return realm.Secret{}, err
	}

	return selectableSecrets[selected], nil
}
