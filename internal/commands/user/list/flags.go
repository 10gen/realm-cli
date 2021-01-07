package list

import (
	"fmt"
	"strings"
)

const (
	flagState      = "state"
	flagStateShort = "s"
	flagStateUsage = `select the state of users to list, available options: ["enabled", "disabled"]`

	flagStatus      = "status-pending"
	flagStatusShort = "p"
	flagStatusUsage = `select the state of users to list, available options: ["enabled", "disabled"]`

	flagProviderTypes      = "provider"
	flagProviderTypesShort = "t"
	flagProviderTypesUsage = `todo add description`

	flagUsers      = "users"
	flagUsersShort = "u" //idk what you want this to be
	flagUsersUsage = `todo add description`
)

const (
	providerTypeLocalUserPass string = "local-userpass"
	providerTypeAPIKey        string = "api-key"
	providerTypeFacebook      string = "oauth2-facebook"
	providerTypeGoogle        string = "oauth2-google"
	providerTypeAnonymous     string = "anon-user"
	providerTypeCustom        string = "custom-token"
)

var (
	errInvalidProviderType = func() error {
		allProviderTypes := []string{
			providerTypeLocalUserPass,
			providerTypeAPIKey,
			providerTypeFacebook,
			providerTypeGoogle,
			providerTypeAnonymous,
			providerTypeCustom,
		}
		return fmt.Errorf("unsupported value, use one of [%s] instead", strings.Join(allProviderTypes, ", "))
	}()
)

func isValidProviderTypes(providers []string) bool {
	for _, provider := range providers {
		switch provider {
		case
			providerTypeLocalUserPass,
			providerTypeAPIKey,
			providerTypeFacebook,
			providerTypeGoogle,
			providerTypeAnonymous,
			providerTypeCustom:
			continue
		default:
			return false
		}
	}
	return true
}
