package user

import (
	"fmt"
	"sort"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// CommandDisable is the `user disable` command
type CommandDisable struct {
	inputs      disableInputs
	outputs     userOutputs
	realmClient realm.Client
}

type disableInputs struct {
	cli.ProjectInputs
	multiUserInputs
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
	app, appErr := cli.ResolveApp(ui, cmd.realmClient, cmd.inputs.Filter())
	if appErr != nil {
		return appErr
	}

	resolved, resolvedErr := cmd.inputs.resolveUsers(cmd.realmClient, app.GroupID, app.ID)
	if resolvedErr != nil {
		return resolvedErr
	}

	selected, selectedErr := selectUsers(ui, resolved, "disable")
	if selectedErr != nil {
		return selectedErr
	}
	for _, user := range selected {
		err := cmd.realmClient.DisableUser(app.GroupID, app.ID, user.ID)
		cmd.outputs = append(cmd.outputs, userOutput{user, err})
	}
	return nil
}

// Feedback is the command feedback
func (cmd *CommandDisable) Feedback(profile *cli.Profile, ui terminal.UI) error {
	if len(cmd.outputs) == 0 {
		return ui.Print(terminal.NewTextLog("No users to disable"))
	}
	outputsByProviderType := cmd.outputs.mapByProviderType()
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

func (i *disableInputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	return i.ProjectInputs.Resolve(ui, profile.WorkingDirectory)
}

func userDisableRow(output userOutput, row map[string]interface{}) {
	var enabled bool
	var details string
	if output.err != nil && !output.user.Disabled {
		enabled = true
	}
	if output.err != nil {
		details = output.err.Error()
	}
	row[headerEnabled] = enabled
	row[headerDetails] = details
}
