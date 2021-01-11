package cli

import (
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// Command is an executable CLI command
// This interface maps 1:1 to Cobra's Command.RunE phase
//
// Optionally, a Command may implement any of the other interfaces found below.
// The order of operations is:
//   1. CommandFlagger.Flags: use this hook to register flags to parse
//   2. CommandInputs.Resolve: use this hook to prompt for any flags not provided
//   3. CommandPreparer.Setup: use this hook to use setup the command (e.g. create clients/services)
//   4. Command.Handler: this is the command hook
//   5. CommandResponder.Feedback: use this hook to print feedback to the user after the command has executed
// At any point should an error occur, command execution will terminate
// and the ensuing steps will not be run
type Command interface {
	Handler(profile *Profile, ui terminal.UI) error
}

// CommandFlagger is a hook for commands to register local flags to be parsed
type CommandFlagger interface {
	Flags(fs *pflag.FlagSet)
}

// CommandInputs returns the command inputs
type CommandInputs interface {
	Inputs() InputResolver
}

// InputResolver is an input resolver
type InputResolver interface {
	Resolve(profile *Profile, ui terminal.UI) error
}

// CommandPreparer handles the command setup phase
// This interface maps 1:1 to Cobra's Command.PreRunE phase
type CommandPreparer interface {
	Setup(profile *Profile, ui terminal.UI) error
}

// CommandResponder handles the command feedback phase
// This interface maps 1:1 to Cobra's Command.PostRun phase
type CommandResponder interface {
	Feedback(profile *Profile, ui terminal.UI) error
}

// CommandDefinition is a command's definition that the CommandFactory
// can build a *cobra.Command from
type CommandDefinition struct {
	// Command is the command's implementation
	// If present, this value is used to specify the cobra.Command execution phases
	Command Command

	// SubCommands are the command's sub commands
	// This array is iteratively added to this Cobra command via (cobra.Command).AddCommand
	SubCommands []CommandDefinition

	// Description is the short command description shown in the 'help' output
	// This value maps 1:1 to Cobra's `Short` property
	Description string

	// Help is the long message shown in the 'help <this-command>' output
	// This value maps 1:1 to Cobra's `Long` property
	Help string

	// Use defines how the command is used
	// This value maps 1:1 to Cobra's `Use` property
	Use string

	// Display controls how the command is described in output
	// If left blank, the command's Use value will be used instead
	Display string

	// Aliases is the list of supported aliases for the command
	// This value maps 1:1 to Cobra's `Aliases` property
	Aliases []string
}
