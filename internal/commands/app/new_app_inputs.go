package app

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/AlecAivazis/survey/v2"
)

const (
	flagDeploymentModelDefault = realm.DeploymentModelGlobal
	flagLocationDefault        = realm.LocationVirginia
)

type newAppInputs struct {
	Project         string
	RemoteApp       string
	Name            string
	DeploymentModel realm.DeploymentModel
	Location        realm.Location
	Environment     realm.Environment
	Template        flags.OptionalString
	ConfigVersion   realm.AppConfigVersion
}

func (i *newAppInputs) resolveRemoteApp(ui terminal.UI, rc realm.Client) (realm.App, error) {
	var ra realm.App
	if i.RemoteApp != "" {
		app, err := cli.ResolveApp(ui, rc, realm.AppFilter{App: i.RemoteApp})
		if err != nil {
			return realm.App{}, err
		}
		ra = app
	}
	return ra, nil
}

func (i *newAppInputs) resolveTemplateID(ui terminal.UI, client realm.Client) error {
	if i.Template.String() == "" && ui.AutoConfirm() {
		return nil
	}

	templates, err := client.AllTemplates()
	if err != nil {
		return err
	}

	// do not disrupt application creation flow if templates are not
	// available and user is not specifying a template
	if i.Template.String() == "" && len(templates) == 0 {
		return nil
	}

	if len(templates) == 0 {
		return fmt.Errorf("unable to find template '%s'", i.Template)
	}

	if i.Template.String() != "" {
		for _, template := range templates {
			if template.ID == i.Template.String() {
				i.Template = flags.NewSetOptionalString(template.ID)
				return nil
			}
		}

		return fmt.Errorf("template '%s' not found", i.Template)
	}

	options := make([]string, 0, len(templates)+1)
	templateIDs := make([]string, 0, len(templates)+1)
	options = append(options, "[No Template]: Do Not Use A Template")
	templateIDs = append(templateIDs, "")
	for _, template := range templates {
		options = append(options, fmt.Sprintf("[%s]: %s", template.ID, template.Name))
		templateIDs = append(templateIDs, template.ID)
	}

	var selectedIndex int
	if err := ui.AskOne(
		&selectedIndex,
		&survey.Select{
			Message: "Please select a template from the available options",
			Options: options,
		},
	); err != nil {
		return err
	}

	i.Template = flags.NewSetOptionalString(templateIDs[selectedIndex])

	return nil
}
