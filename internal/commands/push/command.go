package push

import (
	"fmt"
	"strings"
	"time"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/AlecAivazis/survey/v2"
)

const (
	flagLocalPath           = "local"
	flagRemote              = "remote"
	flagIncludeDependencies = "include-dependencies"
	flagIncludeNodeModules  = "include-node-modules"
	flagIncludePackageJSON  = "include-package-json"
	flagIncludeHosting      = "include-hosting"
	flagResetCDNCache       = "reset-cdn-cache"
	flagDryRun              = "dry-run"
)

var (
	warnFailedToDiscardDraft = terminal.NewWarningLog("Failed to discard the draft created for your deployment")
)

// CommandMeta is the command meta for the 'push' command
var CommandMeta = cli.CommandMeta{
	Use:         "push",
	Aliases:     []string{"import"},
	Description: "Imports and deploys changes from your local directory to your Realm app",
	HelpText: `Updates a remote Realm app with your local directory. First, input a Realm app
that you would like changes pushed to. This input can be either the application
Client App ID of an existing Realm app you would like to update, or the Name of
a new Realm app you would like to create. Changes pushed are automatically
deployed.`,
}

// Command is the `push` command
type Command struct {
	inputs inputs
}

// Flags is the command flags
func (cmd *Command) Flags() []flags.Flag {
	return []flags.Flag{
		flags.StringFlag{
			Value: &cmd.inputs.LocalPath,
			Meta: flags.Meta{
				Name: flagLocalPath,
				Usage: flags.Usage{
					Description: "Specify the local filepath of a Realm app to be imported",
				},
			},
		},
		flags.StringFlag{
			Value: &cmd.inputs.RemoteApp,
			Meta: flags.Meta{
				Name: flagRemote,
				Usage: flags.Usage{
					Description: "Specify the name or ID of a remote Realm app to edit",
				},
			},
		},
		flags.BoolFlag{
			Value: &cmd.inputs.IncludeNodeModules,
			Meta: flags.Meta{
				Name: flagIncludeNodeModules,
				Usage: flags.Usage{
					Description: "Import and include Realm app dependencies from a node_modules archive",
					Note:        "The allowed formats are as a directory or compressed into a .zip, .tar, .tar.gz, or .tgz file",
				},
			},
		},
		flags.BoolFlag{
			Value: &cmd.inputs.IncludePackageJSON,
			Meta: flags.Meta{
				Name: flagIncludePackageJSON,
				Usage: flags.Usage{
					Description: "Import and include Realm app dependencies from a package.json file",
				},
			},
		},
		// TODO(REALMC-10088): Remove this flag in realm-cli 3.x
		flags.BoolFlag{
			Value: &cmd.inputs.IncludeDependencies,
			Meta: flags.Meta{
				Name:      flagIncludeDependencies,
				Shorthand: "d",
				Usage: flags.Usage{
					Description: "Import and include Realm app dependencies in the diff from a node_modules archive",
					Note:        "The allowed formats are as a directory or compressed into a .zip, .tar, .tar.gz, or .tgz file",
				},
				Deprecate: fmt.Sprintf("support will be removed in v3.x, please use %q instead", flagIncludeNodeModules),
			},
		},
		flags.BoolFlag{
			Value: &cmd.inputs.IncludeHosting,
			Meta: flags.Meta{
				Name:      flagIncludeHosting,
				Shorthand: "s",
				Usage: flags.Usage{
					Description: "Import and include Realm app hosting files",
				},
			},
		},
		flags.BoolFlag{
			Value: &cmd.inputs.ResetCDNCache,
			Meta: flags.Meta{
				Name:      flagResetCDNCache,
				Shorthand: "c",
				Usage: flags.Usage{
					Description: "Reset the hosting CDN cache of a Realm app",
				},
			},
		},
		flags.BoolFlag{
			Value: &cmd.inputs.DryRun,
			Meta: flags.Meta{
				Name:      flagDryRun,
				Shorthand: "x",
				Usage: flags.Usage{
					Description: "Run without pushing any changes to the Realm server",
				},
			},
		},
		cli.ProjectFlag(&cmd.inputs.Project),
	}
}

// Inputs is the command inputs
func (cmd *Command) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *Command) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	app, err := local.LoadApp(cmd.inputs.LocalPath)
	if err != nil {
		return err
	}

	appRemote, err := cmd.inputs.resolveRemoteApp(ui, clients.Realm, app.Meta)
	if err != nil {
		return err
	}

	if appRemote.GroupID == "" {
		groupID, err := cli.ResolveGroupID(ui, clients.Atlas)
		if err != nil {
			return err
		}
		appRemote.GroupID = groupID
	}

	if appRemote.AppID == "" {
		if cmd.inputs.DryRun {
			ui.Print(
				terminal.NewTextLog("This is a new app. To create a new app, you must omit the 'dry-run' flag to proceed"),
				terminal.NewFollowupLog(terminal.MsgSuggestions, cmd.display(true)),
			)
			return nil
		}

		newApp, proceed, err := createNewApp(ui, clients.Realm, app.RootDir, appRemote.GroupID, app.AppData)
		if err != nil {
			return err
		}
		if !proceed {
			return nil
		}

		ui.Print(terminal.NewTextLog("App created successfully"))

		appRemote.AppID = newApp.ID
		appRemote.ClientAppID = newApp.ClientAppID
	}

	ui.Print(terminal.NewTextLog("Determining changes"))
	appDiffs, err := clients.Realm.Diff(appRemote.GroupID, appRemote.AppID, app.AppData)
	if err != nil {
		return err
	}

	var uploadPathDependencies string
	var dependenciesDiffs realm.DependenciesDiff
	if cmd.inputs.IncludeNodeModules || cmd.inputs.IncludePackageJSON || cmd.inputs.IncludeDependencies {
		appDependencies, err := cmd.inputs.resolveAppDependencies(app.RootDir)
		if err != nil {
			return err
		}

		uploadPath, cleanup, err := appDependencies.PrepareUpload()
		if err != nil {
			return err
		}
		defer cleanup()

		dependenciesDiffs, err = clients.Realm.DiffDependencies(appRemote.GroupID, appRemote.AppID, uploadPath)
		if err != nil {
			return err
		}
		uploadPathDependencies = uploadPath
	}

	hosting, err := local.FindAppHosting(app.RootDir)
	if err != nil {
		return err
	}

	var hostingDiffs local.HostingDiffs
	if cmd.inputs.IncludeHosting {
		appAssets, err := clients.Realm.HostingAssets(appRemote.GroupID, appRemote.AppID)
		if err != nil {
			return err
		}

		hostingDiffs, err = hosting.Diffs(profile.HostingAssetCachePath(), appRemote.AppID, appAssets)
		if err != nil {
			return err
		}
	}

	if len(appDiffs) == 0 && dependenciesDiffs.Len() == 0 && hostingDiffs.Size() == 0 {
		ui.Print(terminal.NewTextLog("Deployed app is identical to proposed version, nothing to do"))
		return nil
	}

	if !ui.AutoConfirm() {
		diffs := make([]string, 0, len(appDiffs)+1+hostingDiffs.Cap())

		diffs = append(diffs, appDiffs...)

		if cmd.inputs.IncludeNodeModules || cmd.inputs.IncludePackageJSON || cmd.inputs.IncludeDependencies {
			diffs = append(diffs, dependenciesDiffs.Strings()...)
		}

		diffs = append(diffs, hostingDiffs.Strings()...)

		// when updating an existing app, if the user has not set the '-y' flag
		// print the app diffs back to the user
		ui.Print(terminal.NewTextLog(
			"The following reflects the proposed changes to your Realm app\n%s",
			strings.Join(diffs, "\n"),
		))
	}

	if cmd.inputs.DryRun {
		ui.Print(
			terminal.NewTextLog("To push these changes, you must omit the 'dry-run' flag to proceed"),
			terminal.NewFollowupLog(terminal.MsgSuggestions, cmd.display(true)),
		)
		return nil
	}

	proceed, err := ui.Confirm("Please confirm the changes shown above")
	if err != nil {
		return err
	}
	if !proceed {
		return nil
	}

	if len(appDiffs) > 0 {
		ui.Print(terminal.NewTextLog("Creating draft"))
		draft, proceed, err := createNewDraft(ui, clients.Realm, appRemote)
		if err != nil {
			return err
		}
		if !proceed {
			return nil
		}

		ui.Print(terminal.NewTextLog("Pushing changes"))
		if err := clients.Realm.Import(appRemote.GroupID, appRemote.AppID, app.AppData); err != nil {
			if err := clients.Realm.DiscardDraft(appRemote.GroupID, appRemote.AppID, draft.ID); err != nil {
				ui.Print(warnFailedToDiscardDraft)
			}
			return err
		}

		ui.Print(terminal.NewTextLog("Deploying draft"))
		if err := deployDraftAndWait(ui, clients.Realm, appRemote, draft.ID); err != nil {
			if err := clients.Realm.DiscardDraft(appRemote.GroupID, appRemote.AppID, draft.ID); err != nil {
				ui.Print(warnFailedToDiscardDraft)
			}
			return err
		}
	}

	if cmd.inputs.IncludePackageJSON || cmd.inputs.IncludeNodeModules || cmd.inputs.IncludeDependencies {
		installDependencies := func() error {
			s := ui.Spinner("Installing dependencies: starting...", terminal.SpinnerOptions{})

			s.Start()
			defer s.Stop()

			if err := clients.Realm.ImportDependencies(appRemote.GroupID, appRemote.AppID, uploadPathDependencies); err != nil {
				return err
			}

			status := realm.DependenciesStatus{State: realm.DependenciesStateCreated}
			for status.State == realm.DependenciesStateCreated {
				var err error
				status, err = clients.Realm.DependenciesStatus(appRemote.GroupID, appRemote.AppID)
				if err != nil {
					return err
				}

				if status.State == realm.DependenciesStateSuccessful || status.State == realm.DependenciesStateFailed {
					break
				}

				s.SetMessage(fmt.Sprintf("Installing dependencies: %s...", status.Message))
				time.Sleep(time.Second)
			}
			if status.State == realm.DependenciesStateFailed {
				return fmt.Errorf("failed to install dependencies: %s", status.Message)
			}
			return nil
		}

		if err := installDependencies(); err != nil {
			return err
		}

		ui.Print(terminal.NewTextLog("Installed dependencies"))
	}

	if cmd.inputs.IncludeHosting {
		s := ui.Spinner("Importing hosting assets...", terminal.SpinnerOptions{})

		importHosting := func() error {
			s.Start()
			defer s.Stop()

			return hosting.UploadHostingAssets(
				clients.Realm,
				appRemote.GroupID,
				appRemote.AppID,
				hostingDiffs,
				func(err error) {
					ui.Print(terminal.NewWarningLog("An error occurred while uploading hosting assets: %s", err.Error()))
				},
			)
		}

		if err := importHosting(); err != nil {
			return err
		}
		ui.Print(terminal.NewTextLog("Import hosting assets"))

		if cmd.inputs.ResetCDNCache {
			s := ui.Spinner("Resetting CDN cache...", terminal.SpinnerOptions{})

			invalidateCache := func() error {
				s.Start()
				defer s.Stop()

				return clients.Realm.HostingCacheInvalidate(appRemote.GroupID, appRemote.AppID, "/*")
			}

			if err := invalidateCache(); err != nil {
				return err
			}
			ui.Print(terminal.NewTextLog("Reset CDN cache"))
		}
	}

	ui.Print(terminal.NewTextLog("Successfully pushed app up: %s", appRemote.ClientAppID))
	return nil
}

func (cmd *Command) display(omitDryRun bool) string {
	return cli.CommandDisplay(CommandMeta.Use, cmd.inputs.args(omitDryRun))
}

func (i *inputs) resolveAppDependencies(rootDir string) (local.Dependencies, error) {
	if i.IncludePackageJSON {
		return local.FindPackageJSON(rootDir)
	}
	return local.FindNodeModules(rootDir)
}

type namer interface{ Name() string }
type locationer interface{ Location() realm.Location }
type deploymentModeler interface{ DeploymentModel() realm.DeploymentModel }
type environmenter interface{ Environment() realm.Environment }
type configVersioner interface{ ConfigVersion() realm.AppConfigVersion }

func createNewApp(ui terminal.UI, realmClient realm.Client, appDirectory, groupID string, appData interface{}) (realm.App, bool, error) {
	if proceed, err := ui.Confirm("Do you wish to create a new app?"); err != nil {
		return realm.App{}, false, err
	} else if !proceed {
		return realm.App{}, false, nil
	}

	var name, location, deploymentModel, environment string
	appConfigVersion := realm.DefaultAppConfigVersion
	if appData != nil {
		if n, ok := appData.(namer); ok {
			name = n.Name()
		}

		if l, ok := appData.(locationer); ok {
			location = l.Location().String()
		}

		if dm, ok := appData.(deploymentModeler); ok {
			deploymentModel = dm.DeploymentModel().String()
		}

		if e, ok := appData.(environmenter); ok {
			environment = e.Environment().String()
		}

		if cv, ok := appData.(configVersioner); ok {
			appConfigVersion = cv.ConfigVersion()
		}
	}

	if name == "" || !ui.AutoConfirm() {
		if err := ui.AskOne(&name, &survey.Input{Message: "App Name", Default: name}); err != nil {
			return realm.App{}, false, err
		}
	}

	if !ui.AutoConfirm() {
		if err := ui.AskOne(
			&location,
			&survey.Select{
				Message: "App Location",
				Options: realm.LocationValues,
				Default: location,
			},
		); err != nil {
			return realm.App{}, false, err
		}
	}

	if !ui.AutoConfirm() {
		if err := ui.AskOne(
			&deploymentModel,
			&survey.Select{
				Message: "App Deployment Model",
				Options: realm.DeploymentModelValues,
				Default: deploymentModel,
			}); err != nil {
			return realm.App{}, false, err
		}
	}

	if !ui.AutoConfirm() {
		if err := ui.AskOne(
			&environment,
			&survey.Select{
				Message: "App Environment",
				Options: realm.EnvironmentValues,
				Default: environment,
			}); err != nil {
			return realm.App{}, false, err
		}
	}

	if proceed, err := ui.Confirm("Please confirm the new app details shown above"); err != nil {
		return realm.App{}, false, err
	} else if !proceed {
		return realm.App{}, false, nil
	}

	app, err := realmClient.CreateApp(
		groupID,
		name,
		realm.AppMeta{
			Location:        realm.Location(location),
			DeploymentModel: realm.DeploymentModel(deploymentModel),
			Environment:     realm.Environment(environment),
		},
	)
	if err != nil {
		return realm.App{}, false, err
	}

	if err := local.AsApp(appDirectory, app, appConfigVersion).WriteConfig(); err != nil {
		return realm.App{}, false, err
	}
	return app, true, nil
}

func createNewDraft(ui terminal.UI, realmClient realm.Client, remote appRemote) (realm.AppDraft, bool, error) {
	draft, draftErr := realmClient.CreateDraft(remote.GroupID, remote.AppID)
	if draftErr == nil {
		return draft, true, nil
	}

	if err, ok := draftErr.(realm.ServerError); !ok || err.Code != realm.ErrCodeDraftAlreadyExists {
		return realm.AppDraft{}, false, draftErr
	}

	existingDraft, existingDraftErr := realmClient.Draft(remote.GroupID, remote.AppID)
	if existingDraftErr != nil {
		return realm.AppDraft{}, false, existingDraftErr
	}

	if !ui.AutoConfirm() {
		if err := diffDraft(ui, realmClient, remote, existingDraft.ID); err != nil {
			return realm.AppDraft{}, false, err
		}

		proceed, proceedErr := ui.Confirm("Would you like to discard this draft?")
		if proceedErr != nil {
			return realm.AppDraft{}, false, proceedErr
		}
		if !proceed {
			return realm.AppDraft{}, false, nil
		}
	}

	if err := realmClient.DiscardDraft(remote.GroupID, remote.AppID, existingDraft.ID); err != nil {
		return realm.AppDraft{}, false, err
	}

	draft, draftErr = realmClient.CreateDraft(remote.GroupID, remote.AppID)
	return draft, true, draftErr
}

func diffDraft(ui terminal.UI, realmClient realm.Client, remote appRemote, draftID string) error {
	diff, diffErr := realmClient.DiffDraft(remote.GroupID, remote.AppID, draftID)
	if diffErr != nil {
		return diffErr
	}

	var logs []terminal.Log
	if !diff.HasChanges() {
		logs = append(logs, terminal.NewTextLog("An empty draft already exists for your app"))
	} else {
		logs = append(logs, terminal.NewListLog("The following draft already exists for your app...", diff.DiffList()...))
		if diff.HostingFilesDiff.HasChanges() {
			logs = append(logs, terminal.NewListLog("With changes to your static hosting files...", diff.HostingFilesDiff.DiffList()...))
		}
		if diff.DependenciesDiff.HasChanges() {
			logs = append(logs, terminal.NewListLog("With changes to your app dependencies...", diff.DependenciesDiff.DiffList()...))
		}
		if diff.GraphQLConfigDiff.HasChanges() {
			logs = append(logs, terminal.NewListLog("With changes to your GraphQL configuration...", diff.GraphQLConfigDiff.DiffList()...))
		}
		if diff.SchemaOptionsDiff.HasChanges() {
			logs = append(logs, terminal.NewListLog("With changes to your app schema...", diff.SchemaOptionsDiff.DiffList()...))
		}
	}
	ui.Print(logs...)
	return nil
}

func deployDraftAndWait(ui terminal.UI, realmClient realm.Client, remote appRemote, draftID string) error {
	deployment, err := realmClient.DeployDraft(remote.GroupID, remote.AppID, draftID)
	if err != nil {
		return err
	}

	s := ui.Spinner("Deploying app changes...", terminal.SpinnerOptions{})

	waitForDeployment := func() error {
		s.Start()
		defer s.Stop()

		for deployment.Status == realm.DeploymentStatusCreated || deployment.Status == realm.DeploymentStatusPending {
			time.Sleep(time.Second)

			deployment, err = realmClient.Deployment(remote.GroupID, remote.AppID, deployment.ID)
			if err != nil {
				return err
			}
		}

		return nil
	}

	if err := waitForDeployment(); err != nil {
		return err
	}

	if deployment.Status == realm.DeploymentStatusFailed {
		ui.Print(terminal.NewWarningLog("Deployment failed"))
		return fmt.Errorf("failed to deploy app: %s", deployment.StatusErrorMessage)
	}

	ui.Print(terminal.NewTextLog("Deployment complete"))
	return nil
}
