package accesslist

import (
	"errors"
	"fmt"
	"sort"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/AlecAivazis/survey/v2"
)

const (
	headerDeleted = "Deleted"
	headerDetails = "Details"
)

var (
	deleteTableHeaders = []string{headerAddress, headerComment, headerDeleted, headerDetails}
)

type deleteAllowedIPOutputs []deleteAllowedIPOutput

type deleteAllowedIPOutput struct {
	allowedIP realm.AllowedIP
	err       error
}
type deleteInputs struct {
	cli.ProjectInputs
	Addresses []string
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
		cli.AppFlagWithContext(&cmd.inputs.App, "to remove an entry in its Access List"),
		cli.ProjectFlag(&cmd.inputs.Project),
		cli.ProductFlag(&cmd.inputs.Products),
		flags.StringSliceFlag{
			Value: &cmd.inputs.Addresses,
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

	tableRows := make([]map[string]interface{}, 0, len(outputs))
	for _, output := range outputs {
		row := map[string]interface{}{
			headerAddress: output.allowedIP.Address,
			headerComment: output.allowedIP.Comment,
		}
		tableRows = append(tableRows, row)
		tableRowDelete(output, row)
	}

	ui.Print(terminal.NewTableLog(
		fmt.Sprintf("Deleted %d IP address(es) and CIDR block(s)", len(outputs)),
		deleteTableHeaders,
		tableRows...,
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
	if len(allowedIPs) == 0 {
		return nil, nil
	}

	if len(i.Addresses) > 0 {
		allowedIPsByAddress := make(map[string]realm.AllowedIP, len(allowedIPs))
		for _, allowedIP := range allowedIPs {
			allowedIPsByAddress[allowedIP.Address] = allowedIP
		}

		allowedIPs := make([]realm.AllowedIP, 0, len(i.Addresses))
		for _, identifier := range i.Addresses {
			if allowedIP, ok := allowedIPsByAddress[identifier]; ok {
				allowedIPs = append(allowedIPs, allowedIP)
			}
		}

		if len(allowedIPs) == 0 {
			return nil, errors.New("unable to find allowed IPs")
		}

		return allowedIPs, nil
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
			Message: "Which IP Address(es) or CIDR block(s) would you like to delete?",
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
	row[headerAddress] = output.allowedIP.Address
	row[headerComment] = output.allowedIP.Comment
	row[headerDeleted] = deleted
}
