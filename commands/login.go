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
	ErrApiKeyRequired = errorf("an API key must be supplied to log in.")
	ErrInvalidApiKey  = errorf("invalid API key.")
)

func init() {
	loginFlagSet = login.InitFlags()
	loginFlagSet.StringVar(&flagLoginApiKey, "api-key", "", "")
}

func loginRun() error {
	if len(loginFlagSet.Args()) > 0 {
		return errorUnknownArg(loginFlagSet.Arg(0))
	}

	apiKey := flagLoginApiKey
	if apiKey == "" {
		return ErrApiKeyRequired
	}
	if !loginValidApiKey(apiKey) {
		return ErrInvalidApiKey
	}

	if config.LoggedIn() {
		return config.ErrAlreadyLoggedIn
	}
	token, err := loginWithApiKey(apiKey)
	if err != nil {
		return err
	}
	return config.LogIn(token)
}

func loginValidApiKey(apiKey string) bool {
	return len(apiKey) == 8 // TODO
}

func loginWithApiKey(apiKey string) (token string, err error) {
	return // TODO
}
