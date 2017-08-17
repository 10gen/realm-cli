package commands

import (
	"fmt"
	"os"

	flag "github.com/ogier/pflag"
)

// Executor handles the logic of utilizing a Command.
type Executor struct {
	Command
}

func (e Executor) Execute() {
	args := os.Args[1:]
	f := flag.NewFlagSet(e.Name(), flag.ExitOnError)
	err := e.Parse(f, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		e.Usage(f)
		return
	}
	err = e.Run()
	if err == ErrShowHelp {
		e.Usage(f)
		return
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
	return
}

func (e Executor) Usage(f *flag.FlagSet) {
	name := e.Name()
	if name == "" {
		fmt.Fprintf(os.Stderr, "Usage:\n")
	} else {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", e.Name())
	}
	f.PrintDefaults()
	fmt.Fprintf(os.Stderr, e.Help())
}
