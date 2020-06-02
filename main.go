// Realm is a tool for command-line administration of MongoDB Realm applications.
package main

import (
	"os"
	"path/filepath"

	"github.com/10gen/realm-cli/commands"
	"github.com/10gen/realm-cli/utils"

	"github.com/mitchellh/cli"
)

func main() {
	c := cli.NewCLI(filepath.Base(os.Args[0]), utils.CLIVersion)
	c.Args = os.Args[1:]

	var ui cli.Ui = &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	c.Commands = map[string]cli.CommandFactory{
		"whoami":         commands.NewWhoamiCommandFactory(ui),
		"login":          commands.NewLoginCommandFactory(ui),
		"logout":         commands.NewLogoutCommandFactory(ui),
		"export":         commands.NewExportCommandFactory(ui),
		"import":         commands.NewImportCommandFactory(ui),
		"diff":           commands.NewDiffCommandFactory(ui),
		"secrets":        commands.NewSecretsCommandFactory(ui),
		"secrets list":   commands.NewSecretsListCommandFactory(ui),
		"secrets add":    commands.NewSecretsAddCommandFactory(ui),
		"secrets update": commands.NewSecretsUpdateCommandFactory(ui),
		"secrets remove": commands.NewSecretsRemoveCommandFactory(ui),
	}

	exitStatus, err := c.Run()
	if err != nil {
		ui.Error(err.Error())
	}

	os.Exit(exitStatus)
}
