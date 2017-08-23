// Package commands provides all commands associated with the CLI, and a means
// of executing them.
package commands

import (
	"github.com/10gen/stitch-cli/local"
	"github.com/10gen/stitch-cli/ui"
	flag "github.com/ogier/pflag"
)

var (
	flagGlobalHelp bool
)

// Command handles the parsing and execution of a command.
type Command struct {
	*flag.FlagSet

	Run func() error

	Name                  string
	ShortUsage, LongUsage string

	Subcommands map[string]*Command
}

// initFlags sets up all global flags and flag-related configuration to be
// associated with every command.
func (c *Command) initFlags() *flag.FlagSet {
	f := flag.NewFlagSet(c.Name, flag.ContinueOnError)
	f.SetInterspersed(true)
	f.Usage = func() {}
	f.BoolVarP(&flagGlobalHelp, "help", "h", false, "")
	f.BoolVar(&ui.ColorEnabled, "color", ui.ColorEnabled, "")
	f.BoolVarP(&ui.Yes, "yes", "y", false, "")
	f.StringVarP(&local.Config, "local-config", "C", "", "")
	c.FlagSet = f
	return f
}

// execute wraps the run method with a check for whether the help flag has been
// set.
func (c *Command) execute() error {
	if flagGlobalHelp {
		return flag.ErrHelp
	}
	return c.Run()
}
