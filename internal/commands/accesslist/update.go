package accesslist

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
	"github.com/AlecAivazis/survey/v2"
)

type updateInputs struct {
	cli.ProjectInputs
	Address    string
	NewAddress string
	Comment    string
}

// CommandMetaUpdate is the command meta for the `accesslist update` command
var CommandMetaUpdate = cli.CommandMeta{
	Use:         "update",
	Display:     "accesslist update",
	Description: "Modify an IP address or CIDR block in the Access List of your Realm app",
	HelpText: `Changes an existing entry from the Access List of your Realm app. You will be
prompted to select an IP address or CIDR block if none is provided in the initial command.`,
}

// CommandCreate is the `accesslist update` command
type CommandUpdate struct {
	inputs updateInputs
}

// Flags is the command flags
func (cmd *CommandUpdate) Flags() []flags.Flag {
	return []flags.Flag{
		cli.AppFlagWithContext(&cmd.inputs.App, "to modify an entry in its Access List"),
		cli.ProjectFlag(&cmd.inputs.Project),
		cli.ProductFlag(&cmd.inputs.Products),
		flags.StringFlag{
			Value: &cmd.inputs.Address,
			Meta: flags.Meta{
				Name: "ip",
				Usage: flags.Usage{
					Description: "Specify the existing IP address or CIDR block that you would like to modify",
				},
			},
		},
		flags.StringFlag{
			Value: &cmd.inputs.NewAddress,
			Meta: flags.Meta{
				Name: "new-ip",
				Usage: flags.Usage{
					Description: "Specify the new IP address or CIDR block that will replace the existing entry",
				},
			},
		},
		flags.StringFlag{
			Value: &cmd.inputs.Comment,
			Meta: flags.Meta{
				Name: "comment",
				Usage: flags.Usage{
					Description: "Add or edit a comment to the IP address or CIDR block that is being modified in the Access List",
					Note:        "This action is optional",
				},
			},
		},
	}
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

	if err := clients.Realm.AllowedIPUpdate(
		app.GroupID,
		app.ID,
		allowedIP.ID,
		cmd.inputs.NewAddress,
		cmd.inputs.Comment,
	); err != nil {
		return err
	}

	ui.Print(terminal.NewTextLog("Successfully updated allowed IP"))
	return nil
}

func (i *updateInputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory, false); err != nil {
		return err
	}

	var questions []*survey.Question

	if i.NewAddress == "" {
		questions = append(questions, &survey.Question{
			Name:   "new-ip",
			Prompt: &survey.Input{Message: "New IP Address or CIDR block"},
		})
	}

	if i.Comment == "" {
		questions = append(questions, &survey.Question{
			Name:   "comment",
			Prompt: &survey.Password{Message: "Comment"},
		})
	}

	if len(questions) > 0 {
		err := ui.Ask(i, questions...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *updateInputs) resolveAllowedIP(ui terminal.UI, allowedIPs []realm.AllowedIP) (realm.AllowedIP, error) {
	if i.Address != "" {
		for _, allowedIP := range allowedIPs {
			if allowedIP.Address == i.Address {
				return allowedIP, nil
			}
		}
		return realm.AllowedIP{}, fmt.Errorf("unable to find allowed IP: %s", i.Address)
	}

	selectableAllowedIPs := map[string]realm.AllowedIP{}
	selectableOptions := make([]string, len(allowedIPs))
	for i, allowedIP := range allowedIPs {
		option := allowedIP.ID
		selectableOptions[i] = option
		selectableAllowedIPs[option] = allowedIP
	}

	var selected string
	if err := ui.AskOne(
		&selected,
		&survey.Select{
			Message: "Which IP address or CIDR block would you like to update?",
			Options: selectableOptions,
		},
	); err != nil {
		return realm.AllowedIP{}, err
	}

	return selectableAllowedIPs[selected], nil
}
