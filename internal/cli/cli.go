package cli

import (
	"fmt"

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
}

// CommandFactory is a command factory
type CommandFactory struct {
	Profile *Profile
	Config  *Config
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

// Build builds a Cobra command from the specified CommandDefinition
func (factory CommandFactory) Build(provider func() CommandDefinition) *cobra.Command {
	cmdDef := provider()

	display := cmdDef.Display
	if display == "" {
		display = cmdDef.Use
	}

	cmd := cobra.Command{
		Use:   cmdDef.Use,
		Short: cmdDef.Description,
		Long:  cmdDef.Help,
		RunE: func(c *cobra.Command, a []string) error {
			ui := newUI(factory.Config.UI, c)

			err := cmdDef.Command.Handler(factory.Profile, ui, a)
			if err != nil {
				return ui.Print(terminal.NewErrorLog(
					fmt.Errorf("%s failed: %w", display, err),
				))
			}
			return nil
		},
	}

	if preparer, ok := cmdDef.Command.(CommandPreparer); ok {
		cmd.PreRunE = func(c *cobra.Command, a []string) error {
			ui := newUI(factory.Config.UI, c)

			err := preparer.Setup(factory.Profile, ui, factory.Config.Command)
			if err != nil {
				return ui.Print(terminal.NewErrorLog(
					fmt.Errorf("%s failed: an error occurred during initialization: %w", display, err),
				))
			}
			return nil
		}
	}

	if responder, ok := cmdDef.Command.(CommandResponder); ok {
		cmd.PostRunE = func(c *cobra.Command, a []string) error {
			ui := newUI(factory.Config.UI, c)

			err := responder.Feedback(factory.Profile, ui)
			if err != nil {
				return ui.Print(terminal.NewErrorLog(
					fmt.Errorf("%s completed successfully, but an error occurred while displaying results: %w", display, err),
				))
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

func newUI(config terminal.UIConfig, cmd *cobra.Command) terminal.UI {
	return terminal.NewUI(config, cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr())
}
