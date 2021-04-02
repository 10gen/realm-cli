// Realm is a tool for command-line administration of MongoDB Realm applications.
package main

import (
	"os"
	"path/filepath"

	"github.com/10gen/realm-cli/utils/telemetry"

	"github.com/10gen/realm-cli/commands"
	"github.com/10gen/realm-cli/utils"

	"github.com/mitchellh/cli"
)

func main() {
	c := cli.NewCLI(filepath.Base(os.Args[0]), utils.CLIVersion)
	c.Args = os.Args[1:]

	service := &telemetry.Service{}

	var ui cli.Ui = &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	c.Commands = map[string]cli.CommandFactory{
		"whoami":         commands.NewWhoamiCommandFactory(ui, service),
		"login":          commands.NewLoginCommandFactory(ui, service),
		"logout":         commands.NewLogoutCommandFactory(ui, service),
		"export":         commands.NewExportCommandFactory(ui, service),
		"import":         commands.NewImportCommandFactory(ui, service),
		"diff":           commands.NewDiffCommandFactory(ui, service),
		"secrets":        commands.NewSecretsCommandFactory(ui, service),
		"secrets list":   commands.NewSecretsListCommandFactory(ui, service),
		"secrets add":    commands.NewSecretsAddCommandFactory(ui, service),
		"secrets update": commands.NewSecretsUpdateCommandFactory(ui, service),
		"secrets remove": commands.NewSecretsRemoveCommandFactory(ui, service),
	}

	exitStatus, err := c.Run()
	if err != nil {
		ui.Error(err.Error())
		service.TrackEvent(telemetry.EventTypeCommandError, telemetry.EventData{
			Key:   telemetry.EventDataKeyError,
			Value: err,
		})
	}

	if exitStatus == 0 {
		service.TrackEvent(telemetry.EventTypeCommandEnd)
	}

	os.Exit(exitStatus)
}
