package delete

import (
	"github.com/10gen/realm-cli/internal/cli"
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
	// Interactive set Status
	if i.Status == shared.StatusTypeInteractive {
		allUserStatuses := []string{shared.StatusTypeConfirmed.String(), shared.StatusTypePending.String()}
		selectedStatuses := []string{}
		err := ui.AskOne(
			&selectedStatuses,
			&survey.MultiSelect{
				Message: "Which user state would you like to filter confirmed users by? Selecting none is equivalent to selecting all.",
				Options: allUserStatuses,
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
		if len(i.ProviderTypes) == 1 && i.ProviderTypes[0] == shared.ProviderTypeInteractive {
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
		if i.State == shared.UserStateTypeInteractive {
			allUserStates := []string{shared.UserStateTypeEnabled.String(), shared.UserStateTypeDisabled.String()}
			selectedStates := []string{}
			err := ui.AskOne(
				&selectedStates,
				&survey.MultiSelect{
					Message: "Which user state would you like to filter confirmed users by? Selecting none is equivalent to selecting all.",
					Options: allUserStates,
				},
			)
			if err != nil {
				return err
			}
			if len(selectedStates) == 1 {
				i.State = shared.UserStateType(selectedStates[0])
			} else {
				i.State = shared.UserStateTypeNil
			}
		}
	}
	return nil
}
