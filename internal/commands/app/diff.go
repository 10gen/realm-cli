package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/AlecAivazis/survey/v2"
)

const (
	flagIncludeNodeModules  = "include-node-modules"
	flagIncludePackageJSON  = "include-package-json"
	flagIncludeDependencies = "include-dependencies"
)

const (
	errDependencyFlagConflictTemplate = `cannot use both "%s" and "%s" at the same time`
)

// CommandMetaDiff is the command meta
var CommandMetaDiff = cli.CommandMeta{
	Use:         "diff",
	Aliases:     []string{},
	Display:     "app diff",
	Description: "Show differences between your local directory and your Realm app",
	HelpText: `Displays file-by-file differences between your local directory and the latest
version of your Realm app. If you have more than one Realm app, you will be
prompted to select a Realm app to view.`,
}

// CommandDiff is the `app diff` command
type CommandDiff struct {
	inputs diffInputs
}

type diffInputs struct {
	LocalPath           string
	RemoteApp           string
	Project             string
	IncludeDependencies bool
	IncludeNodeModules  bool
	IncludePackageJSON  bool
	IncludeHosting      bool
}

// Flags is the command flags
func (cmd *CommandDiff) Flags() []flags.Flag {
	return []flags.Flag{
		flags.StringFlag{
			Value: &cmd.inputs.LocalPath,
			Meta: flags.Meta{
				Name: "local",
				Usage: flags.Usage{
					Description: "Specify the local filepath of a Realm app to diff",
				},
			},
		},
		flags.StringFlag{
			Value: &cmd.inputs.RemoteApp,
			Meta: flags.Meta{
				Name: "remote",
				Usage: flags.Usage{
					Description: "Specify the name or ID of a Realm app to diff",
				},
			},
		},
		flags.BoolFlag{
			Value: &cmd.inputs.IncludeNodeModules,
			Meta: flags.Meta{
				Name: flagIncludeNodeModules,
				Usage: flags.Usage{
					Description: "Include Realm app dependencies in the diff from a node_modules archive",
					Note:        "The allowed formats are as a directory or compressed into a .zip, .tar, .tar.gz, or .tgz file",
				},
			},
		},
		flags.BoolFlag{
			Value: &cmd.inputs.IncludePackageJSON,
			Meta: flags.Meta{
				Name: flagIncludePackageJSON,
				Usage: flags.Usage{
					Description: "Include Realm app dependencies in the diff from a package.json file",
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
					Description: "Include Realm app dependencies in the diff from a node_modules archive",
					Note:        "The allowed formats are as a directory or compressed into a .zip, .tar, .tar.gz, or .tgz file",
				},
				Deprecate: fmt.Sprintf("support will be removed in v3.x, please use %q instead", flagIncludeNodeModules),
			},
		},
		flags.BoolFlag{
			Value: &cmd.inputs.IncludeHosting,
			Meta: flags.Meta{
				Name:      "include-hosting",
				Shorthand: "s",
				Usage: flags.Usage{
					Description: "Include Realm app hosting files in the diff",
				},
			},
		},
		cli.ProjectFlag(&cmd.inputs.Project),
	}
}

// Inputs is the command inputs
func (cmd *CommandDiff) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *CommandDiff) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	app, err := local.LoadApp(cmd.inputs.LocalPath)
	if err != nil {
		return err
	}

	appToDiff, err := cli.ResolveApp(ui, clients.Realm, cli.AppOptions{
		Filter:  realm.AppFilter{GroupID: cmd.inputs.Project, App: cmd.inputs.RemoteApp},
		AppMeta: app.Meta,
	})
	if err != nil {
		return err
	}

	diffs, err := clients.Realm.Diff(appToDiff.GroupID, appToDiff.ID, app.AppData)
	if err != nil {
		return err
	}

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

		dependenciesDiff, err := clients.Realm.DiffDependencies(appToDiff.GroupID, appToDiff.ID, uploadPath)
		if err != nil {
			return err
		}
		diffs = append(diffs, dependenciesDiff.Strings()...)
	}

	if cmd.inputs.IncludeHosting {
		hosting, err := local.FindAppHosting(app.RootDir)
		if err != nil {
			return err
		}

		appAssets, err := clients.Realm.HostingAssets(appToDiff.GroupID, appToDiff.ID)
		if err != nil {
			return err
		}

		hostingDiffs, err := hosting.Diffs(profile.HostingAssetCachePath(), appToDiff.ID, appAssets)
		if err != nil {
			return err
		}

		diffs = append(diffs, hostingDiffs.Strings()...)
	}

	if len(diffs) == 0 {
		// there are no diffs
		ui.Print(terminal.NewTextLog("Deployed app is identical to proposed version"))
		return nil
	}

	ui.Print(terminal.NewTextLog(
		"The following reflects the proposed changes to your Realm app\n%s",
		strings.Join(diffs, "\n"),
	))

	return nil
}

func (i *diffInputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	if i.IncludePackageJSON {
		if i.IncludeNodeModules {
			return fmt.Errorf(errDependencyFlagConflictTemplate, flagIncludeNodeModules, flagIncludePackageJSON)
		}
		if i.IncludeDependencies {
			return fmt.Errorf(errDependencyFlagConflictTemplate, flagIncludeDependencies, flagIncludePackageJSON)
		}
	}

	searchPath := i.LocalPath
	if searchPath == "" {
		searchPath = profile.WorkingDirectory
	}

	searchPathAbs, err := filepath.Abs(searchPath)
	if err != nil {
		return err
	}

	if _, err = os.Stat(searchPathAbs); os.IsNotExist(err) {
		return errProjectInvalid(searchPath, false)
	}

	app, _, err := local.FindApp(searchPath)
	if err != nil {
		return err
	}

	if i.LocalPath == "" && app.RootDir == "" {
		if err := ui.AskOne(&i.LocalPath, &survey.Input{Message: "App filepath (local)"}); err != nil {
			return err
		}

		app, _, err = local.FindApp(i.LocalPath)
		if err != nil {
			return err
		}
	}

	if app.RootDir == "" {
		return errProjectInvalid(i.LocalPath, true)
	}
	i.LocalPath = app.RootDir

	if i.RemoteApp == "" && app.Meta.AppID == "" {
		i.RemoteApp = app.Option()
	}

	return nil
}

func (i *diffInputs) resolveAppDependencies(rootDir string) (local.Dependencies, error) {
	if i.IncludePackageJSON {
		return local.FindPackageJSON(rootDir)
	}
	return local.FindNodeModules(rootDir)
}
