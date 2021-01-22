package secrets

import (
	"fmt"

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
	secrets     []realm.Secret
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

	secrets, FindSecretsErr := cmd.realmClient.FindSecrets(app)
	if FindSecretsErr != nil {
		return FindSecretsErr
	}
	cmd.secrets = secrets
	return nil
}

// Feedback is the command feedback
func (cmd *CommandList) Feedback(profile *cli.Profile, ui terminal.UI) error {
	if len(cmd.secrets) == 0 {
		return ui.Print(terminal.NewTextLog("No available secrets to show"))
	}

	var logs []terminal.Log
	logs = append(logs, terminal.NewTableLog(
		fmt.Sprintf("Found %d secrets", len(cmd.secrets)),
		secretTableHeaders(),
		secretTableRows(cmd.secrets)...,
	))
	return ui.Print(logs...)
}

func secretTableHeaders() []string {
	var headers []string
	headers = append(
		headers,
		headerID,
		headerName,
	)
	return headers
}

func secretTableRows(secrets []realm.Secret) []map[string]interface{} {
	secretTableRows := make([]map[string]interface{}, 0, len(secrets))
	for _, secret := range secrets {
		secretTableRows = append(secretTableRows, map[string]interface{}{
			headerName: secret.Name,
			headerID:   secret.ID,
		})
	}
	return secretTableRows
}

func (i *listInputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory); err != nil {
		return err
	}

	return nil
}
