package commands

import (
	"errors"

	"github.com/10gen/stitch-cli/config"
	flag "github.com/ogier/pflag"
)

var (
	login Command = new(loginCommand)

	loginHelp = ""

	ErrApiKeyRequired = errors.New("An API key must be supplied to log in.")
	ErrInvalidApiKey  = errors.New("Invalid API key.")
)

type loginCommand struct {
	apiKey string
}

func (lc *loginCommand) Name() string {
	return ""
}

func (lc *loginCommand) Parse(f *flag.FlagSet, args []string) error {
	f.StringVar(&lc.apiKey, "api-key", "", "api key for a MongoDB Cloud account")

	f.Parse(args)
	if lc.apiKey == "" {
		return ErrApiKeyRequired
	}
	if !loginValidApiKey(lc.apiKey) {
		return ErrInvalidApiKey
	}
	return nil
}

func (lc *loginCommand) Help() string {
	return loginHelp
}

func (lc *loginCommand) Run() error {
	if config.LoggedIn() {
		return config.ErrAlreadyLoggedIn
	}
	token, err := loginWithApiKey(lc.apiKey)
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
