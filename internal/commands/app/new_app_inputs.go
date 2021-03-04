package app

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
)

const (
	flagProject      = "project"
	flagProjectUsage = "the MongoDB cloud project id"

	flagRemote      = "remote"
	flagRemoteUsage = "choose an application or template to initialize the Realm app with"

	flagName      = "name"
	flagNameShort = "n"
	flagNameUsage = "set the name of the Realm app to be initialized"

	flagDeploymentModel        = "deployment-model"
	flagDeploymentModelShort   = "d"
	flagDeploymentModelUsage   = `select the Realm app's deployment model, available options: ["global", "local"]`
	flagDeploymentModelDefault = realm.DeploymentModelGlobal

	flagLocation        = "location"
	flagLocationShort   = "l"
	flagLocationUsage   = `select the Realm app's location, available options: ["US-VA", "local"]`
	flagLocationDefault = realm.LocationVirginia
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
	Remote          string
	Name            string
	DeploymentModel realm.DeploymentModel
	Location        realm.Location
}

func (i *newAppInputs) resolveRemoteApp(ui terminal.UI, rc realm.Client) (appRemote, error) {
	var r appRemote
	if i.Remote != "" {
		remote, err := cli.ResolveApp(ui, rc, realm.AppFilter{App: i.Remote})
		if err != nil {
			return appRemote{}, err
		}
		r.GroupID = remote.GroupID
		r.AppID = remote.ID
	}
	return r, nil
}
