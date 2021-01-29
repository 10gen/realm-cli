package user

import (
	"errors"
	"fmt"
	"sort"

	"github.com/10gen/realm-cli/internal/app"
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
	"github.com/AlecAivazis/survey/v2"

	"github.com/spf13/pflag"
)

// CommandDelete is the `user delete` command
type CommandDelete struct {
	inputs      deleteInputs
	outputs     []userOutput
	realmClient realm.Client
}

type deleteInputs struct {
	app.ProjectInputs
	State         realm.UserState
	ProviderTypes []string
	Pending       bool
	Users         []string
}

func (i *deleteInputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory); err != nil {
		return err
	}
	return nil
}

// ResolveUsers will use the provided Realm client to resolve the users specified by the realm.App through inputs
func (i *deleteInputs) ResolveUsers(ui terminal.UI, client realm.Client, app realm.App) ([]realm.User, error) {
	filter := realm.UserFilter{
		IDs:       i.Users,
		State:     i.State,
		Pending:   i.Pending,
		Providers: realm.NewAuthProviderTypes(i.ProviderTypes...),
	}
	foundUsers, usersErr := client.FindUsers(app.GroupID, app.ID, filter)
	if usersErr != nil {
		return nil, usersErr
	}
	if len(i.Users) > 0 {
		if len(foundUsers) == 0 {
			return nil, errors.New("no users found")
		}
		return foundUsers, nil
	}

	selectableUsers := map[string]realm.User{}
	selectableUserOptions := make([]string, len(foundUsers))
	for idx, user := range foundUsers {
		var apt realm.AuthProviderType
		if len(user.Identities) > 0 {
			apt = user.Identities[0].ProviderType
		}
		opt := displayUser(apt, user)
		selectableUserOptions[idx] = opt
		selectableUsers[opt] = user
	}
	var selectedUsers []string
	askErr := ui.AskOne(
		&selectedUsers,
		&survey.MultiSelect{
			Message: "Which user(s) would you like to delete?",
			Options: selectableUserOptions,
		},
	)
	if askErr != nil {
		return nil, askErr
	}
	users := make([]realm.User, len(selectedUsers))
	for idx, user := range selectedUsers {
		users[idx] = selectableUsers[user]
	}
	return users, nil
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
	app, appErr := app.Resolve(ui, cmd.realmClient, cmd.inputs.Filter())
	if appErr != nil {
		return appErr
	}
	users, usersErr := cmd.inputs.ResolveUsers(ui, cmd.realmClient, app)
	if usersErr != nil {
		return usersErr
	}
	for _, user := range users {
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
			append(userTableHeaders(apt), headerDeleted, headerDetails),
			userTableRows(apt, outputs, userDeleteRow)...,
		))
	}
	return ui.Print(logs...)
}

func userDeleteRow(output userOutput, row map[string]interface{}) {
	if output.err != nil {
		row[headerDeleted] = false
		row[headerDetails] = output.err.Error()
	} else {
		row[headerDeleted] = true
		row[headerDetails] = ""
	}
}
