package secrets

// Flag names and usages across the secrets commands
const (
	flagName            = "name"
	flagNameShort       = "n"
	flagNameUsageCreate = "the name of the secret to add to your Realm App"
	flagNameUsageUpdate = "the new name for the secret"

	flagValue            = "value"
	flagValueShort       = "v"
	flagValueUsageCreate = "the value of the secret to add to your Realm App"
	flagValueUsageUpdate = "the new value for the secret"

	flagSecret            = "secret"
	flagSecretShort       = "s"
	flagSecretUsageUpdate = "ID or name of the secret to update"
	flagSecretUsageDelete = "set the list of secrets to delete by ID or Name"
)
