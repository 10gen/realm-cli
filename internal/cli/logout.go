package cli

import (
	"github.com/10gen/realm-cli/internal/terminal"
)

// LogoutCommand creates the 'logout' command
func LogoutCommand() CommandDefinition {
	return CommandDefinition{
		Command:     &logoutCommand{},
		Use:         "logout",
		Description: "Terminate the current user’s session",
		Help:        "logout", // TODO(REALMC-7429): add help text description
	}
}

type logoutCommand struct {
}

func (cmd *logoutCommand) Handler(profile *Profile, ui terminal.UI, args []string) error {
	profile.ClearSession()
	return profile.Save()
}

func (cmd *logoutCommand) Feedback(profile *Profile, ui terminal.UI) error {
	return ui.Print(terminal.NewTextLog("Successfully logged out."))
}
