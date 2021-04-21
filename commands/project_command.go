package commands

import (
	"github.com/10gen/realm-cli/utils/telemetry"

	"github.com/mitchellh/cli"
)

const (
	flagProjectIDName = "project-id"
)

// NewProjectCommand returns a new *ProjectCommand
func NewProjectCommand(name string, ui cli.Ui, telemetryService *telemetry.Service) *ProjectCommand {
	return &ProjectCommand{
		BaseCommand: &BaseCommand{
			Name:             name,
			UI:               ui,
			TelemetryService: telemetryService,
		},
	}
}

// ProjectCommand handles the parsing and execution of an Atlas project-based command.
type ProjectCommand struct {
	*BaseCommand

	flagProjectID string
}

func (pc *ProjectCommand) run(args []string) error {
	if pc.FlagSet == nil {
		pc.NewFlagSet()
	}

	pc.FlagSet.StringVar(&pc.flagProjectID, flagProjectIDName, "", "")

	if err := pc.BaseCommand.run(args); err != nil {
		return err
	}

	return nil
}

// Help defines help documentation for parameters that apply to project commands
func (pc *ProjectCommand) Help() string {
	return `

  --project-id [string]
	The Atlas Project ID.` +
		pc.BaseCommand.Help()
}
