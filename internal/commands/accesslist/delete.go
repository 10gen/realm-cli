package accesslist

import (
	"fmt"
	"sort"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/pflag"
)

type deleteInputs struct {
	cli.ProjectInputs
	IPAddress string
}

// CommandMetaDelete is the command meta for the `accesslist delete` command
var CommandMetaDelete = cli.CommandMeta{
	Use:     "delete",
	Display: "accesslist delete",
	Hidden:  true,
}

// CommandDelete for the ip access delete command
type CommandDelete struct {
	inputs deleteInputs
}

// Flags is the command flags
func (cmd *CommandDelete) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)
	fs.StringVar(&cmd.inputs.IPAddress, flagIP, "", flagIPUsageDelete)
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
		ui.Print(terminal.NewTextLog("No ip addresses to delete"))
		return nil
	}

	outputs := make(allowedIPOutputs, len(selected))
	for i, allowedIP := range selected {
		err := clients.Realm.AllowedIPDelete(app.GroupID, app.ID, allowedIP.ID)
		outputs[i] = allowedIPOutput{allowedIP, err}
	}

	sort.SliceStable(outputs, func(i, j int) bool {
		return outputs[i].err != nil && outputs[j].err == nil
	})

	ui.Print(terminal.NewTableLog(
		fmt.Sprintf("Deleted %d ip address(es)", len(outputs)),
		tableHeaders,
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

func (i *deleteInputs) resolveAllowedIP(ui terminal.UI, accessList realm.AccessList) ([]realm.AllowedIP, error) {
	if i.IPAddress != "" {
		for _, allowedIP := range accessList.AllowedIPs {
			if allowedIP.IPAddress == i.IPAddress {
				allowedIPs := make([]realm.AllowedIP, 0, 1)
				return allowedIPs, nil
			}
		}
		return nil, fmt.Errorf("unable to find allowed IP: %s", i.IPAddress)
	}

	ipOptions := make([]string, 0, len(accessList.AllowedIPs))
	allowedIPsByOption := map[string]realm.AllowedIP{}
	for _, allowedIP := range accessList.AllowedIPs {
		ipOption := allowedIP.IPAddress

		ipOptions = append(ipOptions, ipOption)
		allowedIPsByOption[ipOption] = allowedIP
	}
	var selections []string
	if err := ui.AskOne(
		&selections,
		&survey.MultiSelect{
			Message: "Which ip address(es) would you like to delete?",
			Options: ipOptions,
		},
	); err != nil {
		return nil, err
	}

	allowedIPs := make([]realm.AllowedIP, 0, len(selections))
	for _, selection := range selections {
		allowedIPs = append(allowedIPs, allowedIPsByOption[selection])
	}

	return allowedIPs, nil
}

func tableRowDelete(output allowedIPOutput, row map[string]interface{}) {
	deleted := false
	if output.err != nil {
		row[headerDetails] = output.err.Error()
	} else {
		deleted = true
	}
	row[headerDeleted] = deleted
}
