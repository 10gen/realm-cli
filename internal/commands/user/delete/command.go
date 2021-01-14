package delete

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/commands/user/shared"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
	"github.com/spf13/pflag"
)

// Command is the `user delete` command
var Command = cli.CommandDefinition{
	Use:         "delete",
	Description: "Delete the user(s) from a Realm application",
	Help:        "user delete",
	Command:     &command{},
}

type command struct {
	inputs      inputs
	outputs     outputs
	realmClient realm.Client
}

type outputs struct {
	failed []error
}

func (cmd *command) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)

	fs.StringSliceVarP(&cmd.inputs.Users, flagUsers, flagUsersShort, []string{}, flagUsersUsage)
	fs.VarP(
		flags.NewEnumSet(&cmd.inputs.ProviderTypes, shared.ValidProviderTypes),
		shared.FlagProvider,
		shared.FlagProviderShort,
		shared.FlagProviderUsage,
	)
	fs.Lookup(shared.FlagProvider).NoOptDefVal = shared.ProviderTypeInteractive
	fs.VarP(&cmd.inputs.State, shared.FlagStateType, shared.FlagStateTypeShort, shared.FlagStateTypeUsage)
	fs.Lookup(shared.FlagProvider).NoOptDefVal = shared.ProviderTypeInteractive
	fs.VarP(&cmd.inputs.Status, shared.FlagStatusType, shared.FlagStatusTypeShort, shared.FlagStatusTypeUsage)
	fs.Lookup(shared.FlagStatusType).NoOptDefVal = shared.StatusTypeInteractive.String()
}

func (cmd *command) Inputs() cli.InputResolver {
	return &cmd.inputs
}

func (cmd *command) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = realm.NewAuthClient(profile.RealmBaseURL(), profile.Session())
	return nil
}

func (cmd *command) Handler(profile *cli.Profile, ui terminal.UI) error {
	app, err := cli.ResolveApp(ui, cmd.realmClient, cmd.inputs.Filter())
	if err != nil {
		return err
	}

	var users []string
	if len(cmd.inputs.Users) < 1 {
		users, err = cmd.inputs.ResolveUsers(ui, cmd.realmClient, app)
		if err != nil {
			return err
		}
	} else {
		users = cmd.inputs.Users
	}

	for _, userID := range users {
		err := cmd.realmClient.DeleteUser(app.GroupID, app.ID, userID)
		if err != nil {
			cmd.outputs.failed = append(cmd.outputs.failed, fmt.Errorf("failed to delete user (%s): %s", userID, err))
		}
	}
	return nil
}

func (cmd *command) Feedback(profile *cli.Profile, ui terminal.UI) error {
	if len(cmd.outputs.failed) == 0 {
		return ui.Print(terminal.NewTextLog("Successfully deleted all selected users!"))
	}

	return ui.Print(terminal.NewListLog("Unable to delete the following users", cmd.outputs.failed))
}
