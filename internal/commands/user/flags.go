package user

import (
	"fmt"
	"strings"

	"github.com/10gen/realm-cli/internal/utils/flags"
)

const (
	flagState      = "state"
	flagStateShort = "s"
	flagStateUsage = `select the state of users to list, available options: ["enabled", "disabled"]`

	flagPending      = "pending"
	flagPendingShort = "p"
	flagPendingUsage = `select to show users with pending status`

	flagStatus      = "status"
	flagStatusShort = "p"
	flagStatusUsage = `select the state of users to fiilter by, available options: ["enabled", "disabled"]`

	flagProvider      = "provider"
	flagProviderShort = "t"
	flagProviderUsage = `set the provider types for which to filter the list of app users with, available options: ` +
		`["local-userpass", "api-key", "oauth2-facebook", "oauth2-google", "oauth2-apple", ` +
		`"anon-user", "custom-token", "custom-function"]`

	flagUser            = "user"
	flagUserShort       = "u"
	flagUserUsage       = `set the user ids for which to filter the list of app users with`
	flagUserDeleteUsage = `set the user ids for which to delete from the app`
)

const (
	providerTypeLocalUserPass  = "local-userpass"
	providerTypeAPIKey         = "api-key"
	providerTypeFacebook       = "oauth2-facebook"
	providerTypeGoogle         = "oauth2-google"
	providerTypeAnonymous      = "anon-user"
	providerTypeCustom         = "custom-token"
	providerTypeApple          = "oauth2-apple"
	providerTypeCustomFunction = "custom-function"
)

var (
	validProviderTypes = []string{
		providerTypeLocalUserPass,
		providerTypeAPIKey,
		providerTypeFacebook,
		providerTypeGoogle,
		providerTypeAnonymous,
		providerTypeCustom,
		providerTypeApple,
		providerTypeCustomFunction,
	}
)

// statusType enumset
type statusType string

// String returns the status type display
func (st statusType) String() string { return string(st) }

// Type returns the statusType type
func (st statusType) Type() string { return "string" }

// Set validates and sets the status type value
func (st *statusType) Set(val string) error {
	newStatusType := statusType(val)

	if !isValidStatusType(newStatusType) {
		return errInvalidStatusType
	}

	*st = newStatusType
	return nil
}

// set of statusType(s)
const (
	statusTypeEmpty     statusType = ""
	statusTypeConfirmed statusType = "confirmed"
	statusTypePending   statusType = "pending"
)

// statusType error msg
var (
	errInvalidStatusType = func() error {
		allStatusTypes := []string{statusTypeConfirmed.String(), statusTypePending.String()}
		return fmt.Errorf("unsupported value, use one of [%s] instead", strings.Join(allStatusTypes, ", "))
	}()
)

// isValidStatusType checks if a valid statusType
func isValidStatusType(st statusType) bool {
	switch st {
	case
		statusTypeEmpty, // allow statusType to be optional
		statusTypeConfirmed,
		statusTypePending:
		return true
	}
	return false
}

// userStateType is a Realm application user state
type userStateType string

// String returns the user state string
func (us userStateType) String() string { return string(us) }

// Type returns the user state type
func (us userStateType) Type() string { return flags.TypeString }

// Set validates and sets the user state value
func (us *userStateType) Set(val string) error {
	newUserStateType := userStateType(val)

	if !isValidUserStateType(newUserStateType) {
		return errInvalidUserStateType
	}

	*us = newUserStateType
	return nil
}

// set of supported user state values
const (
	userStateTypeEmpty    userStateType = ""
	userStateTypeEnabled  userStateType = "enabled"
	userStateTypeDisabled userStateType = "disabled"
)

var (
	errInvalidUserStateType = func() error {
		allUserStateTypeTypes := []string{userStateTypeEnabled.String(), userStateTypeDisabled.String()}
		return fmt.Errorf("unsupported value, use one of [%s] instead", strings.Join(allUserStateTypeTypes, ", "))
	}()
)

func isValidUserStateType(us userStateType) bool {
	switch us {
	case
		userStateTypeEmpty, // allow state to be optional
		userStateTypeEnabled,
		userStateTypeDisabled:
		return true
	}
	return false
}
