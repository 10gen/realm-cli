package create

import (
	"fmt"
	"strings"

	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/AlecAivazis/survey/v2/core"
)

const (
	flagUserType      = "type"
	flagUserTypeShort = "t"
	flagUserTypeUsage = `select the type of user to create, available options: ["api-key", "email"]`

	flagEmail      = "email"
	flagEmailShort = "u"
	flagEmailUsage = "sets the email of the user to be created"

	flagPassword      = "password"
	flagPasswordShort = "p"
	flagPasswordUsage = "sets the password of the user to be created"

	flagAPIKeyName      = "name"
	flagAPIKeyNameShort = "n"
	flagAPIKeyNameUsage = "sets the name of the api key to be created"
)

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
