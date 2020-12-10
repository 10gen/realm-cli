package user

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/commands/user/create"
	"github.com/10gen/realm-cli/internal/commands/user/list"
)

// Command defines the `user` command
var Command = cli.CommandDefinition{
	Use:         "user",
	Aliases:     []string{"users"},
	Description: "Manage the users of your MongoDB Realm application",
	Help:        "user",
	SubCommands: []cli.CommandDefinition{list.Command, create.Command},
}
