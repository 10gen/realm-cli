package user

import (
	"fmt"
	"sort"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/spf13/pflag"
)

// CommandDelete is the `user delete` command
type CommandDelete struct {
	inputs      deleteInputs
	outputs     userOutputs
	realmClient realm.Client
}

type deleteInputs struct {
	cli.ProjectInputs
	multiUserInputs
}

// Flags is the command flags
func (cmd *CommandDelete) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)
	fs.StringSliceVarP(&cmd.inputs.Users, flagUser, flagUserShort, []string{}, flagUserDeleteUsage)
	fs.VarP(
		flags.NewEnumSet(&cmd.inputs.ProviderTypes, validAuthProviderTypes()),
		flagProvider,
		flagProviderShort,
		flagProviderUsage,
	)
	fs.VarP(&cmd.inputs.State, flagState, flagStateShort, flagStateUsage)
	fs.BoolVarP(&cmd.inputs.Pending, flagPending, flagPendingShort, false, flagPendingUsage)
}

// Inputs is the command inputs
func (cmd *CommandDelete) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Setup is the command setup
func (cmd *CommandDelete) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = profile.RealmAuthClient()
	return nil
}

// Handler is the command handler
func (cmd *CommandDelete) Handler(profile *cli.Profile, ui terminal.UI) error {
	app, appErr := cli.ResolveApp(ui, cmd.realmClient, cmd.inputs.Filter())
	if appErr != nil {
		return appErr
	}

	resolved, resolvedErr := cmd.inputs.resolveUsers(cmd.realmClient, app.GroupID, app.ID)
	if resolvedErr != nil {
		return resolvedErr
	}

	selected, selectedErr := cmd.inputs.selectUsers(ui, resolved, "delete")
	if selectedErr != nil {
		return selectedErr
	}
	for _, user := range selected {
		err := cmd.realmClient.DeleteUser(app.GroupID, app.ID, user.ID)
		cmd.outputs = append(cmd.outputs, userOutput{user: user, err: err})
	}
	return nil
}

// Feedback is the command feedback
func (cmd *CommandDelete) Feedback(profile *cli.Profile, ui terminal.UI) error {
	if len(cmd.outputs) == 0 {
		return ui.Print(terminal.NewTextLog("No users to delete"))
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
			append(userTableHeaders(apt), headerDeleted, headerDetails),
			userTableRows(apt, outputs, userDeleteRow)...,
		))
	}
	return ui.Print(logs...)
}

func (i *deleteInputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory); err != nil {
		return err
	}
	return nil
}

func userDeleteRow(output userOutput, row map[string]interface{}) {
	var deleted bool
	var details string
	if output.err != nil {
		details = output.err.Error()
	} else {
		deleted = true
	}
	row[headerDeleted] = deleted
	row[headerDetails] = details
}
