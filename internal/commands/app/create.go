package app

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/telemetry"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

// CommandMetaCreate is the command meta for the `app create` command
var CommandMetaCreate = cli.CommandMeta{
	Use:         "create",
	Display:     "app create",
	Description: "Create a new app (or a template app) from your current working directory and deploy it to the Realm server",
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
func (cmd *CommandCreate) Flags() []flags.Flag {
	return []flags.Flag{
		remoteAppFlag(&cmd.inputs.RemoteApp),
		flags.StringFlag{
			Value: &cmd.inputs.LocalPath,
			Meta: flags.Meta{
				Name: flagLocalPathCreate,
				Usage: flags.Usage{
					Description: "Specify the local filepath of a Realm app to be created",
				},
			},
		},
		nameFlag(&cmd.inputs.Name),
		locationFlag(&cmd.inputs.Location),
		deploymentModelFlag(&cmd.inputs.DeploymentModel),
		environmentFlag(&cmd.inputs.Environment),
		flags.StringSliceFlag{
			Value: &cmd.inputs.Clusters,
			Meta: flags.Meta{
				Name: flagCluster,
				Usage: flags.Usage{
					Description: "Link Atlas cluster(s) to your Realm app",
					Note:        "Only one cluster can be linked during app creation if creating a template app",
				},
			},
		},
		flags.StringSliceFlag{
			Value: &cmd.inputs.ClusterServiceNames,
			Meta: flags.Meta{
				Name: flagClusterServiceName,
				Usage: flags.Usage{
					Description: "Specify the Realm app Service name to reference your Atlas cluster",
					Note:        "Service names will be overwritten when creating a template app",
				},
			},
		},
		flags.StringSliceFlag{
			Value: &cmd.inputs.ServerlessInstances,
			Meta: flags.Meta{
				Name: flagServerlessInstance,
				Usage: flags.Usage{
					Description: "Link Atlas Serverless instance(s) to your Realm app",
					Note:        "Serverless instances cannot be used to create template apps",
				},
			},
		},
		flags.StringSliceFlag{
			Value: &cmd.inputs.ServerlessInstanceServiceNames,
			Meta: flags.Meta{
				Name: flagServerlessInstanceServiceName,
				Usage: flags.Usage{
					Description: "Specify the Realm app Service name to reference your Atlas Serverless instance",
				},
			},
		},
		flags.StringSliceFlag{
			Value: &cmd.inputs.Datalakes,
			Meta: flags.Meta{
				Name: flagDatalake,
				Usage: flags.Usage{
					Description: "Link Atlas data lake(s) to your Realm app",
					Note:        "Data lakes cannot be used to create template apps",
				},
			},
		},
		flags.StringSliceFlag{
			Value: &cmd.inputs.DatalakeServiceNames,
			Meta: flags.Meta{
				Name: flagDatalakeServiceName,
				Usage: flags.Usage{
					Description: "Specify the Realm app Service name to reference your Atlas data lake",
				},
			},
		},
		flags.StringFlag{
			Value: &cmd.inputs.Template,
			Meta: flags.Meta{
				Name: flagTemplate,
				Usage: flags.Usage{
					Description:   "Create your Realm app from an available template",
					AllowedValues: realm.AllowedTemplates,
				},
			},
		},
		flags.BoolFlag{
			Value: &cmd.inputs.DryRun,
			Meta: flags.Meta{
				Name:      flagDryRun,
				Shorthand: "x",
				Usage: flags.Usage{
					Description: "Run without writing any changes to the local filepath or pushing any changes to the Realm server",
				},
			},
		},
		cli.ProjectFlag(&cmd.inputs.Project),
		cli.ConfigVersionFlag(&cmd.inputs.ConfigVersion, flagConfigVersionDescription),
	}
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

	if err := cmd.inputs.resolveTemplateID(clients.Realm); err != nil {
		return err
	}

	dsClusters, dsClustersMissing, err := cmd.inputs.resolveClusters(ui, clients.Atlas, groupID)
	if err != nil {
		return err
	}

	var dsServerlessInstances []dataSourceCluster
	var dsServerlessInstancesMissing []string
	if len(cmd.inputs.ServerlessInstances) > 0 {
		dsServerlessInstances, dsServerlessInstancesMissing, err = cmd.inputs.resolveServerlessInstances(ui, clients.Atlas, groupID)
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

	nonExistingDataSources := make([]string, 0, len(dsClustersMissing)+len(dsDatalakesMissing)+len(dsServerlessInstancesMissing))
	for _, missingCluster := range dsClustersMissing {
		nonExistingDataSources = append(nonExistingDataSources, fmt.Sprintf("'%s'", missingCluster))
	}
	for _, missingServerlessInstance := range dsServerlessInstancesMissing {
		nonExistingDataSources = append(nonExistingDataSources, fmt.Sprintf("'%s'", missingServerlessInstance))
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

	if cmd.inputs.Template == "" {
		return cmd.handleCreateApp(profile, ui, clients, groupID, rootDir, appRemote, dsClusters, dsServerlessInstances, dsDatalakes)
	}
	return cmd.handleCreateTemplateApp(profile, ui, clients, groupID, rootDir, dsClusters, dsDatalakes)
}

func (cmd CommandCreate) handleCreateApp(
	profile *user.Profile,
	ui terminal.UI,
	clients cli.Clients,
	groupID string,
	rootDir string,
	appRemote realm.App,
	dsClusters []dataSourceCluster,
	dsServerlessInstances []dataSourceCluster,
	dsDatalakes []dataSourceDatalake,
) error {
	if cmd.inputs.DryRun {
		logs := make([]terminal.Log, 0, 4)
		var appCreatedText string
		if appRemote.GroupID == "" && appRemote.ID == "" {
			appCreatedText = fmt.Sprintf("A minimal Realm app would be created at %s", rootDir)
		} else {
			appCreatedText = fmt.Sprintf("A Realm app based on the Realm app '%s' would be created at %s", cmd.inputs.RemoteApp, rootDir)
		}

		logs = append(logs, terminal.NewTextLog(appCreatedText))

		for i, cluster := range dsClusters {
			logs = append(logs, terminal.NewTextLog("The cluster '%s' would be linked as data source '%s'", cmd.inputs.Clusters[i], cluster.Name))
		}
		for i, serverlessInstance := range dsServerlessInstances {
			logs = append(logs, terminal.NewTextLog("The serverless instance '%s' would be linked as data source '%s'", cmd.inputs.ServerlessInstances[i], serverlessInstance.Name))
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
	if appRemote.GroupID == "" && appRemote.ID == "" {
		appLocal = local.NewApp(
			rootDir,
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

		if err := local.WriteZip(rootDir, zipPkg); err != nil {
			return err
		}

		appLocal, err = local.LoadApp(rootDir)
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

	for _, dsServerlessInstance := range dsServerlessInstances {
		local.AddDataSource(appLocal.AppData, map[string]interface{}{
			"name": dsServerlessInstance.Name,
			"type": dsServerlessInstance.Type,
			"config": map[string]interface{}{
				"clusterName":         dsServerlessInstance.Config.ClusterName,
				"readPreference":      dsServerlessInstance.Config.ReadPreference,
				"wireProtocolEnabled": dsServerlessInstance.Config.WireProtocolEnabled,
			},
			"version": dsServerlessInstance.Version,
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

	if err := appLocal.LoadData(appLocal.RootDir); err != nil {
		return err
	}

	if err := clients.Realm.Import(appRealm.GroupID, appRealm.ID, appLocal.AppData); err != nil {
		return err
	}

	output := newAppOutputs{
		AppID:    appRealm.ClientAppID,
		Filepath: rootDir,
		URL:      fmt.Sprintf("%s/groups/%s/apps/%s/dashboard", profile.RealmBaseURL(), appRealm.GroupID, appRealm.ID),
	}

	for _, dsCluster := range dsClusters {
		output.Clusters = append(output.Clusters, dataSourceOutputs{dsCluster.Name})
	}

	for _, dsServerlessInstance := range dsServerlessInstances {
		output.ServerlessInstances = append(output.ServerlessInstances, dataSourceOutputs{dsServerlessInstance.Name})
	}

	for _, dsDatalake := range dsDatalakes {
		output.Datalakes = append(output.Datalakes, dataSourceOutputs{dsDatalake.Name})
	}

	ui.Print(terminal.NewJSONLog("Successfully created app", output))
	ui.Print(terminal.NewFollowupLog("Check out your app", fmt.Sprintf("cd ./%s && %s app describe", cmd.inputs.LocalPath, cli.Name)))
	return nil
}

// handleCreateTemplateApp creates a template app and writes it to the user's disk.
// This function does not take in serverless instances because they are not compatible with template apps.
func (cmd CommandCreate) handleCreateTemplateApp(
	profile *user.Profile,
	ui terminal.UI,
	clients cli.Clients,
	groupID string,
	rootDir string,
	dsClusters []dataSourceCluster,
	dsDatalakes []dataSourceDatalake,
) error {
	if cmd.inputs.DryRun {
		// +2 indicates that there are two logs in addition to the data lake and cluster ones
		logs := make([]terminal.Log, 0, len(dsClusters)+len(dsDatalakes)+2)
		logs = append(logs, terminal.NewTextLog(
			"A Realm app would be created at %s using the '%s' template",
			rootDir,
			cmd.inputs.Template,
		))

		for _, cluster := range dsClusters {
			logs = append(logs, terminal.NewTextLog("The cluster '%s' would be linked as data source '%s'", cluster.Config.ClusterName, cluster.Name))
		}
		for _, datalake := range dsDatalakes {
			logs = append(logs, terminal.NewTextLog("The data lake '%s' would be linked as data source '%s'", datalake.Config.DatalakeName, datalake.Name))
		}
		logs = append(logs, terminal.NewFollowupLog("To create this app run", cmd.display(true)))
		ui.Print(logs...)
		return nil
	}

	if len(dsClusters) != 1 {
		return errors.New("must specify a cluster when creating app from template")
	}
	createAppMetadata := realm.AppMeta{
		Location:        cmd.inputs.Location,
		DeploymentModel: cmd.inputs.DeploymentModel,
		Environment:     cmd.inputs.Environment,
		Template:        cmd.inputs.Template,
		DataSource:      dsClusters[0],
	}

	createTemplateApp := func() (realm.App, error) {
		s := ui.Spinner("Creating template app...", terminal.SpinnerOptions{})
		s.Start()
		defer s.Stop()

		appRealm, err := clients.Realm.CreateApp(
			groupID,
			cmd.inputs.Name,
			createAppMetadata,
		)
		if err != nil {
			return realm.App{}, err
		}
		ui.Print(terminal.NewTextLog("Created template app"))
		return appRealm, nil
	}
	appRealm, err := createTemplateApp()
	if err != nil {
		return err
	}

	_, err = writeTemplateAppToLocal(ui, clients.Realm, appRealm.ID, appRealm.GroupID, cmd.inputs.Template, rootDir)
	if err != nil {
		return err
	}

	output := newAppOutputs{
		AppID:    appRealm.ClientAppID,
		Filepath: rootDir,
		Backend:  filepath.Join(rootDir, local.BackendPath),
		URL:      fmt.Sprintf("%s/groups/%s/apps/%s/dashboard", profile.RealmBaseURL(), appRealm.GroupID, appRealm.ID),
		Clusters: []dataSourceOutputs{{Name: dsClusters[0].Name}},
	}

	frontendDir := filepath.Join(rootDir, local.FrontendPath)
	_, err = os.Stat(frontendDir)
	if err != nil && os.IsNotExist(err) {
		output.Frontends = frontendDir
	}

	pathRelative, err := filepath.Rel(profile.WorkingDirectory, rootDir)
	if err != nil {
		return err
	}
	readmeFile, err := local.FindReadme(rootDir, pathRelative, cmd.inputs.Template)
	if err != nil {
		return err
	}

	// TODO(REALMC-9460): Add better template-app-specific directions for checking out the newly created template app
	ui.Print(terminal.NewJSONLog("Successfully created app", output))
	ui.Print(terminal.NewFollowupLog("Check out your app", fmt.Sprintf("cd ./%s && %s app describe", cmd.inputs.LocalPath, cli.Name)))
	ui.Print(terminal.NewFollowupLog(fmt.Sprintf("View directions on how to run the template app: %s", readmeFile)))

	return nil
}

func (cmd *CommandCreate) display(omitDryRun bool) string {
	return cli.CommandDisplay(CommandMetaCreate.Display, cmd.inputs.args(omitDryRun))
}

func writeTemplateAppToLocal(ui terminal.UI, realmClient realm.Client, appID, groupID, templateID, rootDir string) (local.App, error) {
	backendDir := filepath.Join(rootDir, local.BackendPath)
	frontendDir := filepath.Join(rootDir, local.FrontendPath)

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

	s := ui.Spinner("Downloading client template...", terminal.SpinnerOptions{})

	downloadAndWriteClient := func() error {
		s.Start()
		defer s.Stop()

		zipPkg, ok, err := realmClient.ClientTemplate(
			groupID,
			appID,
			templateID,
		)
		if err != nil {
			return err
		}
		if !ok {
			// The template has no frontend to pull
			return nil
		}
		if err := local.WriteZip(path.Join(frontendDir, templateID), zipPkg); err != nil {
			return err
		}
		return nil
	}

	if err := downloadAndWriteClient(); err != nil {
		return local.App{}, err
	}

	return appLocal, nil
}
