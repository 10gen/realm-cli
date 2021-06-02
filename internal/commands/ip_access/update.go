package ipaccess

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/spf13/pflag"
)

// CommandUpdate is the ip access update command
type CommandUpdate struct {
	inputs updateInputs
}

// Flags is the command flags
func (cmd *CommandUpdate) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)
	fs.StringVar(&cmd.inputs.IPAddress, flagIP, "", flagIPUsageUpdate)
	fs.StringVar(&cmd.inputs.NewIPAddress, flagNewIP, "", flagNewIPUsageUpdate)
	fs.StringVar(&cmd.inputs.Comment, flagComment, "", flagCommentUsageUpdate)
}

// Inputs is the command inputs
func (cmd *CommandUpdate) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *CommandUpdate) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
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

	if err := clients.Realm.UpdateAllowedIP(
		app.GroupID,
		app.ID,
		allowedIP.ID,
		cmd.inputs.NewIPAddress,
		cmd.inputs.Comment,
	); err != nil {
		return err
	}

	ui.Print(terminal.NewTextLog("Successfully updated allowed IP"))
	return nil
}
