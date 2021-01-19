package user

import (
	"fmt"
	"sort"
	"time"

	"github.com/10gen/realm-cli/internal/app"
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/spf13/pflag"
)

// CommandList is the `user list` command
type CommandList struct {
	inputs      listInputs
	realmClient realm.Client
	users       []realm.User
}

type listInputs struct {
	app.ProjectInputs
	UserState     realm.UserState
	Pending       bool
	ProviderTypes []string
	Users         []string
}

// Flags is the command flags
func (cmd *CommandList) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)

	fs.VarP(&cmd.inputs.UserState, flagState, flagStateShort, flagStateUsage)
	fs.BoolVarP(&cmd.inputs.Pending, flagPending, flagPendingShort, false, flagPendingUsage)
	fs.VarP(
		flags.NewEnumSet(&cmd.inputs.ProviderTypes, validProviderTypes),
		flagProvider,
		flagProviderShort,
		flagProviderUsage,
	)
	fs.StringSliceVarP(&cmd.inputs.Users, flagUser, flagUserShort, []string{}, flagUserUsage)
}

// Inputs is the command inputs
func (cmd *CommandList) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Setup is the command setup
func (cmd *CommandList) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = realm.NewAuthClient(profile)
	return nil
}

// Handler is the command handler
func (cmd *CommandList) Handler(profile *cli.Profile, ui terminal.UI) error {
	app, appErr := app.Resolve(ui, cmd.realmClient, cmd.inputs.Filter())
	if appErr != nil {
		return appErr
	}
	users, usersErr := cmd.realmClient.FindUsers(
		app.GroupID,
		app.ID,
		realm.UserFilter{
			State:     cmd.inputs.UserState,
			Pending:   cmd.inputs.Pending,
			Providers: cmd.inputs.ProviderTypes,
			IDs:       cmd.inputs.Users,
		},
	)
	if usersErr != nil {
		return usersErr
	}

	cmd.users = users

	return nil
}

// Feedback is the command feedback
func (cmd *CommandList) Feedback(profile *cli.Profile, ui terminal.UI) error {
	if len(cmd.users) == 0 {
		return ui.Print(terminal.NewTextLog("No available users to show"))
	}

	var usersByProviderType = make(map[string][]realm.User)
	for _, user := range cmd.users {
		for _, identity := range user.Identities {
			usersByProviderType[identity.ProviderType] = append(usersByProviderType[identity.ProviderType], user)
		}
	}

	var logs []terminal.Log
	for providerType, users := range usersByProviderType {
		sort.Slice(users, getUserComparerByLastAuthentication(users))

		logs = append(logs, terminal.NewTableLog(
			fmt.Sprintf("Provider type: %s", providerType),
			userTableHeaders(providerType),
			userTableRows(providerType, users)...,
		))
	}
	return ui.Print(logs...)
}

func getUserComparerByLastAuthentication(users []realm.User) func(i, j int) bool {
	return func(i, j int) bool {
		return users[i].LastAuthenticationDate > users[j].LastAuthenticationDate
	}
}

func userTableHeaders(providerType string) []string {
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
		headerEnabled,
		headerType,
		headerLastAuthenticationDate,
	)
	return headers
}

func userTableRows(providerType string, users []realm.User) []map[string]interface{} {
	userTableRows := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		userTableRows = append(userTableRows, userTableRow(providerType, user))
	}
	return userTableRows
}

func userTableRow(providerType string, user realm.User) map[string]interface{} {
	timeString := "n/a"
	if user.LastAuthenticationDate != 0 {
		timeString = time.Unix(user.LastAuthenticationDate, 0).UTC().String()
	}
	row := map[string]interface{}{
		headerID:                     user.ID,
		headerEnabled:                !user.Disabled,
		headerType:                   user.Type,
		headerLastAuthenticationDate: timeString,
	}
	switch providerType {
	case providerTypeAPIKey:
		row[headerName] = user.Data[userDataName]
	case providerTypeLocalUserPass:
		row[headerEmail] = user.Data[userDataEmail]
	}
	return row
}

func (i *listInputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory); err != nil {
		return err
	}

	return nil
}
