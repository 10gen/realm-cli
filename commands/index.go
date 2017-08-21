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
    stitch [--version] [--help] [-C <PATH>] <COMMAND> [<ARGS>]
`,
	LongUsage: `OPTIONS:
    -C <PATH>
            Run as if stitch was started in PATH instead of the current
            working directory.
    --color true|false
            Enable/disabled colored output. Defaults to coloring based on
            your environment.

SUBCOMMANDS:
for your account:
   login      Authenticate as an administrator
   logout     Deauthenticate
   me         Show your user admin info
   groups     Show what groups you are a member of
   apps       Show what apps you can administrate
   clusters   Show what Atlas clusters you can access
to manage your stitch applications:
   info       Show info about a particular app
   clone      Export a stitch app
   create     Create a new stitch app
   sync       Push changes made locally to a stitch app configuration
   diff       See the difference between the local app configuration and its remote version
   validate   Validate the local app configuration
   migrate    Migrate to a new version of the configuration spec
other subcommands:
   help       Show help for a command
   version    Show the version of this CLI
`,
	Subcommands: map[string]*Command{
		"login":    login,
		"logout":   logout,
		"me":       me,
		"groups":   groups,
		"apps":     apps,
		"clusters": clusters,
		"info":     info,
		"help":     help,
		"version":  version,
	},
}

var (
	indexFlagSet *flag.FlagSet
	indexPtr     *Command // to prevent initialization loops

	flagIndexVersion bool
)

func init() {
	indexPtr = Index

	indexFlagSet = Index.InitFlags()
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
