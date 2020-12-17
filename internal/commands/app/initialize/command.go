package initialize

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// Command is the `app init` command
var Command = cli.CommandDefinition{
	Use:         "init",
	Aliases:     []string{"initialize"},
	Display:     "app init",
	Description: "Initialize a Realm app in your current local directory",
	Help:        "",
	Command:     &command{},
}

type command struct {
	inputs      inputs
	realmClient realm.Client
}

func (cmd *command) Flags(fs *pflag.FlagSet) {
	fs.StringVar(&cmd.inputs.Project, flagProject, "", flagProjectUsage)
	fs.StringVarP(&cmd.inputs.From, flagFrom, flagFromShort, "", flagFromUsage)
	fs.StringVarP(&cmd.inputs.Name, flagName, flagNameShort, "", flagNameUsage)
	fs.VarP(&cmd.inputs.DeploymentModel, flagDeploymentModel, flagDeploymentModelShort, flagDeploymentModelUsage)
	fs.VarP(&cmd.inputs.Location, flagLocation, flagLocationShort, flagLocationUsage)
}

func (cmd *command) Inputs() cli.InputResolver {
	return &cmd.inputs
}

func (cmd *command) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = realm.NewAuthClient(profile.RealmBaseURL(), profile.Session())
	return nil
}

func (cmd *command) Handler(profile *cli.Profile, ui terminal.UI) error {
	from, fromErr := cmd.inputs.resolveFrom(ui, cmd.realmClient)
	if fromErr != nil {
		return fromErr
	}

	if from.IsZero() {
		return cmd.initialize(profile.WorkingDirectory)
	}
	return cmd.initializeFromApp(profile.WorkingDirectory, from.GroupID, from.AppID)
}

func (cmd *command) Feedback(profile *cli.Profile, ui terminal.UI) error {
	return ui.Print(terminal.NewTextLog("Successfully initialized app"))
}

func (cmd *command) initialize(wd string) error {
	appConfig := cli.AppConfig{
		AppData:         cli.AppData{Name: cmd.inputs.Name},
		ConfigVersion:   realm.DefaultAppConfigVersion,
		Location:        cmd.inputs.Location,
		DeploymentModel: cmd.inputs.DeploymentModel,
	}

	data, err := json.MarshalIndent(appConfig, cli.ExportedJSONPrefix, cli.ExportedJSONIndent)
	if err != nil {
		return fmt.Errorf("failed to write app config: %w", err)
	}

	return cli.WriteFile(filepath.Join(wd, realm.FileAppConfig), 0666, bytes.NewReader(data))
}

func (cmd *command) initializeFromApp(wd, groupID, appID string) error {
	_, zipPkg, exportErr := cmd.realmClient.Export(groupID, appID, realm.ExportRequest{IsTemplated: true})
	if exportErr != nil {
		return exportErr
	}
	return cli.WriteZip(wd, zipPkg)
}
