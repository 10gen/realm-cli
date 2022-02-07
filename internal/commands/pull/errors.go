package pull

import (
	"errors"

	"github.com/10gen/realm-cli/internal/utils/cli"
)

var (
	errProjectNotFound = cli.NewErr(
		errors.New("must specify --remote or run command from inside a Realm app directory"),
		cli.ErrNoUsage{},
	)
)
