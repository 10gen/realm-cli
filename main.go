// Stitch is a tool for command-line administration of MongoDB Stitch applications.
package main

import (
	"os"
	"path/filepath"

	"github.com/10gen/stitch-cli/commands"

	"github.com/mitchellh/cli"
)

func main() {
	c := cli.NewCLI(filepath.Base(os.Args[0]), "0.0.1")
	c.Args = os.Args[1:]

	var ui cli.Ui = &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	c.Commands = map[string]cli.CommandFactory{
		"whoami": commands.NewWhoamiCommandFactory(ui),
		"login":  commands.NewLoginCommandFactory(ui),
		"logout": commands.NewLogoutCommandFactory(ui),
		"export": commands.NewExportCommandFactory(ui),
		"import": commands.NewImportCommandFactory(ui),
	}

	exitStatus, err := c.Run()
	if err != nil {
		ui.Error(err.Error())
	}

	os.Exit(exitStatus)
}
