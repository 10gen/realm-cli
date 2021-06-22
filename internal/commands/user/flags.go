package user

import (
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

func providersFlag(value *[]string) flags.CustomFlag {
	return flags.NewStringSetFlag(
		value,
		flags.StringSetOptions{
			ValidValues: []string{
				string(realm.AuthProviderTypeUserPassword),
				string(realm.AuthProviderTypeAPIKey),
				string(realm.AuthProviderTypeFacebook),
				string(realm.AuthProviderTypeGoogle),
				string(realm.AuthProviderTypeAnonymous),
				string(realm.AuthProviderTypeCustomToken),
				string(realm.AuthProviderTypeApple),
				string(realm.AuthProviderTypeCustomFunction),
			},
			Meta: flags.Meta{
				Name: "provider",
				Usage: flags.Usage{
					Description: "Filter the Realm app's users by provider type",
				},
			},
		},
	)
}

func pendingFlag(value *bool) flags.BoolFlag {
	return flags.BoolFlag{
		Value: value,
		Meta: flags.Meta{
			Name: "pending",
			Usage: flags.Usage{
				Description: "View the Realm app's pending users",
			},
		},
	}
}

func stateFlag(value *realm.UserState) flags.CustomFlag { //nolint: interfacer
	return flags.CustomFlag{
		Value: value,
		Meta: flags.Meta{
			Name: "state",
			Usage: flags.Usage{
				Description:   "Filter the Realm app's users by state",
				DefaultValue:  "<none>",
				AllowedValues: []string{"enabled", "disabled"},
			},
		},
	}
}

func usersFlag(value *[]string, description string) flags.StringSliceFlag {
	return flags.StringSliceFlag{
		Value: value,
		Meta: flags.Meta{
			Name:      "user",
			Shorthand: "u",
			Usage: flags.Usage{
				Description: description,
			},
		},
	}
}
