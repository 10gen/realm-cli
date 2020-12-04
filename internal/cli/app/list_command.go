package app

import (
	"errors"
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
		Aliases:     []string{"ls"},
		Command:     &appListCommand{},
		Use:         "list",
		Description: "List the apps of an Atlas project the current user has access to",
		Help:        "list help", // TODO(REALMC-7429): add help text description
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
	session := profile.GetSession()
	cmd.realmClient = realm.NewAuthClient(config.RealmBaseURL, &session)
	return nil
}

func (cmd *appListCommand) Handler(profile *cli.Profile, ui terminal.UI, args []string) error {
	userProfile, err := cmd.realmClient.GetUserProfile()
	if err != nil {
		return fmt.Errorf("Something unexpected happened: %s", err.Error())
	}

	appList, err := cmd.realmClient.FindProjectAppByClientAppID(userProfile.FilterGroupIDsFromUserProfile(cmd.project), cmd.app)
	if err != nil {
		return fmt.Errorf("Something unexpected happened: %s", err.Error())
	}

	if len(appList) == 0 && len(cmd.app) > 0 {
		return errors.New("Found no matches, try changing the --app input")
	}

	cmd.appListResult = appList
	return nil
}

func (cmd *appListCommand) Feedback(profile *cli.Profile, ui terminal.UI) error {
	// REALMC-7156 fix this printing to be formatted
	return ui.Print(terminal.NewTextLog(fmt.Sprintf("results are: %v", cmd.appListResult)))
}
