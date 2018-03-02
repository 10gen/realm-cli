package commands

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/10gen/stitch-cli/models"
	u "github.com/10gen/stitch-cli/user"
	"github.com/10gen/stitch-cli/utils"

	"github.com/mitchellh/cli"
)

// NewExportCommandFactory returns a new cli.CommandFactory given a cli.Ui
func NewExportCommandFactory(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		workingDirectory, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		return &ExportCommand{
			workingDirectory:  workingDirectory,
			exportToDirectory: utils.WriteZipToDir,
			BaseCommand: &BaseCommand{
				Name: "export",
				UI:   ui,
			},
		}, nil
	}
}

// ExportCommand is used to export a Stitch App
type ExportCommand struct {
	*BaseCommand

	workingDirectory  string
	exportToDirectory func(dest string, zipData io.Reader) error

	flagAppID  string
	flagOutput string
}

// Help returns long-form help information for this command
func (ec *ExportCommand) Help() string {
	return `Export a stitch application to a local directory.

REQUIRED:
  --app-id [string]
	The App ID for your app (i.e. the name of your app followed by a unique suffix, like "my-app-nysja")

OPTIONS:
  -o, --output [string]
	Directory to write the exported configuration. Defaults to "<app_name>_<timestamp>"` +
		ec.BaseCommand.Help()
}

// Synopsis returns a one-liner description for this command
func (ec *ExportCommand) Synopsis() string {
	return `Export a stitch application to a local directory.`
}

// Run executes the command
func (ec *ExportCommand) Run(args []string) int {
	set := ec.NewFlagSet()

	set.StringVar(&ec.flagAppID, flagAppIDName, "", "")
	set.StringVarP(&ec.flagOutput, "output", "o", "", "")

	if err := ec.BaseCommand.run(args); err != nil {
		ec.UI.Error(err.Error())
		return 1
	}

	if err := ec.run(); err != nil {
		ec.UI.Error(err.Error())
		return 1
	}

	return 0
}

func (ec *ExportCommand) run() error {
	if ec.flagAppID == "" {
		return errAppIDRequired
	}

	user, err := ec.User()
	if err != nil {
		return err
	}

	if !user.LoggedIn() {
		return u.ErrNotLoggedIn
	}

	if dir, err := utils.GetDirectoryContainingFile(ec.workingDirectory, models.AppConfigFileName); err == nil {
		return fmt.Errorf("cannot export within config directory %q", dir)
	}

	stitchClient, err := ec.StitchClient()
	if err != nil {
		return err
	}

	app, err := stitchClient.FetchAppByClientAppID(ec.flagAppID)
	if err != nil {
		return err
	}

	filename, body, err := stitchClient.Export(app.GroupID, app.ID)
	if err != nil {
		return err
	}

	defer body.Close()

	if ec.flagOutput != "" {
		filename = ec.flagOutput
	} else {
		filename = filename[:strings.LastIndex(filename, "_")]
	}

	return ec.exportToDirectory(filename, body)
}
