package create

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
)

// input field names, per survey
const (
	inputFieldEmail      = "email"
	inputFieldPassword   = "password"
	inputFieldAPIKeyName = "apiKeyName"
)

type inputs struct {
	cli.ProjectAppInputs
	UserType   userType
	Email      string
	Password   string
	APIKeyName string
}

func (i *inputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	if err := i.ProjectAppInputs.Resolve(ui, profile.WorkingDirectory); err != nil {
		return err
	}

	if i.UserType == userTypeNil && i.APIKeyName == "" && i.Email == "" {
		err := ui.AskOne(
			&i.UserType,
			&survey.Select{
				Message: "Which auth provider type are you creating a user for?",
				Options: []string{userTypeAPIKey.String(), userTypeEmailPassword.String()},
			},
		)
		if err != nil {
			return err
		}
	} else if i.APIKeyName != "" {
		i.UserType = userTypeAPIKey
	} else if i.Email != "" {
		i.UserType = userTypeEmailPassword
	}

	var questions []*survey.Question

	switch i.UserType {
	case userTypeAPIKey:
		if i.APIKeyName == "" {
			questions = append(questions, &survey.Question{
				Name:   inputFieldAPIKeyName,
				Prompt: &survey.Input{Message: "API Key Name"},
			})
		}
	case userTypeEmailPassword:
		if i.Email == "" {
			questions = append(questions, &survey.Question{
				Name:   inputFieldEmail,
				Prompt: &survey.Input{Message: "New Email"},
			})
		}
		if i.Password == "" {
			questions = append(questions, &survey.Question{
				Name:   inputFieldPassword,
				Prompt: &survey.Password{Message: "New Password"},
			})
		}
	}

	if len(questions) > 0 {
		if err := ui.Ask(i, questions...); err != nil {
			return err
		}
	}

	return nil
}
