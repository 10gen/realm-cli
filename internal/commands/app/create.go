package app

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

var (
	flagDirectory      = "app-dir"
	flagDirectoryShort = "c"
	flagDirectoryUsage = "the directory to create your new Realm app, defaults to Realm app name"

	flagDataSource      = "data-source"
	flagDataSourceShort = "s"
	flagDataSourceUsage = "atlas cluster to back your Realm app, defaults to first available"

	flagDryRun      = "dry-run"
	flagDryRunShort = "x"
	flagDryRunUsage = "include to run without writing any changes to the file system or import/export the new Realm app" // Does this sound correct?

	flagAppVersion      = "app-version"
	flagAppVersionUsage = "specify the app config version to pull changes down as"
)

type createInputs struct {
	initInputs
	Directory  string
	DataSource string
	DryRun     bool
	AppVersion realm.AppConfigVersion
}

func (i *createInputs) resolveDirectory(client realm.Client, profile *cli.Profile) (string, error) {
	dir := i.Directory
	if dir == "" {
		dir = i.Name
	}
	fullPath := path.Join(profile.WorkingDirectory, dir)
	fi, statErr := os.Stat(fullPath)
	if statErr != nil {
		return "", statErr
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		_, appOK, appErr := local.FindApp(fullPath)
		if appErr != nil { // do we care about this case?
			return "", appErr
		}
		if appOK {
			return "", fmt.Errorf("A Realm app already exists at %s", fullPath)
		}
		return dir, nil
	}
	return dir, nil
}

func (i *createInputs) resolveDataSource(client realm.Client, groupID, appID string) (string, error) {
	clusters, err := client.ListClusters(groupID, appID)
	if err != nil {
		return "", err
	}
	var dsCluster string
	for _, cluster := range clusters {
		if (i.DataSource == "" && cluster.State == "IDLE") || i.DataSource == cluster.Name {
			dsCluster = cluster.Name
			break
		}
	}
	if dsCluster == "" {
		if i.DataSource != "" {
			return "", fmt.Errorf("Unable to find the %s cluster", i.DataSource)
		}
		return "", fmt.Errorf("Unable to find any available cluster for Group ID %s", groupID) // Probably better to use the group/project name?
	}
	return dsCluster, nil
}

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
	fs.StringVarP(&cmd.inputs.DataSource, flagDataSource, flagDataSourceShort, "", flagDataSourceUsage)
	fs.BoolVarP(&cmd.inputs.DryRun, flagDryRun, flagDryRunShort, false, flagDryRunUsage)

	fs.Var(&cmd.inputs.AppVersion, flagAppVersion, flagAppVersionUsage) // I think I need this?
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
	if cmd.inputs.Project == "" {
		groupID, groupErr := cli.ResolveGroupID(ui, cmd.atlasClient)
		if groupErr != nil {
			return groupErr
		}
		cmd.inputs.Project = groupID
	}

	from, fromErr := cmd.inputs.resolveFrom(ui, cmd.realmClient)
	if fromErr != nil {
		return fromErr
	}

	apps, appsErr := cmd.realmClient.FindApps(realm.AppFilter{from.GroupID, from.AppID})
	if appsErr != nil {
		return appsErr
	}
	if len(apps) == 1 {
		cmd.inputs.Name = apps[0].Name
	}

	dir, dirErr := cmd.inputs.resolveDirectory(cmd.realmClient, profile)
	if dirErr != nil {
		return dirErr
	}

	if cmd.inputs.DryRun {
		if cmd.inputs.DataSource != "" {
			// verify data source
			// do we need to use atlas client to do this? since there is not an app id for the realm route
		}
		// Do we verify if data source is not specified?
		return nil
	}

	if from.IsZero() {
		localApp := local.NewApp(
			dir,
			cmd.inputs.Name,
			cmd.inputs.Location,
			cmd.inputs.DeploymentModel,
		)
		configErr := localApp.WriteConfig()
		if configErr != nil {
			return configErr
		}
		newApp, newAppErr := cmd.realmClient.CreateApp(cmd.inputs.Project, cmd.inputs.Name, realm.AppMeta{cmd.inputs.Location, cmd.inputs.DeploymentModel})
		if newAppErr != nil {
			return newAppErr
		}
		importErr := cmd.realmClient.Import(
			newApp.GroupID,
			newApp.ID,
			localApp,
		)
		if importErr != nil {
			return importErr
		}
	} else {
		var appVersion realm.AppConfigVersion
		if cmd.inputs.AppVersion == realm.AppConfigVersionZero {
			appVersion = realm.AppConfigVersion20210101
		}
		zipPkgName, zipPkg, exportErr := cmd.realmClient.Export(
			from.GroupID,
			from.AppID,
			realm.ExportRequest{ConfigVersion: appVersion},
		)
		if exportErr != nil {
			return exportErr
		}
		if idx := strings.LastIndex(zipPkgName, "_"); idx != -1 {
			zipPkgName = zipPkgName[:idx]
		}
		zipPkgPath := filepath.Join(dir, zipPkgName) // Is this the correct path
		if writeErr := local.WriteZip(zipPkgPath, zipPkg); writeErr != nil {
			return writeErr
		}
	}

	app, appErr := local.LoadApp(dir)
	if appErr != nil {
		return appErr
	}
	dsCluster, dsClusterErr := cmd.inputs.resolveDataSource(cmd.realmClient, cmd.inputs.Project, app.ID())
	if dsClusterErr != nil {
		return dsClusterErr
	}
	service := realm.ServiceDescData{
		Config: map[string]interface{}{
			"clusterName":         dsCluster,
			"readPreference":      "primary",
			"wireProtocolEnabled": false,
		},
		Name: cmd.inputs.Name + "_cluster",
		Type: "mongodb-atlas",
	}
	_, dsErr := cmd.realmClient.CreateAppService(cmd.inputs.Project, app.ID(), service) // for output possibly
	if dsErr != nil {
		return dsErr
	}

	var appVersion realm.AppConfigVersion
	if cmd.inputs.AppVersion == realm.AppConfigVersionZero {
		appVersion = realm.AppConfigVersion20210101
	}
	zipPkgName, zipPkg, exportErr := cmd.realmClient.Export(
		cmd.inputs.Project,
		app.ID(),
		realm.ExportRequest{ConfigVersion: appVersion},
	)
	if exportErr != nil {
		return exportErr
	}
	if idx := strings.LastIndex(zipPkgName, "_"); idx != -1 {
		zipPkgName = zipPkgName[:idx]
	}
	zipPkgPath := filepath.Join(dir, zipPkgName) // Is this the correct path
	if writeErr := local.WriteZip(zipPkgPath, zipPkg); writeErr != nil {
		return writeErr
	}

	return nil

}

// Feedback is the command feedback
func (cmd *CommandCreate) Feedback(profile *cli.Profile, ui terminal.UI) error {
	return ui.Print(terminal.NewTextLog("Successfully created an app"))
}
