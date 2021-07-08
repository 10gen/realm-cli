package user

import (
	"fmt"
	"strings"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
)

// input field names, per survey
const (
	inputCreateFieldEmail      = "email"
	inputCreateFieldPassword   = "password"
	inputCreateFieldAPIKeyName = "apiKeyName"
)

type createInputs struct {
	cli.ProjectInputs
	UserType   userType
	Email      string
	Password   string
	APIKeyName string
}

func (i *createInputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory, false); err != nil {
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
				Name:   inputCreateFieldAPIKeyName,
				Prompt: &survey.Input{Message: "API Key Name"},
			})
		}
	case userTypeEmailPassword:
		if i.Email == "" {
			questions = append(questions, &survey.Question{
				Name:   inputCreateFieldEmail,
				Prompt: &survey.Input{Message: "New Email"},
			})
		}
		if i.Password == "" {
			questions = append(questions, &survey.Question{
				Name:   inputCreateFieldPassword,
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

type userType string

// String returns the user type display
func (ut userType) String() string { return string(ut) }

// Type returns the userType type
func (ut userType) Type() string { return flags.TypeString }

// Set validates and sets the user type value
func (ut *userType) Set(val string) error {
	newUserType := userType(val)

	if !isValidUserType(newUserType) {
		return errInvalidUserType
	}

	*ut = newUserType
	return nil
}

// WriteAnswer validates and sets the user type value
func (ut *userType) WriteAnswer(name string, value interface{}) error {
	var newUserType userType

	switch v := value.(type) {
	case core.OptionAnswer:
		newUserType = userType(v.Value)
	}

	if !isValidUserType(newUserType) {
		return errInvalidUserType
	}
	*ut = newUserType
	return nil
}

const (
	userTypeNil           userType = ""
	userTypeAPIKey        userType = "api-key"
	userTypeEmailPassword userType = "email"
)

var (
	errInvalidUserType = func() error {
		allUserTypes := []string{userTypeAPIKey.String(), userTypeEmailPassword.String()}
		return fmt.Errorf("unsupported value, use one of [%s] instead", strings.Join(allUserTypes, ", "))
	}()
)

func isValidUserType(ut userType) bool {
	switch ut {
	case
		userTypeNil, // allow userType to be optional
		userTypeAPIKey,
		userTypeEmailPassword:
		return true
	}
	return false
}
