package app

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/spf13/pflag"
)

// CommandMetaCreate is the command meta for the `app create` command
var CommandMetaCreate = cli.CommandMeta{
	Use:         "create",
	Display:     "app create",
	Description: "Create a new app from your current working directory and deploy it to the Realm server",
	HelpText: `Creates a new Realm app by saving your configuration files in a local directory
and deploying the new app to the Realm server. This command will create a new
directory for your project.

You can specify a "--remote" flag to create a Realm app from an existing app;
if you do not specify a "--remote" flag, the CLI will create a default Realm app.

NOTE: To create a Realm app without deploying it, use "app init".`,
}

// CommandCreate is the `app create` command
type CommandCreate struct {
	inputs createInputs
}

// Flags is the command flags
func (cmd *CommandCreate) Flags(fs *pflag.FlagSet) {
	fs.StringVar(&cmd.inputs.LocalPath, flagLocalPathCreate, "", flagLocalPathCreateUsage)
	fs.StringVarP(&cmd.inputs.Name, flagName, flagNameShort, "", flagNameUsage)
	fs.StringVar(&cmd.inputs.RemoteApp, flagRemote, "", flagRemoteUsage)
	fs.VarP(&cmd.inputs.Location, flagLocation, flagLocationShort, flagLocationUsage)
	fs.VarP(&cmd.inputs.DeploymentModel, flagDeploymentModel, flagDeploymentModelShort, flagDeploymentModelUsage)
	fs.VarP(&cmd.inputs.Environment, flagEnvironment, flagEnvironmentShort, flagEnvironmentUsage)
	fs.StringVar(&cmd.inputs.Cluster, flagCluster, "", flagClusterUsage)
	fs.StringVar(&cmd.inputs.DataLake, flagDataLake, "", flagDataLakeUsage)
	fs.BoolVarP(&cmd.inputs.DryRun, flagDryRun, flagDryRunShort, false, flagDryRunUsage)

	fs.StringVar(&cmd.inputs.Project, flagProject, "", flagProjectUsage)
	flags.MarkHidden(fs, flagProject)

	fs.Var(&cmd.inputs.ConfigVersion, flagConfigVersion, flagConfigVersionUsage)
	flags.MarkHidden(fs, flagConfigVersion)
}

// Inputs is the command inputs
func (cmd *CommandCreate) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *CommandCreate) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	appRemote, err := cmd.inputs.resolveRemoteApp(ui, clients.Realm)
	if err != nil {
		return err
	}

	groupID := cmd.inputs.Project

	if groupID == "" {
		id, err := cli.ResolveGroupID(ui, clients.Atlas, appRemote.GroupID)
		if err != nil {
			return err
		}
		groupID = id
	}
	err = cmd.inputs.resolveName(ui, clients.Realm, appRemote.GroupID, appRemote.ClientAppID)
	if err != nil {
		return err
	}

	dir, err := cmd.inputs.resolveLocalPath(ui, profile.WorkingDirectory)
	if err != nil {
		return err
	}

	var dsCluster dataSourceCluster
	if cmd.inputs.Cluster != "" {
		dsCluster, err = cmd.inputs.resolveCluster(clients.Atlas, groupID)
		if err != nil {
			return err
		}
	}

	var dsDataLake dataSourceDataLake
	if cmd.inputs.DataLake != "" {
		dsDataLake, err = cmd.inputs.resolveDataLake(clients.Atlas, groupID)
		if err != nil {
			return err
		}
	}

	if cmd.inputs.DryRun {
		logs := make([]terminal.Log, 0, 4)
		if appRemote.IsZero() {
			logs = append(logs, terminal.NewTextLog("A minimal Realm app would be created at %s", dir))
		} else {
			logs = append(logs, terminal.NewTextLog("A Realm app based on the Realm app '%s' would be created at %s", cmd.inputs.RemoteApp, dir))
		}
		if dsCluster.Name != "" {
			logs = append(logs, terminal.NewTextLog("The cluster '%s' would be linked as data source '%s'", cmd.inputs.Cluster, dsCluster.Name))
		}
		if dsDataLake.Name != "" {
			logs = append(logs, terminal.NewTextLog("The data lake '%s' would be linked as data source '%s'", cmd.inputs.DataLake, dsDataLake.Name))
		}
		logs = append(logs, terminal.NewFollowupLog("To create this app run", cmd.display(true)))
		ui.Print(logs...)
		return nil
	}

	appRealm, err := clients.Realm.CreateApp(
		groupID,
		cmd.inputs.Name,
		realm.AppMeta{cmd.inputs.Location, cmd.inputs.DeploymentModel, cmd.inputs.Environment},
	)
	if err != nil {
		return err
	}

	var appLocal local.App

	if appRemote.IsZero() {
		appLocal = local.NewApp(
			dir,
			appRealm.ClientAppID,
			cmd.inputs.Name,
			cmd.inputs.Location,
			cmd.inputs.DeploymentModel,
			cmd.inputs.Environment,
			cmd.inputs.ConfigVersion,
		)
		local.AddAuthProvider(appLocal.AppData, "api-key", map[string]interface{}{
			"name":     "api-key",
			"type":     "api-key",
			"disabled": true,
		})
	} else {
		_, zipPkg, err := clients.Realm.Export(
			appRemote.GroupID,
			appRemote.ID,
			realm.ExportRequest{},
		)
		if err != nil {
			return err
		}
		if err := local.WriteZip(dir, zipPkg); err != nil {
			return err
		}
		appLocal, err = local.LoadApp(dir)
		if err != nil {
			return err
		}
	}

	if dsCluster.Name != "" {
		local.AddDataSource(appLocal.AppData, map[string]interface{}{
			"name": dsCluster.Name,
			"type": dsCluster.Type,
			"config": map[string]interface{}{
				"clusterName":         dsCluster.Config.ClusterName,
				"readPreference":      dsCluster.Config.ReadPreference,
				"wireProtocolEnabled": dsCluster.Config.WireProtocolEnabled,
			},
		})
	}
	if dsDataLake.Name != "" {
		local.AddDataSource(appLocal.AppData, map[string]interface{}{
			"name": dsDataLake.Name,
			"type": dsDataLake.Type,
			"config": map[string]interface{}{
				"dataLakeName": dsDataLake.Config.DataLakeName,
			},
		})
	}

	if err := appLocal.Write(); err != nil {
		return err
	}

	if err := appLocal.Load(); err != nil {
		return err
	}

	if err := clients.Realm.Import(appRealm.GroupID, appRealm.ID, appLocal.AppData); err != nil {
		return err
	}

	headers := []string{"Info", "Details"}
	rows := make([]map[string]interface{}, 0, 5)
	rows = append(rows, map[string]interface{}{"Info": "Client App ID", "Details": appRealm.ClientAppID})
	rows = append(rows, map[string]interface{}{"Info": "Realm Directory", "Details": dir})
	rows = append(rows, map[string]interface{}{"Info": "Realm UI", "Details": fmt.Sprintf("%s/groups/%s/apps/%s/dashboard", profile.RealmBaseURL(), appRealm.GroupID, appRealm.ID)})
	if dsCluster.Name != "" {
		rows = append(rows, map[string]interface{}{"Info": "Data Source (Cluster)", "Details": dsCluster.Name})
	}
	if dsDataLake.Name != "" {
		rows = append(rows, map[string]interface{}{"Info": "Data Source (Data Lake)", "Details": dsDataLake.Name})
	}

	ui.Print(terminal.NewTableLog("Successfully created app", headers, rows...))
	ui.Print(terminal.NewFollowupLog("Check out your app", fmt.Sprintf("cd ./%s && %s app describe", cmd.inputs.LocalPath, cli.Name)))
	return nil
}

func (cmd *CommandCreate) display(omitDryRun bool) string {
	return cli.CommandDisplay(CommandMetaCreate.Display, cmd.inputs.args(omitDryRun))
}
