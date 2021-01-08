package list

import (
	"fmt"
	"strings"
)

const (
	flagState      = "state"
	flagStateShort = "s"
	flagStateUsage = `select the state of users to list, available options: ["enabled", "disabled"]`

	flagPending      = "pending"
	flagPendingShort = "p"
	flagPendingUsage = `select to show users with pending status`

	flagProvider      = "provider"
	flagProviderShort = "t"
	flagProviderUsage = `set the provider types for which to filter the list of app users with, available options: ["local-userpass", "api-key", "oauth2-facebook", "oauth2-google", "anon-user", "custom-token"]`

	flagUser      = "user"
	flagUserShort = "u"
	flagUserUsage = `set the user ids for which to filter the list of app users with`
)

const (
	providerTypeLocalUserPass = "local-userpass"
	providerTypeAPIKey        = "api-key"
	providerTypeFacebook      = "oauth2-facebook"
	providerTypeGoogle        = "oauth2-google"
	providerTypeAnonymous     = "anon-user"
	providerTypeCustom        = "custom-token"
)

var (
	errInvalidProviderType = func() error {
		return fmt.Errorf(
			"unsupported value, use one of [%s] instead",
			strings.Join(
				[]string{
					providerTypeLocalUserPass,
					providerTypeAPIKey,
					providerTypeFacebook,
					providerTypeGoogle,
					providerTypeAnonymous,
					providerTypeCustom,
				},
				", ",
			),
		)
	}()
)

func isEachProviderTypeValid(providers []string) bool {
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
