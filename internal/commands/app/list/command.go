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
	apps        []realm.App
	inputs      cli.ProjectAppInputs
	realmClient realm.Client
}

func (cmd *command) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)
}

func (cmd *command) Setup(profile *cli.Profile, ui terminal.UI, appData cli.AppData) error {
	cmd.realmClient = realm.NewAuthClient(profile.RealmBaseURL(), profile.Session())
	return nil
}

func (cmd *command) Handler(profile *cli.Profile, ui terminal.UI) error {
	apps, appsErr := cmd.realmClient.FindApps(realm.AppFilter{cmd.inputs.Project, cmd.inputs.App})
	if appsErr != nil {
		return fmt.Errorf("failed to get apps: %w", appsErr)
	}

	cmd.apps = apps
	return nil
}

// TODO(REALMC-7574): print list of apps
func (cmd *command) Feedback(profile *cli.Profile, ui terminal.UI) error {
	return ui.Print(terminal.NewTextLog(fmt.Sprintf("results are: %v", cmd.apps)))
}
