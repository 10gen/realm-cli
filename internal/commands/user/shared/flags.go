package shared

import (
	"fmt"
	"strings"

	"github.com/10gen/realm-cli/internal/utils/flags"
)

// Shared flag variables across command
const (
	FlagProvider      = "provider"
	FlagProviderShort = "t"
	FlagProviderUsage = `set the provider types for which to filter the list of app users with, available options: ` +
		`["local-userpass", "api-key", "oauth2-facebook", "oauth2-google", "oauth2-apple", ` +
		`"anon-user", "custom-token", "custom-function"], for interactive provide flag without value`

	FlagStateType      = "state"
	FlagStateTypeShort = "s"
	FlagStateTypeUsage = `select the state of users to fiilter by, available options: ` +
		`["enabled", "disabled"], for interactive provide flag without value`

	FlagStatusType      = "pending"
	FlagStatusTypeShort = "p"
	FlagStatusTypeUsage = `select the users' status, available options: ["confirmed", "pending"], ` +
		`for interactive provide flag without value`
)

// Provider Types to filter users by
const (
	ProviderTypeLocalUserPass  = "local-userpass"
	ProviderTypeAPIKey         = "api-key"
	ProviderTypeFacebook       = "oauth2-facebook"
	ProviderTypeGoogle         = "oauth2-google"
	ProviderTypeAnonymous      = "anon-user"
	ProviderTypeCustom         = "custom-token"
	ProviderTypeApple          = "oauth2-apple"
	ProviderTypeCustomFunction = "custom-function"
	ProviderTypeInteractive    = "interactive"
)

// All valid Provider Types to filter users by
var (
	ValidProviderTypes = []string{
		ProviderTypeLocalUserPass,
		ProviderTypeAPIKey,
		ProviderTypeFacebook,
		ProviderTypeGoogle,
		ProviderTypeAnonymous,
		ProviderTypeCustom,
		ProviderTypeApple,
		ProviderTypeCustomFunction,
	}
)

// isValidProviderType checks string for valid Provider Types
func isValidProviderType(pt string) bool {
	switch pt {
	case
		ProviderTypeLocalUserPass,
		ProviderTypeAPIKey,
		ProviderTypeFacebook,
		ProviderTypeGoogle,
		ProviderTypeAnonymous,
		ProviderTypeCustom,
		ProviderTypeApple,
		ProviderTypeCustomFunction,
		ProviderTypeInteractive: // allows the provider type to be provided as flag only
		return true
	}
	return false
}

// StatusType enumset
type StatusType string

// String returns the status type display
func (st StatusType) String() string { return string(st) }

// Type returns the StatusType type
func (st StatusType) Type() string { return "string" }

// Set validates and sets the status type value
func (st *StatusType) Set(val string) error {
	newStatusType := StatusType(val)

	if !isValidStatusType(newStatusType) {
		return errInvalidStatusType
	}

	*st = newStatusType
	return nil
}

// set of StatusType(s)
const (
	StatusTypeNil         StatusType = ""
	StatusTypeConfirmed   StatusType = "confirmed"
	StatusTypePending     StatusType = "pending"
	StatusTypeInteractive StatusType = "interactive"
)

// All valid StatusTypes to filter users by
var (
	ValidStatusTypes = []StatusType{
		StatusTypeConfirmed,
		StatusTypePending,
	}
)

// StatusType error msg
var (
	errInvalidStatusType = func() error {
		allStatusTypes := []string{StatusTypeConfirmed.String(), StatusTypePending.String()}
		return fmt.Errorf("unsupported value, use one of [%s] instead", strings.Join(allStatusTypes, ", "))
	}()
)

// isValidStatusType checks if a valid StatusType
func isValidStatusType(st StatusType) bool {
	switch st {
	case
		StatusTypeNil, // allow StatusType to be optional
		StatusTypeConfirmed,
		StatusTypePending,
		StatusTypeInteractive: // allow StatusType to be provided as flag only
		return true
	}
	return false
}

// UserStateType is a Realm application user state
type UserStateType string

// String returns the user state string
func (us UserStateType) String() string { return string(us) }

// Type returns the user state type
func (us UserStateType) Type() string { return flags.TypeString }

// Set validates and sets the user state value
func (us *UserStateType) Set(val string) error {
	newUserStateType := UserStateType(val)

	if !isValidUserStateType(newUserStateType) {
		return errInvalidUserStateType
	}

	*us = newUserStateType
	return nil
}

// set of supported user state values
const (
	UserStateTypeNil         UserStateType = ""
	UserStateTypeEnabled     UserStateType = "enabled"
	UserStateTypeDisabled    UserStateType = "disabled"
	UserStateTypeInteractive UserStateType = "interactive"
)

// All valid user states to filter users by
var (
	ValidUserStateTypes = []UserStateType{
		UserStateTypeEnabled,
		UserStateTypeDisabled,
	}
)

var (
	errInvalidUserStateType = func() error {
		allUserStateTypeTypes := []string{UserStateTypeEnabled.String(), UserStateTypeDisabled.String()}
		return fmt.Errorf("unsupported value, use one of [%s] instead", strings.Join(allUserStateTypeTypes, ", "))
	}()
)

func isValidUserStateType(us UserStateType) bool {
	switch us {
	case
		UserStateTypeNil, // allow state to be optional
		UserStateTypeEnabled,
		UserStateTypeDisabled,
		UserStateTypeInteractive: // allow UserStateType to be provided as flag only
		return true
	}
	return false
}
