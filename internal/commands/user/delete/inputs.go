package delete

import (
	"fmt"
	"strings"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/commands/user/shared"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
)

// input field names, per survey
const (
	inputFieldUser      = "user"
	inputFieldProviders = "provider"
	inputFieldState     = "state"
	inputFieldStatus    = "status"
)

type inputs struct {
	cli.ProjectAppInputs
	State             realm.UserState
	Users             []string
	ProviderTypes     []string
	Status            statusType
	InteractiveFilter bool
}

func (i *inputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	if err := i.ProjectAppInputs.Resolve(ui, profile.WorkingDirectory); err != nil {
		return err
	}

	var err error
	// Interactive set Status
	if i.InteractiveFilter {
		allUserStatuses := []string{statusTypeConfirmed.String(), statusTypePending.String()}
		selectedStatuses := []string{}
		defaultStatuses := []string{}
		if i.Status != statusTypeNil {
			newStatusType := statusType(i.Status)
			err = newStatusType.Set(i.Status.String())
			if err != nil {
				return err
			}
			defaultStatuses = append(defaultStatuses, i.Status.String())
		}
		err := ui.AskOne(
			&selectedStatuses,
			&survey.MultiSelect{
				Message: "Which user state would you like to filter confirmed users by? Selecting none is equivalent to selecting all.",
				Options: allUserStatuses,
				Default: defaultStatuses,
			},
		)
		if err != nil {
			return err
		}
		if len(selectedStatuses) == 1 {
			i.Status = statusType(selectedStatuses[0])
		} else {
			i.Status = statusTypeNil
		}
	}
	if i.Status == statusTypeConfirmed || i.Status == statusTypeNil {
		// Interactive set Providers
		if i.InteractiveFilter {
			err := ui.AskOne(
				&i.ProviderTypes,
				&survey.MultiSelect{
					Message: "Which provider(s) would you like to filter confirmed users by? Selecting none is equivalent to selecting all.",
					Options: shared.ValidProviderTypes,
					Default: i.ProviderTypes,
				},
			)
			if err != nil {
				return err
			}
		}

		// Interactive set State
		if i.InteractiveFilter {
			allUserStates := []string{realm.UserStateEnabled.String(), realm.UserStateDisabled.String()}
			selectedStates := []string{}
			defaultStates := []string{}
			if i.State != realm.UserStateNil {
				newStateType := realm.UserState(i.State)
				err = newStateType.Set(i.State.String())
				if err != nil {
					return err
				}
				defaultStates = append(defaultStates, i.State.String())
			}
			err := ui.AskOne(
				&selectedStates,
				&survey.MultiSelect{
					Message: "Which user state would you like to filter confirmed users by? Selecting none is equivalent to selecting all.",
					Options: allUserStates,
					Default: defaultStates,
				},
			)
			if err != nil {
				return err
			}
			if len(selectedStates) == 1 {
				i.State = realm.UserState(selectedStates[0])
			} else {
				i.State = realm.UserStateNil
			}
		}
	}
	return nil
}

func (i *inputs) ResolveUsers(ui terminal.UI, client realm.Client, app realm.App) error {
	if i.Users == nil || len(i.Users) < 1 {
		var err error

		userFilter := realm.UserFilter{}
		if i.Status != statusTypeNil {
			userFilter.Pending = i.Status == statusTypePending
		}
		if len(i.ProviderTypes) > 0 {
			userFilter.Providers = i.ProviderTypes
		}
		if i.State != realm.UserStateNil {
			userFilter.State = i.State
		}
		selectableUsersSet := make(map[string]realm.User)
		if i.Status == statusTypeConfirmed || i.Status == statusTypeNil {
			foundUsers, err := client.FindUsers(app.GroupID, app.ID, userFilter)
			if err != nil {
				return err
			}
			for _, user := range foundUsers {
				selectableUsersSet[user.ID] = user
			}
		}
		if i.Status == statusTypePending || i.Status == statusTypeNil {
			foundUsers, err := client.FindUsers(app.GroupID, app.ID, userFilter)
			if err != nil {
				return err
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
			case shared.ProviderTypeAnonymous:
				selectableUserOptions[x] = fmt.Sprintf(userOptionPattern, "Anonymous", user.ID)
			case shared.ProviderTypeLocalUserPass:
				selectableUserOptions[x] = fmt.Sprintf(userOptionPattern, user.Data["email"], user.ID)
			case shared.ProviderTypeAPIKey:
				selectableUserOptions[x] = fmt.Sprintf(userOptionPattern, user.Data["name"], user.ID)
			case shared.ProviderTypeApple:
				selectableUserOptions[x] = fmt.Sprintf(userOptionPattern, "Apple", user.ID)
			case shared.ProviderTypeGoogle:
				selectableUserOptions[x] = fmt.Sprintf(userOptionPattern, "Google", user.ID)
			case shared.ProviderTypeFacebook:
				selectableUserOptions[x] = fmt.Sprintf(userOptionPattern, "Facebook", user.ID)
			case shared.ProviderTypeCustom:
				selectableUserOptions[x] = fmt.Sprintf(userOptionPattern, "Custom JWT", user.ID)
			case shared.ProviderTypeCustomFunction:
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
			return err
		}
		for _, userPattern := range selectedUsers {
			i.Users = append(i.Users, strings.Split(userPattern, " - ")[1])
		}
	}
	return nil
}
