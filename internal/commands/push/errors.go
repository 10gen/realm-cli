package push

import (
	"fmt"
)

type errProjectInvalid struct {
	path       string
	pathExists bool
}

func (err errProjectInvalid) Error() string {
	if !err.pathExists {
		return fmt.Sprintf("directory '%s' does not exist", err.path)
	}
	return fmt.Sprintf("directory '%s' is not a supported Realm app project", err.path)
}

func (err errProjectInvalid) DisableUsage() struct{} { return struct{}{} }
