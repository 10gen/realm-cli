package app

import (
	"errors"
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
)

const (
	flagRemoteAppNew      = "remote"
	flagRemoteAppNewUsage = "Specify the name or ID of a remote Realm app to clone"

	flagName      = "name"
	flagNameShort = "n"
	flagNameUsage = "Name your new Realm app"

	flagDeploymentModel        = "deployment-model"
	flagDeploymentModelShort   = "d"
	flagDeploymentModelUsage   = `Select the Realm app's deployment model (Default value: <none>; Allowed values: <none>, "GLOBAL", "LOCAL")`
	flagDeploymentModelDefault = realm.DeploymentModelGlobal

	flagLocation        = "location"
	flagLocationShort   = "l"
	flagLocationUsage   = `Select the Realm app's location (Default value: <none>; Allowed values: <none>, "US-VA", "US-OR", "DE-FF", "IE", "AU", "IN-MB", "SG")`
	flagLocationDefault = realm.LocationVirginia

	flagEnvironment      = "environment"
	flagEnvironmentShort = "e"
	flagEnvironmentUsage = `Select the Realm app's environment (Default value: <none>; Allowed values: <none>, "development", "testing", "qa", "production")`

	flagProject      = "project"
	flagProjectUsage = "Specify the ID of a MongoDB Atlas project"

	flagConfigVersion      = "config-version"
	flagConfigVersionUsage = "Specify the app config version to export as"
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

	templates, err := client.AllTemplates()
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

func (i createInputs) resolveInitialTemplateDataSource(ui terminal.UI, dataLakes []dataSourceDatalake, clusters []dataSourceCluster) (interface{}, error) {
	if len(dataLakes)+len(clusters) == 0 {
		return nil, errors.New("cannot create a template without an initial data source")
	}

	if len(dataLakes)+len(clusters) == 1 {
		if len(clusters) > 0 {
			return clusters[0], nil
		}
		return dataLakes[0], nil
	}

	// If linking multiple data sources, prompt the user for which data source to write template app schema onto
	initialTemplateDataSources := make([]interface{}, 0, len(clusters)+len(dataLakes))
	options := make([]string, 0, len(clusters)+len(dataLakes))

	for _, cluster := range clusters {
		initialTemplateDataSources = append(initialTemplateDataSources, cluster)
		options = append(options, fmt.Sprintf("[Cluster]: %s", cluster.Name))
	}
	for _, datalake := range dataLakes {
		initialTemplateDataSources = append(initialTemplateDataSources, datalake)
		options = append(options, fmt.Sprintf("[Data Lake]: %s", datalake.Name))
	}

	var selectedIndex int
	if err := ui.AskOne(
		&selectedIndex,
		&survey.Select{
			Message: "Please choose a data source to write template app schema to:",
			Options: options,
		},
	); err != nil {
		return nil, err
	}

	return initialTemplateDataSources[selectedIndex], nil
}
