package app

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// CommandInit is the `app init` command
type CommandInit struct {
	inputs      initInputs
	atlasClient atlas.Client
	realmClient realm.Client
}

// Flags is the command flags
func (cmd *CommandInit) Flags(fs *pflag.FlagSet) {
	fs.StringVar(&cmd.inputs.Project, flagProject, "", flagProjectUsage)
	fs.StringVarP(&cmd.inputs.From, flagFrom, flagFromShort, "", flagFromUsage)
	fs.StringVarP(&cmd.inputs.Name, flagName, flagNameShort, "", flagNameUsage)
	fs.VarP(&cmd.inputs.DeploymentModel, flagDeploymentModel, flagDeploymentModelShort, flagDeploymentModelUsage)
	fs.VarP(&cmd.inputs.Location, flagLocation, flagLocationShort, flagLocationUsage)
}

// Inputs is the command inputs
func (cmd *CommandInit) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Setup is the command setup
func (cmd *CommandInit) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.atlasClient = profile.AtlasAuthClient()
	cmd.realmClient = profile.RealmAuthClient()
	return nil
}

// Handler is the command handler
func (cmd *CommandInit) Handler(profile *cli.Profile, ui terminal.UI) error {
	from, fromErr := cmd.inputs.resolveFrom(ui, cmd.realmClient, cmd.atlasClient)
	if fromErr != nil {
		return fromErr
	}

	if from.IsZero() {
		/*
			TODO(REALMC-7886): initialize also the following:
			  - auth/custom_user_data.json: { "enabled": false }
			  - auth/providers.json: {}
			  - data_sources/
			  - http_endpoints/
			  - sync/config.json: { "development_mode_enabled": false }

			this logic probably wants to live in local.App, where ConfigVersion actually determines what gets written/initialized
		*/
		return local.NewApp(
			profile.WorkingDirectory,
			cmd.inputs.Name,
			cmd.inputs.Location,
			cmd.inputs.DeploymentModel,
		).WriteConfig()
	}

	_, zipPkg, exportErr := cmd.realmClient.Export(
		from.GroupID,
		from.AppID,
		realm.ExportRequest{IsTemplated: true},
	)
	if exportErr != nil {
		return exportErr
	}
	return local.WriteZip(profile.WorkingDirectory, zipPkg)
}

// Feedback is the command feedback
func (cmd *CommandInit) Feedback(profile *cli.Profile, ui terminal.UI) error {
	return ui.Print(terminal.NewTextLog("Successfully initialized app"))
}
