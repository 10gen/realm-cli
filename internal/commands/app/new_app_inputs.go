package app

import (
	"fmt"
	"strings"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
)

const (
	flagName      = "name"
	flagNameShort = "n"
	flagNameUsage = "set the name of the new Realm app"

	flagRemoteAppNew      = "remote"
	flagRemoteAppNewUsage = "choose an application to build the new Realm app from"

	flagDeploymentModel        = "deployment-model"
	flagDeploymentModelShort   = "d"
	flagDeploymentModelDefault = realm.DeploymentModelGlobal

	flagLocation        = "location"
	flagLocationShort   = "l"
	flagLocationDefault = realm.LocationVirginia

	flagEnvironment      = "environment"
	flagEnvironmentShort = "e"

	flagProject      = "project"
	flagProjectUsage = "the MongoDB cloud project id"

	flagConfigVersion      = "config-version"
	flagConfigVersionUsage = "the config version of the Realm app structure; defaults to latest stable config version"
)

var (
	flagDeploymentModelUsage = fmt.Sprintf("select the Realm app's deployment model, available options: [%s]", strings.Join(realm.DeploymentModelValues, ", "))
	flagLocationUsage        = fmt.Sprintf("select the Realm app's location, available options: [%s]", strings.Join(realm.LocationValues, ", "))
	flagEnvironmentUsage     = fmt.Sprintf("select the Realm app's environment, available options: [%s]", strings.Join(realm.EnvironmentValues, ", "))
)

type newAppInputs struct {
	Project         string
	RemoteApp       string
	Name            string
	DeploymentModel realm.DeploymentModel
	Location        realm.Location
	Environment     realm.Environment
	Template        string
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
	if i.Template == "" && ui.AutoConfirm() {
		return nil
	}

	templates, err := client.Templates()
	if err != nil {
		return err
	}

	// do not disrupt application creation flow if templates are not
	// available and user is not specifying a template
	if i.Template == "" && len(templates) == 0 {
		return nil
	}

	if len(templates) == 0 {
		return fmt.Errorf("unable to find template '%s'", i.Template)
	}

	if i.Template != "" {
		for _, template := range templates {
			if template.ID == i.Template {
				i.Template = template.ID
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

	i.Template = templateIDs[selectedIndex]

	return nil
}
