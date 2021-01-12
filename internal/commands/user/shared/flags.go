package shared

// Shared flag variables across command
const (
	FlagProvider      = "provider"
	FlagProviderShort = "t"
	FlagProviderUsage = `set the provider types for which to filter the list of app users with, available options: ["local-userpass", "api-key", "oauth2-facebook", "oauth2-google", "anon-user", "custom-token"]`

	FlagStateType      = "state"
	FlagStateTypeShort = "s"
	FlagStateTypeUsage = `select the state of users to fiilter by, available options: ["enabled", "disabled"]`
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

// Checks string for valid Provider Types
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
