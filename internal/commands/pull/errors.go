package pull

import (
	"errors"

	"github.com/10gen/realm-cli/internal/cli/feedback"
)

var (
	errProjectNotFound = feedback.NewErr(
		errors.New("must specify --remote or run command from inside a Realm app directory"),
		feedback.ErrNoUsage{},
	)
)
