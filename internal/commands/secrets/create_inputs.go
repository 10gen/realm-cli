package secrets

import (
	"github.com/10gen/realm-cli/internal/app"
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/AlecAivazis/survey/v2"
)

type createInputs struct {
	app.ProjectInputs
	Name  string
	Value string
}

func (i *createInputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	if i.Name == "" {
		if err := ui.AskOne(&i.Name, &survey.Input{Message: "Secret Name"}); err != nil {
			return err
		}
	}
	if i.Value == "" {
		if err := ui.AskOne(&i.Value, &survey.Input{Message: "Secret Value"}); err != nil {
			return err
		}
	}

	return nil
}
