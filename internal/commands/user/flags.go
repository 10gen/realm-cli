package user

const (
	flagState      = "state"
	flagStateUsage = `Filter the Realm app's users by state (Default value: <none>; Allowed values: <none>, "enabled", "disabled")`

	flagPending      = "pending"
	flagPendingUsage = `View the Realm app's pending users`

	flagProvider      = "provider"
	flagProviderUsage = `Filter the Realm app's users by provider type (Default value: <none>; Allowed values: <none>, "anon-user", "api-key", "local-userpass", "oauth2-google", "oauth2-apple", "oauth2-facebook", "custom-token", "custom-function")`

	flagUser             = "user"
	flagUserShort        = "u"
	flagUserListUsage    = "Filter the Realm app's users by ID(s)"
	flagUserDeleteUsage  = "Specify the Realm app's users' ID(s) to delete"
	flagUserDisableUsage = "Specify the Realm app's users' ID(s) to disable"
	flagUserEnableUsage  = "Specify the Realm app's users' ID(s) to enable"
	flagUserRevokeUsage  = "Specify the Realm app's users' ID(s) to revoke sessions for"
)
