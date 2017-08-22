package commands

import (
	"fmt"

	"github.com/10gen/stitch-cli/config"
	flag "github.com/ogier/pflag"
)

var me = &Command{
	Run:  meRun,
	Name: "me",
	ShortUsage: `
USAGE:
    stitch me [--help] [--fetch] [<SPECIFIER>]
`,
	LongUsage: `Show your user info.

ARGS:
    <SPECIFIER>
            One of "name", "email", or "api-key" to get that particular field of information.

OPTIONS:
    --fetch
            Retrieve and update user info according to stitch servers.
`,
}

var (
	meFlagSet *flag.FlagSet

	flagMeFetch bool
)

func init() {
	meFlagSet = me.initFlags()
	meFlagSet.BoolVar(&flagMeFetch, "fetch", false, "")
}

func meRun() error {
	args := clustersFlagSet.Args()
	var specifier string
	if len(args) > 0 {
		specifier = args[0]
		args = args[1:]
	}
	if len(args) > 0 {
		return errUnknownArg(args[0])
	}

	if !config.LoggedIn() {
		return config.ErrNotLoggedIn
	}
	if flagMeFetch {
		err := config.Fetch()
		if err != nil {
			return err
		}
	}

	user := config.User()
	if specifier == "" {
		items := []kv{
			{key: "name", value: user.Name},
			{key: "email", value: user.Email},
			{key: "api-key", value: user.APIKey},
		}
		printKV(items)
		return nil
	}
	switch specifier {
	case "name":
		fmt.Println(user.Name)
	case "email":
		fmt.Println(user.Email)
	case "api-key":
		fmt.Println(user.APIKey)
	default:
		return errUnknownArg(specifier)
	}
	return nil
}
