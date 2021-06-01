package ip_access

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/spf13/pflag"
)

// CommandDelete for the ip access delete command
type CommandDelete struct {
	inputs deleteInputs
}

// Flags is the command flags
func (cmd *CommandDelete) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)
	fs.StringVar(&cmd.inputs.IP, flagIP, "", flagIPUsageDelete)
}

// Inputs is the command inputs
func (cmd *CommandDelete) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *CommandDelete) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	app, err := cli.ResolveApp(ui, clients.Realm, cmd.inputs.Filter())
	if err != nil {
		return err
	}

	allowedIPs, err := clients.Realm.AllowedIPs(app.GroupID, app.ID)
	if err != nil {
		return err
	}

	allowedIP, err := cmd.inputs.resolveAllowedIP(ui, allowedIPs)
	if err != nil {
		return err
	}

	if err := clients.Realm.DeleteAllowedIP(
		app.GroupID,
		app.ID,
		allowedIP.IP,
	); err != nil {
		return err
	}

	ui.Print(terminal.NewTextLog("Successfully deleted allowed IP"))
	return nil
}
