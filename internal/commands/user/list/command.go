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
	outputs     outputs
	realmClient realm.Client
	users       []realm.User
}

type outputs struct {
	apiKey realm.APIKey
	user   realm.User
}

func (cmd *command) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)

	fs.Var(&cmd.inputs.StateValue, flagState, flagStateUsage)
	fs.BoolVar(&cmd.inputs.Pending, flagStatus, false, flagStatusUsage)
	fs.Var(newProviderTypesValue(&cmd.inputs.ProviderTypes), flagProviderTypes, flagProviderTypesUsage)
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
			State:     cmd.inputs.StateValue.getUserState(),
			Pending:   cmd.inputs.Pending,
			Providers: cmd.inputs.ProviderTypes,
			//todo add users filter
		},
	)
	if usersErr != nil {
		return usersErr
	}

	cmd.users = users

	return nil
}

const (
	headerID                     = "ID"
	headerName                   = "Name"
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
			name := user.Data["name"]
			if name == nil { //what else should I use here?
				name = user.Data["email"]
			}
			timeString := ""
			if user.LastAuthenticationDate != 0 {
				timeString = time.Unix(user.LastAuthenticationDate, 0).String()
			}
			userTable = append(userTable, map[string]interface{}{
				headerID:                     user.ID,
				headerName:                   name,
				headerEnabled:                !user.Disabled,
				headerType:                   user.Type,
				headerLastAuthenticationDate: timeString,
			})
		}
		logs = append(logs, terminal.NewTableLog(
			fmt.Sprintf("Provider type: %s", provider),
			[]string{headerName, headerID, headerEnabled, headerType, headerLastAuthenticationDate},
			userTable...,
		))
	}

	return ui.Print(logs...)
}
