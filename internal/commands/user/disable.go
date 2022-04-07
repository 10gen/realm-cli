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

// CommandMetaDisable is the command meta for the `user disable` command
var CommandMetaDisable = cli.CommandMeta{
	Use:         "disable",
	Display:     "user disable",
	Description: "Disable an application User of your Realm app",
	HelpText: `Deactivates a User on your Realm app. A User that has been disabled will not be
allowed to log in, even if they provide valid credentials.`,
}

// CommandDisable is the `user disable` command
type CommandDisable struct {
	inputs disableInputs
}

// Flags is the command flags
func (cmd *CommandDisable) Flags() []flags.Flag {
	return []flags.Flag{
		cli.AppFlagWithContext(&cmd.inputs.App, "to disable its users"),
		cli.ProjectFlag(&cmd.inputs.Project),
		cli.ProductFlag(&cmd.inputs.Products),
		usersFlag(&cmd.inputs.Users, "Specify the Realm app's users' ID(s) to disable"),
	}
}

// Inputs is the command inputs
func (cmd *CommandDisable) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *CommandDisable) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
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

	users, err := cmd.inputs.selectUsers(ui, found, "disable")
	if err != nil {
		return err
	}

	outputs := make(userOutputs, 0, len(users))
	for _, user := range users {
		err := clients.Realm.DisableUser(app.GroupID, app.ID, user.ID)
		outputs = append(outputs, userOutput{user, err})
	}

	if len(outputs) == 0 {
		ui.Print(terminal.NewTextLog("No users to disable"))
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
			append(tableHeaders(providerType), headerEnabled, headerDetails),
			tableRows(providerType, o, tableRowDisable)...,
		))
	}
	ui.Print(logs...)
	return nil
}

type disableInputs struct {
	cli.ProjectInputs
	multiUserInputs
}

func (i *disableInputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	return i.ProjectInputs.Resolve(ui, profile.WorkingDirectory, false)
}

func tableRowDisable(output userOutput, row map[string]interface{}) {
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
