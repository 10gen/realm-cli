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

	"github.com/briandowns/spinner"
	"github.com/spf13/pflag"
)

// Command is the `pull` command
type Command struct {
	inputs inputs
}

// Flags is the command flags
func (cmd *Command) Flags(fs *pflag.FlagSet) {
	fs.StringVarP(&cmd.inputs.To, flagTo, flagToShort, "", flagToUsage)
	fs.StringVar(&cmd.inputs.Project, flagProject, "", flagProjectUsage)
	fs.StringVarP(&cmd.inputs.From, flagFrom, flagFromShort, "", flagFromUsage)
	fs.Var(&cmd.inputs.AppVersion, flagAppVersion, flagAppVersionUsage)
	fs.BoolVarP(&cmd.inputs.DryRun, flagDryRun, flagDryRunShort, false, flagDryRunUsage)
	fs.BoolVarP(&cmd.inputs.IncludeDependencies, flagIncludeDependencies, flagIncludeDependenciesShort, false, flagIncludeDependenciesUsage)
	fs.BoolVarP(&cmd.inputs.IncludeHosting, flagIncludeHosting, flagIncludeHostingShort, false, flagIncludeHostingUsage)
}

// Inputs is the command inputs
func (cmd *Command) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *Command) Handler(profile *cli.Profile, ui terminal.UI, clients cli.Clients) error {
	from, err := cmd.inputs.resolveFrom(ui, clients.Realm)
	if err != nil {
		return err
	}

	pathTarget, zipPkg, err := cmd.doExport(profile, clients.Realm, from.GroupID, from.AppID)
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

			archiveName, archivePkg, err := clients.Realm.ExportDependencies(from.GroupID, from.AppID)
			if err != nil {
				return err
			}

			return local.WriteFile(
				filepath.Join(pathTarget, local.NameFunctions, archiveName),
				0755,
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

			appAssets, err := clients.Realm.HostingAssets(from.GroupID, from.AppID)
			if err != nil {
				return err
			}

			return local.WriteHostingAssets(clients.HostingAsset, pathTarget, from.GroupID, from.AppID, appAssets)
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
	name, zipPkg, exportErr := realmClient.Export(
		groupID,
		appID,
		realm.ExportRequest{ConfigVersion: cmd.inputs.AppVersion},
	)
	if exportErr != nil {
		return "", nil, exportErr
	}

	to := cmd.inputs.To
	if to == "" {
		if idx := strings.LastIndex(name, "_"); idx != -1 {
			name = name[:idx]
		}
		to = name
	}

	relativeToPath := to
	if filepath.IsAbs(to) {
		var err error
		relativeToPath, err = filepath.Rel(profile.WorkingDirectory, to)
		if err != nil {
			return "", nil, err
		}
	}

	target := filepath.Join(profile.WorkingDirectory, relativeToPath)

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
