package pull

import (
	"archive/zip"
	"path/filepath"
	"strings"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// Command is the `pull` command
type Command struct {
	inputs inputs
}

// Flags is the command flags
func (cmd *Command) Flags(fs *pflag.FlagSet) {
	fs.StringVarP(&cmd.inputs.Target, flagTarget, flagTargetShort, "", flagTargetUsage)
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

	path, zipPkg, err := cmd.doExport(profile, clients.Realm, from.GroupID, from.AppID)
	if err != nil {
		return err
	}

	if cmd.inputs.DryRun {
		ui.Print(
			terminal.NewTextLog("No changes were written to your file system"),
			terminal.NewDebugLog("Contents would have been written to: %s", path),
		)
		return nil
	}

	if err := local.WriteZip(path, zipPkg); err != nil {
		return err
	}

	if cmd.inputs.IncludeDependencies {
		archiveName, archivePkg, err := clients.Realm.ExportDependencies(from.GroupID, from.AppID)
		if err != nil {
			return err
		}

		archivePath := filepath.Join(path, local.NameFunctions, archiveName)
		if err := local.WriteFile(archivePath, 0755, archivePkg); err != nil {
			return err
		}
	}

	// TODO(REALMC-7177): include hosting

	ui.Print(terminal.NewTextLog("Successfully pulled down Realm app to your local filesystem"))
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

	path := cmd.inputs.Target
	if path == "" {
		if idx := strings.LastIndex(name, "_"); idx != -1 {
			name = name[:idx]
		}
		path = filepath.Join(profile.WorkingDirectory, name)
	}

	return path, zipPkg, nil
}
