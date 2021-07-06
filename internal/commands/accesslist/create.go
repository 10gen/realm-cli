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
	Address    string
	Comment    string
	UseCurrent bool
	AllowAll   bool
}

// CommandMetaCreate is the command meta for the `accessList create` command
var CommandMetaCreate = cli.CommandMeta{
	Use:         "create",
	Aliases:     []string{"add"},
	Display:     "accessList create",
	Description: "Create an IP address or CIDR block in the Access List of your Realm app",
	HelpText:    "Adds a new entry into the Access List of your Realm app. You will be prompted to input an IP address or CIDR block if none is provided in the initial command.",
	Hidden:      true,
}

// CommandCreate is the ip access create command
type CommandCreate struct {
	inputs createInputs
}

// Flags is the command flags
func (cmd *CommandCreate) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs, "")

	fs.StringVar(&cmd.inputs.Address, flagIP, "", flagIPUsageCreate)
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

	if cmd.inputs.AllowAll {
		cmd.inputs.Address = "0.0.0.0"
	}

	allowedIP, err := clients.Realm.AllowedIPCreate(app.GroupID, app.ID, cmd.inputs.Address, cmd.inputs.Comment, cmd.inputs.UseCurrent)
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

	if i.Address == "" {
		if i.UseCurrent && i.AllowAll {
			return errors.New("when you are using this command, you can only provide one IP address or CIDR block at a time")
		}
	} else {
		if i.UseCurrent || i.AllowAll {
			return errors.New("when you are using this command, you can only provide one IP address or CIDR block at a time")
		}
	}
	return nil
}
