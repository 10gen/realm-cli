package app

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

// CommandMetaInit is the command meta for the `app init` command
var CommandMetaInit = cli.CommandMeta{
	Use:         "init",
	Aliases:     []string{"initialize"},
	Display:     "app init",
	Description: "Initialize a Realm app in your current working directory",
	HelpText: `Initializes a new Realm app by saving your configuration files in your current
working directory.

You can specify a "--remote" flag to initialize a Realm app from an existing app;
if you do not specify a "--remote" flag, the CLI will initialize a default
Realm app.

NOTE: To create a new Realm app and have it deployed, use "app create".`,
}

// CommandInit is the `app init` command
type CommandInit struct {
	inputs initInputs
}

// Flags is the command flags
func (cmd *CommandInit) Flags() []flags.Flag {
	return []flags.Flag{
		remoteAppFlag(&cmd.inputs.RemoteApp),
		nameFlag(&cmd.inputs.Name),
		locationFlag(&cmd.inputs.Location),
		deploymentModelFlag(&cmd.inputs.DeploymentModel),
		environmentFlag(&cmd.inputs.Environment),
		cli.ProjectFlag(&cmd.inputs.Project),
		cli.ConfigVersionFlag(&cmd.inputs.ConfigVersion, flagConfigVersionDescription),
	}
}

// Inputs is the command inputs
func (cmd *CommandInit) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *CommandInit) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	appRemote, err := cmd.inputs.resolveRemoteApp(ui, clients.Realm)
	if err != nil {
		return err
	}

	if appRemote.GroupID == "" && appRemote.ID == "" {
		if err := cmd.writeAppFromScratch(profile.WorkingDirectory); err != nil {
			return err
		}
	} else {
		if err := cmd.writeAppFromExisting(profile.WorkingDirectory, clients.Realm, appRemote.GroupID, appRemote.ID); err != nil {
			return err
		}
	}

	ui.Print(terminal.NewTextLog("Successfully initialized app"))
	return nil
}

func (cmd *CommandInit) writeAppFromScratch(wd string) error {
	appLocal := local.NewApp(wd,
		"", // no app id yet
		cmd.inputs.Name,
		cmd.inputs.Location,
		cmd.inputs.DeploymentModel,
		cmd.inputs.Environment,
		cmd.inputs.ConfigVersion,
	)
	local.AddAuthProvider(appLocal.AppData, "api-key", map[string]interface{}{
		"name":     "api-key",
		"type":     "api-key",
		"disabled": true,
	})
	return appLocal.Write()
}

func (cmd *CommandInit) writeAppFromExisting(wd string, realmClient realm.Client, groupID, appID string) error {
	_, zipPkg, err := realmClient.Export(groupID, appID, realm.ExportRequest{IsTemplated: true})
	if err != nil {
		return err
	}

	return local.WriteZip(wd, zipPkg)
}
