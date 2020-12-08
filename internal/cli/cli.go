package cli

import (
	"fmt"
	"log"
	"os"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/flags"
	"github.com/10gen/realm-cli/internal/telemetry"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

// Command is an executable CLI command
// This interface maps 1:1 to Cobra's Command.RunE phase
type Command interface {
	Handler(profile *Profile, ui terminal.UI, args []string) error
}

// CommandDefinition is a command's definition that the CommandFactory
// can build a *cobra.Command from
type CommandDefinition struct {
	Command

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
type CommandFactory interface {
	Build(func() CommandDefinition) *cobra.Command
	Close()
	Setup()
	Run(cmd *cobra.Command)
	SetGlobalFlags(fs *flag.FlagSet)
}

type commandFactory struct {
	config    *Config
	profile   *Profile
	ui        terminal.UI
	inReader  *os.File
	outWriter *os.File
	errWriter *os.File
	errLogger *log.Logger
	tracker   telemetry.Tracker
}

// Config is the global CLI config
type Config struct {
	CommandConfig
	terminal.UIConfig
}

// CommandConfig holds the global config for a CLI command
type CommandConfig struct {
	RealmBaseURL  string
	TelemetryMode telemetry.Mode
}

// NewCommandFactory creates a new command factory
func NewCommandFactory() CommandFactory {
	errLogger := log.New(os.Stderr, "UTC ERROR ", log.Ltime|log.Lmsgprefix)

	config := new(Config)

	profile, profileErr := NewDefaultProfile()
	if profileErr != nil {
		errLogger.Fatal(profileErr)
	}

	return &commandFactory{
		config:    config,
		profile:   profile,
		errLogger: errLogger,
	}
}

func (factory *commandFactory) configureTelemetry() error {
	telemetry.ConfigureEvents(factory.profile.GetUser().PublicAPIKey)
	telemetryMode := factory.config.TelemetryMode
	if telemetryMode == telemetry.ModeNil {
		telemetryMode = factory.profile.GetTelemetryMode()
	}
	factory.tracker = telemetry.NewTracker(telemetryMode)
	factory.profile.SetTelemetryMode(telemetryMode)
	return factory.profile.Save()
}

func (factory *commandFactory) Setup() {
	if err := factory.profile.Load(); err != nil {
		factory.errLogger.Fatal(err)
	}

	if err := factory.configureTelemetry(); err != nil {
		factory.errLogger.Fatal(err)
	}

	if filepath := factory.config.OutputTarget; filepath != "" {
		f, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
		if err != nil {
			factory.errLogger.Fatal(fmt.Errorf("failed to open target file: %w", err))
		}
		factory.outWriter = f
	}
}

func (factory *commandFactory) Close() {
	if factory.config.OutputTarget != "" {
		factory.outWriter.Close()
	}
}

type suppressUsageError struct {
	error
}

func (factory *commandFactory) Run(cmd *cobra.Command) {
	if err := cmd.Execute(); err != nil {
		if _, ok := err.(suppressUsageError); !ok {
			fmt.Println(cmd.UsageString())
		}

		if factory.ui == nil {
			factory.errLogger.Fatal(err)
		}

		if printErr := factory.ui.Print(terminal.NewErrorLog(err)); printErr != nil {
			factory.errLogger.Fatal(err) // log the original failure
		}

		os.Exit(1)
	}
}

// Build builds a Cobra command from the specified CommandDefinition
func (factory *commandFactory) Build(provider func() CommandDefinition) *cobra.Command {
	command := provider()

	display := command.Display
	if display == "" {
		display = command.Use
	}

	cmd := cobra.Command{
		Use:   command.Use,
		Short: command.Description,
		Long:  command.Help,
		RunE: func(c *cobra.Command, a []string) error {
			factory.tracker.Track(telemetry.NewCommandStartEvent(display))
			err := command.Handler(factory.profile, factory.ui, a)
			if err != nil {
				factory.tracker.Track(telemetry.NewCommandErrorEvent(display, err))
				return suppressUsageError{fmt.Errorf("%s failed: %w", display, err)}
			}
			factory.tracker.Track(telemetry.NewCommandCompleteEvent(display))
			return nil
		},
	}

	cmd.PersistentPreRun = func(c *cobra.Command, a []string) {
		factory.ensureUI()
		cmd.SetIn(factory.inReader)
		cmd.SetOut(factory.outWriter)
		cmd.SetErr(factory.errWriter)
	}

	if command, ok := command.Command.(CommandPreparer); ok {
		cmd.PreRunE = func(c *cobra.Command, a []string) error {
			err := command.Setup(factory.profile, factory.ui, factory.config.CommandConfig)
			if err != nil {
				return fmt.Errorf("%s setup failed: %w", display, err)
			}
			return nil
		}
	}

	if command, ok := command.Command.(CommandResponder); ok {
		cmd.PostRunE = func(c *cobra.Command, a []string) error {
			err := command.Feedback(factory.profile, factory.ui)
			if err != nil {
				return suppressUsageError{fmt.Errorf("%s completed, but displaying results failed: %w", display, err)}
			}
			return nil
		}
	}

	if command, ok := command.Command.(CommandFlagger); ok {
		command.RegisterFlags(cmd.Flags())
	}

	return &cmd
}

func (factory *commandFactory) SetGlobalFlags(fs *flag.FlagSet) {
	fs.StringVarP(&factory.profile.Name, flags.Profile, flags.ProfileShort, DefaultProfile, flags.ProfileUsage)
	fs.BoolVar(&factory.config.DisableColors, flags.DisableColors, false, flags.DisableColorsUsage)
	fs.VarP(&factory.config.OutputFormat, flags.OutputFormat, flags.OutputFormatShort, flags.OutputFormatUsage)
	fs.StringVarP(&factory.config.OutputTarget, flags.OutputTarget, flags.OutputTargetShort, "", flags.OutputTargetUsage)
	fs.StringVar(&factory.config.RealmBaseURL, flags.RealmBaseURL, realm.DefaultBaseURL, flags.RealmBaseURLUsage)
	fs.VarP(&factory.config.TelemetryMode, flags.TelemetryMode, flags.TelemetryModeShort, flags.TelemetryModeUsage)
}

func (factory *commandFactory) ensureUI() {
	if factory.inReader == nil {
		factory.inReader = os.Stdin
	}

	if factory.outWriter == nil {
		factory.outWriter = os.Stdout
	}

	if factory.errWriter == nil {
		if factory.config.OutputTarget != "" {
			factory.errWriter = factory.outWriter
		} else {
			factory.errWriter = os.Stderr
		}
	}

	if factory.ui == nil {
		factory.ui = terminal.NewUI(factory.config.UIConfig, factory.inReader, factory.outWriter, factory.errWriter)
	}
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
	RegisterFlags(fs *flag.FlagSet)
}
