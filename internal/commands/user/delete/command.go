package delete

import (
	"fmt"
	"sort"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/commands/user/shared"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
	"github.com/spf13/pflag"
)

// Command is the `user delete` command
var Command = cli.CommandDefinition{
	Use:         "delete",
	Description: "Delete the user(s) from a Realm application",
	Help:        "user delete",
	Command:     &command{},
}

type command struct {
	inputs      inputs
	outputs     []output
	realmClient realm.Client
}

type output struct {
	user realm.User
	err  error
}

func (cmd *command) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)

	fs.StringSliceVarP(&cmd.inputs.Users, flagUsers, flagUsersShort, []string{}, flagUsersUsage)
	fs.VarP(
		flags.NewEnumSet(&cmd.inputs.ProviderTypes, shared.ValidProviderTypes),
		shared.FlagProvider,
		shared.FlagProviderShort,
		shared.FlagProviderUsage,
	)
	fs.Lookup(shared.FlagProvider).NoOptDefVal = shared.ProviderTypeInteractive
	fs.VarP(&cmd.inputs.State, shared.FlagStateType, shared.FlagStateTypeShort, shared.FlagStateTypeUsage)
	fs.Lookup(shared.FlagProvider).NoOptDefVal = shared.ProviderTypeInteractive
	fs.VarP(&cmd.inputs.Status, shared.FlagStatusType, shared.FlagStatusTypeShort, shared.FlagStatusTypeUsage)
	fs.Lookup(shared.FlagStatusType).NoOptDefVal = shared.StatusTypeInteractive.String()
}

func (cmd *command) Inputs() cli.InputResolver {
	return &cmd.inputs
}

func (cmd *command) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = realm.NewAuthClient(profile.RealmBaseURL(), profile.Session())
	return nil
}

func (cmd *command) Handler(profile *cli.Profile, ui terminal.UI) error {
	app, err := cli.ResolveApp(ui, cmd.realmClient, cmd.inputs.Filter())
	if err != nil {
		return err
	}

	var users []realm.User
	if len(cmd.inputs.Users) < 1 {
		users, err = cmd.inputs.ResolveUsers(ui, cmd.realmClient, app)
		if err != nil {
			return err
		}
	} else {
		for _, userID := range cmd.inputs.Users {
			foundUser, err := cmd.realmClient.FindUsers(app.GroupID, app.ID, realm.UserFilter{IDs: []string{userID}})
			if err != nil {
				return fmt.Errorf("Unable to find user with ID %s", userID)
			}
			users = append(users, foundUser[0])
		}
	}

	for _, user := range users {
		err = cmd.realmClient.DeleteUser(app.GroupID, app.ID, user.ID)
		cmd.outputs = append(cmd.outputs, output{user: user, err: err})
	}
	return nil
}

func (cmd *command) Feedback(profile *cli.Profile, ui terminal.UI) error {
	if len(cmd.outputs) == 0 {
		return ui.Print(terminal.NewTextLog("No users to delete"))
	}

	var outputByProviderType = make(map[string][]output)
	for _, output := range cmd.outputs {
		for _, identity := range output.user.Identities {
			outputByProviderType[identity.ProviderType] = append(outputByProviderType[identity.ProviderType], output)
		}
	}
	var logs []terminal.Log
	for providerType, outputs := range outputByProviderType {

		sort.Slice(outputs, getOutputComparerBySuccess(outputs))

		logs = append(logs, terminal.NewTableLog(
			fmt.Sprintf("Provider type: %s", providerType),
			userTableHeaders(providerType),
			userTableRows(providerType, outputs)...,
		))
	}
	return ui.Print(logs...)
}

func getOutputComparerBySuccess(outputs []output) func(i, j int) bool {
	return func(i, j int) bool {
		return !(outputs[i].err != nil && outputs[i].err == nil)
	}
}

func userTableHeaders(providerType string) []string {
	var headers []string
	switch providerType {
	case shared.ProviderTypeAPIKey:
		headers = append(headers, shared.HeaderName)
	case shared.ProviderTypeLocalUserPass:
		headers = append(headers, shared.HeaderEmail)
	}
	headers = append(
		headers,
		shared.HeaderID,
		shared.HeaderType,
		shared.HeaderDeleted,
		shared.HeaderDetails,
	)
	return headers
}

func userTableRows(providerType string, outputs []output) []map[string]interface{} {
	userTableRows := make([]map[string]interface{}, 0, len(outputs))
	for _, output := range outputs {
		userTableRows = append(userTableRows, userTableRow(providerType, output))
	}
	return userTableRows
}

func userTableRow(providerType string, output output) map[string]interface{} {
	msg := "n/a"
	if output.err != nil {
		msg = output.err.Error()
	}
	row := map[string]interface{}{
		shared.HeaderID:      output.user.ID,
		shared.HeaderType:    output.user.Type,
		shared.HeaderDeleted: output.err == nil,
		shared.HeaderDetails: msg,
	}
	switch providerType {
	case shared.ProviderTypeAPIKey:
		row[shared.HeaderName] = output.user.Data[shared.UserDataName]
	case shared.ProviderTypeLocalUserPass:
		row[shared.HeaderEmail] = output.user.Data[shared.UserDataEmail]
	}
	return row
}
