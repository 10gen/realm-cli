package app

import (
	"errors"

	"github.com/10gen/realm-cli/internal/utils/cli"
)

func errProjectExists(path string) error {
	var suffix string
	if path != "" {
		suffix = " at " + path
	}
	return cli.NewErr(errors.New("a project already exists"+suffix), cli.ErrNoUsage{})
}
