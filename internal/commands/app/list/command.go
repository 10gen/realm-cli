package list

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// Command is the `app list` command
var Command = cli.CommandDefinition{
	Use:         "list",
	Aliases:     []string{"ls"},
	Display:     "app list",
	Description: "List the MongoDB Realm applications associated with the current user",
	Help:        "list help", // TODO(REALMC-7429): add help text description
	Command:     &command{},
}

type command struct {
	app           string
	project       string
	appListResult []realm.App

	realmClient realm.Client
}

func (cmd *command) RegisterFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&cmd.project, flagProject, flagProjectShort, "", flagProjectUsage)
	fs.StringVarP(&cmd.app, flagApp, flagAppShort, "", flagAppUsage)
}

func (cmd *command) Setup(profile *cli.Profile, ui terminal.UI, ctx cli.Context) error {
	cmd.realmClient = realm.NewAuthClient(ctx.RealmBaseURL, profile.GetSession())
	return nil
}

func (cmd *command) Handler(profile *cli.Profile, ui terminal.UI, args []string) error {
	var appList []realm.App
	var err error
	if cmd.project != "" {
		appList, err = cmd.realmClient.GetApps(cmd.project)
	} else {
		appList, err = cmd.realmClient.GetAppsForUser()
	}
	if err != nil {
		return fmt.Errorf("failed to get apps: %s", err)
	}

	cmd.appListResult = appList
	return nil
}

func (cmd *command) Feedback(profile *cli.Profile, ui terminal.UI) error {
	// REALMC-7156 fix this printing to be formatted
	return ui.Print(terminal.NewTextLog(fmt.Sprintf("results are: %v", cmd.appListResult)))
}
