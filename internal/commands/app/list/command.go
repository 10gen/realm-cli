package list

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/profile"
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

func (cmd *command) Setup(profile *profile.Profile, ui terminal.UI, appData cli.AppData) error {
	cmd.realmClient = realm.NewAuthClient(profile.GetRealmBaseURL(), profile)
	return nil
}

func (cmd *command) Handler(profile *profile.Profile, ui terminal.UI) error {
	apps, appsErr := cmd.realmClient.FindApps(realm.AppFilter{cmd.inputs.Project, cmd.inputs.App})
	if appsErr != nil {
		return fmt.Errorf("failed to get apps: %w", appsErr)
	}

	cmd.apps = apps
	return nil
}

func (cmd *command) Feedback(profile *profile.Profile, ui terminal.UI) error {
	if len(cmd.apps) == 0 {
		return ui.Print(terminal.NewTextLog("No available apps to show"))
	}
	apps := make([]interface{}, 0, len(cmd.apps))
	for _, app := range cmd.apps {
		apps = append(apps, app)
	}
	return ui.Print(terminal.NewListLog(fmt.Sprintf("Found %d apps", len(apps)), apps...))
}
