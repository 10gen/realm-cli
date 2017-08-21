package commands

import (
	"github.com/10gen/stitch-cli/config"
	"github.com/10gen/stitch-cli/ui"
	flag "github.com/ogier/pflag"
)

var groups = &Command{
	Run:  groupsRun,
	Name: "groups",
	ShortUsage: `
USAGE:
    stitch groups [--help] [<PERM>]
`,
	LongUsage: `Show what groups you are a member of.

ARGS:
    <PERM>
			One of "rw" or "r" to view only groups for which you have exactly
			the given permissions.

OPTIONS:
    --hide-perms
            Hide the permissions associated with each group.
`,
}

var (
	groupsFlagSet *flag.FlagSet

	flagGroupsHidePerms bool
)

func init() {
	groupsFlagSet = groups.InitFlags()
	groupsFlagSet.BoolVar(&flagGroupsHidePerms, "hide-perms", false, "")
}

func groupsRun() error {
	args := groupsFlagSet.Args()
	var perms permissions = permissionsAny
	if len(args) > 0 {
		switch args[0] {
		case permissionsRW, permissionsR:
			perms = permissions(args[0])
		default:
			return errorf("invalid permissions %q.", args[0])
		}
		args = args[1:]
	}
	if len(args) > 0 {
		return errorUnknownArg(args[0])
	}
	if !config.LoggedIn() {
		return config.ErrNotLoggedIn
	}

	groups := groupsAll()
	if len(groups) == 0 {
		// Should this be an error? It's useful for the status code.
		return errorf("you are not a member of any groups.")
	}

	if perms != permissionsAny {
		filteredGroups := make([]permissionedItem, 0)
		for _, item := range groups {
			if item.perms == perms {
				filteredGroups = append(filteredGroups, item)
			}
		}
		groups = filteredGroups
	}

	for i, p := range groups {
		groups[i].name = ui.Color(ui.Group, p.name)
	}

	hidePerms := flagGroupsHidePerms || perms != permissionsAny
	printPermissionedItems(groups, hidePerms)
	return nil
}

func groupsAll() []permissionedItem {
	// TODO
	return []permissionedItem{
		{permissionsRW, "group-1"},
		{permissionsRW, "group-2"},
		{permissionsR, "group-3"},
	}
}
