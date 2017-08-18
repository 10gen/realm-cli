package commands

import "fmt"

func Errorf(format string, a ...interface{}) error {
	return fmt.Errorf("stitch: "+format, a...)
}

func ErrorUnknownArg(arg string) error {
	return Errorf("unknown argument %q.", arg)
}
