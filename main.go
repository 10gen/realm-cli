// Realm is a tool for command-line administration of MongoDB Realm applications.
package main

import (
	"os"
	"path/filepath"

	"github.com/10gen/realm-cli/commands"
	"github.com/10gen/realm-cli/utils"
	"github.com/10gen/realm-cli/utils/telemetry"

	"github.com/mitchellh/cli"
)

func main() {
	c := cli.NewCLI(filepath.Base(os.Args[0]), utils.CLIVersion)
	c.Args = os.Args[1:]

	telemetryService := &telemetry.Service{}

	var ui cli.Ui = &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	c.Commands = map[string]cli.CommandFactory{
		"whoami":         commands.NewWhoamiCommandFactory(ui, telemetryService),
		"login":          commands.NewLoginCommandFactory(ui, telemetryService),
		"logout":         commands.NewLogoutCommandFactory(ui, telemetryService),
		"export":         commands.NewExportCommandFactory(ui, telemetryService),
		"import":         commands.NewImportCommandFactory(ui, telemetryService),
		"diff":           commands.NewDiffCommandFactory(ui, telemetryService),
		"secrets":        commands.NewSecretsCommandFactory(ui, telemetryService),
		"secrets list":   commands.NewSecretsListCommandFactory(ui, telemetryService),
		"secrets add":    commands.NewSecretsAddCommandFactory(ui, telemetryService),
		"secrets update": commands.NewSecretsUpdateCommandFactory(ui, telemetryService),
		"secrets remove": commands.NewSecretsRemoveCommandFactory(ui, telemetryService),
	}

	exitStatus, err := c.Run()
	if err != nil {
		ui.Error(err.Error())
		telemetryService.TrackEvent(telemetry.EventTypeCommandError, telemetry.EventData{
			Key:   telemetry.EventDataKeyError,
			Value: err,
		})
	}

	if exitStatus == 0 {
		telemetryService.TrackEvent(telemetry.EventTypeCommandEnd)
	}

	os.Exit(exitStatus)
}
