package commands

import (
	"github.com/mitchellh/cli"
)

// NewLogoutCommandFactory returns a new cli.CommandFactory given a cli.Ui
func NewLogoutCommandFactory(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &LogoutCommand{
			BaseCommand: &BaseCommand{
				Name: "logout",
				UI:   ui,
			},
		}, nil
	}
}

// LogoutCommand deauthenticates a user and clears out their auth credentials from storage
type LogoutCommand struct {
	*BaseCommand
}

// Synopsis returns a one-liner description for this command
func (lc *LogoutCommand) Synopsis() string {
	return `Deauthenticate as an administrator.`
}

// Help returns long-form help information for this command
func (lc *LogoutCommand) Help() string {
	return lc.Synopsis()
}

// Run executes the command
func (lc *LogoutCommand) Run(args []string) int {
	if err := lc.BaseCommand.run(args); err != nil {
		lc.UI.Error(err.Error())
		return 1
	}

	if err := lc.storage.Clear(); err != nil {
		lc.UI.Error(err.Error())
		return 1
	}

	return 0
}
