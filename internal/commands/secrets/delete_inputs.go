package secrets

import "github.com/10gen/realm-cli/internal/cli"

const (
	flagSecret      = "secret"
	flagSecretShort = "s"
	flagSecretUsage = "set the list of secrets to delete by ID or Name"
)

type deleteInputs struct {
	cli.ProjectInputs
	secrets []string
}