package cli

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/telemetry"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CommandFactory is a command factory
type CommandFactory struct {
	profile          *Profile
	ui               terminal.UI
	uiConfig         terminal.UIConfig
	inReader         *os.File
	outWriter        *os.File
	errWriter        *os.File
	errLogger        *log.Logger
	telemetryService *telemetry.Service
}

// NewCommandFactory creates a new command factory
func NewCommandFactory() *CommandFactory {
	errLogger := log.New(os.Stderr, "UTC ERROR ", log.Ltime|log.Lmsgprefix)

	profile, profileErr := NewDefaultProfile()
	if profileErr != nil {
		errLogger.Fatal(profileErr)
	}

	return &CommandFactory{
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
		if command, ok := command.Command.(CommandFlags); ok {
			command.Flags(cmd.Flags())
		}

		cmd.PersistentPreRun = func(c *cobra.Command, a []string) {
			factory.ensureUI()
			cmd.SetIn(factory.inReader)
			cmd.SetOut(factory.outWriter)
			cmd.SetErr(factory.errWriter)

			if err := factory.profile.resolveFlags(); err != nil {
				factory.ui.Print(terminal.NewErrorLog(err))
				os.Exit(1)
			}

			telemetryService, err := telemetry.NewService(
				factory.profile.telemetryMode,
				factory.profile.User().PublicAPIKey,
				factory.profile.telemetryWriteKey,
				factory.errLogger,
				display,
			)
			if err != nil {
				factory.ui.Print(terminal.NewErrorLog(err))
			}
			factory.telemetryService = telemetryService
		}

		if command, ok := command.Command.(CommandInputs); ok {
			cmd.PreRunE = func(c *cobra.Command, a []string) error {
				if err := command.Inputs().Resolve(factory.profile, factory.ui); err != nil {
					return fmt.Errorf("%s setup failed: %w", display, err)
				}
				return nil
			}
		}

		cmd.RunE = func(c *cobra.Command, a []string) error {
			factory.telemetryService.TrackEvent(telemetry.EventTypeCommandStart)

			err := command.Command.Handler(factory.profile, factory.ui, Clients{
				Realm:        realm.NewAuthClient(factory.profile.RealmBaseURL(), factory.profile), // TODO(REALMC-8185): make this accept factory.profile.Session()
				Atlas:        atlas.NewAuthClient(factory.profile.AtlasBaseURL(), factory.profile.User()),
				HostingAsset: local.NewHostingAssetClient(http.Get),
			})
			if err != nil {
				factory.telemetryService.TrackEvent(
					telemetry.EventTypeCommandError,
					telemetry.EventData{Key: telemetry.EventDataKeyErr, Value: err},
				)
				return fmt.Errorf("%s failed: %w", display, errDisableUsage{err})
			}

			factory.telemetryService.TrackEvent(telemetry.EventTypeCommandComplete)
			return nil
		}
	}

	return &cmd
}

// Close closes the command factory
func (factory *CommandFactory) Close() {
	if factory.uiConfig.OutputTarget != "" {
		factory.outWriter.Close()
	}
}

// Run executes the command
func (factory *CommandFactory) Run(cmd *cobra.Command) {
	if err := cmd.Execute(); err != nil {
		handleUsage(cmd, err)

		if factory.ui == nil {
			factory.errLogger.Fatal(err)
		}

		logs := []terminal.Log{terminal.NewErrorLog(err)}
		if e, ok := err.(CommandSuggester); ok {
			logs = append(logs, terminal.NewFollowupLog(terminal.MsgSuggestedCommands, e.SuggestedCommands()))
		}
		if e, ok := err.(LinkReferrer); ok {
			logs = append(logs, terminal.NewFollowupLog(terminal.MsgSuggestedCommands, e.ReferenceLinks()))
		}

		factory.ui.Print(logs...)
		os.Exit(1)
	}
}

// SetGlobalFlags sets the global flags
func (factory *CommandFactory) SetGlobalFlags(fs *pflag.FlagSet) {
	// cli profile
	fs.StringVarP(&factory.profile.Name, flagProfile, flagProfileShort, DefaultProfile, flagProfileUsage)
	fs.StringVar(&factory.profile.atlasBaseURL, flagAtlasBaseURL, "", flagAtlasBaseURLUsage)
	fs.StringVar(&factory.profile.realmBaseURL, flagRealmBaseURL, "", flagRealmBaseURLUsage)
	fs.VarP(&factory.profile.telemetryMode, telemetry.FlagMode, telemetry.FlagModeShort, telemetry.FlagModeUsage)
	fs.StringVarP(&factory.profile.telemetryWriteKey, telemetry.FlagWriteKey, telemetry.FlagWriteKeyShort, "", telemetry.FlagWriteKeyUsage)
	// cli ui
	fs.BoolVarP(&factory.uiConfig.AutoConfirm, terminal.FlagAutoConfirm, terminal.FlagAutoConfirmShort, false, terminal.FlagAutoConfirmUsage)
	fs.BoolVar(&factory.uiConfig.DisableColors, terminal.FlagDisableColors, false, terminal.FlagDisableColorsUsage)
	fs.VarP(&factory.uiConfig.OutputFormat, terminal.FlagOutputFormat, terminal.FlagOutputFormatShort, terminal.FlagOutputFormatUsage)
	fs.StringVarP(&factory.uiConfig.OutputTarget, terminal.FlagOutputTarget, terminal.FlagOutputTargetShort, "", terminal.FlagOutputTargetUsage)
}

// Setup initializes the command factory
func (factory *CommandFactory) Setup() {
	if err := factory.profile.Load(); err != nil {
		factory.errLogger.Fatal(err)
	}

	if filepath := factory.uiConfig.OutputTarget; filepath != "" {
		f, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
		if err != nil {
			factory.errLogger.Fatal(fmt.Errorf("failed to open target file: %w", err))
		}
		factory.outWriter = f
	}
}

func (factory *CommandFactory) ensureUI() {
	if factory.inReader == nil {
		factory.inReader = os.Stdin
	}

	if factory.outWriter == nil {
		factory.outWriter = os.Stdout
	}

	if factory.errWriter == nil {
		if factory.uiConfig.OutputTarget != "" {
			factory.errWriter = factory.outWriter
		} else {
			factory.errWriter = os.Stderr
		}
	}

	if factory.ui == nil {
		factory.ui = terminal.NewUI(factory.uiConfig, factory.inReader, factory.outWriter, factory.errWriter, factory.errLogger)
	}
}

func handleUsage(cmd *cobra.Command, err error) {
	if _, ok := errors.Unwrap(err).(DisableUsage); ok {
		return
	}
	fmt.Println(cmd.UsageString())
}
