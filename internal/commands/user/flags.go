package user

const (
	flagState      = "state"
	flagStateUsage = `select the state of users to list, available options: ["enabled", "disabled"]`

	flagPending      = "pending"
	flagPendingUsage = `include to show users with pending status`

	flagProvider      = "provider"
	flagProviderUsage = `set the provider types for which to filter the list of app users with, available options: ` +
		`["local-userpass", "api-key", "oauth2-facebook", "oauth2-google", "oauth2-apple", ` +
		`"anon-user", "custom-token", "custom-function"]`

	flagUser             = "user"
	flagUserShort        = "u"
	flagUserListUsage    = `set the user ids for which to filter the list of app users with`
	flagUserDeleteUsage  = `set the user ids for which to delete in the app`
	flagUserDisableUsage = `set the user ids for which to disable in the app`
	flagUserEnableUsage  = `set the user ids for which to enable in the app`
	flagUserRevokeUsage  = `set the user ids for which to revoke sessions from`
)
