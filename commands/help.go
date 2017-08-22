package commands

import (
	"fmt"

	flag "github.com/ogier/pflag"
)

var help = &Command{
	Run:  helpRun,
	Name: "help",
	ShortUsage: `
USAGE:
    stitch help [--help] <COMMAND>
    stitch help -g|--guide
`,
	LongUsage: `Get help for usage of this CLI.
	
OPTIONS:
    -g, --guide
            Shows a guide to using this CLI.
`,
}

var (
	helpFlagSet *flag.FlagSet

	flagHelpGuide bool
)

func init() {
	helpFlagSet = help.InitFlags()
	helpFlagSet.BoolVarP(&flagHelpGuide, "guide", "g", false, "")
}

const helpGuide = `stitch: the CLI for MongoDB Stitch.

Using the stitch CLI starts with logging in. Aquire your login token from the
web UI and log in on the CLI:

    $ stitch login --api-key TOKEN

Once you've logged in, you can get some basic information about yourself:

    $ stitch me
    name: Charlie Programmer
    email: charlie.programmer@example.com

You can see which groups you are a part of and what permissions you have for
each group:

    $ stitch groups
    rw   group-1
    rw   group-2
    r    group-3

And you can see your apps:

    $ stitch apps
    group-1:
        rw   platespace-prod-ffxys
        rw   platespace-stg-asdfu
    group-2:
        rw   todoapp-fooba
    group-3:
        r    blog-qwertt

Or just the apps within a particular group:

    $ stitch apps group-1
    rw   platespace-prod-ffxys
    rw   platespace-stg-asdfu


The 'info' command is useful for getting information on a particular app:
`

func helpRun() error {
	// TODO: use pager if tty
	flagGlobalHelp = true
	args := helpFlagSet.Args()
	if len(args) > 0 {
		// get help page for subcommand
		cmd := indexPtr
		ok := true
		for ; len(args) > 0; args = args[1:] {
			cmd, ok = cmd.Subcommands[args[0]]
			if !ok {
				break
			}
		}
		if !ok {
			return errUnknownArg(args[0])
		}
		Executor{cmd}.Usage()
		return nil
	}
	if flagHelpGuide {
		fmt.Println(helpGuide)
		return nil
	}
	Executor{indexPtr}.Usage()
	return nil
}
