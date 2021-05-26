package whoami

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"
)

// CommandMeta is the command meta for the `whoami` command
var CommandMeta = cli.CommandMeta{
	Use:         "whoami",
	Description: "Display information about the current user",
	// TODO(REALMC-8832): this is an example of where standardizing cli, comamnd and flag names
	// into a shared package would be helpful, to reduce coupling the command packages to each other
	// (since this HelptText creates a "whoami depends on login" package cycle)
	HelpText: `Displays a table that includes your Public and redacted Private Atlas
programmatic API Key (e.g. ********-****-****-****-3ba985aa367a). No session
data will be surfaced if you are not logged in.

NOTE: To log in and authenticate your session, use "realm-cli login"`,
}

// Command is the `whoami` command
type Command struct{}

// Handler is the command handler
func (cmd *Command) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	user := profile.Credentials()
	session := profile.Session()

	if user.PrivateAPIKey == "" {
		ui.Print(terminal.NewTextLog("No user is currently logged in"))
		return nil
	}

	if session.AccessToken == "" {
		ui.Print(terminal.NewTextLog("The user, %s, is not currently logged in", user.PublicAPIKey))
		return nil
	}

	ui.Print(terminal.NewTextLog("Currently logged in user: %s (%s)", user.PublicAPIKey, user.RedactedPrivateAPIKey()))
	return nil
}
