package commands

import "fmt"

func errorf(format string, a ...interface{}) error {
	return fmt.Errorf("stitch: "+format, a...)
}

func errorUnknownArg(arg string) error {
	return errorf("unknown argument %q.", arg)
}

func errorNotInGroup(group string) error {
	return errorf("you are not a member of the group %q.", group)
}

func errorClusterNotExistsForGroup(group, cluster string) error {
	return errorf("the cluster %q does not exist for group %q", cluster, group)
}
