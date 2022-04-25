package user

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

// CommandMetaCreate is the command meta for the `user create` command
var CommandMetaCreate = cli.CommandMeta{
	Use:         "create",
	Display:     "user create",
	Description: "Create an application user for your Realm app",
	HelpText: `Adds a new User to your Realm app. You can create a User for the following
enabled Auth Providers: "Email/Password", or "API Key".`,
}

// CommandCreate is the `user create` command
type CommandCreate struct {
	inputs createInputs
}

// Flags is the command flags
func (cmd *CommandCreate) Flags() []flags.Flag {
	return []flags.Flag{
		cli.AppFlagWithContext(&cmd.inputs.App, "to create its users"),
		cli.ProjectFlag(&cmd.inputs.Project),
		cli.ProductFlag(&cmd.inputs.Products),
		flags.CustomFlag{
			Value: &cmd.inputs.UserType,
			Meta: flags.Meta{
				Name: "type",
				Usage: flags.Usage{
					Description:  "Select the type of user to create",
					DefaultValue: "<none>",
					AllowedValues: []string{
						string(userTypeAPIKey),
						string(userTypeEmailPassword),
					},
				},
			},
		},
		flags.StringFlag{
			Value: &cmd.inputs.APIKeyName,
			Meta: flags.Meta{
				Name: "name",
				Usage: flags.Usage{
					Description: "Specify the name of the new API Key",
				},
			},
		},
		flags.StringFlag{
			Value: &cmd.inputs.Email,
			Meta: flags.Meta{
				Name: "email",
				Usage: flags.Usage{
					Description: "Specify the email of the new user",
				},
			},
		},
		flags.StringFlag{
			Value: &cmd.inputs.Password,
			Meta: flags.Meta{
				Name: "password",
				Usage: flags.Usage{
					Description: "Specify the password of the new user",
				},
			},
		},
	}
}

// Inputs is the command inputs
func (cmd *CommandCreate) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *CommandCreate) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	app, err := cli.ResolveApp(ui, clients.Realm, cli.AppOptions{
		AppMeta: cmd.inputs.AppMeta,
		Filter:  cmd.inputs.Filter(),
	})
	if err != nil {
		return err
	}

	switch cmd.inputs.UserType {
	case userTypeAPIKey:
		apiKey, err := clients.Realm.CreateAPIKey(app.GroupID, app.ID, cmd.inputs.APIKeyName)
		if err != nil {
			return fmt.Errorf("failed to create API Key: %s", err)
		}

		ui.Print(terminal.NewJSONLog(
			"Successfully created API Key",
			newUserAPIKeyOutputs{
				newUserOutputs: newUserOutputs{
					ID:      apiKey.ID,
					Enabled: !apiKey.Disabled,
				},
				Name: apiKey.Name,
				Key:  apiKey.Key,
			},
		))
	case userTypeEmailPassword:
		user, err := clients.Realm.CreateUser(app.GroupID, app.ID, cmd.inputs.Email, cmd.inputs.Password)
		if err != nil {
			return fmt.Errorf("failed to create user: %s", err)
		}

		ui.Print(terminal.NewJSONLog(
			"Successfully created user",
			newUserEmailOutputs{
				newUserOutputs: newUserOutputs{
					ID:      user.ID,
					Enabled: !user.Disabled,
				},
				Email: user.Data["email"],
				Type:  user.Type,
			},
		))
	}

	return nil
}
