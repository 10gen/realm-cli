package user

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/app"
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// CommandCreate is the `user create` command
type CommandCreate struct {
	inputs      createInputs
	outputs     outputs
	realmClient realm.Client
}

type outputs struct {
	apiKey realm.APIKey
	user   realm.User
}

// Flags is the command flags
func (cmd *CommandCreate) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)

	fs.VarP(&cmd.inputs.UserType, flagUserType, flagUserTypeShort, flagUserTypeUsage)
	fs.StringVarP(&cmd.inputs.Email, flagEmail, flagEmailShort, "", flagEmailUsage)
	fs.StringVarP(&cmd.inputs.Password, flagPassword, flagPasswordShort, "", flagPasswordUsage)
	fs.StringVarP(&cmd.inputs.APIKeyName, flagAPIKeyName, flagAPIKeyNameShort, "", flagAPIKeyNameUsage)
}

// Inputs is the command inputs
func (cmd *CommandCreate) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Setup is the command setup
func (cmd *CommandCreate) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = realm.NewAuthClient(profile)
	return nil
}

// Handler is the command handler
func (cmd *CommandCreate) Handler(profile *cli.Profile, ui terminal.UI) error {
	app, appErr := app.Resolve(ui, cmd.realmClient, cmd.inputs.Filter())
	if appErr != nil {
		return appErr
	}

	switch cmd.inputs.UserType {
	case userTypeAPIKey:
		apiKey, err := cmd.realmClient.CreateAPIKey(app.GroupID, app.ID, cmd.inputs.APIKeyName)
		if err != nil {
			return fmt.Errorf("failed to create api key: %s", err)
		}
		cmd.outputs.apiKey = apiKey
	case userTypeEmailPassword:
		user, err := cmd.realmClient.CreateUser(app.GroupID, app.ID, cmd.inputs.Email, cmd.inputs.Password)
		if err != nil {
			return fmt.Errorf("failed to create user: %s", err)
		}
		cmd.outputs.user = user
	}

	return nil
}

// Feedback is the command feedback
func (cmd *CommandCreate) Feedback(profile *cli.Profile, ui terminal.UI) error {
	switch cmd.inputs.UserType {
	case userTypeAPIKey:
		return ui.Print(terminal.NewTableLog(
			"Successfully created api key",
			[]string{headerID, headerEnabled, headerName, headerAPIKey},
			map[string]interface{}{
				headerID:      cmd.outputs.apiKey.ID,
				headerEnabled: !cmd.outputs.apiKey.Disabled,
				headerName:    cmd.outputs.apiKey.Name,
				headerAPIKey:  cmd.outputs.apiKey.Key,
			},
		))
	case userTypeEmailPassword:
		return ui.Print(terminal.NewTableLog(
			"Successfully created user",
			[]string{headerID, headerEnabled, headerEmail, headerType},
			map[string]interface{}{
				headerID:      cmd.outputs.user.ID,
				headerEnabled: !cmd.outputs.user.Disabled,
				headerEmail:   cmd.outputs.user.Data["email"],
				headerType:    cmd.outputs.user.Type,
			},
		))
	}
	return nil
}
