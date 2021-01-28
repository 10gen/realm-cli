package push

import (
	"strings"
	"time"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/pflag"
)

// Command is the `push` command
type Command struct {
	inputs      inputs
	outputs     outputs
	atlasClient atlas.Client
	realmClient realm.Client
}

type outputs struct {
	appCreated   bool
	diffRejected bool
	draftKept    bool
	noDiffs      bool
}

// Flags is the command flags
func (cmd *Command) Flags(fs *pflag.FlagSet) {
	fs.StringVarP(&cmd.inputs.AppDirectory, flagAppDirectory, flagAppDirectoryShort, "", flagAppDirectoryUsage)
	fs.StringVar(&cmd.inputs.Project, flagProject, "", flagProjectUsage)
	fs.StringVarP(&cmd.inputs.To, flagTo, flagToShort, "", flagToUsage)
	fs.BoolVarP(&cmd.inputs.AsNew, flagAsNew, flagAsNewShort, false, flagAsNewUsage)
	fs.BoolVarP(&cmd.inputs.DryRun, flagDryRun, flagDryRunShort, false, flagDryRunUsage)
	fs.BoolVarP(&cmd.inputs.IncludeDependencies, flagIncludeDependencies, flagIncludeDependenciesShort, false, flagIncludeDependenciesUsage)
	fs.BoolVarP(&cmd.inputs.IncludeHosting, flagIncludeHosting, flagIncludeHostingShort, false, flagIncludeHostingUsage)
	fs.BoolVarP(&cmd.inputs.ResetCDNCache, flagResetCDNCache, flagResetCDNCacheShort, false, flagResetCDNCacheUsage)
}

// Inputs is the command inputs
func (cmd *Command) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Setup is the command setup
func (cmd *Command) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.atlasClient = profile.AtlasAuthClient()
	cmd.realmClient = profile.RealmAuthClient()
	return nil
}

// Handler is the command handler
func (cmd *Command) Handler(profile *cli.Profile, ui terminal.UI) error {
	app, appErr := local.LoadApp(cmd.inputs.AppDirectory)
	if appErr != nil {
		return appErr
	}

	to, toErr := cmd.inputs.resolveTo(ui, cmd.realmClient)
	if toErr != nil {
		return toErr
	}

	if to.GroupID == "" {
		groupID, err := cli.ResolveGroupID(ui, cmd.atlasClient)
		if err != nil {
			return err
		}
		to.GroupID = groupID
	}

	if to.AppID == "" {
		if cmd.inputs.DryRun {
			return nil
		}

		app, err := cmd.createNewApp(ui, to.GroupID, app)
		if err != nil {
			return err
		}
		if app.ID == "" {
			return nil
		}
		to.AppID = app.ID
		cmd.outputs.appCreated = true
	}

	diffs, diffsErr := cmd.realmClient.Diff(to.GroupID, to.AppID, app)
	if diffsErr != nil {
		return diffsErr
	}

	if cmd.inputs.IncludeHosting {
		// TODO(REALMC-7177): diff hosting changes
		diffs = append(diffs, "Import hosting")
	}

	if cmd.inputs.IncludeDependencies {
		diffs = append(diffs, "Import dependencies")
	}

	if len(diffs) == 0 {
		cmd.outputs.noDiffs = true
		return nil
	}

	if !ui.AutoConfirm() && !cmd.outputs.appCreated {
		// when updating an existing app, if the user has not set the '-y' flag
		// print the app diffs back to the user
		if err := ui.Print(terminal.NewTextLog(
			"The following reflects the proposed changes to your Realm app\n%s",
			strings.Join(diffs, "\n"),
		)); err != nil {
			return err
		}
	}

	if cmd.inputs.DryRun {
		return nil
	}

	proceed, confirmErr := ui.Confirm("Please confirm the changes shown above")
	if confirmErr != nil {
		return confirmErr
	}
	if !proceed {
		cmd.outputs.diffRejected = true
		return nil
	}

	draft, draftProceed, draftErr := createNewDraft(ui, cmd.realmClient, to)
	if draftErr != nil {
		return draftErr
	}
	if !draftProceed {
		cmd.outputs.draftKept = true
		return nil
	}

	if err := cmd.realmClient.Import(to.GroupID, to.AppID, app); err != nil {
		return err
	}

	if err := deployDraftAndWait(ui, cmd.realmClient, to, draft.ID); err != nil {
		return err
	}

	// TODO(REALMC-7177): import hosting

	// TODO(REALMC-7868): import dependencies

	return nil
}

// Feedback is the command feedback
func (cmd *Command) Feedback(profile *cli.Profile, ui terminal.UI) error {
	if cmd.outputs.noDiffs {
		return ui.Print(terminal.NewTextLog("Deployed app is identical to proposed version, nothing to do"))
	}
	if cmd.outputs.diffRejected || cmd.outputs.draftKept {
		return ui.Print(terminal.NewTextLog("No changes were pushed to your Realm application"))
	}

	if cmd.outputs.appCreated {
		if cmd.inputs.DryRun {
			return ui.Print(
				terminal.NewTextLog("This is a new app. To create a new app, you must omit the 'dry-run' flag to proceed"),
				terminal.NewSuggestedCommandsLog(cmd.commandString(true)),
			)
		}
		return ui.Print(
			terminal.NewTextLog("This is a new app. You must create a new app to proceed"),
		)
	}

	return ui.Print(terminal.NewTextLog("Successfully pushed app changes"))
}

type namer interface{ Name() string }
type locationer interface{ Location() realm.Location }
type deploymentModeler interface{ DeploymentModel() realm.DeploymentModel }

func (cmd *Command) createNewApp(ui terminal.UI, groupID string, appData interface{}) (realm.App, error) {
	if proceed, err := ui.Confirm("Do you wish to create a new app?"); err != nil {
		return realm.App{}, err
	} else if !proceed {
		return realm.App{}, nil
	}

	var name string
	if n, ok := appData.(namer); ok {
		name = n.Name()
	}
	if name == "" || !ui.AutoConfirm() {
		if err := ui.AskOne(&name, &survey.Input{Message: "App Name", Default: name}); err != nil {
			return realm.App{}, err
		}
	}

	var location string
	if l, ok := appData.(locationer); ok {
		location = l.Location().String()
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
			return realm.App{}, err
		}
	}

	var deploymentModel string
	if dm, ok := appData.(deploymentModeler); ok {
		deploymentModel = dm.DeploymentModel().String()
	}
	if !ui.AutoConfirm() {
		if err := ui.AskOne(
			&deploymentModel,
			&survey.Select{
				Message: "App Deployment Model",
				Options: realm.DeploymentModelValues,
				Default: deploymentModel,
			}); err != nil {
			return realm.App{}, err
		}
	}

	app, appErr := cmd.realmClient.CreateApp(
		groupID,
		name,
		realm.AppMeta{Location: realm.Location(location), DeploymentModel: realm.DeploymentModel(deploymentModel)},
	)
	if appErr != nil {
		return realm.App{}, appErr
	}

	if appErr != nil {
		return realm.App{}, appErr
	}
	cmd.outputs.appCreated = true

	if err := local.AsApp(cmd.inputs.AppDirectory, app).WriteConfig(); err != nil {
		return realm.App{}, err
	}

	return app, nil
}

func createNewDraft(ui terminal.UI, realmClient realm.Client, to to) (realm.AppDraft, bool, error) {
	draft, draftErr := realmClient.CreateDraft(to.GroupID, to.AppID)
	if draftErr == nil {
		return draft, true, nil
	}

	if err, ok := draftErr.(realm.ServerError); !ok || err.Code != realm.ErrCodeDraftAlreadyExists {
		return realm.AppDraft{}, false, draftErr
	}

	existingDraft, existingDraftErr := realmClient.Draft(to.GroupID, to.AppID)
	if existingDraftErr != nil {
		return realm.AppDraft{}, false, existingDraftErr
	}

	if !ui.AutoConfirm() {
		if err := diffDraft(ui, realmClient, to, existingDraft.ID); err != nil {
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

	if err := realmClient.DiscardDraft(to.GroupID, to.AppID, existingDraft.ID); err != nil {
		return realm.AppDraft{}, false, err
	}

	draft, draftErr = realmClient.CreateDraft(to.GroupID, to.AppID)
	return draft, true, draftErr
}

func diffDraft(ui terminal.UI, realmClient realm.Client, to to, draftID string) error {
	diff, diffErr := realmClient.DiffDraft(to.GroupID, to.AppID, draftID)
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
	return ui.Print(logs...)
}

func deployDraftAndWait(ui terminal.UI, realmClient realm.Client, to to, draftID string) error {
	deployment, deploymentErr := realmClient.DeployDraft(to.GroupID, to.AppID, draftID)
	if deploymentErr != nil {
		return deploymentErr
	}

	for deployment.Status == realm.DeploymentStatusCreated || deployment.Status == realm.DeploymentStatusPending {
		// TODO(REALMC-7867): replace this Print statement with a spinner & status message (which goes away after the function completes)
		if err := ui.Print(terminal.NewTextLog("Checking on the status of your deployment...")); err != nil {
			return err
		}
		time.Sleep(time.Second)

		deployment, deploymentErr = realmClient.Deployment(to.GroupID, to.AppID, deployment.ID)
		if deploymentErr != nil {
			if err := realmClient.DiscardDraft(to.GroupID, to.AppID, draftID); err != nil {
				if e := ui.Print(terminal.NewWarningLog("We failed to discard the draft we created for your deployment")); e != nil {
					return e
				}
			}
			return deploymentErr
		}
	}
	return nil
}

func (cmd *Command) commandString(omitDryRun bool) string {
	sb := strings.Builder{}
	sb.WriteString("realm-cli push")

	// TODO(REALMC-7866): make this more accurate based on the inputs provided

	if cmd.inputs.DryRun && !omitDryRun {
		sb.WriteString(" --dry-run")
	}

	return sb.String()
}
