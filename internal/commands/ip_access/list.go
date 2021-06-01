package ip_access

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/spf13/pflag"
)

type CommandList struct {
	inputs listInputs
}

type listInputs struct {
	cli.ProjectInputs
}

func (cmd *CommandList) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags((fs))
}

func (cmd *CommandList) Inputs() cli.InputResolver {
	return &cmd.inputs
}

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

	// ui.Print(terminal.NewTableLog(
	// 	fmt.Sprintf("Found %d allowed IPs", len(allowedIPs)),
	// 	tableHeaders(),
	// 	tableRowsList(secrets)...,
	// ))
	return nil
}

func (i *listInputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory, false); err != nil {
		return err
	}

	return nil
}
