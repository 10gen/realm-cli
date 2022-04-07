package accesslist

import (
	"errors"
	"fmt"
	"strings"

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
prompted to select an IP address or CIDR block if none are provided in the
initial command.`,
}

// CommandDelete is the `accesslist delete` command
type CommandDelete struct {
	inputs deleteInputs
}

// Flags is the command flags
func (cmd *CommandDelete) Flags() []flags.Flag {
	return []flags.Flag{
		cli.AppFlagWithContext(&cmd.inputs.App, "to remove entries from its Access List"),
		cli.ProjectFlag(&cmd.inputs.Project),
		cli.ProductFlag(&cmd.inputs.Products),
		flags.StringSliceFlag{
			Value: &cmd.inputs.Addresses,
			Meta: flags.Meta{
				Name: "ip",
				Usage: flags.Usage{
					Description: "Specify the IP address(es) or CIDR block(s) to delete",
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
	app, err := cli.ResolveApp(ui, clients.Realm, cli.AppOptions{
		AppMeta: cmd.inputs.AppMeta,
		Filter:  cmd.inputs.Filter(),
	})
	if err != nil {
		return err
	}

	allowedIPs, err := clients.Realm.AllowedIPs(app.GroupID, app.ID)
	if err != nil {
		return err
	}

	selectedAllowedIPs, err := cmd.inputs.resolveAllowedIP(ui, allowedIPs)
	if err != nil {
		return err
	}

	if len(selectedAllowedIPs) == 0 {
		return errors.New("no IP addresses or CIDR blocks to delete")
	}

	outputs := make([]deleteAllowedIPOutput, len(selectedAllowedIPs))
	for i, allowedIP := range selectedAllowedIPs {
		err := clients.Realm.AllowedIPDelete(app.GroupID, app.ID, allowedIP.ID)
		outputs[i] = deleteAllowedIPOutput{allowedIP, err}
	}

	tableRows := make([]map[string]interface{}, 0, len(outputs))
	for _, output := range outputs {
		row := map[string]interface{}{
			headerAddress: output.allowedIP.Address,
			headerComment: output.allowedIP.Comment,
		}
		var deleted bool
		var err error
		if output.err != nil {
			err = output.err
		} else {
			deleted = true
		}

		row[headerDeleted] = deleted
		row[headerDetails] = err
		tableRows = append(tableRows, row)
	}

	ui.Print(terminal.NewTableLog(
		fmt.Sprintf("Deleted %d IP address(es) and/or CIDR block(s)", len(outputs)),
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

		selectedAllowedIPs := make([]realm.AllowedIP, 0, len(i.Addresses))
		var notFoundAddresses []string
		for _, address := range i.Addresses {
			allowedIP, ok := allowedIPsByAddress[address]
			if !ok {
				notFoundAddresses = append(notFoundAddresses, address)
				continue
			}
			selectedAllowedIPs = append(selectedAllowedIPs, allowedIP)
		}

		if len(notFoundAddresses) > 0 {
			return nil, errors.New(
				"unable to find IP address(es) and/or CIDR block(s): " +
					strings.Join(notFoundAddresses, ", "),
			)
		}

		return selectedAllowedIPs, nil
	}

	// user did not provide any IP addresses and/or CIDR blocks to delete
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
			Message: "Which IP Address(es) and/or CIDR block(s) would you like to delete?",
			Options: addressOptions,
		},
	); err != nil {
		return nil, err
	}

	selectedAllowedIPs := make([]realm.AllowedIP, 0, len(selections))
	for _, selection := range selections {
		selectedAllowedIPs = append(selectedAllowedIPs, allowedIPsByOption[selection])
	}
	return selectedAllowedIPs, nil
}
