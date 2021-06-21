package secrets

// Flag names and usages across the secrets commands
const (
	flagName            = "name"
	flagNameShort       = "n"
	flagNameUsageCreate = "Name the secret"
	flagNameUsageUpdate = "Re-name the secret"

	flagValue            = "value"
	flagValueShort       = "v"
	flagValueUsageCreate = "Specify the secret value"
	flagValueUsageUpdate = "Specify the new secret value"

	flagSecret            = "secret"
	flagSecretShort       = "s"
	flagSecretUsageUpdate = "Specify the name or ID of the secret to update"
	flagSecretUsageDelete = "Speicfy the name or ID of the secret to delete"
)
