package app

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// CommandCreate is the `app create` command
type CommandCreate struct {
	inputs createInputs
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

// Handler is the command handler
func (cmd *CommandCreate) Handler(profile *cli.Profile, ui terminal.UI, clients cli.Clients) error {
	from, err := cmd.inputs.resolveFrom(ui, clients.Realm)
	if err != nil {
		return err
	}

	var groupID = cmd.inputs.Project
	if from.IsZero() {
		if groupID == "" {
			id, err := cli.ResolveGroupID(ui, clients.Atlas)
			if err != nil {
				return err
			}
			groupID = id
		}
	} else {
		groupID = from.GroupID
	}

	err = cmd.inputs.resolveName(ui, clients.Realm, from)
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
		_, zipPkg, err := clients.Realm.Export(
			from.GroupID,
			from.AppID,
			realm.ExportRequest{},
		)
		if err != nil {
			return err
		}
		if err := local.WriteZip(dir, zipPkg); err != nil {
			return err
		}
	}

	loadedApp, err := local.LoadApp(dir)
	if err != nil {
		return err
	}

	newApp, err := clients.Realm.CreateApp(groupID, cmd.inputs.Name, realm.AppMeta{cmd.inputs.Location, cmd.inputs.DeploymentModel})
	if err != nil {
		return err
	}

	// TODO(REALMC-8135): Implement data-source flag for app create command
	// dsCluster, dsClusterErr := cmd.inputs.resolveDataSource(clients.Realm, cmd.inputs.Project, app.ID())
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
	// _, dsErr := clients.Realm.CreateAppService(cmd.inputs.Project, app.ID(), service) // for output possibly
	// if dsErr != nil {
	// 	return dsErr
	// }
	// Won't need if import works
	// zipPkgName, zipPkg, exportErr := clients.Realm.Export(
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

	if err := clients.Realm.Import(
		newApp.GroupID,
		newApp.ID,
		loadedApp,
	); err != nil {
		return err
	}

	headers := []string{"Info", "Details"}
	rows := []map[string]interface{}{
		{
			"Info":    "Client App ID",
			"Details": newApp.ClientAppID,
		},
		{
			"Info":    "Realm Directory",
			"Details": dir,
		},
		{
			"Info":    "Realm UI",
			"Details": fmt.Sprintf("%s/groups/%s/apps/%s/dashboard", profile.RealmBaseURL(), newApp.GroupID, newApp.ID),
		},
		{
			"Info":    "Check out your app",
			"Details": fmt.Sprintf("cd ./%s && realm-cli app describe", newApp.Name),
		},
	}

	ui.Print(terminal.NewTableLog("Successfully created app", headers, rows...))
	return nil
}
