package app

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/telemetry"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/briandowns/spinner"
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
	fs.StringVar(&cmd.inputs.RemoteApp, flagRemoteAppNew, "", flagRemoteAppNewUsage)
	fs.StringVar(&cmd.inputs.LocalPath, flagLocalPathCreate, "", flagLocalPathCreateUsage)
	fs.StringVarP(&cmd.inputs.Name, flagName, flagNameShort, "", flagNameUsage)
	fs.VarP(&cmd.inputs.Location, flagLocation, flagLocationShort, flagLocationUsage)
	fs.VarP(&cmd.inputs.DeploymentModel, flagDeploymentModel, flagDeploymentModelShort, flagDeploymentModelUsage)
	fs.VarP(&cmd.inputs.Environment, flagEnvironment, flagEnvironmentShort, flagEnvironmentUsage)
	fs.StringSliceVar(&cmd.inputs.Clusters, flagCluster, []string{}, flagClusterUsage)
	fs.StringSliceVar(&cmd.inputs.ClusterServiceNames, flagClusterServiceName, []string{}, flagClusterServiceNameUsage)
	fs.StringSliceVar(&cmd.inputs.Datalakes, flagDatalake, []string{}, flagDatalakeUsage)
	fs.StringSliceVar(&cmd.inputs.DatalakeServiceNames, flagDatalakeServiceName, []string{}, flagDatalakeServiceNameUsage)
	fs.StringVar(&cmd.inputs.Template, flagTemplate, "", flagTemplateUsage)
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

// AdditionalTrackedFields adds any additional fields to our tracking service. In this case, we will apply the template id if in use
func (cmd *CommandCreate) AdditionalTrackedFields() []telemetry.EventData {
	if cmd.inputs.Template == "" {
		return nil
	}
	return []telemetry.EventData{
		{Key: telemetry.EventDataKeyTemplate, Value: cmd.inputs.Template},
	}
}

// Handler is the command handler
func (cmd *CommandCreate) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	appRemote, err := cmd.inputs.resolveRemoteApp(ui, clients.Realm)
	if err != nil {
		return err
	}

	groupID := cmd.inputs.Project
	if groupID == "" {
		groupID = appRemote.GroupID
	}
	if groupID == "" {
		groupID, err = cli.ResolveGroupID(ui, clients.Atlas)
		if err != nil {
			return err
		}
	}

	err = cmd.inputs.resolveName(ui, clients.Realm, appRemote.GroupID, appRemote.ClientAppID)
	if err != nil {
		return err
	}

	rootDir, err := cmd.inputs.resolveLocalPath(ui, profile.WorkingDirectory)
	if err != nil {
		return err
	}

	if err := cmd.inputs.resolveTemplateID(ui, clients.Realm); err != nil {
		return err
	}

	var dsClusters []dataSourceCluster
	var dsClustersMissing []string
	if len(cmd.inputs.Clusters) > 0 {
		dsClusters, dsClustersMissing, err = cmd.inputs.resolveClusters(ui, clients.Atlas, groupID)
		if err != nil {
			return err
		}
	}

	var dsDatalakes []dataSourceDatalake
	var dsDatalakesMissing []string
	if len(cmd.inputs.Datalakes) > 0 {
		dsDatalakes, dsDatalakesMissing, err = cmd.inputs.resolveDatalakes(ui, clients.Atlas, groupID)
		if err != nil {
			return err
		}
	}

	nonExistingDataSources := make([]string, 0, len(dsClustersMissing)+len(dsDatalakesMissing))
	for _, missingCluster := range dsClustersMissing {
		nonExistingDataSources = append(nonExistingDataSources, fmt.Sprintf("'%s'", missingCluster))
	}
	for _, missingDatalake := range dsDatalakesMissing {
		nonExistingDataSources = append(nonExistingDataSources, fmt.Sprintf("'%s'", missingDatalake))
	}

	if len(nonExistingDataSources) > 0 {
		ui.Print(terminal.NewWarningLog("Note: The following data sources were not linked because they could not be found: %s", strings.Join(nonExistingDataSources, ", ")))
		proceed, err := ui.Confirm("Would you still like to create the app?")
		if err != nil {
			return err
		}
		if !proceed {
			return nil
		}
	}

	// If using a template, we nest backendDir under rootDir and export the
	// backend code there alongside a sibling directory containing the frontend
	// code. Otherwise, all code is exported in rootDir
	backendDir := rootDir
	if cmd.inputs.Template != "" {
		backendDir = path.Join(rootDir, local.BackendPath)
	}

	if cmd.inputs.DryRun {
		logs := make([]terminal.Log, 0, 4)
		var appCreatedText string
		if appRemote.GroupID == "" && appRemote.ID == "" {
			appCreatedText = fmt.Sprintf("A minimal Realm app would be created at %s", backendDir)
		} else {
			appCreatedText = fmt.Sprintf("A Realm app based on the Realm app '%s' would be created at %s", cmd.inputs.RemoteApp, backendDir)
		}

		if cmd.inputs.Template != "" {
			appCreatedText = fmt.Sprintf("%s using the '%s' template", appCreatedText, cmd.inputs.Template)
		}

		logs = append(logs, terminal.NewTextLog(appCreatedText))

		for i, cluster := range dsClusters {
			logs = append(logs, terminal.NewTextLog("The cluster '%s' would be linked as data source '%s'", cmd.inputs.Clusters[i], cluster.Name))
		}
		for i, datalake := range dsDatalakes {
			logs = append(logs, terminal.NewTextLog("The data lake '%s' would be linked as data source '%s'", cmd.inputs.Datalakes[i], datalake.Name))
		}
		logs = append(logs, terminal.NewFollowupLog("To create this app run", cmd.display(true)))
		ui.Print(logs...)
		return nil
	}

	createAppMetadata := realm.AppMeta{
		Location:        cmd.inputs.Location,
		DeploymentModel: cmd.inputs.DeploymentModel,
		Environment:     cmd.inputs.Environment,
		Template:        cmd.inputs.Template,
	}

	// choose a data source to import template app schema data onto
	if cmd.inputs.Template != "" {
		initialDataSource, err := cmd.inputs.resolveInitialTemplateDataSource(ui, dsDatalakes, dsClusters)
		if err != nil {
			return err
		}
		createAppMetadata.DataSource = initialDataSource
	}

	appRealm, err := clients.Realm.CreateApp(
		groupID,
		cmd.inputs.Name,
		createAppMetadata,
	)
	if err != nil {
		return err
	}

	var appLocal local.App
	if cmd.inputs.Template != "" {
		appLocal, err = createFromTemplate(clients.Realm, appRealm.ID, appRealm.GroupID, cmd.inputs.Template, backendDir, rootDir)
		if err != nil {
			return err
		}
	} else if appRemote.GroupID == "" && appRemote.ID == "" {
		appLocal = local.NewApp(
			backendDir,
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

		if err := local.WriteZip(backendDir, zipPkg); err != nil {
			return err
		}

		appLocal, err = local.LoadApp(backendDir)
		if err != nil {
			return err
		}
	}

	for _, dsCluster := range dsClusters {
		local.AddDataSource(appLocal.AppData, map[string]interface{}{
			"name": dsCluster.Name,
			"type": dsCluster.Type,
			"config": map[string]interface{}{
				"clusterName":         dsCluster.Config.ClusterName,
				"readPreference":      dsCluster.Config.ReadPreference,
				"wireProtocolEnabled": dsCluster.Config.WireProtocolEnabled,
			},
			"version": dsCluster.Version,
		})

	}

	for _, dsDatalake := range dsDatalakes {
		local.AddDataSource(appLocal.AppData, map[string]interface{}{
			"name": dsDatalake.Name,
			"type": dsDatalake.Type,
			"config": map[string]interface{}{
				"dataLakeName": dsDatalake.Config.DatalakeName,
			},
			"version": dsDatalake.Version,
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

	output := newAppOutputs{
		AppID:    appRealm.ClientAppID,
		Filepath: backendDir,
		URL:      fmt.Sprintf("%s/groups/%s/apps/%s/dashboard", profile.RealmBaseURL(), appRealm.GroupID, appRealm.ID),
	}

	for _, dsCluster := range dsClusters {
		output.Clusters = append(output.Clusters, dataSourceOutputs{dsCluster.Name})
	}

	for _, dsDatalake := range dsDatalakes {
		output.Datalakes = append(output.Datalakes, dataSourceOutputs{dsDatalake.Name})
	}

	ui.Print(terminal.NewJSONLog("Successfully created app", output))
	ui.Print(terminal.NewFollowupLog("Check out your app", fmt.Sprintf("cd ./%s && %s app describe", cmd.inputs.LocalPath, cli.Name)))
	return nil
}

func (cmd *CommandCreate) display(omitDryRun bool) string {
	return cli.CommandDisplay(CommandMetaCreate.Display, cmd.inputs.args(omitDryRun))
}

func createFromTemplate(realmClient realm.Client, appID, groupID, templateID, backendDir, rootDir string) (local.App, error) {
	_, zipPkg, err := realmClient.Export(
		groupID,
		appID,
		realm.ExportRequest{},
	)
	if err != nil {
		return local.App{}, err
	}

	if err := local.WriteZip(backendDir, zipPkg); err != nil {
		return local.App{}, err
	}

	appLocal, err := local.LoadApp(backendDir)
	if err != nil {
		return local.App{}, err
	}

	s := spinner.New(terminal.SpinnerCircles, 250*time.Millisecond)
	s.Suffix = " Downloading client template..."

	downloadAndWriteClient := func() error {
		s.Start()
		defer s.Stop()

		zipPkg, err := realmClient.ClientTemplate(
			groupID,
			appID,
			templateID,
		)
		if err != nil {
			return err
		}

		if err := local.WriteZip(path.Join(rootDir, local.FrontendPath), zipPkg); err != nil {
			return err
		}

		return nil
	}

	if err := downloadAndWriteClient(); err != nil {
		return local.App{}, err
	}

	return appLocal, nil
}
