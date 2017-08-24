# MongoDB Stitch CLI (not yet ready for use)

Use the `mock` build tag to prevent actual calls to the atlas API, using
mock data instead.

TODO:
- finish implementing command logic (see TODO comments)
- vendor in dependencies
- fix `login` bug where `--api-key 12345678` fails but `--api-key=12345678` works
- support --json in more commands (currently just `info` command)
- shell completion
- integration tests

### Basic Usage

via `stitch help`:

```
USAGE:
    stitch [--version] [-h|--help] [-C <CONFIG>] <COMMAND> [<ARGS>]

OPTIONS:
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
```

### Guide

via `stitch help --guide`:

```
stitch: the CLI for MongoDB Stitch.

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
```
