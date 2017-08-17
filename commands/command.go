package commands

import (
	"errors"
	"fmt"
	"strings"

	flag "github.com/ogier/pflag"
)

var (
	// ErrShowHelp is an error used to inform the caller that the help/usage should
	// be shown.
	ErrShowHelp = errors.New("")
)

// Command handles the parsing and execution of a command.
type Command interface {
	// Name is the name of the command.
	Name() string
	// Parse sets up flags and parses flags.
	Parse(f *flag.FlagSet, args []string) error
	// Help gives general help text to be supplemented by usage based on Parse.
	Help() string
	// Run executes the command. If the returned error is an ErrShowHelp, help/usage will be shown.
	Run() error
}

// SuperCommand is a Command that can hold subcommands.
type SuperCommand struct {
	Command
	SubCommands map[string]Command

	SubCommandHelp []struct {
		Name, Help string
	}

	active             Command
	activeIsSubCommand bool
}

func (supcmd *SuperCommand) Name() string {
	if supcmd.active == nil {
		return supcmd.Command.Name()
	}
	return supcmd.active.Name()
}

// Parse reads raw arguments into the command.
func (supcmd *SuperCommand) Parse(f *flag.FlagSet, args []string) error {
	supcmd.active = supcmd.Command
	if len(args) > 0 {
		for name, subcmd := range supcmd.SubCommands {
			if name == args[0] {
				supcmd.activeIsSubCommand = true
				supcmd.active = subcmd
				args = args[1:]
			}
		}
	}
	return supcmd.active.Parse(f, args)
}

// Help returns the help for the command and an overview of its subcommands.
func (supcmd *SuperCommand) Help() string {
	if supcmd.activeIsSubCommand {
		return supcmd.active.Help()
	}
	hsup := supcmd.Command.Help()
	if hsup != "" {
		hsup += "\n\n"
	}
	var hsub string
	if supcmd.SubCommandHelp != nil {
		h := subCommandUsageFormat(supcmd.SubCommandHelp)
		hsub = fmt.Sprintf("\nSubcommands:\n%s", h)
	} else if len(supcmd.SubCommands) > 0 {
		hsubs := []string{"", "Subcommands:"}
		for name := range supcmd.SubCommands {
			hsubs = append(hsubs, "  "+name)
		}
		hsub = strings.Join(hsubs, "\n")
	}
	return fmt.Sprintf("%s%s", hsup, hsub)
}

// Run executes the active command or subcommand.
// Parse must be called before Run.
func (supcmd *SuperCommand) Run() error {
	return supcmd.active.Run()
}

// SimpleCommand is a Command that does no parsing.
type SimpleCommand struct {
	// N is the command name.
	N string
	// H is the help text.
	H string
	// F is the function to run.
	F func(args []string) error

	args []string
}

// Name returns the command name.
func (sc *SimpleCommand) Name() string {
	return sc.N
}

// Parse stores the given arguments.
func (sc *SimpleCommand) Parse(f *flag.FlagSet, args []string) error {
	sc.args = args
	return nil
}

// Help returns the help text.
func (sc *SimpleCommand) Help() string {
	return sc.H
}

// Run calls the underlying function.
func (sc *SimpleCommand) Run() error {
	return sc.F(sc.args)
}

// EmptyCommand is a Command that does nothing.
type EmptyCommand struct{}

// Name returns the empty string.
func (ec EmptyCommand) Name() string {
	return ""
}

// Parse does nothing.
func (ec EmptyCommand) Parse(*flag.FlagSet, []string) error {
	return nil
}

// Help returns the empty string.
func (ec EmptyCommand) Help() string {
	return ""
}

// Run does nothing.
func (ec EmptyCommand) Run() error {
	return nil
}
