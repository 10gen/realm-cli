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

// CommandRevoke is the `user revoke` command
type CommandRevoke struct {
	inputs      revokeInputs
	outputs     []userOutput
	realmClient realm.Client
}

type revokeInputs struct {
	app.ProjectInputs
	usersInputs
}

func (i *revokeInputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory); err != nil {
		return err
	}
	return nil
}

// Flags is the command flags
func (cmd *CommandRevoke) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)
	fs.StringSliceVarP(&cmd.inputs.Users, flagUser, flagUserShort, []string{}, flagUserRevokeUsage)
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
func (cmd *CommandRevoke) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Setup is the command setup
func (cmd *CommandRevoke) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = profile.RealmAuthClient()
	return nil
}

// Handler is the command handler
func (cmd *CommandRevoke) Handler(profile *cli.Profile, ui terminal.UI) error {
	app, appErr := app.Resolve(ui, cmd.realmClient, cmd.inputs.Filter())
	if appErr != nil {
		return appErr
	}
	users, usersErr := cmd.inputs.ResolveUsers(ui, cmd.realmClient, app)
	if usersErr != nil {
		return usersErr
	}
	for _, user := range users {
		err := cmd.realmClient.RevokeUserSessions(app.GroupID, app.ID, user.ID)
		cmd.outputs = append(cmd.outputs, userOutput{user: user, err: err})
	}
	return nil
}

// Feedback is the command feedback
func (cmd *CommandRevoke) Feedback(profile *cli.Profile, ui terminal.UI) error {
	if len(cmd.outputs) == 0 {
		return ui.Print(terminal.NewTextLog("No users to revoke, try changing the --user input"))
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
			userRevokeTableHeaders(apt),
			userRevokeTableRows(apt, outputs)...,
		))
	}
	return ui.Print(logs...)
}

func userRevokeTableHeaders(authProviderType realm.AuthProviderType) []string {
	var headers []string
	switch authProviderType {
	case realm.AuthProviderTypeAPIKey:
		headers = append(headers, headerName)
	case realm.AuthProviderTypeUserPassword:
		headers = append(headers, headerEmail)
	}
	headers = append(
		headers,
		headerID,
		headerType,
		headerRevoked,
		headerDetails,
	)
	return headers
}

func userRevokeTableRows(authProviderType realm.AuthProviderType, outputs []userOutput) []map[string]interface{} {
	rows := make([]map[string]interface{}, 0, len(outputs))
	for _, output := range outputs {
		rows = append(rows, userRevokeTableRow(authProviderType, output))
	}
	return rows
}

func userRevokeTableRow(authProviderType realm.AuthProviderType, output userOutput) map[string]interface{} {
	var details string
	if output.err != nil {
		details = output.err.Error()
	}
	success := "no"
	if output.err == nil {
		success = "yes"
	}
	row := map[string]interface{}{
		headerID:      output.user.ID,
		headerType:    output.user.Type,
		headerRevoked: success,
		headerDetails: details,
	}
	switch authProviderType {
	case realm.AuthProviderTypeAPIKey:
		row[headerName] = output.user.Data[userDataName]
	case realm.AuthProviderTypeUserPassword:
		row[headerEmail] = output.user.Data[userDataEmail]
	}
	return row
}
