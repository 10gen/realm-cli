package app

import (
	"errors"

	"github.com/10gen/realm-cli/internal/cli/feedback"
)

func errProjectExists(path string) error {
	var suffix string
	if path != "" {
		suffix = " at " + path
	}
	return feedback.NewErr(errors.New("a project already exists"+suffix), feedback.ErrNoUsage{})
}
