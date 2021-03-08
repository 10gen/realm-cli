package app

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// CommandInit is the `app init` command
type CommandInit struct {
	inputs initInputs
}

// Flags is the command flags
func (cmd *CommandInit) Flags(fs *pflag.FlagSet) {
	fs.StringVar(&cmd.inputs.Project, flagProject, "", flagProjectUsage)
	fs.StringVar(&cmd.inputs.RemoteApp, flagRemote, "", flagRemoteUsage)
	fs.StringVarP(&cmd.inputs.Name, flagName, flagNameShort, "", flagNameUsage)
	fs.VarP(&cmd.inputs.DeploymentModel, flagDeploymentModel, flagDeploymentModelShort, flagDeploymentModelUsage)
	fs.VarP(&cmd.inputs.Location, flagLocation, flagLocationShort, flagLocationUsage)
}

// Inputs is the command inputs
func (cmd *CommandInit) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *CommandInit) Handler(profile *cli.Profile, ui terminal.UI, clients cli.Clients) error {
	appRemote, err := cmd.inputs.resolveRemoteApp(ui, clients.Realm)
	if err != nil {
		return err
	}

	if appRemote.IsZero() {
		if err := cmd.writeAppFromScratch(profile.WorkingDirectory); err != nil {
			return err
		}
	} else {
		if err := cmd.writeAppFromExisting(profile.WorkingDirectory, clients.Realm, appRemote.GroupID, appRemote.AppID); err != nil {
			return err
		}
	}

	ui.Print(terminal.NewTextLog("Successfully initialized app"))
	return nil
}

func (cmd *CommandInit) writeAppFromScratch(wd string) error {
	/*
		TODO(REALMC-7886): initialize also the following:
			- auth/custom_user_data.json: { "enabled": false }
			- auth/providers.json: {}
			- data_sources/
			- http_endpoints/
			- sync/config.json: { "development_mode_enabled": false }

		this logic probably wants to live in local.App, where ConfigVersion actually determines what gets written/initialized
	*/
	return local.NewApp(wd,
		"", // no app id yet
		cmd.inputs.Name,
		cmd.inputs.Location,
		cmd.inputs.DeploymentModel,
	).WriteConfig()
}

func (cmd *CommandInit) writeAppFromExisting(wd string, realmClient realm.Client, groupID, appID string) error {
	_, zipPkg, err := realmClient.Export(groupID, appID, realm.ExportRequest{IsTemplated: true})
	if err != nil {
		return err
	}

	return local.WriteZip(wd, zipPkg)
}
