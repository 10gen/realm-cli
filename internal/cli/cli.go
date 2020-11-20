package cli

import (
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Command is an executable CLI command
// This interface maps 1:1 to Cobra's Command.RunE phase
type Command interface {
	Handler(profile *Profile, ui terminal.UI, args []string) error
}

// CommandDefinition is a command's definition that the CommandFactory
// can build a *cobra.Command from
type CommandDefinition struct {
	Command Command

	// Description is the short description shown in the command's help text
	// This value maps 1:1 to Cobra's `short` property
	Description string

	// Usage is the one-line usage message
	// This value maps 1:1 to Cobra's `use` property
	Usage string
}

// CommandFactory is a command factory
type CommandFactory struct {
	Profile *Profile
	Config  Config
}

// Build builds a Cobra command from the specified CommandDefinition
func (factory CommandFactory) Build(provider func() CommandDefinition) *cobra.Command {
	cmdDef := provider()

	cmd := cobra.Command{
		Use:   cmdDef.Usage,
		Short: cmdDef.Description,
		RunE: func(c *cobra.Command, a []string) error {
			return cmdDef.Command.Handler(
				factory.Profile,
				newCobraUI(factory.Config.UI, c),
				a,
			)
		},
	}

	if preparer, ok := cmdDef.Command.(CommandPreparer); ok {
		cmd.PreRunE = func(c *cobra.Command, a []string) error {
			return preparer.Setup(
				factory.Profile,
				newCobraUI(factory.Config.UI, c),
				factory.Config.Command,
			)
		}
	}

	if responder, ok := cmdDef.Command.(CommandResponder); ok {
		cmd.PostRunE = func(c *cobra.Command, a []string) error {
			err := responder.Feedback(
				factory.Profile,
				newCobraUI(factory.Config.UI, c),
			)
			if err != nil {
				return err // TODO(REALMC-7340): handle this error gracefully
			}
			return nil
		}
	}

	if flagger, ok := cmdDef.Command.(CommandFlagger); ok {
		flagger.RegisterFlags(cmd.Flags())
	}

	return &cmd
}

// CommandPreparer handles the command setup phase
// This interface maps 1:1 to Cobra's Command.PreRunE phase
type CommandPreparer interface {
	Setup(profile *Profile, ui terminal.UI, config CommandConfig) error
}

// CommandResponder handles the command feedback phase
// This interface maps 1:1 to Cobra's Command.PostRun phase
type CommandResponder interface {
	Feedback(profile *Profile, ui terminal.UI) error
}

// CommandFlagger is a hook for commands to register local flags to be parsed
type CommandFlagger interface {
	RegisterFlags(fs *pflag.FlagSet)
}

// Config is the global CLI config
type Config struct {
	Command CommandConfig
	UI      terminal.UIConfig
}

// CommandConfig holds the global config for a CLI command
type CommandConfig struct {
	RealmBaseURL string
}

func newCobraUI(config terminal.UIConfig, cmd *cobra.Command) terminal.UI {
	return terminal.NewUI(config, cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr())
}
