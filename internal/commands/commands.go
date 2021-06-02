package commands

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/commands/app"
	"github.com/10gen/realm-cli/internal/commands/function"
	"github.com/10gen/realm-cli/internal/commands/ip_access"
	"github.com/10gen/realm-cli/internal/commands/login"
	"github.com/10gen/realm-cli/internal/commands/logout"
	"github.com/10gen/realm-cli/internal/commands/logs"
	"github.com/10gen/realm-cli/internal/commands/pull"
	"github.com/10gen/realm-cli/internal/commands/push"
	"github.com/10gen/realm-cli/internal/commands/schema"
	"github.com/10gen/realm-cli/internal/commands/secrets"
	"github.com/10gen/realm-cli/internal/commands/user"
	"github.com/10gen/realm-cli/internal/commands/whoami"
)

// set of commands
var (
	Login = cli.CommandDefinition{
		Command:     &login.Command{},
		CommandMeta: login.CommandMeta,
	}

	Logout = cli.CommandDefinition{
		Command:     &logout.Command{},
		CommandMeta: logout.CommandMeta,
	}

	Whoami = cli.CommandDefinition{
		Command:     &whoami.Command{},
		CommandMeta: whoami.CommandMeta,
	}

	Push = cli.CommandDefinition{
		Command:     &push.Command{},
		CommandMeta: push.CommandMeta,
	}

	Pull = cli.CommandDefinition{
		Command:     &pull.Command{},
		CommandMeta: pull.CommandMeta,
	}

	App = cli.CommandDefinition{
		CommandMeta: cli.CommandMeta{
			Use:         "apps",
			Aliases:     []string{"app"},
			Description: "Manage the Realm apps associated with the current user",
		},
		SubCommands: []cli.CommandDefinition{
			{
				Command:     &app.CommandInit{},
				CommandMeta: app.CommandMetaInit,
			},
			{
				Command:     &app.CommandCreate{},
				CommandMeta: app.CommandMetaCreate,
			},
			{
				Command:     &app.CommandList{},
				CommandMeta: app.CommandMetaList,
			},
			{
				Command:     &app.CommandDelete{},
				CommandMeta: app.CommandMetaDelete,
			},
			{
				Command:     &app.CommandDiff{},
				CommandMeta: app.CommandMetaDiff,
			},
			{
				Command:     &app.CommandDescribe{},
				CommandMeta: app.CommandMetaDescribe,
			},
		},
	}

	User = cli.CommandDefinition{
		CommandMeta: cli.CommandMeta{
			Use:         "users",
			Aliases:     []string{"user"},
			Description: "Manage the Users of your Realm app",
		},
		SubCommands: []cli.CommandDefinition{
			{
				Command:     &user.CommandCreate{},
				CommandMeta: user.CommandMetaCreate,
			},
			{
				Command:     &user.CommandList{},
				CommandMeta: user.CommandMetaList,
			},
			{
				Command:     &user.CommandDisable{},
				CommandMeta: user.CommandMetaDisable,
			},
			{
				Command:     &user.CommandEnable{},
				CommandMeta: user.CommandMetaEnable,
			},
			{
				Command:     &user.CommandRevoke{},
				CommandMeta: user.CommandMetaRevoke,
			},
			{
				Command:     &user.CommandDelete{},
				CommandMeta: user.CommandMetaDelete,
			},
		},
	}

	Secrets = cli.CommandDefinition{
		CommandMeta: cli.CommandMeta{
			Use:         "secrets",
			Aliases:     []string{"secret"},
			Description: "Manage the Secrets of your Realm app",
		},
		SubCommands: []cli.CommandDefinition{
			{
				Command:     &secrets.CommandCreate{},
				CommandMeta: secrets.CommandMetaCreate,
			},
			{
				Command:     &secrets.CommandList{},
				CommandMeta: secrets.CommandMetaList,
			},
			{
				Command:     &secrets.CommandUpdate{},
				CommandMeta: secrets.CommandMetaUpdate,
			},
			{
				Command:     &secrets.CommandDelete{},
				CommandMeta: secrets.CommandMetaDelete,
			},
		},
	}

	Function = cli.CommandDefinition{
		CommandMeta: cli.CommandMeta{
			Use:         "function",
			Aliases:     []string{"functions"},
			Description: "Interact with the Functions of your Realm app",
		},
		SubCommands: []cli.CommandDefinition{
			{
				Command:     &function.CommandRun{},
				CommandMeta: function.CommandMetaRun,
			},
		},
	}

	Logs = cli.CommandDefinition{
		CommandMeta: cli.CommandMeta{
			Use:         "logs",
			Aliases:     []string{"log"},
			Description: "Interact with the Logs of your Realm app",
		},
		SubCommands: []cli.CommandDefinition{
			{
				Command:     &logs.CommandList{},
				CommandMeta: logs.CommandMetaList,
			},
		},
	}

	Schema = cli.CommandDefinition{
		CommandMeta: cli.CommandMeta{
			Use:         "schema",
			Aliases:     []string{"schemas"},
			Description: "Manage the Schemas of your Realm app",
		},
		SubCommands: []cli.CommandDefinition{
			{
				Command:     &schema.CommandDatamodels{},
				CommandMeta: schema.CommandMetaDatamodels,
			},
		},
	}

	IPAccess = cli.CommandDefinition{
		Use:         "IP access",
		Description: "Manage IP access",
		SubCommands: []cli.CommandDefinition{
			{
				Command:     &ip_access.CommandCreate{},
				Use:         "create",
				Aliases:     []string{"add"},
				Display:     "allowed IP create",
				Description: "Create an allowed IP in your Realm app",
				Help:        "Adds a new IP access to your Realm app. You will need to provde an IP Address and an optional comment.",
			},
			{
				Command:     &ip_access.CommandList{},
				Use:         "list",
				Aliases:     []string{"ls"},
				Display:     "allowed IPs list",
				Description: "List the allowed IPs in your Realm app",
				Help:        "Displays a list of the allowed IPs in your Realm app.",
			},
			{
				Command:     &ip_access.CommandUpdate{},
				Use:         "update",
				Display:     "allowed IP update",
				Description: "Update an allowed IP in your Realm app",
				Help:        "Modifies the value of an allowed IP in your Realm app.",
			},
			{
				Command:     &ip_access.CommandDelete{},
				Use:         "delete",
				Display:     "allowed IP delete",
				Description: "Delete an allowed IP from your Realm app",
				Help:        "Removes an allowed IP from your Realm app.",
			},
		},
	}
)
