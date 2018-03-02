package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/10gen/stitch-cli/api"
	"github.com/10gen/stitch-cli/models"
	u "github.com/10gen/stitch-cli/user"
	"github.com/10gen/stitch-cli/utils"

	"github.com/mitchellh/cli"
	"github.com/mitchellh/go-homedir"
)

const (
	importFlagPath        = "path"
	importFlagStrategy    = "strategy"
	importFlagAppName     = "app-name"
	importFlagGroupID     = "group-id"
	importStrategyMerge   = "merge"
	importStrategyReplace = "replace"
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

	writeToDirectory func(dest string, zipData io.Reader, overwrite bool) error
	workingDirectory string

	flagAppID    string
	flagAppPath  string
	flagAppName  string
	flagGroupID  string
	flagStrategy string
}

// Help returns long-form help information for this command
func (ic *ImportCommand) Help() string {
	return `Import and deploy a stitch application from a local directory.

REQUIRED:
  --app-id [string]
	The App ID for your app (i.e. the name of your app followed by a unique suffix, like "my-app-nysja").

  --app-name [string]
	The name of your app to be used if app is to be created new.

OPTIONS:
  --path [string]
	A path to the local directory containing your app.

  --group-id [string]
	The Atlas Group ID.

  --strategy [merge|replace] (default: merge)
	How your app should be imported.

	merge - import and overwrite existing entities while preserving those that exist on Stitch. Secrets missing will not be lost.
	replace - like merge but does not preserve entities missing from the local directory's app configuration.
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
	set.StringVar(&ic.flagAppPath, importFlagPath, "", "")
	set.StringVar(&ic.flagGroupID, importFlagGroupID, "", "")
	set.StringVar(&ic.flagAppName, importFlagAppName, "", "")
	set.StringVar(&ic.flagStrategy, importFlagStrategy, importStrategyMerge, "")

	if err := ic.BaseCommand.run(args); err != nil {
		ic.UI.Error(err.Error())
		return 1
	}

	if ic.flagStrategy != importStrategyMerge && ic.flagStrategy != importStrategyReplace {
		ic.UI.Error(fmt.Sprintf("unknown import strategy %q; accepted values are [%s|%s]", ic.flagStrategy, importStrategyMerge, importStrategyReplace))
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
	var appNotFound bool
	if err != nil {
		switch err.(type) {
		case api.ErrAppNotFound:
			appNotFound = true
			if appInstanceData.AppID == "" {
				err = errors.New("this app does not exist yet")
			}
		default:
			return err
		}
	}

	var skipDiff bool

	if appNotFound {
		skipDiff = true
		ic.flagStrategy = importStrategyReplace

		var wantedNewApp bool
		app, wantedNewApp, err = ic.askCreateEmptyApp(err.Error(), appInstanceData.AppName, stitchClient)
		if err != nil {
			return err
		}
		if !wantedNewApp {
			return nil
		}
	}

	// Diff changes unless -y flag has been provided
	if !ic.flagYes && !skipDiff {
		diffs, err := stitchClient.Diff(app.GroupID, app.ID, appData, ic.flagStrategy)
		if err != nil {
			return err
		}

		for _, diff := range diffs {
			ic.UI.Info(diff)
		}

		confirm, err := ic.AskYesNo("Please confirm the changes shown above:")
		if err != nil {
			return err
		}

		if !confirm {
			return nil
		}
	}

	if err := stitchClient.Import(app.GroupID, app.ID, appData, ic.flagStrategy); err != nil {
		return err
	}

	// re-fetch imported app to sync IDs
	_, body, err := stitchClient.Export(app.GroupID, app.ID)
	if err != nil {
		return fmt.Errorf("failed to sync app with local directory after import: %s", err)
	}

	defer body.Close()

	if err := ic.writeToDirectory(appPath, body, true); err != nil {
		return fmt.Errorf("failed to sync app with local directory after import: %s", err)
	}

	return nil
}

func (ic *ImportCommand) askCreateEmptyApp(query string, defaultAppName string, stitchClient api.StitchClient) (*models.App, bool, error) {
	if ic.flagAppName != "" {
		defaultAppName = ic.flagAppName
	}

	confirm, err := ic.AskYesNo(fmt.Sprintf("%s: would you like to create a new app?", query))
	if err != nil {
		return nil, false, err
	}

	if !confirm {
		return nil, false, nil
	}
	groupID, err := ic.Ask("Atlas Group ID", ic.flagGroupID)
	if err != nil {
		return nil, false, err
	}

	appName, err := ic.Ask("App name", defaultAppName)
	if err != nil {
		return nil, false, err
	}

	apps, err := stitchClient.FetchAppsByGroupID(groupID)
	if err != nil {
		return nil, false, err
	}

	for _, app := range apps {
		if app.Name == appName {
			return nil, false, fmt.Errorf("app already exists with name %q", appName)
		}
	}

	app, err := stitchClient.CreateEmptyApp(groupID, appName)
	if err != nil {
		return nil, false, err
	}

	ic.UI.Info(fmt.Sprintf("New app created and imported: %s", app.ClientAppID))
	return app, true, nil
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
		appInstanceData.AppName = appInstanceDataFromDotfile.AppName
	}

	return nil
}
