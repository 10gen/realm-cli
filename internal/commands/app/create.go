package app

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// set of supported `app create` command strings
const (
	CommandCreateDisplay = "app create"
)

// CommandCreate is the `app create` command
type CommandCreate struct {
	inputs createInputs
}

// Flags is the command flags
func (cmd *CommandCreate) Flags(fs *pflag.FlagSet) {
	fs.StringVar(&cmd.inputs.Project, flagProject, "", flagProjectUsage)
	fs.StringVarP(&cmd.inputs.Name, flagName, flagNameShort, "", flagNameUsage)
	fs.StringVarP(&cmd.inputs.From, flagFrom, flagFromShort, "", flagFromUsage)
	fs.StringVarP(&cmd.inputs.Directory, flagDirectory, flagDirectoryShort, "", flagDirectoryUsage)
	fs.VarP(&cmd.inputs.DeploymentModel, flagDeploymentModel, flagDeploymentModelShort, flagDeploymentModelUsage)
	fs.VarP(&cmd.inputs.Location, flagLocation, flagLocationShort, flagLocationUsage)
	fs.StringVar(&cmd.inputs.Cluster, flagCluster, "", flagClusterUsage)
	fs.StringVar(&cmd.inputs.DataLake, flagDataLake, "", flagDataLakeUsage)
	fs.BoolVarP(&cmd.inputs.DryRun, flagDryRun, flagDryRunShort, false, flagDryRunUsage)
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
		logs := make([]terminal.Log, 0, 5)
		if from.IsZero() {
			logs = append(logs, terminal.NewTextLog("A minimal Realm app would be created at %s", dir))
		} else {
			logs = append(logs, terminal.NewTextLog("A Realm app based on the Realm app %s would be created at %s", cmd.inputs.From, dir))
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

	if dsCluster.Name != "" || dsDataLake.Name != "" {
		var dataSourceDir string
		switch loadedApp.ConfigVersion() {
		case
			realm.AppConfigVersion20200603,
			realm.AppConfigVersion20180301:
			dataSourceDir = local.NameServices
		default:
			dataSourceDir = local.NameDataSources
		}
		if dsCluster.Name != "" {
			data, err := local.MarshalJSON(dsCluster)
			if err != nil {
				return err
			}
			path := filepath.Join(dataSourceDir, dsCluster.Name, local.FileConfig.String())
			err = local.WriteFile(filepath.Join(dir, path), 0666, bytes.NewReader(data))
			if err != nil {
				return err
			}
		}
		if dsDataLake.Name != "" {
			data, err := local.MarshalJSON(dsDataLake)
			if err != nil {
				return err
			}
			path := filepath.Join(dataSourceDir, dsDataLake.Name, local.FileConfig.String())
			err = local.WriteFile(filepath.Join(dir, path), 0666, bytes.NewReader(data))
			if err != nil {
				return err
			}
		}
		if err = loadedApp.Load(); err != nil {
			return err
		}
	}

	if err := clients.Realm.Import(newApp.GroupID, newApp.ID, loadedApp.AppData); err != nil {
		return err
	}

	headers := []string{"Info", "Details"}
	rows := make([]map[string]interface{}, 0, 5)
	rows = append(rows, map[string]interface{}{"Info": "Client App ID", "Details": newApp.ClientAppID})
	rows = append(rows, map[string]interface{}{"Info": "Realm Directory", "Details": dir})
	rows = append(rows, map[string]interface{}{"Info": "Realm UI", "Details": fmt.Sprintf("%s/groups/%s/apps/%s/dashboard", profile.RealmBaseURL(), newApp.GroupID, newApp.ID)})
	if dsCluster.Name != "" {
		rows = append(rows, map[string]interface{}{"Info": "Data Source (Cluster)", "Details": dsCluster.Name})
	}
	if dsDataLake.Name != "" {
		rows = append(rows, map[string]interface{}{"Info": "Data Source (Data Lake)", "Details": dsDataLake.Name})
	}

	ui.Print(terminal.NewTableLog("Successfully created app", headers, rows...))
	ui.Print(terminal.NewFollowupLog("Check out your app", fmt.Sprintf("cd ./%s && %s app describe", cmd.inputs.Directory, cli.Name)))
	return nil
}

func (cmd *CommandCreate) display(omitDryRun bool) string {
	return cli.CommandDisplay(CommandCreateDisplay, cmd.inputs.args(omitDryRun))
}
