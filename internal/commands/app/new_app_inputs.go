package app

import (
	"fmt"
	"strings"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
)

const (
	flagName      = "name"
	flagNameShort = "n"
	flagNameUsage = "set the name of the new Realm app"

	flagRemote      = "remote"
	flagRemoteUsage = "choose an application to build the new Realm app from"

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
