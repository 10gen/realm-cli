package commands

import (
	"github.com/10gen/stitch-cli/config"
	flag "github.com/ogier/pflag"
)

var me = &Command{
	Run:  meRun,
	Name: "me",
	ShortUsage: `
USAGE:
    stitch me [--help] [<SPECIFIER>]
`,
	LongUsage: `Show your user info.

ARGS:
    <SPECIFIER>
            One of "name", "email", or "api-key" to get that particular field of information.
`,
}

var (
	meFlagSet *flag.FlagSet
)

func init() {
	meFlagSet = me.InitFlags()
}

func meRun() error {
	args := clustersFlagSet.Args()
	var specifier string
	if len(args) > 0 {
		switch args[0] {
		case "name", "email", "api-key":
			specifier = args[0]
		default:
			return errorf("invalid permissions %q.", args[0])
		}
		args = args[1:]
	}
	if len(args) > 0 {
		return errorUnknownArg(args[0])
	}

	if !config.LoggedIn() {
		return config.ErrNotLoggedIn
	}
	items := meInfo()
	if specifier != "" {
		for _, item := range items {
			if item.key == specifier {
				printSingleKV(item)
				break
			}
		}
	} else {
		printKV(items)
	}
	return nil
}

func meInfo() []kv {
	// TODO
	return []kv{
		{key: "name", value: "Charlie Programmer"},
		{key: "email", value: "charlie.programmer@example.com"},
		{key: "api-key", value: "ONETWOTHREEFOUR"},
	}
}
