package create

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/terminal"
)

// Command defines the `user create` command
var Command = cli.CommandDefinition{
	Use:         "create",
	Description: "Create a user for a Realm application",
	Help:        "user create",
	Command:     &command{},
}

type command struct{}

func (cmd *command) Handler(profile *cli.Profile, ui terminal.UI, args []string) error {
	return nil
}
