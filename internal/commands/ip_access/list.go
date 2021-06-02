package ipaccess

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/spf13/pflag"
)

// CommandMetaList is the command meta for the `ip-access list` command
var CommandMetaList = cli.CommandMeta{
	Use:     "list",
	Aliases: []string{"ls"},
	Display: "allowed IPs list",
}

// CommandList is the ip access list command
type CommandList struct {
	inputs listInputs
}

type listInputs struct {
	cli.ProjectInputs
}

// Flags are the command flags
func (cmd *CommandList) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags((fs))
}

// Inputs are the command inputs
func (cmd *CommandList) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *CommandList) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	app, appErr := cli.ResolveApp(ui, clients.Realm, cmd.inputs.Filter())
	if appErr != nil {
		return appErr
	}

	allowedIPs, allowedIPsErr := clients.Realm.AllowedIPs(app.GroupID, app.ID)
	if allowedIPsErr != nil {
		return allowedIPsErr
	}

	if len(allowedIPs) == 0 {
		ui.Print(terminal.NewTextLog("No available allowed IPs to show"))
		return nil
	}

	ui.Print(terminal.NewTableLog(
		fmt.Sprintf("Found %d allowed IPs", len(allowedIPs)),
		tableHeaders(),
		tableRowsList(allowedIPs)...,
	))
	return nil
}

func tableRowsList(allowedIPs []realm.AllowedIP) []map[string]interface{} {
	rows := make([]map[string]interface{}, 0, len(allowedIPs))
	for _, allowedIP := range allowedIPs {
		rows = append(rows, map[string]interface{}{
			headerIP:      allowedIP.IPAddress,
			headerComment: allowedIP.Comment,
		})
	}
	return rows
}

func (i *listInputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory, false); err != nil {
		return err
	}

	return nil
}
