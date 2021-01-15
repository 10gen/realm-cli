package app

import (
	"fmt"

	appcli "github.com/10gen/realm-cli/internal/app"
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// CommandList is the `app list` command
type CommandList struct {
	apps        []realm.App
	inputs      appcli.ProjectInputs
	realmClient realm.Client
}

// Flags is the command flags
func (cmd *CommandList) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)
}

// Setup is the command setup
func (cmd *CommandList) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = realm.NewAuthClient(profile)
	return nil
}

// Handler is the command handler
func (cmd *CommandList) Handler(profile *cli.Profile, ui terminal.UI) error {
	apps, appsErr := cmd.realmClient.FindApps(realm.AppFilter{cmd.inputs.Project, cmd.inputs.App})
	if appsErr != nil {
		return fmt.Errorf("failed to get apps: %w", appsErr)
	}

	cmd.apps = apps
	return nil
}

// Feedback is the command feedback
func (cmd *CommandList) Feedback(profile *cli.Profile, ui terminal.UI) error {
	if len(cmd.apps) == 0 {
		return ui.Print(terminal.NewTextLog("No available apps to show"))
	}
	apps := make([]interface{}, 0, len(cmd.apps))
	for _, app := range cmd.apps {
		apps = append(apps, app)
	}
	return ui.Print(terminal.NewListLog(fmt.Sprintf("Found %d apps", len(apps)), apps...))
}
