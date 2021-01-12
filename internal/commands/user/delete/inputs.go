package delete

import (
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
	shared.UsersInputs
	Users []string
}

func (i *inputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	if err := i.ProjectAppInputs.Resolve(ui, profile.WorkingDirectory); err != nil {
		return err
	}

	var err error
	// Interactive set Status
	if i.InteractiveFilter {
		allUserStatuses := []string{shared.StatusTypeConfirmed.String(), shared.StatusTypePending.String()}
		selectedStatuses := []string{}
		defaultStatuses := []string{}
		if i.Status != shared.StatusTypeNil {
			newStatusType := shared.StatusType(i.Status)
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
			i.Status = shared.StatusType(selectedStatuses[0])
		} else {
			i.Status = shared.StatusTypeNil
		}
	}
	if i.Status == shared.StatusTypeConfirmed || i.Status == shared.StatusTypeNil {
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
