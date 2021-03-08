package app

import (
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
	flagDeploymentModelUsage   = `select the Realm app's deployment model, available options: ["GLOBAL", "LOCAL"]`
	flagDeploymentModelDefault = realm.DeploymentModelGlobal

	flagLocation        = "location"
	flagLocationShort   = "l"
	flagLocationUsage   = `select the Realm app's location, available options: ["US-VA", "local"]`
	flagLocationDefault = realm.LocationVirginia

	flagProject      = "project"
	flagProjectUsage = "the MongoDB cloud project id"

	flagConfigVersion      = "config-version"
	flagConfigVersionUsage = "the config version of the Realm app structure; defaults to latest stable config version"
)

type appRemote struct {
	GroupID string
	AppID   string
}

func (r appRemote) IsZero() bool {
	return r.GroupID == "" && r.AppID == ""
}

type newAppInputs struct {
	Project         string
	RemoteApp       string
	Name            string
	DeploymentModel realm.DeploymentModel
	Location        realm.Location
	ConfigVersion   realm.AppConfigVersion
}

func (i *newAppInputs) resolveRemoteApp(ui terminal.UI, rc realm.Client) (appRemote, error) {
	var r appRemote
	if i.RemoteApp != "" {
		app, err := cli.ResolveApp(ui, rc, realm.AppFilter{App: i.RemoteApp})
		if err != nil {
			return appRemote{}, err
		}
		r.GroupID = app.GroupID
		r.AppID = app.ID
	}
	return r, nil
}
