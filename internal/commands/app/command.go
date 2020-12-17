package app

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/commands/app/initialize"
	"github.com/10gen/realm-cli/internal/commands/app/list"
)

// Command defines the `app` command
var Command = cli.CommandDefinition{
	Use:         "app",
	Aliases:     []string{"apps"},
	Description: "Manage the apps associated with the current user",
	Help:        "app help", // TODO(REALMC-7429): add help text description
	SubCommands: []cli.CommandDefinition{initialize.Command, list.Command},
}
