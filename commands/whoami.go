package commands

import (
	"fmt"

	"github.com/mitchellh/cli"
)

// NewWhoamiCommandFactory returns a new cli.CommandFactory given a cli.Ui
func NewWhoamiCommandFactory(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &WhoamiCommand{
			BaseCommand: &BaseCommand{
				Name: "whoami",
				UI:   ui,
			},
		}, nil
	}
}

// WhoamiCommand is used to print the name and API key of the current user
type WhoamiCommand struct {
	*BaseCommand
}

// Synopsis returns a one-liner description for this command
func (whoami *WhoamiCommand) Synopsis() string {
	return "Display Current User Info"
}

// Help returns long-form help information for this command
func (whoami *WhoamiCommand) Help() string {
	return `Print the name and API key associated with the current user.

OPTIONS:` + whoami.BaseCommand.Help()
}

// Run executes the command
func (whoami *WhoamiCommand) Run(args []string) int {
	if err := whoami.BaseCommand.run(args); err != nil {
		whoami.UI.Error(err.Error())
		return 1
	}

	user, err := whoami.User()
	if err != nil {
		whoami.UI.Error(err.Error())
		return 1
	}

	message := "no user info available"
	if publicAPIKey := user.PublicAPIKey; publicAPIKey != "" {
		message = fmt.Sprintf("%s [API Key: %s]", publicAPIKey, user.RedactedAPIKey())
	}

	whoami.UI.Info(message)
	return 0
}
