package ip_access

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/AlecAivazis/survey/v2"
)

const (
	createInputFieldAllowedIP        = "ip"
	createInputFieldAllowedIPComment = "comment"
)

type createInputs struct {
	cli.ProjectInputs
	IP      string
	Comment string
}

func (i *createInputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	var questions []*survey.Question

	if i.IP == "" {
		questions = append(questions, &survey.Question{
			Name:   createInputFieldAllowedIP,
			Prompt: &survey.Input{Message: "Allowed IP"},
		})
	}

	if i.Comment == "" {
		questions = append(questions, &survey.Question{
			Name:   createInputFieldAllowedIPComment,
			Prompt: &survey.Input{Message: "Comment"},
		})
	}

	if len(questions) > 0 {
		return ui.Ask(i, questions...)
	}
	return nil
}
