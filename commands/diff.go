package commands

import (
	"io"
	"os"

	"github.com/10gen/realm-cli/models"
	"github.com/10gen/realm-cli/utils"
	"github.com/mitchellh/cli"
)

// NewDiffCommandFactory returns a new cli.CommandFactory given a cli.Ui
func NewDiffCommandFactory(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		workingDirectory, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		return &DiffCommand{
			BaseCommand: &BaseCommand{
				Name: "diff",
				UI:   ui,
			},
			workingDirectory: workingDirectory,
			writeToDirectory: utils.WriteZipToDir,
			writeAppConfigToFile: func(dest string, app models.AppInstanceData) error {
				return app.MarshalFile(dest)
			},
		}, nil
	}
}

// DiffCommand is used to view the changes you would make to the Realm App
type DiffCommand struct {
	*BaseCommand

	writeToDirectory     func(dest string, zipData io.Reader, overwrite bool) error
	writeAppConfigToFile func(dest string, app models.AppInstanceData) error
	workingDirectory     string

	flagAppID          string
	flagAppPath        string
	flagAppName        string
	flagGroupID        string
	flagStrategy       string
	flagIncludeHosting bool
}

// Help returns long-form help information for this command
func (dc *DiffCommand) Help() string {
	return `View your changes to an application from a local directory.

REQUIRED:
  --app-id [string]
	The App ID for your app (i.e. the name of your app followed by a unique suffix, like "my-app-nysja").

OPTIONS:
  --path [string]
	A path to the local directory containing your app.

  --project-id [string]
	The Atlas Project ID.

  --include-hosting
	Upload static assets from "/hosting" directory.
	` +
		dc.BaseCommand.Help()
}

// Synopsis returns a one-liner description for this command
func (dc *DiffCommand) Synopsis() string {
	return `View the changes you would make to the current app without importing the changes.`
}

// Run executes the command
func (dc *DiffCommand) Run(args []string) int {
	flags := dc.NewFlagSet()

	flags.StringVar(&dc.flagAppID, flagAppIDName, "", "")
	flags.StringVar(&dc.flagAppPath, importFlagPath, "", "")
	flags.StringVar(&dc.flagGroupID, flagProjectIDName, "", "")
	flags.BoolVar(&dc.flagIncludeHosting, importFlagIncludeHosting, false, "")

	if err := dc.BaseCommand.run(args); err != nil {
		dc.UI.Error(err.Error())
		return 1
	}

	ic := &ImportCommand{
		BaseCommand: dc.BaseCommand,

		writeToDirectory:     dc.writeToDirectory,
		writeAppConfigToFile: dc.writeAppConfigToFile,
		workingDirectory:     dc.workingDirectory,

		flagAppID:          dc.flagAppID,
		flagAppPath:        dc.flagAppPath,
		flagAppName:        dc.flagAppName,
		flagGroupID:        dc.flagGroupID,
		flagStrategy:       dc.flagStrategy,
		flagIncludeHosting: dc.flagIncludeHosting,
	}

	dryRun := true
	if err := ic.importApp(dryRun); err != nil {
		dc.UI.Error(err.Error())
		return 1
	}
	return 0
}
