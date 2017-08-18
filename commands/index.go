package commands

import (
	"fmt"

	"github.com/10gen/stitch-cli/config"
	flag "github.com/ogier/pflag"
)

// Index is the command for the root stitch command.
var Index = &Command{
	Name: "stitch",
	Run:  indexRun,
	ShortUsage: `
Usage: stitch [--version] [--help] [-C <PATH>] <COMMAND> [<ARGS>]
`,
	LongUsage: `Subcommands for various tasks:

your administration account:
   login      Authenticate as an administrator
   logout     Deauthenticate
   me         Show your user admin info
   groups     Show what groups you are a member of
   apps       Show what apps you can administrate
   clusters   Show what Atlas clusters you can access

manage your stitch applications:
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
		"login":   login,
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

	indexFlagSet = Index.InitFlags()
	indexFlagSet.BoolVar(&flagIndexVersion, "version", false, "")
	indexFlagSet.StringVarP(&config.Chdir, "", "C", "", "")
}

func indexRun() error {
	if len(indexFlagSet.Args()) > 0 {
		return Errorf("%q is not a stitch command.", indexFlagSet.Arg(0))
	}
	if flagIndexVersion {
		fmt.Println(Version)
		return nil
	}
	return flag.ErrHelp
}
