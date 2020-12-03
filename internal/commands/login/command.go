package login

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/pflag"
)

// Command is the `login` command
var Command = cli.CommandDefinition{
	Use:         "login",
	Description: "Authenticate with an Atlas programmatic API Key",
	Help:        "login", // TODO(REALMC-7429): add help text description
	Command:     &command{},
}

type command struct {
	inputs      inputs
	realmClient realm.Client
}

func (cmd *command) Flag(fs *pflag.FlagSet) {
	fs.StringVarP(&cmd.inputs.PublicAPIKey, flagPublicAPIKey, flagPublicAPIKeyShort, "", flagPublicAPIKeyUsage)
	fs.StringVarP(&cmd.inputs.PrivateAPIKey, flagPrivateAPIKey, flagPrivateAPIKeyShort, "", flagPrivateAPIKeyUsage)
}

func (cmd *command) Setup(profile *cli.Profile, ui terminal.UI, ctx cli.Context) error {
	if err := cmd.inputs.resolve(profile, ui); err != nil {
		return fmt.Errorf("failed to resolve inputs: %w", err)
	}

	cmd.realmClient = realm.NewClient(ctx.RealmBaseURL)
	return nil
}

func (cmd *command) Handler(profile *cli.Profile, ui terminal.UI, args []string) error {
	existingUser := profile.GetUser()

	if existingUser.PublicAPIKey != "" && existingUser.PublicAPIKey != cmd.inputs.PublicAPIKey {
		var proceed bool

		err := ui.AskOne(
			&proceed,
			&survey.Confirm{Message: fmt.Sprintf(
				"This action will terminate the existing session for user: %s (%s), would you like to proceed?",
				existingUser.PublicAPIKey,
				existingUser.RedactedPrivateAPIKey(),
			)},
		)
		if err != nil {
			return err
		}

		if !proceed {
			return nil
		}
	}

	profile.SetUser(cmd.inputs.PublicAPIKey, cmd.inputs.PrivateAPIKey)

	auth, authErr := cmd.realmClient.Authenticate(cmd.inputs.PublicAPIKey, cmd.inputs.PrivateAPIKey)
	if authErr != nil {
		return authErr
	}

	profile.SetSession(auth.AccessToken, auth.RefreshToken)
	return profile.Save()
}

func (cmd *command) Feedback(profile *cli.Profile, ui terminal.UI) error {
	return ui.Print(terminal.NewTextLog("Successfully logged in"))
}
