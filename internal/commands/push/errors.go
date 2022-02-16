package push

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli/feedback"
)

func errProjectInvalid(path string, pathExists bool) error {
	var cause error
	if !pathExists {
		cause = fmt.Errorf("directory '%s' does not exist", path)
	} else {
		cause = fmt.Errorf("directory '%s' is not a supported Realm app project", path)
	}

	return feedback.NewErr(cause, feedback.ErrNoUsage{})
}
