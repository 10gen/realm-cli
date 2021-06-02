package ipaccess

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/spf13/pflag"
)

// CommandMetaCreate is the command meta for the `ip-access create` command
var CommandMetaCreate = cli.CommandMeta{
	Use:     "create",
	Aliases: []string{"add"},
	Display: "allowed IP create",
}

// CommandCreate is the ip access create command
type CommandCreate struct {
	inputs createInputs
}

// Flags is the command flags
func (cmd *CommandCreate) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)

	fs.StringVar(&cmd.inputs.IPAddress, flagIP, "", flagIPUsageCreate)
	fs.StringVar(&cmd.inputs.Comment, flagComment, "", flagCommentUsageCreate)
	fs.BoolVar(&cmd.inputs.UseCurrent, flagUseCurrent, false, flagUseCurrentUsageCreate)
	fs.BoolVar(&cmd.inputs.AllowAll, flagAllowAll, false, flagAllowAllUsageCreate)
}

// Inputs is the command inputs
func (cmd *CommandCreate) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *CommandCreate) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	app, err := cli.ResolveApp(ui, clients.Realm, cmd.inputs.Filter())
	if err != nil {
		return err
	}

	allowedIP, err := clients.Realm.CreateAllowedIP(app.GroupID, app.ID, cmd.inputs.IPAddress, cmd.inputs.Comment)
	if err != nil {
		return err
	}

	ui.Print(terminal.NewTextLog("Successfully created allowed IP, id: %s", allowedIP.ID))
	return nil
}
