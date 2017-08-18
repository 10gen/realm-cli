package commands

import (
	"fmt"

	flag "github.com/ogier/pflag"
)

const Version = "0.0.1"

var version = &Command{
	Run:  versionRun,
	Name: "version",
	ShortUsage: `
Usage: stitch version [--help]
`,
	LongUsage: `Get the version of this CLI.`,
}

var (
	versionFlagSet *flag.FlagSet
)

func init() {
	versionFlagSet = version.InitFlags()
}

func versionRun() error {
	if len(versionFlagSet.Args()) > 0 {
		return ErrorUnknownArg(versionFlagSet.Arg(0))
	}
	fmt.Println(Version)
	return nil
}
