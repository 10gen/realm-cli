package app

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/flags"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// ListCommand creates the 'app list' subcommand
func ListCommand() cli.CommandDefinition {
	return cli.CommandDefinition{
		Use:         "list",
		Aliases:     []string{"ls"},
		Description: "List the apps of an Atlas project the current user has access to",
		Help:        "list help", // TODO(REALMC-7429): add help text description
		Command:     &appListCommand{},
	}
}

type appListCommand struct {
	app           string
	project       string
	appListResult []realm.App

	realmClient realm.Client
}

func (cmd *appListCommand) RegisterFlags(fs *pflag.FlagSet) {
	fs.StringVar(&cmd.app, flags.App, "", flags.AppUsage)
	fs.StringVar(&cmd.project, flags.Project, "", flags.ProjectUseage)
}

func (cmd *appListCommand) Setup(profile *cli.Profile, ui terminal.UI, config cli.CommandConfig) error {
	cmd.realmClient = realm.NewAuthClient(config.RealmBaseURL, profile.GetSession())
	return nil
}

func (cmd *appListCommand) Handler(profile *cli.Profile, ui terminal.UI, args []string) error {
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

func (cmd *appListCommand) Feedback(profile *cli.Profile, ui terminal.UI) error {
	// REALMC-7156 fix this printing to be formatted
	return ui.Print(terminal.NewTextLog(fmt.Sprintf("results are: %v", cmd.appListResult)))
}
