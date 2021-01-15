package shared

import (
	"fmt"
	"strings"

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
func (i *UsersInputs) ResolveUsers(ui terminal.UI, client realm.Client, app realm.App) ([]realm.User, error) {
	var err error

	if (len(i.ProviderTypes) == 1 && i.ProviderTypes[0] == ProviderTypeInteractive) || i.State == UserStateTypeInteractive || i.Status == StatusTypeInteractive {
		// Interactive should have already occurred
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
	selectableUsers := map[string]realm.User{}
	selectableUserOptions := make([]string, len(selectableUsersSet))
	sep := " - "
	userOptIndex := 0
	for _, user := range selectableUsersSet {
		var opt string
		switch user.Identities[0].ProviderType {
		case ProviderTypeAnonymous:
			opt = strings.Join([]string{"Anonymous", user.ID}, sep)
		case ProviderTypeLocalUserPass:
			opt = strings.Join([]string{"User/Password", fmt.Sprint(user.Data[UserDataEmail]), user.ID}, sep)
		case ProviderTypeAPIKey:
			opt = strings.Join([]string{"ApiKey", fmt.Sprint(user.Data[UserDataName]), user.ID}, sep)
		case ProviderTypeApple:
			opt = strings.Join([]string{"Apple", user.ID}, sep)
		case ProviderTypeGoogle:
			opt = strings.Join([]string{"Google", user.ID}, sep)
		case ProviderTypeFacebook:
			opt = strings.Join([]string{"Facebook", user.ID}, sep)
		case ProviderTypeCustom:
			opt = strings.Join([]string{"Custom JWT", user.ID}, sep)
		case ProviderTypeCustomFunction:
			opt = strings.Join([]string{"Custom Function", user.ID}, sep)
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
