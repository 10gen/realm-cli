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
		flags.NewEnumSet(&cmd.inputs.ProviderTypes, validProviderTypes),
		flagProvider,
		flagProviderShort,
		flagProviderUsage,
	)
	fs.VarP(&cmd.inputs.State, flagState, flagStateShort, flagStateUsage)
	fs.VarP(&cmd.inputs.Status, flagStatus, flagStatusShort, flagStatusUsage)
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
	app, err := app.Resolve(ui, cmd.realmClient, cmd.inputs.Filter())
	if err != nil {
		return err
	}

	users, err := cmd.inputs.ResolveUsers(ui, cmd.realmClient, app)
	if err != nil {
		return err
	}

	for _, user := range users {
		err = cmd.realmClient.DeleteUser(app.GroupID, app.ID, user.ID)
		cmd.outputs = append(cmd.outputs, userOutput{user: user, err: err})
	}
	return nil
}

// Feedback is the command feedback
func (cmd *CommandDelete) Feedback(profile *cli.Profile, ui terminal.UI) error {
	if len(cmd.outputs) == 0 {
		return ui.Print(terminal.NewTextLog("No users to delete"))
	}

	var outputByProviderType = map[string][]userOutput{}
	for _, output := range cmd.outputs {
		for _, identity := range output.user.Identities {
			outputByProviderType[identity.ProviderType] = append(outputByProviderType[identity.ProviderType], output)
		}
	}
	logs := make([]terminal.Log, len(outputByProviderType))
	logIndex := 0
	for providerType, outputs := range outputByProviderType {

		sort.Slice(outputs, getUserOutputComparerBySuccess(outputs))

		logs[logIndex] = terminal.NewTableLog(
			fmt.Sprintf("Provider type: %s", providerType),
			userDeleteTableHeaders(providerType),
			userDeleteTableRows(providerType, outputs)...,
		)
		logIndex++
	}
	return ui.Print(logs...)
}

func userDeleteTableHeaders(providerType string) []string {
	var headers []string
	switch providerType {
	case providerTypeAPIKey:
		headers = append(headers, headerName)
	case providerTypeLocalUserPass:
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

func userDeleteTableRows(providerType string, outputs []userOutput) []map[string]interface{} {
	userDeleteTableRows := make([]map[string]interface{}, 0, len(outputs))
	for _, output := range outputs {
		userDeleteTableRows = append(userDeleteTableRows, userDeleteTableRow(providerType, output))
	}
	return userDeleteTableRows
}

func userDeleteTableRow(providerType string, output userOutput) map[string]interface{} {
	msg := "n/a"
	if output.err != nil {
		msg = output.err.Error()
	}
	row := map[string]interface{}{
		headerID:      output.user.ID,
		headerType:    output.user.Type,
		headerDeleted: output.err == nil,
		headerDetails: msg,
	}
	switch providerType {
	case providerTypeAPIKey:
		row[headerName] = output.user.Data[userDataName]
	case providerTypeLocalUserPass:
		row[headerEmail] = output.user.Data[userDataEmail]
	}
	return row
}
