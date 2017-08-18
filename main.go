package main

import (
	"os"

	"github.com/10gen/stitch-cli/commands"
)

func main() {
	commands.Executor{commands.Index}.Execute(os.Args[1:])
}
