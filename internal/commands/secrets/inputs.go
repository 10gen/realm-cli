package secrets

// Flag names and usages across the secrets commands
const (
	flagName            = "name"
	flagNameShort       = "n"
	flagNameUsageCreate = "the name of the secret"
	flagNameUsageUpdate = "the new name of the secret"

	flagValue            = "value"
	flagValueShort       = "v"
	flagValueUsageCreate = "the value of the secret"
	flagValueUsageUpdate = "the new value of the secret"

	flagSecret            = "secret"
	flagSecretShort       = "s"
	flagSecretUsageUpdate = "the name or id of the secret to update"
	flagSecretUsageDelete = "the name or id of the secret to delete"
)
