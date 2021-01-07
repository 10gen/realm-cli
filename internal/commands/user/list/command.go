package list

import (
	"fmt"
	"sort"
	"time"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// Command is the `user list` command
var Command = cli.CommandDefinition{
	Use:         "list",
	Description: "List the users of your Realm application",
	Help:        "user list",
	Command:     &command{},
}

type command struct {
	inputs      inputs
	realmClient realm.Client
	users       []realm.User
}

func (cmd *command) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)

	fs.VarP(&cmd.inputs.UserState, flagState, flagStateShort, flagStateUsage)
	fs.BoolVarP(&cmd.inputs.Pending, flagStatus, flagStatusShort, false, flagStatusUsage)
	fs.VarP(newProviderTypesValue(&cmd.inputs.ProviderTypes), flagProviderTypes, flagProviderTypesShort, flagProviderTypesUsage)
	fs.StringSliceVarP(&cmd.inputs.Users, flagUsers, flagUsersShort, []string{}, flagUsersUsage)
}

func (cmd *command) Inputs() cli.InputResolver {
	return &cmd.inputs
}

func (cmd *command) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = realm.NewAuthClient(profile.RealmBaseURL(), profile.Session())
	return nil
}

func (cmd *command) Handler(profile *cli.Profile, ui terminal.UI) error {
	app, appErr := cli.ResolveApp(ui, cmd.realmClient, cmd.inputs.Filter())
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
		return fmt.Errorf("failed to list users: %s", usersErr)
	}

	cmd.users = users

	return nil
}

const (
	headerID                     = "ID"
	headerName                   = "Name"
	headerEmail                  = "Email"
	headerEnabled                = "Enabled"
	headerType                   = "Type"
	headerLastAuthenticationDate = "Last Authentication"
)

func (cmd *command) Feedback(profile *cli.Profile, ui terminal.UI) error {
	if len(cmd.users) == 0 {
		return ui.Print(terminal.NewTextLog("No available users to show"))
	}
	var usersByProvider = make(map[string][]realm.User, 0)
	for _, user := range cmd.users {
		for _, identity := range user.Identities {
			usersByProvider[identity.ProviderType] = append(usersByProvider[identity.ProviderType], user)
		}
	}

	var logs []terminal.Log
	for provider, users := range usersByProvider {
		sort.Slice(
			users,
			func(i, j int) bool {
				return users[i].LastAuthenticationDate < users[j].LastAuthenticationDate
			},
		)
		userTable := make([]map[string]interface{}, 0, len(users))
		for _, user := range users {
			userTable = append(userTable, userTableRow(provider, user))
		}
		logs = append(logs, terminal.NewTableLog(
			fmt.Sprintf("Provider type: %s", provider),
			userTableHeaders(provider),
			userTable...,
		))
	}
	return ui.Print(logs...)
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

func userTableRow(providerType string, user realm.User) map[string]interface{} {
	timeString := ""
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
		row[headerName] = user.Data["name"]
	case providerTypeLocalUserPass:
		row[headerEmail] = user.Data["email"]
	}
	return row
}
