package commands

import (
	"github.com/10gen/stitch-cli/config"
	flag "github.com/ogier/pflag"
)

var clone = &Command{
	Run:  cloneRun,
	Name: "clone",
	ShortUsage: `
USAGE:
    stitch clone [--help] [-o <FILE>] <APP-CLIENT-ID>
`,
	LongUsage: `Export a stitch application to a local file.

ARGS:
    <APP-CLIENT-ID>
            Client ID of the stitch application to clone.

OPTIONS:
    -o, --output <FILE>
	        File to write the exported configuration. Defaults to "stitch.json".
`,
}

var (
	cloneFlagSet *flag.FlagSet

	flagCloneOutput string
)

func init() {
	cloneFlagSet = clone.initFlags()
	cloneFlagSet.StringVarP(&flagCloneOutput, "output", "o", "stitch.json", "")
}

func cloneRun() error {
	args := cloneFlagSet.Args()
	if len(args) == 0 {
		return errorf("missing <APP-CLIENT-ID> argument")
	}
	if len(args) > 1 {
		return errUnknownArg(args[1])
	}
	appClientID := args[0]

	if !config.LoggedIn() {
		return config.ErrNotLoggedIn
	}

	// TODO use appClientID and admin API to get exported app.
	return errAppNotFound(appClientID)
	// payload := []byte(`{"exported": "configuration"}`)
	// return ioutil.WriteFile(flagCloneOutput, payload, 0600)
}
