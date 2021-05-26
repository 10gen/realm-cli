package user

import (
	"fmt"
	"sort"
	"time"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/spf13/pflag"
)

// CommandMetaList is the command meta for the `user list` command
var CommandMetaList = cli.CommandMeta{
	Use:         "list",
	Aliases:     []string{"ls"},
	Description: "List the application users of your Realm app",
	HelpText: `Displays a list of your Realm app's Users' details. The list is grouped by Auth
Provider type and sorted by Last Authentication Date.`,
}

// CommandList is the `user list` command
type CommandList struct {
	inputs listInputs
}

type listInputs struct {
	cli.ProjectInputs
	multiUserInputs
}

// Flags is the command flags
func (cmd *CommandList) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)

	fs.StringSliceVarP(&cmd.inputs.Users, flagUser, flagUserShort, []string{}, flagUserListUsage)
	fs.BoolVar(&cmd.inputs.Pending, flagPending, false, flagPendingUsage)
	fs.Var(&cmd.inputs.State, flagState, flagStateUsage)
	fs.Var(
		flags.NewEnumSet(&cmd.inputs.ProviderTypes, validAuthProviderTypes()),
		flagProvider,
		flagProviderUsage,
	)
}

// Inputs is the command inputs
func (cmd *CommandList) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *CommandList) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	app, err := cli.ResolveApp(ui, clients.Realm, cmd.inputs.Filter())
	if err != nil {
		return err
	}

	users, err := cmd.inputs.findUsers(clients.Realm, app.GroupID, app.ID)
	if err != nil {
		return err
	}

	outputs := make(userOutputs, 0, len(users))
	for _, user := range users {
		outputs = append(outputs, userOutput{user, err})
	}

	if len(outputs) == 0 {
		ui.Print(terminal.NewTextLog("No available users to show"))
		return nil
	}

	outputsByProviderType := outputs.byProviderType()

	logs := make([]terminal.Log, 0, len(outputsByProviderType))
	for _, providerType := range realm.ValidAuthProviderTypes {
		o := outputsByProviderType[providerType]
		if len(o) == 0 {
			continue
		}

		sort.Slice(o, getUserComparerByLastAuthentication(o))

		logs = append(logs, terminal.NewTableLog(
			fmt.Sprintf("Provider type: %s", providerType.Display()),
			append(tableHeaders(providerType), headerEnabled, headerLastAuthenticationDate),
			tableRows(providerType, o, tableRowList)...,
		))
	}

	ui.Print(logs...)
	return nil
}

func getUserComparerByLastAuthentication(outputs []userOutput) func(i, j int) bool {
	return func(i, j int) bool {
		return outputs[i].user.LastAuthenticationDate > outputs[j].user.LastAuthenticationDate
	}
}

func (i *listInputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	return i.ProjectInputs.Resolve(ui, profile.WorkingDirectory, false)
}

func tableRowList(output userOutput, row map[string]interface{}) {
	timeString := "n/a"
	if output.user.LastAuthenticationDate != 0 {
		timeString = time.Unix(output.user.LastAuthenticationDate, 0).UTC().String()
	}
	row[headerLastAuthenticationDate] = timeString
	row[headerEnabled] = !output.user.Disabled
}
