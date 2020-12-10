package list

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/terminal"
)

// Command is the `user list` command
var Command = cli.CommandDefinition{
	Use:         "list",
	Description: "List the users of your Realm application",
	Help:        "user list",
	Command:     &command{},
}

type command struct{}

func (cmd *command) Handler(profile *cli.Profile, ui terminal.UI, args []string) error {
	return nil
}
