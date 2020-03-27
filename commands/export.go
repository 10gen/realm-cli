package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/10gen/stitch-cli/api"
	"github.com/10gen/stitch-cli/models"
	u "github.com/10gen/stitch-cli/user"
	"github.com/10gen/stitch-cli/utils"

	"github.com/mitchellh/cli"
	"github.com/mitchellh/go-homedir"
)

const numWorkers = 4

// NewExportCommandFactory returns a new cli.CommandFactory given a cli.Ui
func NewExportCommandFactory(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		workingDirectory, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		return &ExportCommand{
			workingDirectory:     workingDirectory,
			exportToDirectory:    utils.WriteZipToDir,
			writeFileToDirectory: utils.WriteFileToDir,
			getAssetAtURL:        getAssetAtURL,
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

	workingDirectory     string
	exportToDirectory    func(dest string, zipData io.Reader, overwrite bool) error
	writeFileToDirectory func(dest string, data io.Reader) error
	getAssetAtURL        func(url string) (io.ReadCloser, error)

	flagProjectID           string
	flagAppID               string
	flagOutput              string
	flagAsTemplate          bool
	flagIncludeHosting      bool
	flagIncludeDependencies bool
	flagForSourceControl    bool
}

// Help returns long-form help information for this command
func (ec *ExportCommand) Help() string {
	return `Export a stitch application to a local directory.

REQUIRED:
  --app-id [string]
	The App ID for your app (i.e. the name of your app followed by a unique suffix, like "my-app-nysja")

OPTIONS:
  --project-id [string]
	Lookup apps associated with this project id, as opposed to ids associated with the current user profile.

  -o [string], --output [string]
	Directory to write the exported configuration. Defaults to "<app_name>_<timestamp>"

  --as-template
	Indicate that the application should be exported as a template.

  --for-source-control
	Indicate that the application should be exported for source control.

  --include-dependencies
	Download dependencies associated with this project

  --include-hosting
	Download static assets associated with this project` +
		ec.BaseCommand.Help()
}

// Synopsis returns a one-liner description for this command
func (ec *ExportCommand) Synopsis() string {
	return `Export a stitch application to a local directory.`
}

// Run executes the command
func (ec *ExportCommand) Run(args []string) int {
	set := ec.NewFlagSet()

	set.StringVar(&ec.flagProjectID, flagProjectIDName, "", "")
	set.StringVar(&ec.flagAppID, flagAppIDName, "", "")
	set.StringVar(&ec.flagOutput, "output", "", "")
	set.StringVar(&ec.flagOutput, "o", "", "")
	set.BoolVar(&ec.flagAsTemplate, "as-template", false, "")
	set.BoolVar(&ec.flagForSourceControl, "for-source-control", false, "")
	set.BoolVar(&ec.flagIncludeDependencies, "include-dependencies", false, "")
	set.BoolVar(&ec.flagIncludeHosting, "include-hosting", false, "")

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

	if dir, getErr := utils.GetDirectoryContainingFile(ec.workingDirectory, models.AppConfigFileName); getErr == nil {
		return fmt.Errorf("cannot export within config directory %q", dir)
	}

	stitchClient, err := ec.StitchClient()
	if err != nil {
		return err
	}

	var app *models.App
	if ec.flagProjectID == "" {
		app, err = stitchClient.FetchAppByClientAppID(ec.flagAppID)
		if err != nil {
			return err
		}
	} else {
		app, err = stitchClient.FetchAppByGroupIDAndClientAppID(ec.flagProjectID, ec.flagAppID)
		if err != nil {
			return err
		}
	}

	exportStrategy := api.ExportStrategyNone
	if ec.flagAsTemplate {
		exportStrategy = api.ExportStrategyTemplate
	} else if ec.flagForSourceControl {
		exportStrategy = api.ExportStrategySourceControl
	}

	filename, body, err := stitchClient.Export(app.GroupID, app.ID, exportStrategy)
	if err != nil {
		return err
	}
	defer body.Close()

	if ec.flagOutput != "" {
		filename, err = homedir.Expand(ec.flagOutput)
		if err != nil {
			return err
		}
	} else if lastUnderscoreIdx := strings.LastIndex(filename, "_"); lastUnderscoreIdx != -1 {
		filename = filename[:lastUnderscoreIdx]
	}

	if err := ec.exportToDirectory(filename, body, false); err != nil {
		return err
	}

	if ec.flagIncludeDependencies {
		depArchive, depBody, err := stitchClient.ExportDependencies(app.GroupID, app.ID)
		if err != nil {
			return err
		}
		defer depBody.Close()
		functionsDir := filepath.Join(filename, utils.FunctionsRoot, depArchive)
		err = ec.writeFileToDirectory(functionsDir, depBody)
		if err != nil {
			return err
		}

	}

	if ec.flagIncludeHosting {
		if err := exportStaticHostingAssets(stitchClient, ec, filename, app); err != nil {
			return err
		}
	}
	return nil
}
