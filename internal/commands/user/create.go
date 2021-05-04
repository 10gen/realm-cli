package user

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// CommandCreate is the `user create` command
type CommandCreate struct {
	inputs createInputs
}

// Flags is the command flags
func (cmd *CommandCreate) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)

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

		ui.Print(terminal.NewTableLog(
			"Successfully created api key",
			[]string{headerID, headerEnabled, headerName, headerAPIKey},
			map[string]interface{}{
				headerID:      apiKey.ID,
				headerEnabled: !apiKey.Disabled,
				headerName:    apiKey.Name,
				headerAPIKey:  apiKey.Key,
			},
		))
	case userTypeEmailPassword:
		user, err := clients.Realm.CreateUser(app.GroupID, app.ID, cmd.inputs.Email, cmd.inputs.Password)
		if err != nil {
			return fmt.Errorf("failed to create user: %s", err)
		}

		ui.Print(terminal.NewTableLog(
			"Successfully created user",
			[]string{headerID, headerEnabled, headerEmail, headerType},
			map[string]interface{}{
				headerID:      user.ID,
				headerEnabled: !user.Disabled,
				headerEmail:   user.Data["email"],
				headerType:    user.Type,
			},
		))
	}

	return nil
}
