package cli

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/10gen/realm-cli/internal/cli/feedback"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/telemetry"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CommandFactory is a command factory
type CommandFactory struct {
	profile          *user.Profile
	ui               terminal.UI
	uiConfig         terminal.UIConfig
	inReader         *os.File
	outWriter        *os.File
	errWriter        *os.File
	telemetryService telemetry.Service
}

// NewCommandFactory creates a new command factory
func NewCommandFactory() (*CommandFactory, error) {
	profile, err := user.NewDefaultProfile()
	if err != nil {
		return nil, err
	}

	return &CommandFactory{profile: profile}, nil
}

// Build builds a Cobra command from the specified CommandDefinition
func (factory *CommandFactory) Build(command CommandDefinition) *cobra.Command {
	display := command.Display
	if display == "" {
		display = command.Use
	}

	var aliasHelp string
	if len(command.Aliases) == 1 {
		aliasHelp = fmt.Sprintf(" (alias: %s)", command.Aliases[0])
	} else if len(command.Aliases) > 1 {
		aliasHelp = fmt.Sprintf(" (aliases: %s)", strings.Join(command.Aliases, ", "))
	}

	cmd := cobra.Command{
		Use:     command.Use,
		Short:   command.Description + aliasHelp,
		Long:    command.Description + "\n\n" + command.HelpText,
		Aliases: command.Aliases,
		Hidden:  command.Hidden,
	}

	cmd.InheritedFlags().SortFlags = false // ensures command usage text displays global flags unsorted

	for _, subCommand := range command.SubCommands {
		cmd.AddCommand(factory.Build(subCommand))
	}

	if command.Command != nil {

		if command, ok := command.Command.(CommandFlags); ok {
			fs := cmd.Flags()
			fs.SortFlags = false // ensures command flags are added unsorted

			for _, flag := range command.Flags() {
				flag.Register(fs)
			}
		}

		cmd.PersistentPreRun = func(c *cobra.Command, a []string) {
			factory.ensureUI()
			cmd.SetIn(factory.inReader)
			cmd.SetOut(factory.outWriter)
			cmd.SetErr(factory.errWriter)

			if err := factory.profile.ResolveFlags(); err != nil {
				factory.ui.Print(terminal.NewErrorLog(err))
				os.Exit(1)
			}

			factory.telemetryService = telemetry.NewService(
				factory.profile.Flags.TelemetryMode,
				factory.profile.Credentials().PublicAPIKey,
				display,
				Version,
			)

			factory.checkForNewVersion(http.DefaultClient)
		}

		if command, ok := command.Command.(CommandInputs); ok {
			cmd.PreRunE = func(c *cobra.Command, a []string) error {
				if err := command.Inputs().Resolve(factory.profile, factory.ui); err != nil {
					return feedback.WrapErr(display+" setup failed: %w", err)
				}
				return nil
			}
		}

		cmd.RunE = func(c *cobra.Command, a []string) error {
			var additionalFields []telemetry.EventData
			if additionalTracker, ok := command.Command.(telemetry.AdditionalTracker); ok {
				additionalFields = additionalTracker.AdditionalTrackedFields()
			}

			factory.telemetryService.TrackEvent(
				telemetry.EventTypeCommandStart,
				additionalFields...,
			)

			err := command.Command.Handler(factory.profile, factory.ui, Clients{
				Realm:        realm.NewAuthClient(factory.profile.RealmBaseURL(), factory.profile),
				Atlas:        atlas.NewAuthClient(factory.profile.AtlasBaseURL(), factory.profile.Credentials()),
				HostingAsset: http.DefaultClient,
			})
			if err != nil {
				factory.telemetryService.TrackEvent(
					telemetry.EventTypeCommandError,
					append(additionalFields, telemetry.EventDataError(err)...)...,
				)
				return feedback.WrapErr(display+" failed: %w", err, feedback.ErrNoUsage{})
			}

			factory.telemetryService.TrackEvent(
				telemetry.EventTypeCommandComplete,
				additionalFields...,
			)
			return nil
		}
	}

	return &cmd
}

// Run executes the command
func (factory *CommandFactory) Run(cmd *cobra.Command) int {
	defer factory.close()

	err := cmd.Execute()
	if err == nil {
		return 0
	}

	handleUsage(cmd, err)

	if factory.ui == nil {
		log.Print(err)
	} else {
		logs := []terminal.Log{terminal.NewErrorLog(err)}

		var suggester feedback.ErrSuggester
		if errors.As(err, &suggester) && len(suggester.Suggestions()) > 0 {
			logs = append(logs, terminal.NewFollowupLog(terminal.MsgSuggestions, suggester.Suggestions()...))
		}

		var linkReferrer feedback.ErrLinkReferrer
		if errors.As(err, &linkReferrer) && len(linkReferrer.ReferenceLinks()) > 0 {
			logs = append(logs, terminal.NewFollowupLog(terminal.MsgReferenceLinks, linkReferrer.ReferenceLinks()...))
		}

		factory.ui.Print(logs...)
	}

	return 1
}

// SetGlobalFlags sets the global flags
func (factory *CommandFactory) SetGlobalFlags(fs *pflag.FlagSet) {
	fs.SortFlags = false // ensures global flags are added unsorted

	// profile flags
	fs.StringVar(&factory.profile.Name, user.FlagProfile, user.DefaultProfile, user.FlagProfileUsage)
	fs.Var(&factory.profile.Flags.TelemetryMode, telemetry.FlagMode, telemetry.FlagModeUsage)

	// ui flags
	fs.StringVarP(&factory.uiConfig.OutputTarget, terminal.FlagOutputTarget, terminal.FlagOutputTargetShort, "", terminal.FlagOutputTargetUsage)
	fs.VarP(&factory.uiConfig.OutputFormat, terminal.FlagOutputFormat, terminal.FlagOutputFormatShort, terminal.FlagOutputFormatUsage)
	fs.BoolVar(&factory.uiConfig.DisableColors, terminal.FlagDisableColors, false, terminal.FlagDisableColorsUsage)
	fs.BoolVarP(&factory.uiConfig.AutoConfirm, terminal.FlagAutoConfirm, terminal.FlagAutoConfirmShort, false, terminal.FlagAutoConfirmUsage)

	// hidden flags
	fs.StringVar(&factory.profile.Flags.AtlasBaseURL, user.FlagAtlasBaseURL, "", user.FlagAtlasBaseURLUsage)
	flags.MarkHidden(fs, user.FlagAtlasBaseURL)

	fs.StringVar(&factory.profile.Flags.RealmBaseURL, user.FlagRealmBaseURL, "", user.FlagRealmBaseURLUsage)
	flags.MarkHidden(fs, user.FlagRealmBaseURL)
}

// Setup initializes the command factory
func (factory *CommandFactory) Setup() {
	if err := factory.profile.Load(); err != nil {
		log.Fatal(err)
	}

	if filepath := factory.uiConfig.OutputTarget; filepath != "" {
		f, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to open target file: %w", err))
		}
		factory.outWriter = f
	}
}

func (factory *CommandFactory) close() {
	if factory.telemetryService != nil {
		factory.telemetryService.Close()
	}

	if factory.uiConfig.OutputTarget != "" {
		factory.outWriter.Close()
	}
}

func (factory *CommandFactory) checkForNewVersion(client VersionManifestClient) {
	now := time.Now()
	lastVersionCheck := factory.profile.LastVersionCheck()

	// check once per day
	if now.Year() == lastVersionCheck.Year() && now.Month() == lastVersionCheck.Month() && now.Day() == lastVersionCheck.Day() {
		return
	}

	v, err := checkVersion(client)
	if err != nil {
		factory.telemetryService.TrackEvent(
			telemetry.EventTypeCommandError,
			telemetry.EventDataError(err)...,
		)
		return
	}

	defer factory.profile.SetLastVersionCheck(now)

	if v.Semver == "" {
		return
	}

	factory.ui.Print(
		terminal.NewWarningLog("New version (v%s) of CLI available: %s", v.Semver, v.URL),
		terminal.NewDebugLog("Note: This is the only time this alert will display today"),
		terminal.NewFollowupLog(
			"To install",
			fmt.Sprintf("npm install -g mongodb-%s@v%s", Name, v.Semver),
			fmt.Sprintf("curl -o ./%s %s && chmod +x ./%s", Name, v.URL, Name),
		),
	)

	factory.telemetryService.TrackEvent(telemetry.EventTypeCommandVersionCheck)
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
		factory.ui = terminal.NewUI(factory.uiConfig, factory.inReader, factory.outWriter, factory.errWriter)
	}
}

func handleUsage(cmd *cobra.Command, err error) {
	var usageHider feedback.ErrUsageHider
	if errors.As(err, &usageHider) {
		if usageHider.HideUsage() {
			return
		}
	}
	fmt.Println(cmd.UsageString())
}
