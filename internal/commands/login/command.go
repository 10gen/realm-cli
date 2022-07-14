package login

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

// CommandMeta is the command meta for the `login` command
var CommandMeta = cli.CommandMeta{
	Use:         "login",
	Description: "Log the CLI into Realm using a MongoDB Cloud API Key",
	HelpText: `Begins an authenticated session with Realm. To get a MongoDB Cloud API Key, open
your Realm app in the Realm UI. Navigate to "Deployment" in the left navigation
menu, and select the "Export App" tab. From there, create a programmatic API Key
to authenticate your realm-cli session.`,
}

// Command is the `login` command
type Command struct {
	inputs inputs
}

// Flags is the command flags
func (cmd *Command) Flags() []flags.Flag {
	return []flags.Flag{
		flags.StringFlag{
			Value:        &cmd.inputs.AuthType,
			DefaultValue: authTypeCloud,
			Meta: flags.Meta{
				Name:   "auth-type",
				Hidden: true,
				Usage: flags.Usage{
					Description: "Specify the public portion of your Atlas programmatic API Key",
				},
			},
		},
		flags.StringFlag{
			Value: &cmd.inputs.PublicAPIKey,
			Meta: flags.Meta{
				Name: "api-key",
				Usage: flags.Usage{
					Description: "Specify the public portion of your Atlas programmatic API Key",
				},
			},
		},
		flags.StringFlag{
			Value: &cmd.inputs.PrivateAPIKey,
			Meta: flags.Meta{
				Name: "private-api-key",
				Usage: flags.Usage{
					Description: "Specify the private portion of your Atlas programmatic API Key",
				},
			},
		},
		flags.StringFlag{
			Value: &cmd.inputs.Username,
			Meta: flags.Meta{
				Name:   "username",
				Hidden: true,
				Usage: flags.Usage{
					Description: "Specify the username of your local Realm credentials",
					Note:        "This is only useful with a locally running Realm server",
				},
			},
		},
		flags.StringFlag{
			Value: &cmd.inputs.Password,
			Meta: flags.Meta{
				Name:   "password",
				Hidden: true,
				Usage: flags.Usage{
					Description: "Specify the password of your local Realm credentials",
					Note:        "This is only useful with a locally running Realm server",
				},
			},
		},
		flags.BoolFlag{
			Value: &cmd.inputs.Browser,
			Meta: flags.Meta{
				Name: "browser",
				Usage: flags.Usage{
					Description: "Direct browser to project access page to create a new API Key for logging into a project",
					Note:        "Can not be used in combination with login credentials",
				},
			},
		},
	}
}

// Inputs is the command inputs
func (cmd *Command) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *Command) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	proceed, err := cmd.checkExistingUser(profile, ui)
	if err != nil {
		return err
	}
	if !proceed {
		return nil
	}

	creds := user.Credentials{
		PublicAPIKey:  cmd.inputs.PublicAPIKey,
		PrivateAPIKey: cmd.inputs.PrivateAPIKey,
		Username:      cmd.inputs.Username,
		Password:      cmd.inputs.Password,
	}
	profile.SetCredentials(creds)

	session, err := clients.Realm.Authenticate(realmAuthType(cmd.inputs.AuthType), creds)
	if err != nil {
		return err
	}

	profile.SetSession(user.Session{session.AccessToken, session.RefreshToken})
	if err := profile.Save(); err != nil {
		return err
	}

	ui.Print(terminal.NewTextLog("Successfully logged in"))
	return nil
}

func (cmd *Command) checkExistingUser(profile *user.Profile, ui terminal.UI) (bool, error) {
	u := profile.Credentials()

	var existingCredentialsName, existingCredentialsSecret string

	if u.PublicAPIKey != "" && u.PublicAPIKey != cmd.inputs.PublicAPIKey {
		existingCredentialsName = u.PublicAPIKey
		existingCredentialsSecret = user.RedactKey(u.PrivateAPIKey)
	} else if u.Username != "" && u.Username != cmd.inputs.Username {
		existingCredentialsName = u.Username
		existingCredentialsSecret = user.Redact(u.Password)
	}

	if existingCredentialsName == "" {
		return true, nil
	}
	return ui.Confirm(
		"This action will terminate the existing session for user: %s (%s), would you like to proceed?",
		existingCredentialsName,
		existingCredentialsSecret,
	)
}
