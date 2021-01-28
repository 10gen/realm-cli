package user

import (
	"fmt"
	"sort"
	"time"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/spf13/pflag"
)

// CommandList is the `user list` command
type CommandList struct {
	inputs      listInputs
	realmClient realm.Client
	outputs     userOutputs
}

type listInputs struct {
	cli.ProjectInputs
	UserState     realm.UserState
	Pending       bool
	ProviderTypes []string
	Users         []string
}

// Flags is the command flags
func (cmd *CommandList) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)

	fs.VarP(&cmd.inputs.UserState, flagState, flagStateShort, flagStateUsage)
	fs.BoolVarP(&cmd.inputs.Pending, flagPending, flagPendingShort, false, flagPendingUsage)
	fs.VarP(
		flags.NewEnumSet(&cmd.inputs.ProviderTypes, validAuthProviderTypes()),
		flagProvider,
		flagProviderShort,
		flagProviderUsage,
	)
	fs.StringSliceVarP(&cmd.inputs.Users, flagUser, flagUserShort, []string{}, flagUserUsage)
}

// Inputs is the command inputs
func (cmd *CommandList) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Setup is the command setup
func (cmd *CommandList) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = profile.RealmAuthClient()
	return nil
}

// Handler is the command handler
func (cmd *CommandList) Handler(profile *cli.Profile, ui terminal.UI) error {
	app, appErr := cli.ResolveApp(ui, cmd.realmClient, cmd.inputs.Filter())
	if appErr != nil {
		return appErr
	}
	users, usersErr := cmd.realmClient.FindUsers(
		app.GroupID,
		app.ID,
		realm.UserFilter{
			State:     cmd.inputs.UserState,
			Pending:   cmd.inputs.Pending,
			Providers: realm.NewAuthProviderTypes(cmd.inputs.ProviderTypes...),
			IDs:       cmd.inputs.Users,
		},
	)
	if usersErr != nil {
		return usersErr
	}
	cmd.outputs = make([]userOutput, 0, len(users))
	for _, user := range users {
		cmd.outputs = append(cmd.outputs, userOutput{user, nil})
	}

	return nil
}

// Feedback is the command feedback
func (cmd *CommandList) Feedback(profile *cli.Profile, ui terminal.UI) error {
	if len(cmd.outputs) == 0 {
		return ui.Print(terminal.NewTextLog("No available users to show"))
	}
	outputsByProviderType := cmd.outputs.mapByProviderType()
	logs := make([]terminal.Log, 0, len(outputsByProviderType))
	for _, apt := range realm.ValidAuthProviderTypes {
		outputs := outputsByProviderType[apt]
		if len(outputs) == 0 {
			continue
		}
		sort.Slice(outputs, getUserComparerByLastAuthentication(outputs))

		logs = append(logs, terminal.NewTableLog(
			fmt.Sprintf("Provider type: %s", apt.Display()),
			append(userTableHeaders(apt), headerEnabled, headerLastAuthenticationDate),
			userTableRows(apt, outputs, userListRow)...,
		))
	}
	return ui.Print(logs...)
}

func getUserComparerByLastAuthentication(outputs []userOutput) func(i, j int) bool {
	return func(i, j int) bool {
		return outputs[i].user.LastAuthenticationDate > outputs[j].user.LastAuthenticationDate
	}
}

func (i *listInputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	return i.ProjectInputs.Resolve(ui, profile.WorkingDirectory)
}

func userListRow(output userOutput, row map[string]interface{}) {
	timeString := "n/a"
	if output.user.LastAuthenticationDate != 0 {
		timeString = time.Unix(output.user.LastAuthenticationDate, 0).UTC().String()
	}
	row[headerLastAuthenticationDate] = timeString
	row[headerEnabled] = !output.user.Disabled
}
