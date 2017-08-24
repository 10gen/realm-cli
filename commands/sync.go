package commands

import (
	"fmt"

	"github.com/10gen/stitch-cli/app"
	"github.com/10gen/stitch-cli/config"
	"github.com/10gen/stitch-cli/local"
	"github.com/10gen/stitch-cli/ui"

	flag "github.com/ogier/pflag"
)

var sync = &Command{
	Run:  syncRun,
	Name: "sync",
	ShortUsage: `
USAGE:
    stitch sync [-C <CONFIG>] [-s <STRATEGY>] [-y] [--help]
`,
	LongUsage: `Push changes made locally to a stitch app configuration.

OPTIONS:
    -C, --local-config <CONFIG>
            Set the stitch config file. Defaults to looking for stitch.json
            recursively up from the current working directory.

    -s, --strategy <STRATEGY>
            One of 'replace' or 'merge'. Replace will completely substitute
            the app configuration with the local config, except for secrets
            (which are not overwritten/deleted unless their containing
            entity is deleted). Merge has more complicated logic TODO.
            Default is 'merge'.

    -y      Suppress diff output and skip confirmation prompts.
`, // TODO: docs for --strategy=merge
}

var (
	syncFlagSet *flag.FlagSet

	flagSyncStrategy string
)

func init() {
	syncFlagSet = sync.initFlags()
	syncFlagSet.StringVarP(&flagSyncStrategy, "strategy", "s", "merge", "")
}

func syncRun() error {
	args := syncFlagSet.Args()
	if len(args) > 0 {
		return errUnknownArg(args[0])
	}
	switch flagSyncStrategy {
	case "merge", "replace":
	default:
		return errorf("invalid sync strategy %q", flagSyncStrategy)
	}
	if !config.LoggedIn() {
		return config.ErrNotLoggedIn
	}
	localApp, ok := local.GetApp()
	if !ok {
		return errorf("could not find local app config (use -C to use a file other than \"stitch.json\")")
	}

	// TODO: use admin SDK to first export the existing remote app.
	// return errorf("could not find app with group %q and ID %q", localApp.Group, localApp.ID)
	remoteApp := app.App{
		// must not change
		Group:    localApp.Group,
		ID:       localApp.ID,
		ClientID: localApp.ClientID,
	}

	if !ui.Yes {
		d := app.Diff(remoteApp, localApp)
		fmt.Println(d)
		if !ui.Ask("are you sure you want to push this config? (this is an irreversible operation)") {
			return nil
		}
	}

	// TODO: use admin SDK to attempt import with the given strategy.
	return nil
}
