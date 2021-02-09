package user

import (
	"fmt"
	"sort"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/spf13/pflag"
)

// CommandRevoke is the `user revoke` command
type CommandRevoke struct {
	inputs      revokeInputs
	outputs     userOutputs
	realmClient realm.Client
}

type revokeInputs struct {
	cli.ProjectInputs
	multiUserInputs
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
	app, appErr := cli.ResolveApp(ui, cmd.realmClient, cmd.inputs.Filter())
	if appErr != nil {
		return appErr
	}

	users, userErr := cmd.inputs.resolveUsers(ui, cmd.realmClient, app)
	if userErr != nil {
		return userErr
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
		return ui.Print(terminal.NewTextLog("No users to revoke sessions for"))
	}
	outputsByProviderType := cmd.outputs.mapByProviderType()
	logs := make([]terminal.Log, 0, len(outputsByProviderType))
	for _, apt := range realm.ValidAuthProviderTypes {
		outputs := outputsByProviderType[apt]
		if len(outputs) == 0 {
			continue
		}
		sort.SliceStable(outputs, getUserOutputComparerBySuccess(outputs))
		logs = append(logs, terminal.NewTableLog(
			fmt.Sprintf("Provider type: %s", apt.Display()),
			append(userTableHeaders(apt), headerRevoked, headerDetails),
			userTableRows(apt, outputs, userRevokeRow)...,
		))
	}
	return ui.Print(logs...)
}

func userRevokeRow(output userOutput, row map[string]interface{}) {
	var revoked bool
	var details string
	if output.err != nil {
		details = output.err.Error()
	} else {
		revoked = true
	}
	row[headerRevoked] = revoked
	row[headerDetails] = details
}

func (i revokeInputs) resolveUsers(ui terminal.UI, realmClient realm.Client, app realm.App) ([]realm.User, error) {
	found, foundErr := i.findUsers(realmClient, app.GroupID, app.ID)
	if foundErr != nil {
		return nil, foundErr
	}

	return i.selectUsers(ui, found, "revoke")
}
