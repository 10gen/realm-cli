package accesslist

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

// CommandMetaList is the command meta for the `accesslist list` command
var CommandMetaList = cli.CommandMeta{
	Use:         "list",
	Aliases:     []string{"ls"},
	Display:     "accesslist list",
	Description: "List the allowed entries in the Access List of your Realm app",
	HelpText:    `This will display the IP addresses/CIDR blocks in your Access List`,
}

// CommandList is the `accesslist list` command
type CommandList struct {
	inputs listInputs
}

type listInputs struct {
	cli.ProjectInputs
}

// Flags are the command flags
func (cmd *CommandList) Flags() []flags.Flag {
	return []flags.Flag{
		cli.AppFlagWithContext(&cmd.inputs.App, "to list its allowed IPs"),
		cli.ProjectFlag(&cmd.inputs.Project),
		cli.ProductFlag(&cmd.inputs.Products),
	}
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

	accessList, accessListErr := clients.Realm.AllowedIPs(app.GroupID, app.ID)
	if accessListErr != nil {
		return accessListErr
	}

	if len(accessList.AllowedIPs) == 0 {
		ui.Print(terminal.NewTextLog("No available allowed IPs to show"))
		return nil
	}

	ui.Print(terminal.NewTableLog(
		fmt.Sprintf("Found %d allowed IPs", len(accessList.AllowedIPs)),
		tableHeaders,
		tableRowsList(accessList.AllowedIPs)...,
	))
	return nil
}

func tableRowsList(allowedIPs []realm.AllowedIP) []map[string]interface{} {
	rows := make([]map[string]interface{}, 0, len(allowedIPs))
	for _, allowedIP := range allowedIPs {
		rows = append(rows, map[string]interface{}{
			headerAddress: allowedIP.Address,
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
