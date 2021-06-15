package user

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
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
func (cmd *CommandCreate) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs, "to create its users")

	fs.Var(&cmd.inputs.UserType, flagUserType, flagUserTypeUsage)
	fs.StringVar(&cmd.inputs.APIKeyName, flagAPIKeyName, "", flagAPIKeyNameUsage)
	fs.StringVar(&cmd.inputs.Email, flagEmail, "", flagEmailUsage)
	fs.StringVar(&cmd.inputs.Password, flagPassword, "", flagPasswordUsage)
}

// Inputs is the command inputs
func (cmd *CommandCreate) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *CommandCreate) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	app, err := cli.ResolveApp(ui, clients.Realm, cmd.inputs.Filter())
	if err != nil {
		return err
	}

	switch cmd.inputs.UserType {
	case userTypeAPIKey:
		apiKey, err := clients.Realm.CreateAPIKey(app.GroupID, app.ID, cmd.inputs.APIKeyName)
		if err != nil {
			return fmt.Errorf("failed to create api key: %s", err)
		}

		ui.Print(terminal.NewJSONLog(
			"Successfully created api key",
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
