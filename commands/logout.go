package commands

import (
	"github.com/10gen/realm-cli/utils/telemetry"
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
	return lc.Synopsis() + `

OPTIONS:` +
		lc.BaseCommand.Help()
}

// Run executes the command
func (lc *LogoutCommand) Run(args []string) int {
	lc.service.TrackEvent(telemetry.EventTypeCommandStart)
	if err := lc.BaseCommand.run(args); err != nil {
		lc.UI.Error(err.Error())
		lc.service.TrackEvent(telemetry.EventTypeCommandError,
			telemetry.EventData{
				Key:   telemetry.EventDataKeyError,
				Value: err,
			})
		return 1
	}

	if err := lc.storage.Clear(); err != nil {
		lc.UI.Error(err.Error())
		lc.service.TrackEvent(telemetry.EventTypeCommandError,
			telemetry.EventData{
				Key:   telemetry.EventDataKeyError,
				Value: err,
			})
		return 1
	}
	lc.service.TrackEvent(telemetry.EventTypeCommandEnd)
	return 0
}
