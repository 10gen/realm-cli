package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/10gen/stitch-cli/models"
	u "github.com/10gen/stitch-cli/user"
	"github.com/10gen/stitch-cli/utils"

	"github.com/mitchellh/cli"
	"github.com/mitchellh/go-homedir"
)

// NewImportCommandFactory returns a new cli.CommandFactory given a cli.Ui
func NewImportCommandFactory(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		workingDirectory, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		return &ImportCommand{
			BaseCommand: &BaseCommand{
				Name: "import",
				UI:   ui,
			},
			workingDirectory: workingDirectory,
			writeToDirectory: utils.WriteZipToDir,
		}, nil
	}
}

// ImportCommand is used to import a Stitch App
type ImportCommand struct {
	*BaseCommand

	writeToDirectory func(dest string, zipData io.Reader) error
	workingDirectory string

	flagAppID   string
	flagAppPath string
}

// Help returns long-form help information for this command
func (ic *ImportCommand) Help() string {
	return `Import and deploy a stitch application from a local directory.

REQUIRED:
  --app-id [string]
	The App ID for your app (i.e. the name of your app followed by a unique suffix, like "my-app-nysja")

OPTIONS:
  --path [string]
	A path to the local directory containing your app
	` +
		ic.BaseCommand.Help()
}

// Synopsis returns a one-liner description for this command
func (ic *ImportCommand) Synopsis() string {
	return `Import and deploy a stitch application from a local directory.`
}

// Run executes the command
func (ic *ImportCommand) Run(args []string) int {
	set := ic.NewFlagSet()

	set.StringVar(&ic.flagAppID, flagAppIDName, "", "")
	set.StringVar(&ic.flagAppPath, "path", "", "")

	if err := ic.BaseCommand.run(args); err != nil {
		ic.UI.Error(err.Error())
		return 1
	}

	if err := ic.importApp(); err != nil {
		ic.UI.Error(err.Error())
		return 1
	}

	return 0
}

func (ic *ImportCommand) importApp() error {
	user, err := ic.User()
	if err != nil {
		return err
	}

	if !user.LoggedIn() {
		return u.ErrNotLoggedIn
	}

	appPath, err := ic.resolveAppDirectory()
	if err != nil {
		return err
	}

	appInstanceData, err := ic.resolveAppInstanceData(appPath)
	if err != nil {
		return err
	}

	loadedApp, err := utils.UnmarshalFromDir(appPath)
	if err != nil {
		return err
	}

	appData, err := json.Marshal(loadedApp)
	if err != nil {
		return err
	}

	stitchClient, err := ic.StitchClient()
	if err != nil {
		return err
	}

	app, err := stitchClient.FetchAppByClientAppID(appInstanceData.AppID)
	if err != nil {
		return err
	}

	// Diff changes unless -y flag has been provided
	if !ic.flagYes {
		diffs, err := stitchClient.Diff(app.GroupID, app.ID, appData)
		if err != nil {
			return err
		}

		for _, diff := range diffs {
			ic.UI.Info(diff)
		}

		confirm, err := ic.Ask("Please confirm the changes shown above:")
		if err != nil {
			return err
		}

		if !confirm {
			return nil
		}
	}

	if err := stitchClient.Import(app.GroupID, app.ID, appData); err != nil {
		return err
	}

	// re-fetch imported app to sync IDs
	_, body, err := stitchClient.Export(app.GroupID, app.ID)
	if err != nil {
		return fmt.Errorf("failed to sync app with local directory after import: %s", err)
	}

	defer body.Close()

	if err := ic.writeToDirectory(appPath, body); err != nil {
		return fmt.Errorf("failed to sync app with local directory after import: %s", err)
	}

	return nil
}

func (ic *ImportCommand) resolveAppDirectory() (string, error) {
	if ic.flagAppPath != "" {
		path, err := homedir.Expand(ic.flagAppPath)
		if err != nil {
			return "", err
		}

		if _, err := os.Stat(path); err != nil {
			return "", errors.New("directory does not exist")
		}
		return path, nil
	}

	return utils.GetDirectoryContainingFile(ic.workingDirectory, models.AppConfigFileName)
}

// resolveAppInstanceData loads data for an app from a .stitch file located in the provided directory path
func (ic *ImportCommand) resolveAppInstanceData(path string) (*models.AppInstanceData, error) {
	appInstanceData := &models.AppInstanceData{
		AppID: ic.flagAppID,
	}

	if appInstanceData.AppID == "" {
		if err := mergeAppInstanceDataFromPath(appInstanceData, path); err != nil {
			return nil, err
		}
	}

	if appInstanceData.AppID == "" {
		return nil, errAppIDRequired
	}

	return appInstanceData, nil
}

func mergeAppInstanceDataFromPath(appInstanceData *models.AppInstanceData, path string) error {
	var appInstanceDataFromDotfile models.AppInstanceData
	err := appInstanceDataFromDotfile.UnmarshalFile(path)

	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return err
	}

	if appInstanceData.AppID == "" {
		appInstanceData.AppID = appInstanceDataFromDotfile.AppID
	}

	return nil
}
