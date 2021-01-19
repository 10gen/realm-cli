package commands

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/commands/app"
	"github.com/10gen/realm-cli/internal/commands/login"
	"github.com/10gen/realm-cli/internal/commands/logout"
	"github.com/10gen/realm-cli/internal/commands/user"
	"github.com/10gen/realm-cli/internal/commands/whoami"
)

// set of commands
var (
	Login = cli.CommandDefinition{
		Use:         "login",
		Description: "Authenticate with an Atlas programmatic API Key",
		Help:        "login", // TODO(REALMC-7429): add help text description
		Command:     &login.Command{},
	}
	Logout = cli.CommandDefinition{
		Use:         "logout",
		Description: "Terminate the current user’s session",
		Help:        "logout", // TODO(REALMC-7429): add help text description
		Command:     &logout.Command{},
	}
	Whoami = cli.CommandDefinition{
		Use:         "whoami",
		Description: "Display the current user's details",
		Help:        "whoami", // TODO(REALMC-7429): add help text description
		Command:     &whoami.Command{},
	}

	App = cli.CommandDefinition{
		Use:         "app",
		Aliases:     []string{"apps"},
		Description: "Manage the apps associated with the current user",
		Help:        "app help", // TODO(REALMC-7429): add help text description
		SubCommands: []cli.CommandDefinition{
			cli.CommandDefinition{
				Use:         "init",
				Aliases:     []string{"initialize"},
				Display:     "app init",
				Description: "Initialize a Realm app in your current local directory",
				Help:        "",
				Command:     &app.CommandInit{},
			},
			cli.CommandDefinition{
				Use:         "list",
				Aliases:     []string{"ls"},
				Display:     "app list",
				Description: "List the MongoDB Realm applications associated with the current user",
				Help:        "list help", // TODO(REALMC-7429): add help text description
				Command:     &app.CommandList{},
			},
		},
	}

	User = cli.CommandDefinition{
		Use:         "user",
		Aliases:     []string{"users"},
		Description: "Manage the users of your MongoDB Realm application",
		Help:        "user",
		SubCommands: []cli.CommandDefinition{
			cli.CommandDefinition{
				Use:         "create",
				Display:     "user create",
				Description: "Create a user for a Realm application",
				Help:        "user create",
				Command:     &user.CommandCreate{},
			},
			cli.CommandDefinition{
				Use:         "delete",
				Display:     "user delete",
				Description: "Delete users from a Realm application",
				Help:        "user delete",
				Command:     &user.CommandDelete{},
			},
			cli.CommandDefinition{
				Use:         "list",
				Description: "List the users of your Realm application",
				Help:        "user list",
				Command:     &user.CommandList{},
			},
		},
	}
)
