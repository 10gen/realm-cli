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

// NewCommandDisable returns a *CommandUserState which is the `user disable` command
func NewCommandDisable() *CommandUserState {
	return &CommandUserState{}
}

// NewCommandEnable returns a *CommandUserState which is the `user enable` command
func NewCommandEnable() *CommandUserState {
	return &CommandUserState{userEnable: true}
}

// CommandUserState is the `user enable/disable` command
type CommandUserState struct {
	userEnable  bool
	inputs      userStateInputs
	outputs     []userOutput
	realmClient realm.Client
}

type userStateInputs struct {
	app.ProjectInputs
	usersInputs
}

func (i *userStateInputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory); err != nil {
		return err
	}

	return nil
}

// Flags is the command flags
func (cmd *CommandUserState) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)
	var flagUsage string
	if cmd.userEnable {
		flagUsage = flagUserEnableUsage
	} else {
		flagUsage = flagUserDisableUsage
	}
	fs.StringSliceVarP(&cmd.inputs.Users, flagUser, flagUserShort, []string{}, flagUsage)
	fs.VarP(
		flags.NewEnumSet(&cmd.inputs.ProviderTypes, validAuthProviderTypes()),
		flagProvider,
		flagProviderShort,
		flagProviderUsage,
	)
}

// Inputs is the command inputs
func (cmd *CommandUserState) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Setup is the command setup
func (cmd *CommandUserState) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = profile.RealmAuthClient()
	return nil
}

// Handler is the command handler
func (cmd *CommandUserState) Handler(profile *cli.Profile, ui terminal.UI) error {
	app, err := app.Resolve(ui, cmd.realmClient, cmd.inputs.Filter())
	if err != nil {
		return err
	}

	cmd.inputs.Pending = false
	users, usersErr := cmd.inputs.ResolveUsers(ui, cmd.realmClient, app)
	if usersErr != nil {
		return usersErr
	}

	for _, user := range users {
		var err error
		if cmd.userEnable {
			err = cmd.realmClient.EnableUser(app.GroupID, app.ID, user.ID)
		} else {
			err = cmd.realmClient.DisableUser(app.GroupID, app.ID, user.ID)
		}
		cmd.outputs = append(cmd.outputs, userOutput{user: user, err: err})
	}
	return nil
}

// Feedback is the command feedback
func (cmd *CommandUserState) Feedback(profile *cli.Profile, ui terminal.UI) error {
	if len(cmd.outputs) == 0 {
		msg := "No users to "
		if cmd.userEnable {
			msg += "enable"
		} else {
			msg += "disable"
		}
		return ui.Print(terminal.NewTextLog(msg))
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
			userStateTableHeaders(apt),
			userStateTableRows(apt, outputs, cmd.userEnable)...,
		))
	}
	return ui.Print(logs...)
}

func userStateTableHeaders(authProviderType realm.AuthProviderType) []string {
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
		headerEnabled,
		headerDetails,
	)
	return headers
}

func userStateTableRows(authProviderType realm.AuthProviderType, outputs []userOutput, enableUser bool) []map[string]interface{} {
	userStateTableRows := make([]map[string]interface{}, 0, len(outputs))
	for _, output := range outputs {
		userStateTableRows = append(userStateTableRows, userStateTableRow(authProviderType, output, enableUser))
	}
	return userStateTableRows
}

func userStateTableRow(authProviderType realm.AuthProviderType, output userOutput, enableUser bool) map[string]interface{} {
	var details string
	if output.err != nil {
		details = output.err.Error()
	}
	var success bool
	if enableUser {
		success = output.err == nil
	} else {
		success = output.err != nil
	}
	row := map[string]interface{}{
		headerID:      output.user.ID,
		headerType:    output.user.Type,
		headerEnabled: success,
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
