package cli

import (
	"fmt"
	"log"
	"os"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/telemetry"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CommandFactory is a command factory
type CommandFactory struct {
	config           *Config
	profile          *Profile
	ui               terminal.UI
	inReader         *os.File
	outWriter        *os.File
	errWriter        *os.File
	errLogger        *log.Logger
	telemetryService *telemetry.Service
}

// NewCommandFactory creates a new command factory
func NewCommandFactory() *CommandFactory {
	errLogger := log.New(os.Stderr, "UTC ERROR ", log.Ltime|log.Lmsgprefix)

	config := new(Config)

	profile, profileErr := NewDefaultProfile()
	if profileErr != nil {
		errLogger.Fatal(profileErr)
	}

	return &CommandFactory{
		config:    config,
		profile:   profile,
		errLogger: errLogger,
	}
}

// Build builds a Cobra command from the specified CommandDefinition
func (factory *CommandFactory) Build(command CommandDefinition) *cobra.Command {
	display := command.Display
	if display == "" {
		display = command.Use
	}

	cmd := cobra.Command{
		Use:     command.Use,
		Short:   command.Description,
		Long:    command.Help,
		Aliases: command.Aliases,
	}

	for _, subCommand := range command.SubCommands {
		cmd.AddCommand(factory.Build(subCommand))
	}

	if command.Command != nil {
		if command, ok := command.Command.(CommandFlagger); ok {
			command.Flag(cmd.Flags())
		}

		cmd.PersistentPreRun = func(c *cobra.Command, a []string) {
			factory.ensureUI()
			cmd.SetIn(factory.inReader)
			cmd.SetOut(factory.outWriter)
			cmd.SetErr(factory.errWriter)

			factory.ensureTelemetry(display)
		}

		if command, ok := command.Command.(CommandPreparer); ok {
			cmd.PreRunE = func(c *cobra.Command, a []string) error {
				err := command.Setup(factory.profile, factory.ui, factory.config.Context)
				if err != nil {
					return fmt.Errorf("%s setup failed: %w", display, err)
				}
				return nil
			}
		}

		cmd.RunE = func(c *cobra.Command, a []string) error {
			factory.telemetryService.TrackEvent(telemetry.EventTypeCommandStart)

			err := command.Command.Handler(factory.profile, factory.ui, a)
			if err != nil {
				factory.telemetryService.TrackEvent(
					telemetry.EventTypeCommandError,
					telemetry.EventData{Key: telemetry.EventDataKeyErr, Value: err},
				)
				return suppressUsageError{fmt.Errorf("%s failed: %w", display, err)}
			}

			factory.telemetryService.TrackEvent(telemetry.EventTypeCommandComplete)
			return nil
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
	}

	return &cmd
}

// Close closes the command factory
func (factory *CommandFactory) Close() {
	if factory.config.OutputTarget != "" {
		factory.outWriter.Close()
	}
}

// Run executes the command
func (factory *CommandFactory) Run(cmd *cobra.Command) {
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

// TODO(REALMC-7429): fill out the flag usages
const (
	flagAutoConfirm      = "yes"
	flagAutoConfirmShort = "y"
	flagAutoConfirmUsage = "set to automatically proceed through command confirmations"

	flagProfile      = "profile"
	flagProfileShort = "i"
	flagProfileUsage = "this is the --profile, -p usage"
)

// SetGlobalFlags sets the global flags
func (factory *CommandFactory) SetGlobalFlags(fs *pflag.FlagSet) {
	// cli profile
	fs.StringVarP(&factory.profile.Name, flagProfile, flagProfileShort, DefaultProfile, flagProfileUsage)

	// cli context
	fs.BoolVarP(&factory.config.AutoConfirm, flagAutoConfirm, flagAutoConfirmShort, false, flagAutoConfirmUsage)
	fs.StringVar(&factory.config.RealmBaseURL, realm.FlagBaseURL, realm.DefaultBaseURL, realm.FlagBaseURLUsage)
	fs.VarP(&factory.config.TelemetryMode, telemetry.FlagMode, telemetry.FlagModeShort, telemetry.FlagModeUsage)

	// cli ui
	fs.BoolVar(&factory.config.DisableColors, terminal.FlagDisableColors, false, terminal.FlagDisableColorsUsage)
	fs.VarP(&factory.config.OutputFormat, terminal.FlagOutputFormat, terminal.FlagOutputFormatShort, terminal.FlagOutputFormatUsage)
	fs.StringVarP(&factory.config.OutputTarget, terminal.FlagOutputTarget, terminal.FlagOutputTargetShort, "", terminal.FlagOutputTargetUsage)
}

// Setup initializes the command factory
func (factory *CommandFactory) Setup() {
	if err := factory.profile.Load(); err != nil {
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

func (factory *CommandFactory) ensureTelemetry(command string) {
	telemetryMode := factory.config.TelemetryMode
	existingTelemetryMode := factory.profile.GetTelemetryMode()

	if telemetryMode == telemetry.ModeNil {
		telemetryMode = existingTelemetryMode
	}

	if telemetryMode != existingTelemetryMode {
		factory.profile.SetTelemetryMode(telemetryMode)

		if err := factory.profile.Save(); err != nil {
			printErr := factory.ui.Print(terminal.NewErrorLog(err))
			if printErr != nil {
				factory.errLogger.Fatal(err) // log the original failure
			}
			os.Exit(1)
		}
	}

	factory.telemetryService = telemetry.NewService(
		telemetryMode,
		factory.profile.GetUser().PublicAPIKey,
		command,
	)
}

func (factory *CommandFactory) ensureUI() {
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

type suppressUsageError struct {
	error
}
