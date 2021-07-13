package app

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	// TODO REALMC-9228 Re-enable prompting for template selection
	_ "github.com/AlecAivazis/survey/v2"
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
