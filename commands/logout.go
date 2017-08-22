package commands

import (
	"github.com/10gen/stitch-cli/config"
	flag "github.com/ogier/pflag"
)

var logout = &Command{
	Run:  logoutRun,
	Name: "logout",
	ShortUsage: `
USAGE:
    stitch logout [--help]
`,
	LongUsage: `Deauthenticate as an administrator.
`,
}

var (
	logoutFlagSet *flag.FlagSet
)

func init() {
	logoutFlagSet = logout.initFlags()
}

func logoutRun() error {
	if len(logoutFlagSet.Args()) > 0 {
		return errUnknownArg(logoutFlagSet.Arg(0))
	}
	if !config.LoggedIn() {
		return config.ErrNotLoggedIn
	}
	return config.LogOut()
}
