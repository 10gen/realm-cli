package accesslist

import (
	"errors"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/AlecAivazis/survey/v2"
)

var (
	errTooManyAddressess = errors.New("must only provide one IP address or CIDR block at a time")
)

type createInputs struct {
	cli.ProjectInputs
	Address    string
	Comment    string
	UseCurrent bool
	AllowAll   bool
}

// CommandMetaCreate is the command meta for the `accesslist create` command
var CommandMetaCreate = cli.CommandMeta{
	Use:         "create",
	Display:     "accesslist create",
	Description: "Create an IP address or CIDR block in the Access List for your Realm app",
	HelpText: `You will be prompted to input an IP address or CIDR block if none is provided in
the initial command.`,
}

// CommandCreate is the `accesslist create` command
type CommandCreate struct {
	inputs createInputs
}

// Flags is the command flags
func (cmd *CommandCreate) Flags() []flags.Flag {
	return []flags.Flag{
		cli.AppFlagWithContext(&cmd.inputs.App, "to create an entry in its Access List"),
		cli.ProjectFlag(&cmd.inputs.Project),
		cli.ProductFlag(&cmd.inputs.Products),
		flags.StringFlag{
			Value: &cmd.inputs.Address,
			Meta: flags.Meta{
				Name: "ip",
				Usage: flags.Usage{
					Description: "Specify the IP address or CIDR block that you would like to add",
				},
			},
		},
		flags.StringFlag{
			Value: &cmd.inputs.Comment,
			Meta: flags.Meta{
				Name: "comment",
				Usage: flags.Usage{
					Description: "Add a comment to the IP address or CIDR block",
					Note:        "This action is optional",
				},
			},
		},
		flags.BoolFlag{
			Value: &cmd.inputs.UseCurrent,
			Meta: flags.Meta{
				Name: "use-current",
				Usage: flags.Usage{
					Description: "Add your current IP address to your Access List",
				},
			},
		},
		flags.BoolFlag{
			Value: &cmd.inputs.AllowAll,
			Meta: flags.Meta{
				Name: "allow-all",
				Usage: flags.Usage{
					Description: "Allows all IP addresses to access your Realm app",
					Note:        `“0.0.0.0/0” will be added as an entry`,
				},
			},
		},
	}
}

// Inputs is the command inputs
func (cmd *CommandCreate) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *CommandCreate) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	app, err := cli.ResolveApp(ui, clients.Realm, cli.AppOptions{
		AppMeta: cmd.inputs.AppMeta,
		Filter:  cmd.inputs.Filter(),
	})
	if err != nil {
		return err
	}

	allowedIP, err := clients.Realm.AllowedIPCreate(app.GroupID, app.ID, cmd.inputs.Address, cmd.inputs.Comment, cmd.inputs.UseCurrent)
	if err != nil {
		return err
	}

	ui.Print(terminal.NewTextLog("Successfully created allowed IP, id: %s", allowedIP.ID))
	return nil
}

func (i *createInputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory, true); err != nil {
		return err
	}

	if i.AllowAll {
		if i.Address != "" {
			return errTooManyAddressess
		}
		i.Address = "0.0.0.0"
	}

	if i.Address != "" && i.UseCurrent {
		return errTooManyAddressess
	}

	if i.Address == "" && !i.UseCurrent {
		// TODO(REALMC-9532): validate the user does not enter an empty string
		if err := ui.AskOne(&i.Address, &survey.Input{Message: "IP Address"}); err != nil {
			return err
		}
	}

	return nil
}
