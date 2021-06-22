package login

import (
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/AlecAivazis/survey/v2"
)

const (
	inputFieldPublicAPIKey  = "publicAPIKey"
	inputFieldPrivateAPIKey = "privateAPIKey"
)

type inputs struct {
	PublicAPIKey  string
	PrivateAPIKey string
}

func (i *inputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	user := profile.Credentials()
	var questions []*survey.Question

	if i.PublicAPIKey == "" {
		if user.PublicAPIKey == "" {
			questions = append(questions, &survey.Question{
				Name:   inputFieldPublicAPIKey,
				Prompt: &survey.Input{Message: "Public API Key", Default: user.PublicAPIKey},
			})
		} else {
			i.PublicAPIKey = user.PublicAPIKey
		}
	}

	if i.PrivateAPIKey == "" {
		if user.PrivateAPIKey == "" {
			questions = append(questions, &survey.Question{
				Name:   inputFieldPrivateAPIKey,
				Prompt: &survey.Password{Message: "Private API Key"},
			})
		} else {
			i.PrivateAPIKey = user.PrivateAPIKey
		}
	}

	if len(questions) > 0 {
		if err := ui.Ask(i, questions...); err != nil {
			return err
		}
	}
	return nil
}
