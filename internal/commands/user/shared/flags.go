package shared

import (
	"fmt"
	"strings"
)

// Shared flag variables across command
const (
	FlagProvider      = "provider"
	FlagProviderShort = "t"
	FlagProviderUsage = `set the provider types for which to filter the list of app users with, available options: ["local-userpass", "api-key", "oauth2-facebook", "oauth2-google", "anon-user", "custom-token"]`

	FlagStateType      = "state"
	FlagStateTypeShort = "s"
	FlagStateTypeUsage = `select the state of users to fiilter by, available options: ["enabled", "disabled"]`

	FlagStatusType      = "pending"
	FlagStatusTypeShort = "p"
	FlagStatusTypeUsage = `select the users' status: ["confirmed", "pending"]`

	FlagInteractive      = "interactive-filter"
	FlagInteractiveShort = "x"
	FlagInteractiveUsage = "filter users interactively"
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

// IsValidProviderType checks string for valid Provider Types
func IsValidProviderType(pt string) bool {
	switch pt {
	case
		ProviderTypeLocalUserPass,
		ProviderTypeAPIKey,
		ProviderTypeFacebook,
		ProviderTypeGoogle,
		ProviderTypeAnonymous,
		ProviderTypeCustom,
		ProviderTypeApple,
		ProviderTypeCustomFunction:
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

	if !IsValidStatusType(newStatusType) {
		return ErrInvalidStatusType
	}

	*st = newStatusType
	return nil
}

// set of StatusType(s)
const (
	StatusTypeNil       StatusType = ""
	StatusTypeConfirmed StatusType = "confirmed"
	StatusTypePending   StatusType = "pending"
)

// StatusType error msg
var (
	ErrInvalidStatusType = func() error {
		allStatusTypes := []string{StatusTypeConfirmed.String(), StatusTypePending.String()}
		return fmt.Errorf("unsupported value, use one of [%s] instead", strings.Join(allStatusTypes, ", "))
	}()
)

// IsValidStatusType checks if a valid StatusType
func IsValidStatusType(st StatusType) bool {
	switch st {
	case
		StatusTypeNil, // allow StatusType to be optional
		StatusTypeConfirmed,
		StatusTypePending:
		return true
	}
	return false
}
