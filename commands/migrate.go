package commands

import (
	flag "github.com/ogier/pflag"
)

var migrate = &Command{
	Run:  migrateRun,
	Name: "migrate",
	ShortUsage: `
USAGE:
    stitch migrate [-C <CONFIG>] [--help]
`,
	LongUsage: `Migrate to a new version of the configuration spec.

OPTIONS:
    -C, --local-config <CONFIG>
            Set the stitch config file. Defaults to looking for stitch.json
            recursively up from the current working directory.
`,
}

var (
	migrateFlagSet *flag.FlagSet
)

func init() {
	migrateFlagSet = migrate.initFlags()
}

func migrateRun() error {
	args := migrateFlagSet.Args()
	if len(args) > 0 {
		return errUnknownArg(args[0])
	}
	return errorf("migration to a new configuration spec is not yet supported")
}
