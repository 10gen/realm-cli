package user

import (
	"fmt"
	"sort"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

// CommandMetaDelete is the command meta for the `user delete` command
var CommandMetaDelete = cli.CommandMeta{
	Use:         "delete",
	Display:     "user delete",
	Description: "Delete an application user from your Realm app",
	HelpText: `You can remove multiple Users at once with the "--user" flag. You can only
specify these Users using their ID values.`,
}

// CommandDelete is the `user delete` command
type CommandDelete struct {
	inputs deleteInputs
}

// Flags is the command flags
func (cmd *CommandDelete) Flags() []flags.Flag {
	return []flags.Flag{
		cli.AppFlagWithContext(&cmd.inputs.App, "to delete its users"),
		cli.ProjectFlag(&cmd.inputs.Project),
		cli.ProductFlag(&cmd.inputs.Products),
		usersFlag(&cmd.inputs.Users, "Specify the Realm app's users' ID(s) to delete"),
		pendingFlag(&cmd.inputs.Pending),
		stateFlag(&cmd.inputs.State),
		providersFlag(&cmd.inputs.ProviderTypes),
	}
}

// Inputs is the command inputs
func (cmd *CommandDelete) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *CommandDelete) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	app, err := cli.ResolveApp(ui, clients.Realm, cli.AppOptions{
		AppMeta: cmd.inputs.AppMeta,
		Filter:  cmd.inputs.Filter(),
	})
	if err != nil {
		return err
	}

	found, err := cmd.inputs.findUsers(clients.Realm, app.GroupID, app.ID)
	if err != nil {
		return err
	}

	users, err := cmd.inputs.selectUsers(ui, found, "delete")
	if err != nil {
		return err
	}

	outputs := make(userOutputs, 0, len(users))
	for _, user := range users {
		err := clients.Realm.DeleteUser(app.GroupID, app.ID, user.ID)
		outputs = append(outputs, userOutput{user, err})
	}

	if len(outputs) == 0 {
		ui.Print(terminal.NewTextLog("No users to delete"))
		return nil
	}

	outputsByProviderType := outputs.byProviderType()

	logs := make([]terminal.Log, 0, len(outputsByProviderType))
	for _, providerType := range realm.ValidAuthProviderTypes {
		o := outputsByProviderType[providerType]
		if len(o) == 0 {
			continue
		}

		sort.SliceStable(o, getUserOutputComparerBySuccess(o))

		logs = append(logs, terminal.NewTableLog(
			fmt.Sprintf("Provider type: %s", providerType.Display()),
			append(tableHeaders(providerType), headerDeleted, headerDetails),
			tableRows(providerType, o, tableRowDelete)...,
		))
	}

	ui.Print(logs...)
	return nil
}

type deleteInputs struct {
	cli.ProjectInputs
	multiUserInputs
}

func (i *deleteInputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory, false); err != nil {
		return err
	}
	return nil
}

func tableRowDelete(output userOutput, row map[string]interface{}) {
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
