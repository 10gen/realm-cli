package ip_access

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/spf13/pflag"
)

type CommandUpdate struct {
	inputs updateInputs
}

// Inputs function for the secrets update command
func (cmd *CommandUpdate) Inputs() cli.InputResolver {
	return &cmd.inputs
}

func (cmd *CommandUpdate) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)
	fs.StringVar(&cmd.inputs.IP, flagIP, "", flagIPUsageUpdate)
	fs.StringVar(&cmd.inputs.NewIP, flagNewIP, "", flagNewIPUsageUpdate)
	fs.StringVar(&cmd.inputs.Comment, flagComment, "", flagCommentUsageUpdate)
}

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
		cmd.inputs.NewIP,
		cmd.inputs.Comment,
	); err != nil {
		return err
	}

	ui.Print(terminal.NewTextLog("Successfully updated allowed IP"))
	return nil
}
