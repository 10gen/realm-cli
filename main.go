// Stitch is a tool for command-line administration of MongoDB Stitch applications.
package main

import (
	"os"

	"github.com/10gen/stitch-cli/commands"
)

func main() {
	err := commands.Executor{commands.Index}.Execute(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
}
