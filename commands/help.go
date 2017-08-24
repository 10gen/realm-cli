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
	helpFlagSet = help.initFlags()
	helpFlagSet.BoolVarP(&flagHelpGuide, "guide", "g", false, "")
}

const helpGuide = `stitch: the CLI for MongoDB Stitch.

Using the stitch CLI starts with logging in. Aquire your login token from the
web UI and log in on the CLI:

    $ stitch login --api-key=TOKEN

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

And you can see your apps, categorized by group:

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


The 'info' command is useful for getting information on a particular app. It
will check locally for a stitch.json app configuration, otherwise it uses the
supplied '--app' option. Additionally, any info subcommand can take the
'--json' flag and output JSON.

    $ stitch info --app platespace-prod-ffxys
    group:    	group-1
    name:     	platespace-prod
    id:       	598dca3bede4017c35942841
    client_id:	platespace-prod-ffxys
    clusters:
    	mongodb-atlas
    services:
    	MongoDB	mongodb-atlas
    	GitHub 	my-github-service
    	HTTP   	my-http-service
    	Slack  	my-slack-service
    	Slack  	my-other-slack-service
    pipelines:
    	my-pipe1
    	my-pipe2
    values:
    	s3bucket
    	admin-phone-number
    authentication:
    	anonymous
    	email
    	facebook
    	api-keys

Some subcommands of info can be very useful for building scripts:

    $ stitch info client-id
    platespace-prod-ffxys

    $ stitch info clusters mongodb-atlas
    mongodb://199.7.91.13:27017/?ssl=true

A new stitch app is created using the 'create' command, in whicn two cases may
occur:
  1. a local config ("stitch.json" or supplied by -i/--input) is found
  2. otherwise, we create an app with defaults and clone it.

    $ stitch create --input platespace_config.json
    successfully created app platespace-demo.
    write updated config to "platespace_config.json"? [y/n]

To export configurations for local management, use the 'clone' command with an
app's client id:

    $ stitch clone -o blog_config.json blog-qwertt

To validate your local configuration, use the 'validate' command. This does
simple validation of the structure of the configuration, and does not guarantee
the config will successfully import.

    $ stitch validate

When you've made changes to a local config, you can compare with what exists on
stitch's servers using the 'diff' command:

    $ stitch diff
    * modified value admin-fav-num from "12" to "721"
    * created service "my-github-service"
    * deleted pipeline "my-old-pipeline"

To push changes to stitch's servers, use the 'sync' command. This can be used
with one of two strategies, 'replace' or 'merge'.

    $ stitch sync --strategy=replace
` // TODO: more details on sync strategies

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
