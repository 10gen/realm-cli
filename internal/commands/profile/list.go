package profile

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

// CommandMetaList is the command meta for the `profile list` command
var CommandMetaList = cli.CommandMeta{
	Use:         "list",
	Aliases:     []string{"ls"},
	Display:     "profiles list",
	Description: "List your profiles",
}

// CommandList is the `profile list` command
type CommandList struct {
	inputs cli.ProjectInputs
}

// Flags is the command flags
func (cmd *CommandList) Flags() []flags.Flag {
	return []flags.Flag{
		cli.ProjectFlag(&cmd.inputs.Project),
		cli.ProductFlag(&cmd.inputs.Products),
	}
}

// Handler is the command handler
func (cmd *CommandList) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	profileMetas, err := user.Profiles()
	if err != nil {
		return err
	}

	names := make([]interface{}, 0)
	for _, v := range profileMetas {
		names = append(names, v.Name)
	}

	ui.Print(terminal.NewListLog(fmt.Sprintf("Found %d profile(s)", len(names)), names...))
	return nil
}
