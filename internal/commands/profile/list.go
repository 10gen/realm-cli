package profile

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"
)

const (
	headerName = "Profile"
	headerKey  = "API Key"
)

// CommandMetaList is the command meta for the `profile list` command
var CommandMetaList = cli.CommandMeta{
	Use:         "list",
	Aliases:     []string{"ls"},
	Display:     "profiles list",
	Description: "List the profiles of your local CLI environment",
}

// CommandList is the `profile list` command
type CommandList struct{}

// Handler is the command handler
func (cmd *CommandList) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	environments, err := local.LoadEnvironments()
	if err != nil {
		return err
	}

	rows := make([]map[string]interface{}, len(environments))
	for _, v := range environments {
		credentials := "(logged out)"
		if v.Credentials.PublicAPIKey != "" && v.Credentials.PrivateAPIKey != "" {
			credentials = fmt.Sprintf("%s (%s)", v.Credentials.PublicAPIKey, user.RedactKey(v.Credentials.PrivateAPIKey))
		}
		rows = append(rows, map[string]interface{}{
			headerName: v.Name,
			headerKey:  credentials,
		})
	}

	ui.Print(terminal.NewTableLog(
		fmt.Sprintf("Found %d profile(s)", len(environments)),
		[]string{headerName, headerKey},
		rows...,
	))

	return nil
}
