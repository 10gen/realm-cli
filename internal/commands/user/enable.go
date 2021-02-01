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

// CommandEnable is the `user enable` command
type CommandEnable struct {
	inputs      enableInputs
	outputs     userOutputs
	realmClient realm.Client
}

type enableInputs struct {
	app.ProjectInputs
	Users []string
}

func (i *enableInputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory); err != nil {
		return err
	}

	return nil
}

// Flags is the command flags
func (cmd *CommandEnable) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)
	fs.StringSliceVarP(&cmd.inputs.Users, flagUser, flagUserShort, []string{}, flagUserEnableUsage)
}

// Inputs is the command inputs
func (cmd *CommandEnable) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Setup is the command setup
func (cmd *CommandEnable) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = profile.RealmAuthClient()
	return nil
}

// Handler is the command handler
func (cmd *CommandEnable) Handler(profile *cli.Profile, ui terminal.UI) error {
	app, err := app.Resolve(ui, cmd.realmClient, cmd.inputs.Filter())
	if err != nil {
		return err
	}

	users, usersErr := cmd.realmClient.FindUsers(app.GroupID, app.ID, realm.UserFilter{IDs: cmd.inputs.Users})
	if usersErr != nil {
		return usersErr
	}

	for _, user := range users {
		err := cmd.realmClient.EnableUser(app.GroupID, app.ID, user.ID)
		cmd.outputs = append(cmd.outputs, userOutput{user: user, err: err})
	}
	return nil
}

// Feedback is the command feedback
func (cmd *CommandEnable) Feedback(profile *cli.Profile, ui terminal.UI) error {
	if len(cmd.outputs) == 0 {
		return ui.Print(terminal.NewTextLog("No users to enable"))
	}
	outputsByProviderType := cmd.outputs.outputsByProviderType()
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
			userTableRows(apt, outputs, userEnableRow)...,
		))
	}
	return ui.Print(logs...)
}

func userEnableRow(output userOutput, row map[string]interface{}) {
	var (
		enabled bool
		details string
	)
	if (output.err == nil && output.user.Disabled) || (output.err != nil && !output.user.Disabled) {
		enabled = true
	}
	if output.err != nil {
		details = output.err.Error()
	}
	row[headerEnabled] = enabled
	row[headerDetails] = details
}