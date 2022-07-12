package login

import (
	"errors"
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

	apiKeysPage = "https://cloud.mongodb.com/go?l=https%3A%2F%2Fcloud.mongodb.com%2Fv2%2F%3Cproject%3E%23access%2FapiKeys"
)

type inputs struct {
	AuthType      string
	PublicAPIKey  string
	PrivateAPIKey string
	Username      string
	Password      string
	Browser       bool
}

func (i *inputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	u := profile.Credentials()
	credentialsProvided := (i.Username != "" && i.Password != "") || (i.PublicAPIKey != "" && i.PrivateAPIKey != "")

	if i.Browser && credentialsProvided {
		return errors.New("credentials will not be authenticated while using browser flag, please login with one or the other")
	}

	if (u == (user.Credentials{}) && !credentialsProvided) || i.Browser {
		if err := ui.OpenBrowser(apiKeysPage); err != nil {
			ui.Print(terminal.NewWarningLog("there was an issue opening your browser"))
		}
	}

	var questions []*survey.Question

	switch {
	case i.Browser:
		questions = []*survey.Question{
			{
				Name:   inputFieldPublicAPIKey,
				Prompt: &survey.Input{Message: "Public API Key"},
			},
			{
				Name:   inputFieldPrivateAPIKey,
				Prompt: &survey.Password{Message: "Private API Key"},
			},
		}
	case i.AuthType == authTypeCloud:
		questions = i.resolveCloudCredentials(u)
	case i.AuthType == authTypeLocal:
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
