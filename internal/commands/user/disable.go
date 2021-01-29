package user

import (
	"fmt"
	"sort"

	"github.com/10gen/realm-cli/internal/app"
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// CommandDisable is the `user disable` command
type CommandDisable struct {
	inputs      disableInputs
	outputs     []userOutput
	realmClient realm.Client
}

type disableInputs struct {
	app.ProjectInputs
	Users []string
}

func (i *disableInputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory); err != nil {
		return err
	}

	return nil
}

func (i *disableInputs) resolveUsers(ui terminal.UI, client realm.Client, app realm.App) ([]realm.User, error) {
	filter := realm.UserFilter{
		IDs: i.Users,
	}
	foundUsers, usersErr := client.FindUsers(app.GroupID, app.ID, filter)
	if usersErr != nil {
		return nil, usersErr
	}
	return foundUsers, nil
}

// Flags is the command flags
func (cmd *CommandDisable) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)
	fs.StringSliceVarP(&cmd.inputs.Users, flagUser, flagUserShort, []string{}, flagUserDisableUsage)
}

// Inputs is the command inputs
func (cmd *CommandDisable) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Setup is the command setup
func (cmd *CommandDisable) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = profile.RealmAuthClient()
	return nil
}

// Handler is the command handler
func (cmd *CommandDisable) Handler(profile *cli.Profile, ui terminal.UI) error {
	app, err := app.Resolve(ui, cmd.realmClient, cmd.inputs.Filter())
	if err != nil {
		return err
	}

	users, usersErr := cmd.inputs.resolveUsers(ui, cmd.realmClient, app)
	if usersErr != nil {
		return usersErr
	}

	for _, user := range users {
		err := cmd.realmClient.DisableUser(app.GroupID, app.ID, user.ID)
		cmd.outputs = append(cmd.outputs, userOutput{user: user, err: err})
	}
	return nil
}

// Feedback is the command feedback
func (cmd *CommandDisable) Feedback(profile *cli.Profile, ui terminal.UI) error {
	if len(cmd.outputs) == 0 {
		msg := "No users to disable"
		return ui.Print(terminal.NewTextLog(msg))
	}
	var outputsByProviderType = map[realm.AuthProviderType][]userOutput{}
	for _, output := range cmd.outputs {
		for _, identity := range output.user.Identities {
			outputsByProviderType[identity.ProviderType] = append(outputsByProviderType[identity.ProviderType], output)
		}
	}
	logs := make([]terminal.Log, 0, len(outputsByProviderType))
	for _, apt := range realm.ValidAuthProviderTypes {
		outputs := outputsByProviderType[apt]
		if len(outputs) == 0 {
			continue
		}
		sort.SliceStable(outputs, getUserOutputComparerBySuccess(outputs))
		logs = append(logs, terminal.NewTableLog(
			fmt.Sprintf("Provider type: %s", apt.Display()),
			append(userTableHeaders(apt), headerEnabled, headerDetails),
			userTableRows(apt, outputs, userDisableRow)...,
		))
	}
	return ui.Print(logs...)
}

func userDisableRow(output userOutput, row map[string]interface{}) {
	if output.err != nil && !output.user.Disabled {
		row[headerEnabled] = true
	} else {
		row[headerEnabled] = false
	}
	if output.err != nil {
		row[headerDetails] = output.err.Error()
	} else {
		row[headerDetails] = ""
	}
}
