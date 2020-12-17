package create

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// Command defines the `user create` command
var Command = cli.CommandDefinition{
	Use:         "create",
	Display:     "user create",
	Description: "Create a user for a Realm application",
	Help:        "user create",
	Command:     &command{},
}

type command struct {
	inputs      inputs
	outputs     outputs
	realmClient realm.Client
}

type outputs struct {
	apiKey realm.APIKey
	user   realm.User
}

func (cmd *command) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)

	fs.VarP(&cmd.inputs.UserType, flagUserType, flagUserTypeShort, flagUserTypeUsage)
	fs.StringVarP(&cmd.inputs.Email, flagEmail, flagEmailShort, "", flagEmailUsage)
	fs.StringVarP(&cmd.inputs.Password, flagPassword, flagPasswordShort, "", flagPasswordUsage)
	fs.StringVarP(&cmd.inputs.APIKeyName, flagAPIKeyName, flagAPIKeyNameShort, "", flagAPIKeyNameUsage)
}

func (cmd *command) Inputs() cli.InputResolver {
	return &cmd.inputs
}

func (cmd *command) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = realm.NewAuthClient(profile.RealmBaseURL(), profile.Session())
	return nil
}

func (cmd *command) Handler(profile *cli.Profile, ui terminal.UI) error {
	app, appErr := cli.ResolveApp(ui, cmd.realmClient, cmd.inputs.Filter())
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

const (
	headerID           = "ID"
	headerEnabled      = "Enabled"
	headerAPIKeyName   = "Name"
	headerAPIKeyAPIKey = "API Key"
	headerUserEmail    = "Email"
	headerUserType     = "Type"
)

func (cmd *command) Feedback(profile *cli.Profile, ui terminal.UI) error {
	switch cmd.inputs.UserType {
	case userTypeAPIKey:
		return ui.Print(terminal.NewTableLog(
			"Successfully created api key",
			[]string{headerID, headerEnabled, headerAPIKeyName, headerAPIKeyAPIKey},
			map[string]interface{}{
				headerID:           cmd.outputs.apiKey.ID,
				headerEnabled:      !cmd.outputs.apiKey.Disabled,
				headerAPIKeyName:   cmd.outputs.apiKey.Name,
				headerAPIKeyAPIKey: cmd.outputs.apiKey.Key,
			},
		))
	case userTypeEmailPassword:
		return ui.Print(terminal.NewTableLog(
			"Successfully created user",
			[]string{headerID, headerEnabled, headerUserEmail, headerUserType},
			map[string]interface{}{
				headerID:        cmd.outputs.user.ID,
				headerEnabled:   !cmd.outputs.user.Disabled,
				headerUserEmail: cmd.outputs.user.Data["email"],
				headerUserType:  cmd.outputs.user.Type,
			},
		))
	}
	return nil
}
