package main

import "github.com/10gen/stitch-cli/commands"

func main() {
	commands.Executor{commands.Index}.Execute()
}
