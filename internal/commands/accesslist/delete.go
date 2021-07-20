package accesslist

import (
	"fmt"
	"sort"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/AlecAivazis/survey/v2"
)

type deleteInputs struct {
	cli.ProjectInputs
	Address string
}

// CommandMetaDelete is the command meta for the `accesslist delete` command
var CommandMetaDelete = cli.CommandMeta{
	Use:         "delete",
	Display:     "accesslist delete",
	Description: "Delete an IP address or CIDR block from the Access List of your Realm app",
	HelpText: `Removes an existing entry from the Access List of your Realm app. You will be
prompted to select an IP address or CIDR block if none is provided in the
initial command.`,
}

// CommandDelete is the `accesslist delete` command
type CommandDelete struct {
	inputs deleteInputs
}

// Flags is the command flags
func (cmd *CommandDelete) Flags() []flags.Flag {
	return []flags.Flag{
		cli.AppFlagWithContext(&cmd.inputs.App, "to modify an entry in its Access List"),
		cli.ProjectFlag(&cmd.inputs.Project),
		cli.ProductFlag(&cmd.inputs.Products),
		flags.StringFlag{
			Value: &cmd.inputs.Address,
			Meta: flags.Meta{
				Name: "ip",
				Usage: flags.Usage{
					Description: "Specify the IP address or CIDR block that you would like to delete",
				},
			},
		},
	}
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

	selected, err := cmd.inputs.resolveAllowedIP(ui, allowedIPs)
	if err != nil {
		return err
	}

	if len(selected) == 0 {
		ui.Print(terminal.NewTextLog("No IP addresses or CIDR blocks to delete"))
		return nil
	}

	outputs := make(deleteAllowedIPOutputs, len(selected))
	for i, allowedIP := range selected {
		err := clients.Realm.AllowedIPDelete(app.GroupID, app.ID, allowedIP.ID)
		outputs[i] = deleteAllowedIPOutput{allowedIP, err}
	}

	sort.SliceStable(outputs, func(i, j int) bool {
		return outputs[i].err != nil && outputs[j].err == nil
	})

	ui.Print(terminal.NewTableLog(
		fmt.Sprintf("Deleted %d IP addresses(s) and CIDR block(s)", len(outputs)),
		tableHeaders(headerDeleted, headerDetails),
		tableRows(outputs, tableRowDelete)...,
	))

	return nil
}

func (i *deleteInputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory, false); err != nil {
		return err
	}
	return nil
}

func (i *deleteInputs) resolveAllowedIP(ui terminal.UI, allowedIPs []realm.AllowedIP) ([]realm.AllowedIP, error) {
	if i.Address != "" {
		for _, allowedIP := range allowedIPs {
			if allowedIP.Address == i.Address {
				allowedIPs := make([]realm.AllowedIP, 0, 1)
				return allowedIPs, nil
			}
		}
		return nil, fmt.Errorf("unable to find allowed IP: %s", i.Address)
	}

	addressOptions := make([]string, 0, len(allowedIPs))
	allowedIPsByOption := map[string]realm.AllowedIP{}
	for _, allowedIP := range allowedIPs {
		addressOption := allowedIP.Address

		addressOptions = append(addressOptions, addressOption)
		allowedIPsByOption[addressOption] = allowedIP
	}

	var selections []string
	if err := ui.AskOne(
		&selections,
		&survey.MultiSelect{
			Message: "Which IP Addresse(s) or CIDR block(s) would you like to delete?",
			Options: addressOptions,
		},
	); err != nil {
		return nil, err
	}

	allowedIPs = make([]realm.AllowedIP, 0, len(selections))
	for _, selection := range selections {
		allowedIPs = append(allowedIPs, allowedIPsByOption[selection])
	}
	return allowedIPs, nil
}

func tableRowDelete(output deleteAllowedIPOutput, row map[string]interface{}) {
	deleted := false
	if output.err != nil {
		row[headerDetails] = output.err.Error()
	} else {
		deleted = true
	}
	row[headerDeleted] = deleted
}
