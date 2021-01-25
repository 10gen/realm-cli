package user

import (
	"errors"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
)

const (
	userDataEmail = "email"
	userDataName  = "name"
)

// usersInputs are the filtering inputs for a user command
type usersInputs struct {
	State         realm.UserState
	ProviderTypes []string
	Pending       bool
	Users         []string
}

// ResolveUsers will use the provided Realm client to resolve the users specified by the realm.App through inputs
func (i *usersInputs) ResolveUsers(ui terminal.UI, client realm.Client, app realm.App) ([]realm.User, error) {
	filter := realm.UserFilter{
		IDs:       i.Users,
		State:     i.State,
		Pending:   i.Pending,
		Providers: realm.StringSliceToProviderTypes(i.ProviderTypes...),
	}
	foundUsers, usersErr := client.FindUsers(app.GroupID, app.ID, filter)
	if usersErr != nil {
		return nil, usersErr
	}
	if len(i.Users) > 0 {
		if len(foundUsers) == 0 {
			return nil, errors.New("No users found")
		}
		return foundUsers, nil
	}

	selectableUsers := map[string]realm.User{}
	selectableUserOptions := make([]string, len(foundUsers))
	for idx, user := range foundUsers {
		var pt realm.ProviderType
		if len(user.Identities) == 0 {
			pt = ""
		} else {
			pt = user.Identities[0].ProviderType
		}
		opt := pt.DisplayUser(user)
		selectableUserOptions[idx] = opt
		selectableUsers[opt] = user
	}
	var selectedUsers []string
	askErr := ui.AskOne(
		&selectedUsers,
		&survey.MultiSelect{
			Message: "Which user(s) would you like to delete?",
			Options: selectableUserOptions,
		},
	)
	if askErr != nil {
		return nil, askErr
	}
	users := make([]realm.User, len(selectedUsers))
	for idx, user := range selectedUsers {
		users[idx] = selectableUsers[user]
	}
	return users, nil
}
