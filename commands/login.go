package commands

import (
	"github.com/10gen/stitch-cli/config"
	flag "github.com/ogier/pflag"
)

var login = &Command{
	Run:  loginRun,
	Name: "login",
	ShortUsage: `
USAGE:
    stitch login [--help] --api-key <TOKEN>
`,
	LongUsage: `Authenticate as an administrator.

OPTIONS:
    --api-key <TOKEN>
	        The API key for a MongoDB Cloud account.
`,
}

var (
	loginFlagSet *flag.FlagSet

	flagLoginApiKey string
)

var (
	ErrApiKeyRequired = errorf("an API key (--api-key=<TOKEN>) must be supplied to log in.")
)

func init() {
	loginFlagSet = login.InitFlags()
	loginFlagSet.StringVar(&flagLoginApiKey, "api-key", "", "")
}

func loginRun() error {
	if len(loginFlagSet.Args()) > 0 {
		return errUnknownArg(loginFlagSet.Arg(0))
	}

	apiKey := flagLoginApiKey
	if apiKey == "" {
		return ErrApiKeyRequired
	}
	return config.LogIn(apiKey)
}
