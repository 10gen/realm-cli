package app

import (
	appcli "github.com/10gen/realm-cli/internal/app"
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
)

const (
	flagProject      = "project"
	flagProjectUsage = "the MongoDB cloud project id"

	flagFrom      = "from"
	flagFromShort = "s"
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

type initInputs struct {
	Project         string
	From            string
	Name            string
	DeploymentModel realm.DeploymentModel
	Location        realm.Location
}

func (i *initInputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	appData, appDataErr := appcli.ResolveData(profile.WorkingDirectory)
	if appDataErr != nil {
		return appDataErr
	}
	if appData.Name != "" {
		return errProjectExists{}
	}

	if i.From == "" {
		if i.Name == "" {
			if err := ui.AskOne(&i.Name, &survey.Input{Message: "App Name"}); err != nil {
				return err
			}
		}
		if i.DeploymentModel == realm.DeploymentModelEmpty {
			i.DeploymentModel = flagDeploymentModelDefault
		}
		if i.Location == realm.LocationEmpty {
			i.Location = flagLocationDefault
		}
	}

	return nil
}

func (i *initInputs) resolveFrom(ui terminal.UI, client realm.Client) (from, error) {
	var f from

	if i.From != "" {
		a, err := appcli.Resolve(ui, client, realm.AppFilter{GroupID: i.Project, App: i.From})
		if err != nil {
			return from{}, err
		}
		f.GroupID = a.GroupID
		f.AppID = a.ID
	}

	return f, nil
}
