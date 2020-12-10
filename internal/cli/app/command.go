package app

import (
	"github.com/10gen/realm-cli/internal/cli"
)

// Command creates the 'app' command
func Command() cli.CommandDefinition {
	return cli.CommandDefinition{
		Use:         "app",
		Aliases:     []string{"apps"},
		Description: "Manage the apps associated with the user",
		Help:        "app help", // TODO(REALMC-7429): add help text description
	}
}
