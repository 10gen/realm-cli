package shared

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/AlecAivazis/survey/v2"
)

// UsersInputs are the filtering inputs for a user command
type UsersInputs struct {
	State         UserStateType
	ProviderTypes []string
	Status        StatusType
}

// ResolveUsers will use the provided Realm client to resolve the users specified by the realm.App through inputs
func (i *UsersInputs) ResolveUsers(ui terminal.UI, client realm.Client, app realm.App) ([]string, error) {
	var err error

	if (len(i.ProviderTypes) == 1 && i.ProviderTypes[0] == ProviderTypeInteractive) || i.State == UserStateTypeInteractive || i.Status == StatusTypeInteractive {
		// Interactive should have already occured
		i.ProviderTypes = []string{}
		i.State = UserStateTypeNil
		i.Status = StatusTypeNil
	}

	userFilter := realm.UserFilter{}
	if len(i.ProviderTypes) > 0 {
		userFilter.Providers = i.ProviderTypes
	}
	if i.State != UserStateTypeNil {
		userFilter.State = realm.UserState(i.State.String())
	}
	selectableUsersSet := map[string]realm.User{}
	if i.Status == StatusTypeConfirmed || i.Status == StatusTypeNil {
		foundUsers, err := client.FindUsers(app.GroupID, app.ID, userFilter)
		if err != nil {
			return nil, err
		}
		for _, user := range foundUsers {
			selectableUsersSet[user.ID] = user
		}
	}
	if i.Status == StatusTypePending || i.Status == StatusTypeNil {
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
	selectableUsers := map[string]string{}
	selectableUserOptions := make([]string, len(selectableUsersSet))
	userOptionPattern := "%s - %s"
	userOptIndex := 0
	for _, user := range selectableUsersSet {
		var opt string
		switch user.Identities[0].ProviderType {
		case ProviderTypeAnonymous:
			opt = fmt.Sprintf(userOptionPattern, "Anonymous", user.ID)
		case ProviderTypeLocalUserPass:
			opt = fmt.Sprintf(userOptionPattern, user.Data["email"], user.ID)
		case ProviderTypeAPIKey:
			opt = fmt.Sprintf(userOptionPattern, user.Data["name"], user.ID)
		case ProviderTypeApple:
			opt = fmt.Sprintf(userOptionPattern, "Apple", user.ID)
		case ProviderTypeGoogle:
			opt = fmt.Sprintf(userOptionPattern, "Google", user.ID)
		case ProviderTypeFacebook:
			opt = fmt.Sprintf(userOptionPattern, "Facebook", user.ID)
		case ProviderTypeCustom:
			opt = fmt.Sprintf(userOptionPattern, "Custom JWT", user.ID)
		case ProviderTypeCustomFunction:
			opt = fmt.Sprintf(userOptionPattern, "Custom Function", user.ID)
		}
		selectableUserOptions[userOptIndex] = opt
		selectableUsers[opt] = user.ID
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
	users := []string{}
	for _, userOption := range selectedUsers {
		users = append(users, selectableUsers[userOption])
	}
	return users, nil
}
