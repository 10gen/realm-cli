// Stitch is a tool for command-line administration of MongoDB Stitch applications.
package main

import (
	"os"

	"github.com/10gen/stitch-cli/commands"

	"github.com/mattn/go-isatty"
	"github.com/mitchellh/cli"
)

func main() {
	c := cli.NewCLI("stitch", "0.0.1")
	c.Args = os.Args[1:]

	var ui cli.Ui = &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	if isatty.IsTerminal(os.Stdout.Fd()) {
		ui = &cli.ColoredUi{
			ErrorColor: cli.UiColorRed,
			WarnColor:  cli.UiColorYellow,
			Ui:         ui,
		}
	}

	c.Commands = map[string]cli.CommandFactory{
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
