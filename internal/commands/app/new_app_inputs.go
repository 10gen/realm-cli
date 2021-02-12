package app

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
)

const (
	flagProject      = "project"
	flagProjectUsage = "the MongoDB cloud project id"

	flagFrom      = "from"
	flagFromShort = "a"
	flagFromUsage = "choose an application or template to initialize the Realm app with"

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

type from struct {
	GroupID string
	AppID   string
}

func (f from) IsZero() bool {
	return f.GroupID == "" && f.AppID == ""
}

type newAppInputs struct {
	Project         string
	From            string
	Name            string
	DeploymentModel realm.DeploymentModel
	Location        realm.Location
}

func (i *newAppInputs) resolveFrom(ui terminal.UI, rc realm.Client) (from, error) {
	var f from
	if i.From != "" {
		app, err := cli.ResolveApp(ui, rc, realm.AppFilter{App: i.From})
		if err != nil {
			return from{}, err
		}
		f.GroupID = app.GroupID
		f.AppID = app.ID
	}
	return f, nil
}
