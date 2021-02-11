package commands

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/commands/app"
	"github.com/10gen/realm-cli/internal/commands/login"
	"github.com/10gen/realm-cli/internal/commands/logout"
	"github.com/10gen/realm-cli/internal/commands/pull"
	"github.com/10gen/realm-cli/internal/commands/push"
	"github.com/10gen/realm-cli/internal/commands/secrets"
	"github.com/10gen/realm-cli/internal/commands/user"
	"github.com/10gen/realm-cli/internal/commands/whoami"
)

// set of commands
var (
	Login = cli.CommandDefinition{
		Command:     &login.Command{},
		Use:         "login",
		Description: "Authenticate with an Atlas programmatic API Key",
		Help:        "login", // TODO(REALMC-7429): add help text description
	}
	Logout = cli.CommandDefinition{
		Command:     &logout.Command{},
		Use:         "logout",
		Description: "Terminate the current user’s session",
		Help:        "logout", // TODO(REALMC-7429): add help text description
	}
	Whoami = cli.CommandDefinition{
		Command:     &whoami.Command{},
		Use:         "whoami",
		Description: "Display the current user's details",
		Help:        "whoami", // TODO(REALMC-7429): add help text description
	}

	Push = cli.CommandDefinition{
		Command:     &push.Command{},
		Use:         "push",
		Aliases:     []string{"import"},
		Description: "Push and deploy changes from your local directory to your Realm app",
		Help: `Push and deploy changes from your local directory to your Realm app

	Updates a remote Realm application with your local directory. First, input a
	Realm app that you would like changes pushed to. This input can be either the
	application Client App ID of an existing Realm app you would like to update, or
	the name of a new Realm app you would like to create. Changes pushed are
	automatically deployed.`,
	}

	Pull = cli.CommandDefinition{
		Command:     &pull.Command{},
		Use:         "pull",
		Aliases:     []string{"export"},
		Description: "Pull the latest version of your Realm app into your local directory",
		Help: `Pull the latest version of your Realm app into your local directory

Updates a remote Realm app with your local directory by pulling changes from the
former into the latter. Input a Realm app that you would like to have changes
pushed from. If applicable, hosting and/or dependencies associated with your
Realm app will be exported as well.`,
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
				Description: "Delete an application user from your Realm app",
				Help:        "Removes a specific user from your Realm app.",
				Command:     &user.CommandDelete{},
			},
			cli.CommandDefinition{
				Use:         "disable",
				Display:     "user disable",
				Description: "Disable an application user of your Realm app",
				Help:        "Deactivates a user on your Realm app. A user that has been disabled will not be allowed to log in, even if they provide valid credentials.",
				Command:     &user.CommandDisable{},
			},
			cli.CommandDefinition{
				Use:         "enable",
				Display:     "user enable",
				Description: "Enable an application user of your Realm app",
				Help:        "Activates a user on your Realm app. A user that has been disabled will not be allowed to log in, even if they provide valid credentials.",
				Command:     &user.CommandEnable{},
			},
			cli.CommandDefinition{
				Use:         "list",
				Description: "List the users of your Realm application",
				Help:        "user list",
				Command:     &user.CommandList{},
			},
			cli.CommandDefinition{
				Use:         "revoke",
				Display:     "user revoke",
				Description: "Revoke an application user’s sessions on your Realm app",
				Help:        "Logs a user out of your Realm app. A user who’s user session has been revoked can log in again if they provide valid credentials.",
				Command:     &user.CommandRevoke{},
			},
		},
	}

	Secrets = cli.CommandDefinition{
		Use:         "secrets",
		Aliases:     []string{"secret"},
		Description: "Manage the secrets of your MongoDB Realm app",
		Help:        "secrets",
		SubCommands: []cli.CommandDefinition{
			cli.CommandDefinition{
				Use:         "create",
				Display:     "secrets create",
				Description: "Create a secret for your Realm app",
				Help:        "Adds a new secret to your Realm app. You will be prompted to name your Secret, and define the value of your Secret.",
				Command:     &secrets.CommandCreate{},
			},
			cli.CommandDefinition{
				Use:         "list",
				Aliases:     []string{"ls"},
				Display:     "secrets list",
				Description: "List the names of secrets in your Realm app",
				Help:        "Displays a list of data tables. Each data table displays the Name and ID of one (1) secret.",
				Command:     &secrets.CommandList{},
			},
		},
	}
)
