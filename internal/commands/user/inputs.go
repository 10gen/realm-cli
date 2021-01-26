package user

import (
	"errors"
	"fmt"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
)

const (
	userDataEmail = "email"
	userDataName  = "name"
)

func validAuthProviderTypes() []interface{} {
	apts := make([]interface{}, 0, len(realm.ValidAuthProviderTypes))
	for _, apt := range realm.ValidAuthProviderTypes {
		apts = append(apts, apt)
	}
	return apts
}

func displayUser(apt realm.AuthProviderType, user realm.User) string {
	sep := " - "
	display := apt.Display()
	displayUserData, err := displayUserData(apt, user)
	if err != nil {
		return ""
	}
	if displayUserData != "" {
		display += sep + displayUserData
	}
	return display + sep + user.ID
}

func displayUserData(apt realm.AuthProviderType, user realm.User) (string, error) {
	var val interface{}
	ok := false
	switch apt {
	case realm.AuthProviderTypeUserPassword:
		val, ok = user.Data["email"]
	case realm.AuthProviderTypeAPIKey:
		val, ok = user.Data["name"]
	default:
		return "", nil
	}
	if ok {
		return fmt.Sprint(val), nil
	}
	return "", errors.New("User does not have ProviderType Data")
}

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
		Providers: realm.NewAuthProviderTypes(i.ProviderTypes...),
	}
	foundUsers, usersErr := client.FindUsers(app.GroupID, app.ID, filter)
	if usersErr != nil {
		return nil, usersErr
	}
	if len(i.Users) > 0 {
		if len(foundUsers) == 0 {
			return nil, errors.New("no users found")
		}
		return foundUsers, nil
	}

	selectableUsers := map[string]realm.User{}
	selectableUserOptions := make([]string, len(foundUsers))
	for idx, user := range foundUsers {
		var apt realm.AuthProviderType
		if len(user.Identities) > 0 {
			apt = user.Identities[0].ProviderType
		}
		opt := displayUser(apt, user)
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
