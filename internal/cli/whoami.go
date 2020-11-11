package cli

import (
	"github.com/10gen/realm-cli/internal/terminal"
)

// WhoamiCommand creates the 'whoami' command
func WhoamiCommand() CommandDefinition {
	return CommandDefinition{
		Command:     &whoamiCommand{},
		Usage:       "Display the current user's details",
		Description: "whoami", // TODO(REALMC-7429): add help text description
	}
}

type whoamiCommand struct {
}

func (cmd *whoamiCommand) Handler(profile *Profile, ui terminal.UI, args []string) error {
	return nil // commands without handlers show help text and usage when ran
}

func (cmd *whoamiCommand) Feedback(profile *Profile, ui terminal.UI) error {
	user := profile.GetUser()
	session := profile.GetSession()

	if user.PublicAPIKey == "" {
		return ui.Print(terminal.TextMessage{"No user is currently logged in."})
	}

	// TODO(REALMC-7339): print details as table, remove TitledJSONDocument once implemented
	// kept it around for "bold text" pattern, which will now be used for table headers
	return ui.Print(
		terminal.TitledJSONDocument{
			Title: "User Credentials",
			JSONDocument: terminal.JSONDocument{
				Data: map[string]interface{}{
					"API Key":         user.PublicAPIKey,
					"Private API Key": user.RedactedPrivateAPIKey(),
				},
			},
		},
		terminal.TitledJSONDocument{
			Title: "Session Info",
			JSONDocument: terminal.JSONDocument{
				Data: map[string]interface{}{
					"Access Token":  session.AccessToken,
					"Refresh Token": session.RefreshToken,
				},
			},
		},
	)
}
