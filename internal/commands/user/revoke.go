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
	inputs revokeInputs
}

// Flags is the command flags
func (cmd *CommandRevoke) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)
	fs.StringSliceVarP(&cmd.inputs.Users, flagUser, flagUserShort, []string{}, flagUserRevokeUsage)
	fs.BoolVar(&cmd.inputs.Pending, flagPending, false, flagPendingUsage)
	fs.Var(&cmd.inputs.State, flagState, flagStateUsage)
	fs.Var(
		flags.NewEnumSet(&cmd.inputs.ProviderTypes, validAuthProviderTypes()),
		flagProvider,
		flagProviderUsage,
	)
}

// Inputs is the command inputs
func (cmd *CommandRevoke) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *CommandRevoke) Handler(profile *cli.Profile, ui terminal.UI, clients cli.Clients) error {
	app, err := cli.ResolveApp(ui, clients.Realm, cmd.inputs.Filter())
	if err != nil {
		return err
	}

	found, err := cmd.inputs.findUsers(clients.Realm, app.GroupID, app.ID)
	if err != nil {
		return err
	}

	users, err := cmd.inputs.selectUsers(ui, found, "revoke")
	if err != nil {
		return err
	}

	outputs := make(userOutputs, 0, len(users))
	for _, user := range users {
		err := clients.Realm.RevokeUserSessions(app.GroupID, app.ID, user.ID)
		outputs = append(outputs, userOutput{user, err})
	}

	if len(outputs) == 0 {
		ui.Print(terminal.NewTextLog("No users to revoke sessions for"))
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
			append(tableHeaders(providerType), headerRevoked, headerDetails),
			tableRows(providerType, o, tableRowRevoke)...,
		))
	}

	ui.Print(logs...)
	return nil
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

func tableRowRevoke(output userOutput, row map[string]interface{}) {
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
