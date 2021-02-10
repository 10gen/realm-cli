package app

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// CommandCreate is the `app create` command
type CommandCreate struct {
	inputs      createInputs
	atlasClient atlas.Client
	realmClient realm.Client
}

// Flags is the command flags
func (cmd *CommandCreate) Flags(fs *pflag.FlagSet) {
	fs.StringVar(&cmd.inputs.Project, flagProject, "", flagProjectUsage)
	fs.StringVarP(&cmd.inputs.Name, flagName, flagNameShort, "", flagNameUsage)
	fs.StringVarP(&cmd.inputs.Directory, flagDirectory, flagDirectoryShort, "", flagDirectoryUsage)
	fs.VarP(&cmd.inputs.DeploymentModel, flagDeploymentModel, flagDeploymentModelShort, flagDeploymentModelUsage)
	fs.VarP(&cmd.inputs.Location, flagLocation, flagLocationShort, flagLocationUsage)
	// TODO(REALMC-8135): Implement data-source flag for app create command
	// fs.StringVarP(&cmd.inputs.DataSource, flagDataSource, flagDataSourceShort, "", flagDataSourceUsage)
	// TODO(REALMC-8134): Implement dry-run for app create command
	// fs.BoolVarP(&cmd.inputs.DryRun, flagDryRun, flagDryRunShort, false, flagDryRunUsage)
}

// Inputs is the command inputs
func (cmd *CommandCreate) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Setup is the command setup
func (cmd *CommandCreate) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.atlasClient = profile.AtlasAuthClient()
	cmd.realmClient = profile.RealmAuthClient()
	return nil
}

// Handler is the command handler
func (cmd *CommandCreate) Handler(profile *cli.Profile, ui terminal.UI) error {
	from, fromErr := cmd.inputs.resolveFrom(ui, cmd.realmClient, cmd.atlasClient)
	if fromErr != nil {
		return fromErr
	}

	var projectID string
	if from.IsZero() {
		var projectIDErr error
		projectID, projectIDErr = cmd.inputs.resolveProject(ui, cmd.atlasClient)
		if projectIDErr != nil {
			return projectIDErr
		}
	} else {
		projectID = from.GroupID
	}

	appName, appNameErr := cmd.inputs.resolveAppName(ui, cmd.realmClient, from)
	if appNameErr != nil {
		return appNameErr
	}

	dir, dirErr := cmd.inputs.resolveDirectory(profile.WorkingDirectory, appName)
	if dirErr != nil {
		return dirErr
	}

	// TODO(REALMC-8134): Implement dry-run for app create command

	if from.IsZero() {
		localApp := local.NewApp(
			dir,
			appName,
			cmd.inputs.Location,
			cmd.inputs.DeploymentModel,
		)
		configErr := localApp.WriteConfig()
		if configErr != nil {
			return configErr
		}
	} else {
		_, zipPkg, exportErr := cmd.realmClient.Export(
			from.GroupID,
			from.AppID,
			realm.ExportRequest{},
		)
		if exportErr != nil {
			return exportErr
		}
		if writeErr := local.WriteZip(dir, zipPkg); writeErr != nil {
			return writeErr
		}
	}

	loadedApp, loadedAppErr := local.LoadApp(dir)
	if loadedAppErr != nil {
		return loadedAppErr
	}

	newApp, newAppErr := cmd.realmClient.CreateApp(projectID, appName, realm.AppMeta{cmd.inputs.Location, cmd.inputs.DeploymentModel})
	if newAppErr != nil {
		return newAppErr
	}

	// TODO(REALMC-8135): Implement data-source flag for app create command
	// dsCluster, dsClusterErr := cmd.inputs.resolveDataSource(cmd.realmClient, cmd.inputs.Project, app.ID())
	// if dsClusterErr != nil {
	// 	return dsClusterErr
	// }
	// service := realm.ServiceDescData{
	// 	Config: map[string]interface{}{
	// 		"clusterName":         dsCluster,
	// 		"readPreference":      "primary",
	// 		"wireProtocolEnabled": false,
	// 	},
	// 	Name: appName + "_cluster",
	// 	Type: "mongodb-atlas",
	// }
	// Potentially try to use Import to create app Service
	// _, dsErr := cmd.realmClient.CreateAppService(cmd.inputs.Project, app.ID(), service) // for output possibly
	// if dsErr != nil {
	// 	return dsErr
	// }
	// Won't need if import works
	// zipPkgName, zipPkg, exportErr := cmd.realmClient.Export(
	// 	cmd.inputs.Project,
	// 	app.ID(),
	// 	realm.ExportRequest{ConfigVersion: realm.DefaultAppConfigVersion},
	// )
	// if exportErr != nil {
	// 	return exportErr
	// }
	// if idx := strings.LastIndex(zipPkgName, "_"); idx != -1 {
	// 	zipPkgName = zipPkgName[:idx]
	// }
	// zipPkgPath := filepath.Join(dir, zipPkgName) // Is this the correct path
	// if writeErr := local.WriteZip(zipPkgPath, zipPkg); writeErr != nil {
	// 	return writeErr
	// }

	importErr := cmd.realmClient.Import(
		newApp.GroupID,
		newApp.ID,
		loadedApp,
	)
	if importErr != nil {
		return importErr
	}

	return nil

}

// Feedback is the command feedback
func (cmd *CommandCreate) Feedback(profile *cli.Profile, ui terminal.UI) error {
	return ui.Print(terminal.NewTextLog("Successfully created app"))
}
