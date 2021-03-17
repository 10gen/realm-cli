package flags

import (
	"fmt"
)

// Arg is a flag arg represented by its name and optional value
type Arg struct {
	Name  string
	Value interface{}
}

func (a Arg) String() string {
	s := " --" + a.Name

	if a.Value == nil {
		return s
	}

	return fmt.Sprintf("%s %v", s, a.Value)
}
