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

func (c *Command) InitFlags() *flag.FlagSet {
	f := flag.NewFlagSet(c.Name, flag.ContinueOnError)
	f.SetInterspersed(true)
	f.Usage = func() {}
	f.BoolVar(&flagGlobalHelp, "help", false, "")
	f.BoolVar(&ui.ColorEnabled, "color", ui.ColorEnabled, "")
	f.StringVarP(&local.Config, "", "C", "", "")
	c.FlagSet = f
	return f
}

func (c *Command) Execute() error {
	if flagGlobalHelp {
		return flag.ErrHelp
	}
	return c.Run()
}
