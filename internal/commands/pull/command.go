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
	inputs      inputs
	outputs     outputs
	realmClient realm.Client
}

type outputs struct {
	destination string
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

// Setup is the command setup
func (cmd *Command) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = profile.RealmAuthClient()
	return nil
}

// Handler is the command handler
func (cmd *Command) Handler(profile *cli.Profile, ui terminal.UI) error {
	from, fromErr := cmd.inputs.resolveFrom(ui, cmd.realmClient)
	if fromErr != nil {
		return fromErr
	}

	path, zipPkg, exportErr := cmd.doExport(profile, from)
	if exportErr != nil {
		return exportErr
	}
	cmd.outputs.destination = path

	if cmd.inputs.DryRun {
		return nil
	}

	if err := local.WriteZip(path, zipPkg); err != nil {
		return err
	}

	if cmd.inputs.IncludeDependencies {
		archiveName, archivePkg, archiveErr := cmd.realmClient.ExportDependencies(from.GroupID, from.AppID)
		if archiveErr != nil {
			return archiveErr
		}

		archivePath := filepath.Join(path, local.NameFunctions, archiveName)
		if err := local.WriteFile(archivePath, 0755, archivePkg); err != nil {
			return err
		}
	}

	// TODO(REALMC-7177): include hosting

	return nil
}

// Feedback is the command feedback
func (cmd *Command) Feedback(profile *cli.Profile, ui terminal.UI) error {
	if cmd.inputs.DryRun {
		return ui.Print(
			terminal.NewTextLog("No changes were written to your file system"),
			terminal.NewDebugLog("Contents would have been written to: %s", cmd.outputs.destination),
		)
	}
	return ui.Print(terminal.NewTextLog("Successfully pulled down Realm app to your local filesystem"))
}

func (cmd *Command) doExport(profile *cli.Profile, f from) (string, *zip.Reader, error) {
	name, zipPkg, exportErr := cmd.realmClient.Export(
		f.GroupID,
		f.AppID,
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
