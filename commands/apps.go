package commands

import (
	"github.com/10gen/stitch-cli/config"
	"github.com/10gen/stitch-cli/ui"
	flag "github.com/ogier/pflag"
)

var apps = &Command{
	Run:  appsRun,
	Name: "apps",
	ShortUsage: `
USAGE:
    stitch apps [--help] [--hide-perms] [<GROUP>]
`,
	LongUsage: `Show what apps you can administrate.

ARGS:
    <GROUP>
            Show only apps from the given MongoDB Cloud group.

OPTIONS:
    --hide-perms
            Hide the permissions associated with each app.
`,
}

var (
	appsFlagSet *flag.FlagSet

	flagAppsHidePerms bool
)

func init() {
	appsFlagSet = apps.InitFlags()
	appsFlagSet.BoolVar(&flagAppsHidePerms, "hide-perms", false, "")
}

func appsRun() error {
	args := appsFlagSet.Args()
	if len(args) > 1 {
		return errUnknownArg(args[0])
	}
	if !config.LoggedIn() {
		return config.ErrNotLoggedIn
	}

	if len(args) == 1 {
		group := args[0]
		if !validGroup(group) {
			return errorf("invalid group %q.", group)
		}

		groupApps, err := appsForGroup(group)
		if err != nil {
			return err
		}

		for i := range groupApps {
			groupApps[i].name = ui.Color(ui.AppClientID, groupApps[i].name)
		}

		printPermissionedItems(groupApps, flagAppsHidePerms)
		return nil
	}

	apps := appsAll()
	if len(apps) == 0 {
		return errorf("you are not administrating any apps.")
	}

	coloredApps := make(map[string][]permissionedItem, len(apps))
	for g, items := range apps {
		for i := range items {
			items[i].name = ui.Color(ui.AppClientID, items[i].name)
		}
		coloredApps[ui.Color(ui.Group, g)] = items
	}

	printGroupedPermissionedItems(coloredApps, flagAppsHidePerms)
	return nil
}

func appsForGroup(group string) ([]permissionedItem, error) {
	// TODO
	switch group {
	case "group-1":
		return []permissionedItem{
			{permissionsRW, "platespace-prod-ffxys"},
			{permissionsRW, "platespace-stg-asdfu"},
		}, nil
	case "group-2":
		return []permissionedItem{
			{permissionsRW, "todoapp-fooba"},
		}, nil
	case "group-3":
		return []permissionedItem{
			{permissionsR, "blog-qwertt"},
		}, nil
	}
	return nil, errorNotInGroup(group)
}

func appsAll() map[string][]permissionedItem {
	// TODO
	return map[string][]permissionedItem{
		"group-1": {
			{permissionsRW, "platespace-prod-ffxys"},
			{permissionsRW, "platespace-stg-asdfu"},
		},
		"group-2": {
			{permissionsRW, "todoapp-fooba"},
		},
		"group-3": {
			{permissionsR, "blog-qwertt"},
		},
	}
}
