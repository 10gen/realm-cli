package user

import (
	"fmt"

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
			if len(foundUser) == 0 || err != nil {
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
	patternShort := "%s - %s"
	patternLong := "%s - %s - %s"
	userOptIndex := 0
	for _, user := range selectableUsersSet {
		var opt string
		if len(user.Identities) == 0 {
			opt = fmt.Sprintf(patternShort, "Unknown", user.ID)
		} else {
			switch user.Identities[0].ProviderType {
			case providerTypeAnonymous:
				opt = fmt.Sprintf(patternShort, "Anonymous", user.ID)
			case providerTypeLocalUserPass:
				opt = fmt.Sprintf(patternLong, "User/Password", fmt.Sprint(user.Data[userDataEmail]), user.ID)
			case providerTypeAPIKey:
				opt = fmt.Sprintf(patternLong, "ApiKey", fmt.Sprint(user.Data[userDataName]), user.ID)
			case providerTypeApple:
				opt = fmt.Sprintf(patternShort, "Apple", user.ID)
			case providerTypeGoogle:
				opt = fmt.Sprintf(patternShort, "Google", user.ID)
			case providerTypeFacebook:
				opt = fmt.Sprintf(patternShort, "Facebook", user.ID)
			case providerTypeCustom:
				opt = fmt.Sprintf(patternShort, "Custom JWT", user.ID)
			case providerTypeCustomFunction:
				opt = fmt.Sprintf(patternShort, "Custom Function", user.ID)
			default:
				opt = fmt.Sprintf(patternShort, "Unknown", user.ID)
			}
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
	users := make([]realm.User, len(selectedUsers))
	for userIndex, userOption := range selectedUsers {
		users[userIndex] = selectableUsers[userOption]
	}
	return users, nil
}
