package commands

import "fmt"

func errorf(format string, a ...interface{}) error {
	return fmt.Errorf("stitch: "+format, a...)
}

func errUnknownArg(arg string) error {
	return errorf("unknown argument %q", arg)
}

func errNotInGroup(group string) error {
	return errorf("you are not a member of the group %q", group)
}

func errClusterNotExistsForGroup(group, cluster string) error {
	return errorf("the cluster %q does not exist for group %q", cluster, group)
}

func errAppNotFound(app string) error {
	return errorf("app %q not found, use `stitch apps` to see which apps you can administer", app)
}
