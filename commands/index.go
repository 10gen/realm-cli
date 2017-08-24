package commands

import (
	"fmt"

	flag "github.com/ogier/pflag"
)

// Index is the command for the root stitch command.
var Index = &Command{
	Name: "stitch",
	Run:  indexRun,
	ShortUsage: `
USAGE:
    stitch [--version] [-h|--help] [-C <CONFIG>] <COMMAND> [<ARGS>]
`,
	LongUsage: `OPTIONS:
    -C, --local-config <CONFIG>
            Set the stitch config file. Defaults to looking for stitch.json
            recursively up from the current working directory.
    --color true|false
            Enable/disabled colored output. Defaults to coloring based on
            your environment.

SUBCOMMANDS:
--- for your account:
   apps       Show what apps you can administrate
   clusters   Show what Atlas clusters you can access
   groups     Show what groups you are a member of
   login      Authenticate as an administrator
   logout     Deauthenticate
   me         Show your user admin info
--- to manage your stitch applications:
   clone      Export a stitch app
   create     Create a new stitch app
   diff       See the difference between the local app configuration and its remote version
   info       Show info about a particular app
   migrate    Migrate to a new version of the configuration spec
   sync       Push changes made locally to a stitch app configuration
   validate   Validate the local app configuration
--- other subcommands:
   help       Show help for a command
   version    Show the version of this CLI
`,
	Subcommands: map[string]*Command{
		// account-related
		"apps":     apps,
		"clusters": clusters,
		"groups":   groups,
		"login":    login,
		"logout":   logout,
		"me":       me,
		// app-related
		"clone":  clone,
		"create": create,
		"diff":   diff,
		"info":   info,
		// other
		"help":    help,
		"version": version,
	},
}

var (
	indexFlagSet *flag.FlagSet
	indexPtr     *Command // to prevent initialization loops

	flagIndexVersion bool
)

func init() {
	indexPtr = Index

	indexFlagSet = Index.initFlags()
	indexFlagSet.BoolVar(&flagIndexVersion, "version", false, "")
}

func indexRun() error {
	if len(indexFlagSet.Args()) > 0 {
		return errorf("%q is not a stitch command.", indexFlagSet.Arg(0))
	}
	if flagIndexVersion {
		fmt.Println(Version)
		return nil
	}
	return flag.ErrHelp
}
