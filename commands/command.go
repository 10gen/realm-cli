package commands

import (
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

func (c *Command) InitFlags() *flag.FlagSet {
	f := flag.NewFlagSet(c.Name, flag.ContinueOnError)
	f.SetInterspersed(true)
	f.Usage = func() {}
	f.BoolVar(&flagGlobalHelp, "help", false, "")
	c.FlagSet = f
	return f
}

func (c *Command) Execute() error {
	if flagGlobalHelp {
		return flag.ErrHelp
	}
	return c.Run()
}
