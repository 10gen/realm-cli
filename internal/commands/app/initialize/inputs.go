package initialize

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
)

type from struct {
	GroupID string
	AppID   string
}

func (f from) IsZero() bool {
	return f.GroupID == "" && f.AppID == ""
}

type inputs struct {
	AppData         cli.AppData
	Project         string
	From            string
	Name            string
	DeploymentModel realm.DeploymentModel
	Location        realm.Location
}

func (i *inputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	appData, appDataErr := cli.ResolveAppData(profile.WorkingDirectory)
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
		if i.DeploymentModel == realm.DeploymentModelNil {
			i.DeploymentModel = flagDeploymentModelDefault
		}
		if i.Location == realm.LocationNil {
			i.Location = flagLocationDefault
		}
	}

	return nil
}

func (i *inputs) resolveFrom(ui terminal.UI, client realm.Client) (from, error) {
	var f from

	if i.From != "" {
		app, err := cli.ResolveApp(ui, client, realm.AppFilter{GroupID: i.Project, App: i.From})
		if err != nil {
			return from{}, err
		}
		f.GroupID = app.GroupID
		f.AppID = app.ID
	}

	return f, nil
}
