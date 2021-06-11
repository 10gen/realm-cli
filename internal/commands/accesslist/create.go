package accesslist

import (
	"errors"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/spf13/pflag"
)

type createInputs struct {
	cli.ProjectInputs
	IPAddress  string
	Comment    string
	UseCurrent bool
	AllowAll   bool
}

// CommandMetaCreate is the command meta for the `accesslist create` command
var CommandMetaCreate = cli.CommandMeta{
	Use:     "create",
	Aliases: []string{"add"},
	Display: "accesslist create",
	Hidden:  true,
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

	allowedIP, err := clients.Realm.AllowedIPCreate(app.GroupID, app.ID, cmd.inputs.IPAddress, cmd.inputs.Comment, cmd.inputs.UseCurrent, cmd.inputs.AllowAll)
	if err != nil {
		return err
	}

	ui.Print(terminal.NewTextLog("Successfully created allowed IP, id: %s", allowedIP.ID))
	return nil
}

func (i *createInputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory, false); err != nil {
		return err
	}

	if i.IPAddress == "" {
		if (!i.UseCurrent && !i.AllowAll) || (i.UseCurrent && i.AllowAll) {
			return errors.New("Must provide an IP Address or one of use-current and allow-all.")
		}
	} else {
		if i.UseCurrent || i.AllowAll {
			return errors.New("Cannot provide both an IP Address and one of use-current or allow-all.")
		}
	}
	return nil
}
