package user

import (
	"fmt"
	"strings"

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
	State         userStateType
	ProviderTypes []string
	Status        statusType
	Users         []string
}

// ResolveUsers will use the provided Realm client to resolve the users specified by the realm.App through inputs
func (i *usersInputs) ResolveUsers(ui terminal.UI, client realm.Client, app realm.App) ([]realm.User, error) {
	var err error

	if len(i.Users) > 0 {
		var users []realm.User
		for _, userID := range i.Users {
			foundUser, err := client.FindUsers(app.GroupID, app.ID, realm.UserFilter{IDs: []string{userID}})
			if err != nil {
				return nil, fmt.Errorf("Unable to find user with ID %s", userID)
			}
			users = append(users, foundUser[0])
		}
		return users, nil
	}

	userFilter := realm.UserFilter{}
	if len(i.ProviderTypes) > 0 {
		userFilter.Providers = i.ProviderTypes
	}
	if i.State != userStateTypeEmpty {
		userFilter.State = realm.UserState(i.State.String())
	}
	selectableUsersSet := map[string]realm.User{}
	if i.Status == statusTypeConfirmed || i.Status == statusTypeEmpty {
		foundUsers, err := client.FindUsers(app.GroupID, app.ID, userFilter)
		if err != nil {
			return nil, err
		}
		for _, user := range foundUsers {
			selectableUsersSet[user.ID] = user
		}
	}
	if i.Status == statusTypePending || i.Status == statusTypeEmpty {
		userFilter.Pending = true
		foundUsers, err := client.FindUsers(app.GroupID, app.ID, userFilter)
		if err != nil {
			return nil, err
		}
		for _, user := range foundUsers {
			selectableUsersSet[user.ID] = user
		}
	}

	// Interactive User Selection
	selectableUsers := map[string]realm.User{}
	selectableUserOptions := make([]string, len(selectableUsersSet))
	sep := " - "
	userOptIndex := 0
	for _, user := range selectableUsersSet {
		var opt string
		switch user.Identities[0].ProviderType {
		case providerTypeAnonymous:
			opt = strings.Join([]string{"Anonymous", user.ID}, sep)
		case providerTypeLocalUserPass:
			opt = strings.Join([]string{"User/Password", fmt.Sprint(user.Data[userDataEmail]), user.ID}, sep)
		case providerTypeAPIKey:
			opt = strings.Join([]string{"ApiKey", fmt.Sprint(user.Data[userDataName]), user.ID}, sep)
		// case providerTypeApple:
		// 	opt = strings.Join([]string{"Apple", user.ID}, sep)
		case providerTypeGoogle:
			opt = strings.Join([]string{"Google", user.ID}, sep)
		case providerTypeFacebook:
			opt = strings.Join([]string{"Facebook", user.ID}, sep)
		case providerTypeCustom:
			opt = strings.Join([]string{"Custom JWT", user.ID}, sep)
			// case providerTypeCustomFunction:
			// 	opt = strings.Join([]string{"Custom Function", user.ID}, sep)
		}
		selectableUserOptions[userOptIndex] = opt
		selectableUsers[opt] = user
		userOptIndex++
	}
	selectedUsers := []string{}
	err = ui.AskOne(
		&selectedUsers,
		&survey.MultiSelect{
			Message: "Which user(s) would you like to delete?",
			Options: selectableUserOptions,
		},
	)
	if err != nil {
		return nil, err
	}
	users := []realm.User{}
	for _, userOption := range selectedUsers {
		users = append(users, selectableUsers[userOption])
	}
	return users, nil
}
