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
	State             realm.UserState
	ProviderTypes     []string
	Status            StatusType
	InteractiveFilter bool
}

// ResolveUsers will use the provided Realm client to resolve the users specified by the realm.App through inputs
func (i *UsersInputs) ResolveUsers(ui terminal.UI, client realm.Client, app realm.App) ([]string, error) {
	var err error

	userFilter := realm.UserFilter{}
	if i.Status != StatusTypeNil {
		userFilter.Pending = i.Status == StatusTypePending
	}
	if len(i.ProviderTypes) > 0 {
		userFilter.Providers = i.ProviderTypes
	}
	if i.State != realm.UserStateNil {
		userFilter.State = i.State
	}
	selectableUsersSet := make(map[string]realm.User)
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
		foundUsers, err := client.FindUsers(app.GroupID, app.ID, userFilter)
		if err != nil {
			return nil, err
		}
		for _, user := range foundUsers {
			selectableUsersSet[user.ID] = user
		}
	}

	// Interactive User Selection
	selectableUserOptions := make([]string, len(selectableUsersSet))
	userOptionPattern := "%s - %s"
	x := 0
	for _, user := range selectableUsersSet {
		switch user.Identities[0].ProviderType {
		case ProviderTypeAnonymous:
			selectableUserOptions[x] = fmt.Sprintf(userOptionPattern, "Anonymous", user.ID)
		case ProviderTypeLocalUserPass:
			selectableUserOptions[x] = fmt.Sprintf(userOptionPattern, user.Data["email"], user.ID)
		case ProviderTypeAPIKey:
			selectableUserOptions[x] = fmt.Sprintf(userOptionPattern, user.Data["name"], user.ID)
		case ProviderTypeApple:
			selectableUserOptions[x] = fmt.Sprintf(userOptionPattern, "Apple", user.ID)
		case ProviderTypeGoogle:
			selectableUserOptions[x] = fmt.Sprintf(userOptionPattern, "Google", user.ID)
		case ProviderTypeFacebook:
			selectableUserOptions[x] = fmt.Sprintf(userOptionPattern, "Facebook", user.ID)
		case ProviderTypeCustom:
			selectableUserOptions[x] = fmt.Sprintf(userOptionPattern, "Custom JWT", user.ID)
		case ProviderTypeCustomFunction:
			selectableUserOptions[x] = fmt.Sprintf(userOptionPattern, "Custom Function", user.ID)
		}
		x++
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
	users := make([]string, len(selectedUsers))
	for x, userPattern := range selectedUsers {
		users[x] = strings.Split(userPattern, " - ")[1]
	}

	return users, nil
}
