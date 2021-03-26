package pull

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/briandowns/spinner"
	"github.com/spf13/pflag"
)

// Command is the `pull` command
type Command struct {
	inputs inputs
}

// Flags is the command flags
func (cmd *Command) Flags(fs *pflag.FlagSet) {
	fs.StringVar(&cmd.inputs.LocalPath, flagLocalPath, "", flagLocalPathUsage)
	fs.StringVar(&cmd.inputs.RemoteApp, flagRemote, "", flagRemoteUsage)
	fs.BoolVarP(&cmd.inputs.IncludeDependencies, flagIncludeDependencies, flagIncludeDependenciesShort, false, flagIncludeDependenciesUsage)
	fs.BoolVarP(&cmd.inputs.IncludeHosting, flagIncludeHosting, flagIncludeHostingShort, false, flagIncludeHostingUsage)
	fs.BoolVarP(&cmd.inputs.DryRun, flagDryRun, flagDryRunShort, false, flagDryRunUsage)

	fs.StringVar(&cmd.inputs.Project, flagProject, "", flagProjectUsage)
	flags.MarkHidden(fs, flagProject)

	fs.Var(&cmd.inputs.AppVersion, flagConfigVersion, flagConfigVersionUsage)
	flags.MarkHidden(fs, flagConfigVersion)
}

// Inputs is the command inputs
func (cmd *Command) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *Command) Handler(profile *cli.Profile, ui terminal.UI, clients cli.Clients) error {
	appRemote, err := cmd.inputs.resolveRemoteApp(ui, clients.Realm)
	if err != nil {
		return err
	}

	pathTarget, zipPkg, err := cmd.doExport(profile, clients.Realm, appRemote.GroupID, appRemote.AppID)
	if err != nil {
		return err
	}

	pathRelative, err := filepath.Rel(profile.WorkingDirectory, pathTarget)
	if err != nil {
		return err
	}

	proceed, err := checkAppDestination(ui, pathTarget)
	if err != nil {
		return err
	} else if !proceed {
		return nil
	}

	if cmd.inputs.DryRun {
		ui.Print(
			terminal.NewTextLog("No changes were written to your file system"),
			terminal.NewDebugLog("Contents would have been written to: %s", pathRelative),
		)
		return nil
	}

	if err := local.WriteZip(pathTarget, zipPkg); err != nil {
		return err
	}
	ui.Print(terminal.NewTextLog("Saved app to disk"))

	if cmd.inputs.IncludeDependencies {
		s := spinner.New(terminal.SpinnerCircles, 250*time.Millisecond)
		s.Suffix = " Fetching dependencies archive..."

		exportDependencies := func() error {
			s.Start()
			defer s.Stop()

			archiveName, archivePkg, err := clients.Realm.ExportDependencies(appRemote.GroupID, appRemote.AppID)
			if err != nil {
				return err
			}

			return local.WriteFile(
				filepath.Join(pathTarget, local.NameFunctions, archiveName),
				0666,
				archivePkg,
			)
		}

		if err := exportDependencies(); err != nil {
			return err
		}
		ui.Print(terminal.NewTextLog("Fetched dependencies archive"))
	}

	if cmd.inputs.IncludeHosting {
		s := spinner.New(terminal.SpinnerCircles, 250*time.Millisecond)
		s.Suffix = " Fetching hosting assets..."

		exportHostingAssets := func() error {
			s.Start()
			defer s.Stop()

			appAssets, err := clients.Realm.HostingAssets(appRemote.GroupID, appRemote.AppID)
			if err != nil {
				return err
			}

			return local.WriteHostingAssets(clients.HostingAsset, pathTarget, appRemote.GroupID, appRemote.AppID, appAssets)
		}

		if err := exportHostingAssets(); err != nil {
			return err
		}
		ui.Print(terminal.NewDebugLog("Fetched hosting assets"))
	}

	ui.Print(terminal.NewTextLog("Successfully pulled app down: %s", pathRelative))
	return nil
}

func (cmd *Command) doExport(profile *cli.Profile, realmClient realm.Client, groupID, appID string) (string, *zip.Reader, error) {
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

func checkAppDestination(ui terminal.UI, path string) (bool, error) {
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
