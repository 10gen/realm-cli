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
			{
				Use:         "init",
				Aliases:     []string{"initialize"},
				Display:     "app init",
				Description: "Initialize a Realm app in your current local directory",
				Help:        "",
				Command:     &app.CommandInit{},
			},
			{
				Use:         "create",
				Display:     "app create",
				Description: "Create a remote Realm app from your current working directory",
				Help: `Creates a new Realm app in the cloud, and saves your configuration files on 
a new disk. This command will create a new directory for your project. 
To initialize a Realm app in your current working directory, use ‘app init’.

You can specify a '--from' input to create a Realm app from 
an existing app. If you do not specify a '--from' input, 
the CLI will create a non-personalized Realm app.`,
				Command: &app.CommandCreate{},
			},
			cli.CommandDefinition{
				Use:         "list",
				Aliases:     []string{"ls"},
				Display:     "app list",
				Description: "List the MongoDB Realm applications associated with the current user",
				Help:        "list help", // TODO(REALMC-7429): add help text description
				Command:     &app.CommandList{},
			},
			{
				Use:         "diff",
				Aliases:     []string{},
				Display:     "app diff",
				Description: "Show differences between your Realm app and your local directory",
				Help: `Displays file-by-file differences between the latest version of your Realm app and your local directory. 
If you have more than one Realm app, you will be prompted to select a Realm app that you would like to 
display from a list of all Realm apps associated with your user profile.`,
				Command: &app.CommandDiff{},
			},
			{
				Use:         "delete",
				Display:     "app delete",
				Description: "Delete a Realm app",
				Help: `If you have more than one Realm app, you will be prompted to select one or multiple app(s) that you would like to delete 
from a list of all your Realm apps. The list includes Realm apps from all projects associated with your user profile.`,
				Command: &app.CommandDelete{},
			},
		},
	}

	User = cli.CommandDefinition{
		Use:         "user",
		Aliases:     []string{"users"},
		Description: "Manage the users of your MongoDB Realm application",
		Help:        "user",
		SubCommands: []cli.CommandDefinition{
			{
				Use:         "create",
				Display:     "user create",
				Description: "Create a user for a Realm application",
				Help:        "user create",
				Command:     &user.CommandCreate{},
			},
			{
				Use:         "delete",
				Display:     "user delete",
				Description: "Delete an application user from your Realm app",
				Help:        "Removes a specific user from your Realm app.",
				Command:     &user.CommandDelete{},
			},
			{
				Use:         "disable",
				Display:     "user disable",
				Description: "Disable an application user of your Realm app",
				Help:        "Deactivates a user on your Realm app. A user that has been disabled will not be allowed to log in, even if they provide valid credentials.",
				Command:     &user.CommandDisable{},
			},
			{
				Use:         "enable",
				Display:     "user enable",
				Description: "Enable an application user of your Realm app",
				Help:        "Activates a user on your Realm app. A user that has been disabled will not be allowed to log in, even if they provide valid credentials.",
				Command:     &user.CommandEnable{},
			},
			{
				Use:         "list",
				Description: "List the users of your Realm application",
				Help:        "user list",
				Command:     &user.CommandList{},
			},
			{
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
			{
				Use:         "create",
				Display:     "secrets create",
				Description: "Create a secret for your Realm app",
				Help:        "Adds a new secret to your Realm app. You will be prompted to name your Secret, and define the value of your Secret.",
				Command:     &secrets.CommandCreate{},
			},
			{
				Use:         "list",
				Aliases:     []string{"ls"},
				Display:     "secrets list",
				Description: "List the names of secrets in your Realm app",
				Help:        "Displays a list of data tables. Each data table displays the Name and ID of one (1) secret.",
				Command:     &secrets.CommandList{},
			},
			{
				Use:         "delete",
				Display:     "secrets delete",
				Description: "Delete a secret from your Realm app",
				Help:        "Removes one or many secrets from your Realm app depending on how many you specify.",
				Command:     &secrets.CommandDelete{},
			},
			cli.CommandDefinition{
				Use:         "update",
				Display:     "secrets update",
				Description: "Update a secret in your Realm app",
				Help:        "Updates one secret from your Realm app. Select the secret to update or provide a name or ID.",
				Command:     &secrets.CommandUpdate{},
			},
		},
	}
)
