package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/10gen/realm-cli/internal/app"
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// CommandInit is the `app init` command
type CommandInit struct {
	inputs      initInputs
	realmClient realm.Client
}

// Flags is the command flags
func (cmd *CommandInit) Flags(fs *pflag.FlagSet) {
	fs.StringVar(&cmd.inputs.Project, flagProject, "", flagProjectUsage)
	fs.StringVarP(&cmd.inputs.From, flagFrom, flagFromShort, "", flagFromUsage)
	fs.StringVarP(&cmd.inputs.Name, flagName, flagNameShort, "", flagNameUsage)
	fs.VarP(&cmd.inputs.DeploymentModel, flagDeploymentModel, flagDeploymentModelShort, flagDeploymentModelUsage)
	fs.VarP(&cmd.inputs.Location, flagLocation, flagLocationShort, flagLocationUsage)
}

// Inputs is the command inputs
func (cmd *CommandInit) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Setup is the command setup
func (cmd *CommandInit) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = realm.NewAuthClient(profile)
	return nil
}

// Handler is the command handler
func (cmd *CommandInit) Handler(profile *cli.Profile, ui terminal.UI) error {
	from, fromErr := cmd.inputs.resolveFrom(ui, cmd.realmClient)
	if fromErr != nil {
		return fromErr
	}

	if from.IsZero() {
		return cmd.initialize(profile.WorkingDirectory)
	}
	return cmd.initializeFromApp(profile.WorkingDirectory, from.GroupID, from.AppID)
}

// Feedback is the command feedback
func (cmd *CommandInit) Feedback(profile *cli.Profile, ui terminal.UI) error {
	return ui.Print(terminal.NewTextLog("Successfully initialized app"))
}

func (cmd *CommandInit) initialize(wd string) error {
	appConfig := app.Config{
		Data:            app.Data{Name: cmd.inputs.Name},
		ConfigVersion:   realm.DefaultAppConfigVersion,
		Location:        cmd.inputs.Location,
		DeploymentModel: cmd.inputs.DeploymentModel,
	}

	data, err := json.MarshalIndent(appConfig, app.ExportedJSONPrefix, app.ExportedJSONIndent)
	if err != nil {
		return fmt.Errorf("failed to write app config: %w", err)
	}

	return cli.WriteFile(filepath.Join(wd, app.FileConfig), 0666, bytes.NewReader(data))
}

func (cmd *CommandInit) initializeFromApp(wd, groupID, appID string) error {
	_, zipPkg, exportErr := cmd.realmClient.Export(groupID, appID, realm.ExportRequest{IsTemplated: true})
	if exportErr != nil {
		return exportErr
	}
	return cli.WriteZip(wd, zipPkg)
}
