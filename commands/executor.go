package commands

import (
	"fmt"
	"os"
	"strings"

	flag "github.com/ogier/pflag"
)

// Executor handles the logic of utilizing a Command.
type Executor struct {
	*Command
}

// Execute is the root entry point for a Command.
func (e Executor) Execute(args []string) error {
	if len(args) > 0 {
		if subcmd, ok := e.Subcommands[args[0]]; ok {
			return Executor{subcmd}.Execute(args[1:])
		}
	}
	err := e.Parse(args)
	if err != nil {
		// Parse already printed the error
		e.Usage()
		return err
	}
	err = e.Command.execute()
	if err == flag.ErrHelp {
		e.Usage()
		return nil
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
	return err
}

// Usage prints the usage associated with the command with varying brevity
// corresponding to whether help was requested.
func (e Executor) Usage() {
	var lines []string
	for _, line := range strings.Split(e.ShortUsage, "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	fmt.Fprintf(os.Stderr, "%s\n", strings.Join(lines, "\n"))
	if !flagGlobalHelp {
		return
	}
	if e.LongUsage != "" {
		fmt.Fprintf(os.Stderr, "\n%s\n", e.LongUsage)
	}
}
