package app

import (
	"errors"
	"strings"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/AlecAivazis/survey/v2"

	"github.com/spf13/pflag"
)

type diffOutputs struct {
	noDiffs bool
}

type diffInputs struct {
	AppDirectory        string
	IncludeDependencies bool
	IncludeHosting      bool
	cli.ProjectInputs
}

// CommandDiff is the `app diff` command
type CommandDiff struct {
	inputs      diffInputs
	outputs     diffOutputs
	atlasClient atlas.Client
	realmClient realm.Client
}

// Flags is the command flags
func (cmd *CommandDiff) Flags(fs *pflag.FlagSet) {
	fs.StringVarP(&cmd.inputs.AppDirectory, flagDirectory, flagDirectoryShort, "", flagDirectoryUsage)
	fs.BoolVarP(&cmd.inputs.IncludeDependencies, flagIncludeDependencies, flagIncludeDependenciesShort, false, flagIncludeDependenciesUsage)
	fs.BoolVarP(&cmd.inputs.IncludeHosting, flagIncludeHosting, flagIncludeHostingShort, false, flagIncludeHostingUsage)

	cmd.inputs.Flags(fs)
}

// Inputs is the command inputs
func (cmd *CommandDiff) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Setup is the command setup
func (cmd *CommandDiff) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = profile.RealmAuthClient()
	return nil
}

// Handler is the command handler
func (cmd *CommandDiff) Handler(profile *cli.Profile, ui terminal.UI) error {
	app, err := local.LoadApp(cmd.inputs.AppDirectory)
	if err != nil {
		return err
	}

	apps, err := cmd.realmClient.FindApps(realm.AppFilter{
		GroupID: cmd.inputs.Project,
		App:     cmd.inputs.App,
	})
	if err != nil {
		return err
	}

	appMap := map[string]realm.App{}
	appNames := make([]string, 0, len(apps))
	appName := ""
	for _, app := range apps {
		if !strings.Contains(cmd.inputs.App, app.Name) {
			continue
		}
		appMap[app.Name] = app
		appNames = append(appNames, app.Name)
		appName = app.Name
	}

	if len(appMap) > 1 {
		err = ui.AskOne(
			&appName,
			&survey.Select{
				Message: "Which app do you want to diff?",
				Options: appNames,
			})
		if err != nil {
			return err
		}
	}

	appToDiff, ok := appMap[appName]
	if !ok {
		return errors.New("app not found")
	}

	diffs, err := cmd.realmClient.Diff(appToDiff.GroupID, appToDiff.ID, app)
	if err != nil {
		return err
	}

	if cmd.inputs.IncludeHosting {
		// TODO(REALMC-7177): diff hosting changes
		diffs = append(diffs, "Import hosting")
	}

	if cmd.inputs.IncludeDependencies {
		diffs = append(diffs, "Import dependencies")
	}

	if len(diffs) == 0 {
		// there are no diffs
		cmd.outputs.noDiffs = true
		return nil
	}

	err = ui.Print(terminal.NewTextLog(
		"The following reflects the proposed changes to your Realm app\n%s",
		strings.Join(diffs, "\n"),
	))
	if err != nil {
		return err
	}

	return nil
}

// Feedback is the command feedback
func (cmd *CommandDiff) Feedback(profile *cli.Profile, ui terminal.UI) error {
	if cmd.outputs.noDiffs {
		return ui.Print(terminal.NewTextLog("Deployed app is identical to proposed version"))
	}

	return ui.Print(terminal.NewTextLog("diff app successful"))
}
