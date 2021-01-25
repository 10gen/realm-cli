package user

import (
	"fmt"
	"sort"

	"github.com/10gen/realm-cli/internal/app"
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"

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
	usersInputs
}

func (i *deleteInputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory); err != nil {
		return err
	}
	return nil
}

// Flags is the command flags
func (cmd *CommandDelete) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)
	fs.StringSliceVarP(&cmd.inputs.Users, flagUser, flagUserShort, []string{}, flagUserDeleteUsage)
	fs.VarP(
		flags.NewEnumSet(&cmd.inputs.ProviderTypes, realm.ValidProviderTypes),
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
	cmd.realmClient = realm.NewAuthClient(profile)
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
	for _, pt := range realm.ValidProviderTypes {
		providerType := realm.AuthProviderType(pt)
		outputs := outputsByProviderType[providerType]
		if len(outputs) == 0 {
			continue
		}
		sort.SliceStable(outputs, getUserOutputComparerBySuccess(outputs))
		logs = append(logs, terminal.NewTableLog(
			fmt.Sprintf("Provider type: %s", providerType.Display()),
			userDeleteTableHeaders(providerType),
			userDeleteTableRows(providerType, outputs)...,
		))
	}
	return ui.Print(logs...)
}

func userDeleteTableHeaders(providerType realm.AuthProviderType) []string {
	var headers []string
	switch providerType {
	case realm.AuthProviderTypeAPIKey:
		headers = append(headers, headerName)
	case realm.AuthProviderTypeUserPassword:
		headers = append(headers, headerEmail)
	}
	headers = append(
		headers,
		headerID,
		headerType,
		headerDeleted,
		headerDetails,
	)
	return headers
}

func userDeleteTableRows(providerType realm.AuthProviderType, outputs []userOutput) []map[string]interface{} {
	rows := make([]map[string]interface{}, 0, len(outputs))
	for _, output := range outputs {
		rows = append(rows, userDeleteTableRow(providerType, output))
	}
	return rows
}

func userDeleteTableRow(providerType realm.AuthProviderType, output userOutput) map[string]interface{} {
	var details string
	if output.err != nil {
		details = output.err.Error()
	}
	row := map[string]interface{}{
		headerID:      output.user.ID,
		headerType:    output.user.Type,
		headerDeleted: output.err == nil,
		headerDetails: details,
	}
	switch providerType {
	case realm.AuthProviderTypeAPIKey:
		row[headerName] = output.user.Data[userDataName]
	case realm.AuthProviderTypeUserPassword:
		row[headerEmail] = output.user.Data[userDataEmail]
	}
	return row
}
