package commands

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/commands/accesslist"
	"github.com/10gen/realm-cli/internal/commands/app"
	"github.com/10gen/realm-cli/internal/commands/function"
	"github.com/10gen/realm-cli/internal/commands/login"
	"github.com/10gen/realm-cli/internal/commands/logout"
	"github.com/10gen/realm-cli/internal/commands/logs"
	"github.com/10gen/realm-cli/internal/commands/profile"
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

	AccessList = cli.CommandDefinition{
		CommandMeta: cli.CommandMeta{
			Use:         "accessList",
			Aliases:     []string{"accesslist", "access-list"},
			Description: "Manage the allowed IP addresses and CIDR blocks of your Realm app",
		},
		SubCommands: []cli.CommandDefinition{
			{
				Command:     &accesslist.CommandCreate{},
				CommandMeta: accesslist.CommandMetaCreate,
			},
			{
				Command:     &accesslist.CommandList{},
				CommandMeta: accesslist.CommandMetaList,
			},
			{
				Command:     &accesslist.CommandUpdate{},
				CommandMeta: accesslist.CommandMetaUpdate,
			},
			{
				Command:     &accesslist.CommandDelete{},
				CommandMeta: accesslist.CommandMetaDelete,
			},
		},
	}

	Profiles = cli.CommandDefinition{
		CommandMeta: cli.CommandMeta{
			Use:         "profiles",
			Aliases:     []string{"profile"},
			Description: "Manage the profiles of your local CLI environment",
			Hidden:      true,
		},
		SubCommands: []cli.CommandDefinition{
			{
				Command:     &profile.CommandList{},
				CommandMeta: profile.CommandMetaList,
			},
		},
	}
)
