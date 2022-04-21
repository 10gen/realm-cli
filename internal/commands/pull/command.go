package pull

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

const (
	flagIncludeNodeModules  = "include-node-modules"
	flagIncludePackageJSON  = "include-package-json"
	flagIncludeDependencies = "include-dependencies"
)

// CommandMeta is the command meta for the `pull` command
var CommandMeta = cli.CommandMeta{
	Use:         "pull",
	Aliases:     []string{"export"},
	Description: "Exports the latest version of your Realm app into your local directory",
	HelpText: `Pulls changes from your remote Realm app into your local directory. If
applicable, Hosting Files and/or Dependencies associated with your Realm app will be
exported as well.`,
}

// Command is the `pull` command
type Command struct {
	inputs inputs
}

// Flags is the command flags
func (cmd *Command) Flags() []flags.Flag {
	return []flags.Flag{
		flags.StringFlag{
			Value: &cmd.inputs.LocalPath,
			Meta: flags.Meta{
				Name: "local",
				Usage: flags.Usage{
					Description: "Specify a local filepath to export a Realm app to",
				},
			},
		},
		flags.StringFlag{
			Value: &cmd.inputs.RemoteApp,
			Meta: flags.Meta{
				Name: "remote",
				Usage: flags.Usage{
					Description: "Specify the name or ID of a remote Realm app to export",
				},
			},
		},
		flags.BoolFlag{
			Value: &cmd.inputs.IncludeNodeModules,
			Meta: flags.Meta{
				Name: flagIncludeNodeModules,
				Usage: flags.Usage{
					Description: "Export and include Realm app dependencies from a node_modules archive",
					Note:        "The allowed formats are as a directory or compressed into a .zip, .tar, .tar.gz, or .tgz file",
				},
			},
		},
		flags.BoolFlag{
			Value: &cmd.inputs.IncludePackageJSON,
			Meta: flags.Meta{
				Name: flagIncludePackageJSON,
				Usage: flags.Usage{
					Description: "Export and include Realm app dependencies from a package.json file",
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
					Description: "Export and include Realm app dependencies from a node_modules archive",
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
					Description: "Export and include Realm app hosting files",
				},
			},
		},
		flags.BoolFlag{
			Value: &cmd.inputs.DryRun,
			Meta: flags.Meta{
				Name:      "dry-run",
				Shorthand: "x",
				Usage: flags.Usage{
					Description: "Run without writing any changes to the local filepath",
				},
			},
		},
		flags.StringSliceFlag{
			Value: &cmd.inputs.TemplateIDs,
			Meta: flags.Meta{
				Name:      "template",
				Shorthand: "t",
				Usage: flags.Usage{
					Description:   "Specify the frontend Template ID(s) to export.",
					Note:          "Specified templates must be compatible with the remote app",
					AllowedValues: realm.AllowedTemplates,
				},
			},
		},
		cli.ProjectFlag(&cmd.inputs.Project),
		cli.ConfigVersionFlag(&cmd.inputs.AppVersion, "Specify the app config version to export as"),
	}
}

// Inputs is the command inputs
func (cmd *Command) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *Command) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	app, err := cmd.inputs.resolveRemoteApp(ui, clients)
	if err != nil {
		return err
	}

	var clientTemplates []clientTemplate
	if app.TemplateID != "" {
		clientTemplates, err = cmd.inputs.resolveClientTemplates(clients.Realm, app.GroupID, app.ID)
		if err != nil {
			return err
		}
	}

	pathProject, zipPkg, err := cmd.doExport(profile, clients.Realm, app.GroupID, app.ID)
	if err != nil {
		return err
	}

	// App path
	proceed, err := checkPathDestination(ui, pathProject)
	if err != nil {
		return err
	} else if !proceed {
		return nil
	}

	pathRelative, err := filepath.Rel(profile.WorkingDirectory, pathProject)
	if err != nil {
		return err
	}

	var pathFrontend string
	pathBackend := pathProject
	if len(clientTemplates) != 0 {
		pathFrontend = filepath.Join(pathProject, local.FrontendPath)
		pathBackend = filepath.Join(pathProject, local.BackendPath)
	}

	if cmd.inputs.DryRun {
		logs := make([]terminal.Log, 0, 3)
		logs = append(logs, terminal.NewTextLog("No changes were written to your file system"))

		if len(clientTemplates) != 0 {
			logs = append(logs,
				terminal.NewDebugLog("App contents would have been written to: %s", filepath.Join(pathRelative, local.BackendPath)),
				terminal.NewDebugLog("Template contents would have been written to: %s", filepath.Join(pathRelative, local.FrontendPath)),
			)
		} else {
			logs = append(logs, terminal.NewDebugLog("Contents would have been written to: %s", pathRelative))
		}

		ui.Print(logs...)
		return nil
	}

	if err := local.WriteZip(pathBackend, zipPkg); err != nil {
		return fmt.Errorf("unable to write app to disk: %s", err)
	}
	ui.Print(terminal.NewTextLog("Saved app to disk"))

	if cmd.inputs.IncludeNodeModules || cmd.inputs.IncludePackageJSON || cmd.inputs.IncludeDependencies {
		logStr := "as a node_modules archive"
		if cmd.inputs.IncludePackageJSON {
			logStr = "as a package.json file"
		}

		s := ui.Spinner(fmt.Sprintf("Fetching dependencies %s...", logStr), terminal.SpinnerOptions{})

		var packageJSONMissing bool

		exportDependencies := func() error {
			s.Start()
			defer s.Stop()

			fileName, file, err := cmd.exportDependencies(clients, app)
			if err != nil {
				return err
			}

			if cmd.inputs.IncludePackageJSON && fileName != local.NamePackageJSON {
				packageJSONMissing = true
			}

			return local.WriteFile(
				filepath.Join(pathBackend, local.NameFunctions, fileName),
				0666,
				file,
			)
		}

		if err := exportDependencies(); err != nil {
			return err
		}

		if packageJSONMissing {
			logStr = "as a node_modules archive"
		}
		ui.Print(terminal.NewTextLog("Fetched dependencies " + logStr))
		if packageJSONMissing {
			ui.Print(terminal.NewWarningLog("The package.json file was not found, a node_modules archive was written instead"))
		}
	}

	if cmd.inputs.IncludeHosting {
		s := ui.Spinner("Fetching hosting assets...", terminal.SpinnerOptions{})

		exportHostingAssets := func() error {
			s.Start()
			defer s.Stop()

			appAssets, err := clients.Realm.HostingAssets(app.GroupID, app.ID)
			if err != nil {
				return err
			}

			return local.WriteHostingAssets(clients.HostingAsset, pathBackend, app.GroupID, app.ID, appAssets)
		}

		if err := exportHostingAssets(); err != nil {
			return err
		}
		ui.Print(terminal.NewDebugLog("Fetched hosting assets"))
	}

	successfulTemplateWrites := make([]interface{}, 0, len(clientTemplates))
	for _, ct := range clientTemplates {
		if err := local.WriteZip(filepath.Join(pathFrontend, ct.id), ct.zipPkg); err != nil {
			return fmt.Errorf("unable to save template '%s' to disk: %s", ct.id, err)
		}
		readmeFile, err := local.FindReadme(pathProject, pathRelative, ct.id)
		if err != nil {
			return err
		}
		successfulTemplateWrites = append(successfulTemplateWrites, fmt.Sprintf("%s: %s", ct.id, readmeFile))
	}

	ui.Print(terminal.NewTextLog("Successfully pulled app down: %s", pathRelative))
	if len(successfulTemplateWrites) != 0 {
		ui.Print(terminal.NewListLog("Successfully saved template(s) to disk", successfulTemplateWrites...))
		ui.Print(terminal.NewTextLog("Navigate to the saved directory to view directions on how to run the template app(s)"))
	}

	return nil
}

func (cmd *Command) doExport(profile *user.Profile, realmClient realm.Client, groupID, appID string) (string, *zip.Reader, error) {
	name, zipPkg, err := realmClient.Export(
		groupID,
		appID,
		realm.ExportRequest{ConfigVersion: cmd.inputs.AppVersion},
	)
	if err != nil {
		return "", nil, err
	}

	pathLocal := cmd.inputs.LocalPath
	if pathLocal == "" {
		if idx := strings.LastIndex(name, "_"); idx != -1 {
			name = name[:idx]
		}
		pathLocal = name
	}

	if filepath.IsAbs(pathLocal) {
		pathLocal, err = filepath.Rel(profile.WorkingDirectory, pathLocal)
		if err != nil {
			return "", nil, err
		}
	}

	target := filepath.Join(profile.WorkingDirectory, pathLocal)

	return target, zipPkg, nil
}

func checkPathDestination(ui terminal.UI, path string) (bool, error) {
	if ui.AutoConfirm() {
		return true, nil
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}

	if !fileInfo.IsDir() {
		return true, nil
	}

	return ui.Confirm("Directory '%s' already exists, do you still wish to proceed?", path)
}

func (cmd *Command) exportDependencies(clients cli.Clients, app realm.App) (string, io.ReadCloser, error) {
	if cmd.inputs.IncludePackageJSON {
		return clients.Realm.ExportDependencies(app.GroupID, app.ID)
	}
	return clients.Realm.ExportDependenciesArchive(app.GroupID, app.ID)
}
