package whoami

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

// CommandMeta is the command meta for the `whoami` command
var CommandMeta = cli.CommandMeta{
	Use:         "whoami",
	Description: "Display information about the current user",
	// TODO(BAAS-14174): this is an example of where standardizing cli, command and flag names
	// into a shared package would be helpful, to reduce coupling the command packages to each other
	// (since this HelpText creates a "whoami depends on login" package cycle)
	HelpText: `Displays a table that includes your Public and redacted Private Atlas
programmatic API Key (e.g. ********-****-****-****-3ba985aa367a). No session
data will be surfaced if you are not logged in.

NOTE: To log in and authenticate your session, use "realm-cli login"`,
}

type inputs struct {
	showProjects bool
}

// Command is the `whoami` command
type Command struct {
	inputs inputs
}

// Flags is the command flags
func (cmd *Command) Flags() []flags.Flag {
	return []flags.Flag{
		flags.BoolFlag{
			Value: &cmd.inputs.showProjects,
			Meta: flags.Meta{
				Name: "show-projects",
				Usage: flags.Usage{
					Description: "Show projects associated with this profile's API Key",
				},
			},
		},
	}
}

const (
	headerProjectID   = "Project ID"
	headerProjectName = "Project Name"
)

// Handler is the command handler
func (cmd *Command) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	u := profile.Credentials()
	sess := profile.Session()

	if u.PrivateAPIKey == "" && u.Password == "" {
		ui.Print(terminal.NewTextLog("No user is currently logged in"))
		return nil
	}

	userDisplay := u.PublicAPIKey
	if userDisplay == "" {
		userDisplay = u.Username
	}

	if sess.AccessToken == "" {
		ui.Print(terminal.NewTextLog("The user, %s, is not currently logged in", userDisplay))
		return nil
	}

	var userSecret string
	if u.PublicAPIKey == "" {
		userSecret = user.Redact(u.Password)
	} else {
		userSecret = user.RedactKey(u.PrivateAPIKey)
	}

	ui.Print(terminal.NewTextLog("Currently logged in user: %s (%s)", userDisplay, userSecret))

	if cmd.inputs.showProjects {
		if groups, err := atlas.AllGroups(clients.Atlas); err == nil {
			tableRows := make([]map[string]interface{}, 0, len(groups))
			for _, group := range groups {
				tableRows = append(tableRows, map[string]interface{}{
					headerProjectID:   group.ID,
					headerProjectName: group.Name,
				})
			}
			ui.Print(terminal.NewTableLog(
				fmt.Sprintf("Projects available (%d)", len(groups)),
				[]string{headerProjectID, headerProjectName},
				tableRows...,
			))
		}
	}

	return nil
}
