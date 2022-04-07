package accesslist

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

const (
	headerAddress = "IP Address"
	headerComment = "Comment"
)

var (
	listTableHeaders = []string{headerAddress, headerComment}
)

// CommandMetaList is the command meta for the `accesslist list` command
var CommandMetaList = cli.CommandMeta{
	Use:         "list",
	Aliases:     []string{"ls"},
	Display:     "accesslist list",
	Description: "List the allowed entries in the Access List of your Realm app",
	HelpText: `This will display the IP addresses and/or CIDR blocks in the Access
List of your Realm app`,
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
		cli.AppFlagWithContext(&cmd.inputs.App, "to list its allowed IP addresses and/or CIDR blocks"),
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
	app, appErr := cli.ResolveApp(ui, clients.Realm, cli.AppOptions{
		AppMeta: cmd.inputs.AppMeta,
		Filter:  cmd.inputs.Filter(),
	})
	if appErr != nil {
		return appErr
	}

	allowedIPs, allowedIPsErr := clients.Realm.AllowedIPs(app.GroupID, app.ID)
	if allowedIPsErr != nil {
		return allowedIPsErr
	}

	if len(allowedIPs) == 0 {
		ui.Print(terminal.NewTextLog("No available allowed IP addresses and/or CIDR blocks to show"))
		return nil
	}

	tableRows := make([]map[string]interface{}, 0, len(allowedIPs))
	for _, allowedIP := range allowedIPs {
		tableRows = append(tableRows, map[string]interface{}{
			headerAddress: allowedIP.Address,
			headerComment: allowedIP.Comment,
		})
	}

	ui.Print(terminal.NewTableLog(
		fmt.Sprintf("Found %d allowed IP address(es) and/or CIDR block(s)", len(allowedIPs)),
		listTableHeaders,
		tableRows...,
	))
	return nil
}

func (i *listInputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory, false); err != nil {
		return err
	}

	return nil
}
