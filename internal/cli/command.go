package cli

import (
	"fmt"
	"strings"

	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
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
// At any point should an error occur, command execution will terminate
// and the ensuing steps will not be run
type Command interface {
	Handler(profile *user.Profile, ui terminal.UI, clients Clients) error
}

// Clients are the CLI clients
type Clients struct {
	Realm        realm.Client
	Atlas        atlas.Client
	HostingAsset local.HostingAssetClient
}

// CommandFlags provides access for commands to register local flags
type CommandFlags interface {
	Flags() []flags.Flag
}

// CommandInputs returns the command inputs
type CommandInputs interface {
	Inputs() InputResolver
}

// InputResolver provides access for command inputs to resolve missing data
type InputResolver interface {
	Resolve(profile *user.Profile, ui terminal.UI) error
}

// CommandDefinition is a command's definition that the CommandFactory
// can build a *cobra.Command from
type CommandDefinition struct {
	CommandMeta

	// Command is the command's implementation
	// If present, this value is used to specify the cobra.Command execution phases
	Command Command

	// SubCommands are the command's sub commands
	// This array is iteratively added to this Cobra command via (cobra.Command).AddCommand
	SubCommands []CommandDefinition
}

// CommandMeta is the command metadata
type CommandMeta struct {
	// Use defines how the command is used
	// This value maps 1:1 to Cobra's `Use` property
	Use string

	// Display controls how the command is described in output
	// If left blank, the command's Use value will be used instead
	Display string

	// Aliases is the list of supported aliases for the command
	// This value maps 1:1 to Cobra's `Aliases` property
	Aliases []string

	// Description is the text shown in the 'help' output of the parent command
	Description string

	// HelpText is the text shown in the 'help' output of the actual command
	// right below the command's description
	HelpText string

	// Hidden controls if this command shows up in the list of available commands
	// This value maps 1:1 to Cobra's `Hidden` property
	Hidden bool
}

// CommandDisplay returns the command display with the provided flags
func CommandDisplay(cmd string, args []flags.Arg) string {
	sb := strings.Builder{}

	fmt.Fprintf(&sb, "%s %s", Name, cmd)

	for _, arg := range args {
		sb.WriteString(arg.String())
	}

	return sb.String()
}
