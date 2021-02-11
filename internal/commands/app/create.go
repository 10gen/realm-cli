package app

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

type createOutputs struct {
	clientAppID string
	dir         string
	uiURL       string
	followUpCmd string
}

// CommandCreate is the `app create` command
type CommandCreate struct {
	inputs      createInputs
	outputs     createOutputs
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
	from, err := cmd.inputs.resolveFrom(ui, cmd.realmClient)
	if err != nil {
		return err
	}

	var groupID = cmd.inputs.Project
	if from.IsZero() {
		if groupID == "" {
			id, err := cli.ResolveGroupID(ui, cmd.atlasClient)
			if err != nil {
				return err
			}
			groupID = id
		}
	} else {
		groupID = from.GroupID
	}

	err = cmd.inputs.resolveName(ui, cmd.realmClient, from)
	if err != nil {
		return err
	}

	dir, err := cmd.inputs.resolveDirectory(profile.WorkingDirectory)
	if err != nil {
		return err
	}

	// TODO(REALMC-8134): Implement dry-run for app create command

	if from.IsZero() {
		localApp := local.NewApp(
			dir,
			cmd.inputs.Name,
			cmd.inputs.Location,
			cmd.inputs.DeploymentModel,
		)
		err := localApp.WriteConfig()
		if err != nil {
			return err
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

	newApp, newAppErr := cmd.realmClient.CreateApp(groupID, cmd.inputs.Name, realm.AppMeta{cmd.inputs.Location, cmd.inputs.DeploymentModel})
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

	err = cmd.realmClient.Import(
		newApp.GroupID,
		newApp.ID,
		loadedApp,
	)
	if err != nil {
		return err
	}

	cmd.outputs = createOutputs{
		clientAppID: newApp.ClientAppID,
		dir:         dir,
		uiURL:       fmt.Sprintf("%s/groups/%s/apps/%s/dashboard", profile.RealmBaseURL(), newApp.GroupID, newApp.ID),
		followUpCmd: fmt.Sprintf("cd ./%s && realm-cli app describe", newApp.Name),
	}

	return nil

}

// Feedback is the command feedback
func (cmd *CommandCreate) Feedback(profile *cli.Profile, ui terminal.UI) error {
	rows := make([]map[string]interface{}, 0, 4)
	rows = append(rows, map[string]interface{}{
		"Info":    "Client App ID",
		"Details": cmd.outputs.clientAppID,
	})
	rows = append(rows, map[string]interface{}{
		"Info":    "Realm Directory",
		"Details": cmd.outputs.dir,
	})
	rows = append(rows, map[string]interface{}{
		"Info":    "Realm UI",
		"Details": cmd.outputs.uiURL,
	})
	rows = append(rows, map[string]interface{}{
		"Info":    "Check out your app",
		"Details": cmd.outputs.followUpCmd,
	})
	return ui.Print(terminal.NewTableLog("Successfully created app",
		[]string{"Info", "Details"},
		rows...))
}
