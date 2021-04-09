package commands

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/commands/app"
	"github.com/10gen/realm-cli/internal/commands/function"
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
		Use:         "login",
		Description: "Log in to your Realm App backend with a MongoDB Cloud API key",
		Help: `Begins an authenticated session with Realm. To get a MongoDB Cloud
API Key, open your Realm app in the Realm UI. Navigate to Deploy in the left
navigation menu, and select the Export App tab. From there, create a
programmatic API key to authenticate your realm-cli session`,
	}

	Logout = cli.CommandDefinition{
		Command:     &logout.Command{},
		Use:         "logout",
		Description: "Log out of your Realm app backend and MongoDB Cloud",
		Help: `Ends the authenticated session and deletes cached access tokens. To
re-authenticate, you must call 'login' with your API Key.`,
	}

	Whoami = cli.CommandDefinition{
		Command:     &whoami.Command{},
		Use:         "whoami",
		Description: "Display information about the current user.",
		Help: `
Displays a table that includes your Public API Key and redacted Private API Key
(e.g. ********-****-****-****-3ba985aa367a). No session data will be surfaced if
you are not logged in.

To log in, authenticate your information using the following command:

realm-cli login --profile <profile-name>`,
	}

	Push = cli.CommandDefinition{
		Command:     &push.Command{},
		Use:         push.CommandUse,
		Aliases:     push.CommandAliases,
		Description: "Import and deploy changes from your local directory to your Realm app",
		Help: `Updates a remote Realm app with your local directory. First, input a Realm app
that you would like changes pushed to. This input can be either the App ID or
Name of an existing Realm app you would like to update, or the name of a new
Realm app you would like to create. Changes pushed are automatically deployed.`,
	}

	Pull = cli.CommandDefinition{
		Command:     &pull.Command{},
		Use:         "export",
		Aliases:     []string{"pull"},
		Description: "Export the latest version of your Realm app into your local directory",
		Help: `Updates your local directory with a remote Realm app by pulling changes from the
latter into the former. Input a Realm app that you would like to have changes
pulled from. If applicable, hosting files and/or dependencies associated with
your Realm app will be exported as well.`,
	}

	App = cli.CommandDefinition{
		Use:         "apps",
		Aliases:     []string{"app"},
		Description: "Manage the Realm apps associated with the current user",
		SubCommands: []cli.CommandDefinition{
			{
				Command:     &app.CommandInit{},
				Use:         "init",
				Aliases:     []string{"initialize"},
				Display:     "app init",
				Description: "Initialize a Realm app in your current working directory",
				Help: `Initializes configuration files and directories to represent a minimal Realm app
within your current working directory. This command only affects your local
environment and does not deploy your app to the Realm servers. To create a new
Realm app and have it deployed, use ‘app create’.

You can specify a ‘--remote’ flag to initialize a Realm app from an existing
app; if you do not specify a '--remote' flag, the CLI will initialize a default
Realm app.`,
			},
			{
				Command:     &app.CommandCreate{},
				Use:         "create",
				Display:     app.CommandCreateDisplay,
				Description: "Create a Realm app in a local directory and deploy it to the Realm server",
				Help: `Creates a new Realm app in the cloud and saves your configuration files in a
local directory.  This command will also deploy the new app to the Realm servers.
This command will create a new directory for your project. To create a Realm app
and not have it deployed, use ‘app init’.

You can specify a '--remote' flag to create a Realm app from an existing app; if
you do not specify a '--remote' flag, the CLI will create a default Realm app.`,
			},
			{
				Command:     &app.CommandList{},
				Use:         "list",
				Aliases:     []string{"ls"},
				Display:     "apps list",
				Description: "List the Realm apps you have access to",
				Help:        `Lists and filters your Realm apps`,
			},
			{
				Command:     &app.CommandDiff{},
				Use:         "diff",
				Aliases:     []string{},
				Display:     "app diff",
				Description: "Show differences between your Realm app and your local directory",
				Help: `Displays file-by-file differences between the latest version of your Realm app
and your local directory. If you have more than one Realm app, you will be
prompted to select a Realm app that you would like to display from a list of all
Realm apps associated with your user profile.`,
			},
			{
				Command:     &app.CommandDelete{},
				Use:         "delete",
				Display:     "app delete",
				Description: "Delete a Realm app",
				Help: `If you have more than one Realm app, you will be prompted to select one or
multiple app(s) that you would like to delete from a list of all your Realm apps.
The list includes Realm apps from all projects associated with your user profile.`,
			},
			{
				Command:     &app.CommandDescribe{},
				Use:         "describe",
				Display:     "app describe",
				Description: "View all of the configured and enabled aspects of your Realm app",
				Help: `Displays information about your Realm app.
If you have more than one Realm app, you will be prompted to select a Realm app that
you would like to view from a list of all Realm apps associated with your user profile.`,
			},
		},
	}

	User = cli.CommandDefinition{
		Use:         "users",
		Aliases:     []string{"user"},
		Description: "Manage the users of your Realm app",
		SubCommands: []cli.CommandDefinition{
			{
				Command:     &user.CommandCreate{},
				Use:         "create",
				Display:     "user create",
				Description: "Create an application user in your Realm app",
				Help: `Adds a new user to your Realm app. You can create a user
				using an Email/Password or an API Key.`,
			},
			{
				Command:     &user.CommandList{},
				Use:         "list",
				Description: "List the application users in your Realm app",
				Help: `Displays a list of your Realm app's users' details. The list is grouped by
auth provider type and sorted by last authentication date.`,
			},
			{
				Command:     &user.CommandDisable{},
				Use:         "disable",
				Display:     "user disable",
				Description: "Disable an application user in your Realm app",
				Help: `Deactivates a user on your Realm app. A user that has been disabled will not be
allowed to log in, even if they provide valid credentials.`,
			},
			{
				Command:     &user.CommandEnable{},
				Use:         "enable",
				Display:     "user enable",
				Description: "Enable an application user in your Realm app",
				Help: `Activates a user on your Realm app. A user that has been enabled will have no
restrictions with logging in.`,
			},
			{
				Command:     &user.CommandRevoke{},
				Use:         "revoke",
				Display:     "user revoke",
				Description: "Revoke an application user’s sessions from your Realm app",
				Help: `Logs a user out of your Realm app. A user who’s user session has been revoked
can log in again if they provide valid credentials.`,
			},
			{
				Command:     &user.CommandDelete{},
				Use:         "delete",
				Display:     "user delete",
				Description: "Delete an application user from your Realm app",
				// TODO(REALMC-7662): Document downstream events after deleting a user
				Help: `Removes a specific user from your Realm app.`,
			},
		},
	}

	Secrets = cli.CommandDefinition{
		Use:         "secrets",
		Aliases:     []string{"secret"},
		Description: "Manage the secrets of your Realm app",
		SubCommands: []cli.CommandDefinition{
			{
				Command:     &secrets.CommandCreate{},
				Use:         "create",
				Display:     "secrets create",
				Description: "Create a secret in your Realm app",
				Help: `Adds a new secret to your Realm app. You will be prompted to name your secret
and define the value of your secret.`,
			},
			{
				Command:     &secrets.CommandList{},
				Use:         "list",
				Aliases:     []string{"ls"},
				Display:     "secrets list",
				Description: "List the secrets in your Realm app",
				Help:        `Displays a list of your Realm app's secrets.`,
			},
			{
				Command:     &secrets.CommandUpdate{},
				Use:         "update",
				Display:     "secret update",
				Description: "Update a secret in your Realm app",
				Help: `Modifies the value of a secret in your Realm app. The name of the secret cannot
be modified.`,
			},
			{
				Command:     &secrets.CommandDelete{},
				Use:         "delete",
				Display:     "secrets delete",
				Description: "Delete a secret from your Realm app",
				Help:        `Removes a secret from your Realm app.`,
			},
		},
	}

	Function = cli.CommandDefinition{
		Use:         "function",
		Aliases:     []string{"functions"},
		Description: "Manage the functions of your Realm app",
		SubCommands: []cli.CommandDefinition{
			{
				Command:     &function.CommandRun{},
				Use:         "run",
				Description: "Run a function from your Realm app",
				Help: `Realm Functions allow you to define and execute server-side logic for your
Realm app. Once you select and run a function for your Realm app, the
following will be displayed:
 - A list of logs, if present
 - The function result as a document
 - A list of error logs, if present
`,
			},
		},
	}

	Logs = cli.CommandDefinition{
		Use:         "logs",
		Aliases:     []string{"log"},
		Description: "Interact with the logs of your Realm app",
		SubCommands: []cli.CommandDefinition{
			{
				Command:     &logs.CommandList{},
				Use:         "list",
				Aliases:     []string{"ls"},
				Display:     "logs list",
				Description: "Lists the logs in your Realm app",
				Help: `Displays a list of your Realm app’s logs sorted by recentness, with most recent
logs appearing towards the bottom.  You can specify a --tail flag to monitor
your logs and follow any newly created logs in real-time.`,
			},
		},
	}

	Schema = cli.CommandDefinition{
		Use:         "schema",
		Aliases:     []string{"schemas"},
		Description: "Manage the schemas of your Realm app",
		SubCommands: []cli.CommandDefinition{
			{
				Command:     &schema.CommandDatamodels{},
				Use:         "datamodels",
				Aliases:     []string{"datamodel"},
				Display:     "schema datamodels",
				Description: "Generate data models based on your schema",
				Help: `Translates your schema’s objects into Realm data models. The data models define
your data as native objects, which can be easily integrated into your
application to use with Realm Sync. Note that you must have a valid JSON schema
before using this command.

With this command, you can:
  - Specify the language with a --language flag
  - Filter which schema objects you’d like to include in your output with --name flags
  - Combine your schema objects into a single output with a --flat flag
  - Omit import groups from your model with a --no-imports flag`,
			},
		},
	}
)
