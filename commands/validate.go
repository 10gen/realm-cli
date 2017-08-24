package commands

import (
	"github.com/10gen/stitch-cli/local"

	flag "github.com/ogier/pflag"
)

var validate = &Command{
	Run:  validateRun,
	Name: "validate",
	ShortUsage: `
USAGE:
    stitch validate [-C <CONFIG>] [--help]
`,
	LongUsage: `Validate the local app configuration.

OPTIONS:
    -C, --local-config <CONFIG>
            Set the stitch config file. Defaults to looking for stitch.json
            recursively up from the current working directory.
`,
}

var (
	validateFlagSet *flag.FlagSet
)

func init() {
	validateFlagSet = validate.initFlags()
}

func validateRun() error {
	args := validateFlagSet.Args()
	if len(args) > 0 {
		return errUnknownArg(args[0])
	}
	_, ok := local.GetApp()
	if !ok {
		return errorf("could not find local app config (use -C to use a file other than \"stitch.json\")")
	}
	return nil
}
