package login

import (
	"fmt"
	"strings"

	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/AlecAivazis/survey/v2"
)

const (
	inputFieldPublicAPIKey  = "publicAPIKey"
	inputFieldPrivateAPIKey = "privateAPIKey"

	inputFieldUsername = "username"
	inputFieldPassword = "password"

	authTypeCloud = "cloud"
	authTypeLocal = "local"
)

type inputs struct {
	AuthType      string
	PublicAPIKey  string
	PrivateAPIKey string
	Username      string
	Password      string
}

func (i *inputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	u := profile.Credentials()

	var questions []*survey.Question

	switch i.AuthType {
	case authTypeCloud:
		questions = i.resolveCloudCredentials(u)
	case authTypeLocal:
		questions = i.resolveLocalCredentials(u)
	default:
		return fmt.Errorf(
			"unsupported login type, use one of [%s] instead",
			strings.Join([]string{authTypeCloud, authTypeLocal}, ", "),
		)
	}

	if len(questions) > 0 {
		if err := ui.Ask(i, questions...); err != nil {
			return err
		}
	}
	return nil
}

func (i *inputs) resolveCloudCredentials(u user.Credentials) []*survey.Question {
	var questions []*survey.Question

	if i.PublicAPIKey == "" {
		if u.PublicAPIKey == "" {
			questions = append(questions, &survey.Question{
				Name:   inputFieldPublicAPIKey,
				Prompt: &survey.Input{Message: "Public API Key", Default: u.PublicAPIKey},
			})
		} else {
			i.PublicAPIKey = u.PublicAPIKey
		}
	}

	if i.PrivateAPIKey == "" {
		if u.PrivateAPIKey == "" {
			questions = append(questions, &survey.Question{
				Name:   inputFieldPrivateAPIKey,
				Prompt: &survey.Password{Message: "Private API Key"},
			})
		} else {
			i.PrivateAPIKey = u.PrivateAPIKey
		}
	}

	return questions
}

func (i *inputs) resolveLocalCredentials(u user.Credentials) []*survey.Question {
	var questions []*survey.Question

	if i.Username == "" {
		if u.Username == "" {
			questions = append(questions, &survey.Question{
				Name:   inputFieldUsername,
				Prompt: &survey.Input{Message: "Username", Default: u.Username},
			})
		} else {
			i.Username = u.Username
		}
	}

	if i.Password == "" {
		if u.Password == "" {
			questions = append(questions, &survey.Question{
				Name:   inputFieldPassword,
				Prompt: &survey.Password{Message: "Password"},
			})
		} else {
			i.Password = u.Password
		}
	}

	return questions
}

func realmAuthType(authType string) string {
	switch authType {
	case authTypeCloud:
		return realm.AuthTypeCloud
	case authTypeLocal:
		return realm.AuthTypeLocal
	}
	return ""
}
