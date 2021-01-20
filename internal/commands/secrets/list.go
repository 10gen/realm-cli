package secrets

import (
	"fmt"
	"time"

	"github.com/10gen/realm-cli/internal/app"
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// CommandList is the `secrets list` command
type CommandList struct {
	inputs      listInputs
	realmClient realm.Client
	values      []realm.Value
}

type listInputs struct {
	app.ProjectInputs
}

// Flags are the command flags
func (cmd *CommandList) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)
}

// Inputs are the command inputs
func (cmd *CommandList) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Setup is the command setup
func (cmd *CommandList) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = realm.NewAuthClient(profile)
	return nil
}

// Handler is the command handler
func (cmd *CommandList) Handler(profile *cli.Profile, ui terminal.UI) error {
	app, appErr := app.Resolve(ui, cmd.realmClient, cmd.inputs.Filter())
	if appErr != nil {
		return appErr
	}

	values, findValuesErr := cmd.realmClient.FindValues(app)
	if findValuesErr != nil {
		return findValuesErr
	}
	cmd.values = values
	return nil
}

// Feedback is the command feedback
func (cmd *CommandList) Feedback(profile *cli.Profile, ui terminal.UI) error {
	if len(cmd.values) == 0 {
		return ui.Print(terminal.NewTextLog("No available secrets to show"))
	}

	var logs []terminal.Log
	logs = append(logs, terminal.NewTableLog(
		fmt.Sprintf("Found %d secrets", len(cmd.values)),
		secretTableHeaders(),
		secretTableRows(cmd.values)...,
	))
	return ui.Print(logs...)
}

func secretTableHeaders() []string {
	var headers []string
	headers = append(
		headers,
		headerName,
		headerID,
		headerSecret,
		headerLastMofified,
	)
	return headers
}

func secretTableRows(values []realm.Value) []map[string]interface{} {
	secretTableRows := make([]map[string]interface{}, 0, len(values))
	for _, value := range values {
		secretTableRows = append(secretTableRows, secretTableRow(value))
	}
	return secretTableRows
}

func secretTableRow(value realm.Value) map[string]interface{} {
	timeString := "n/a"
	if value.LastModified != 0 {
		timeString = time.Unix(value.LastModified, 0).UTC().String()
	}
	row := map[string]interface{}{
		headerName:         value.Name,
		headerID:           value.ID,
		headerLastMofified: timeString,
		headerSecret:       value.Secret,
	}
	return row
}

func (i *listInputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory); err != nil {
		return err
	}

	return nil
}
