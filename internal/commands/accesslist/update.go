package accesslist

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/pflag"
)

type updateInputs struct {
	cli.ProjectInputs
	IPAddress    string
	NewIPAddress string
	Comment      string
}

// CommandMetaUpdate is the command meta for the `accesslist update` command
var CommandMetaUpdate = cli.CommandMeta{
	Use:     "update",
	Display: "accesslist update",
	Hidden:  true,
}

// CommandUpdate is the ip access update command
type CommandUpdate struct {
	inputs updateInputs
}

// Flags is the command flags
func (cmd *CommandUpdate) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)
	fs.StringVar(&cmd.inputs.IPAddress, flagIP, "", flagIPUsageUpdate)
	fs.StringVar(&cmd.inputs.NewIPAddress, flagNewIP, "", flagNewIPUsageUpdate)
	fs.StringVar(&cmd.inputs.Comment, flagComment, "", flagCommentUsageUpdate)
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

	accessList, err := clients.Realm.AllowedIPs(app.GroupID, app.ID)
	if err != nil {
		return err
	}

	allowedIP, err := cmd.inputs.resolveAllowedIP(ui, accessList)
	if err != nil {
		return err
	}

	var questions []*survey.Question

	if cmd.inputs.NewIPAddress == "" {
		questions = append(questions, &survey.Question{
			Name:   "new-ip",
			Prompt: &survey.Input{Message: "New IP Address"},
		})
	}

	if cmd.inputs.Comment == "" {
		questions = append(questions, &survey.Question{
			Name:   "comment",
			Prompt: &survey.Password{Message: "Comment"},
		})
	}

	if len(questions) > 0 {
		err := ui.Ask(cmd.inputs, questions...)
		if err != nil {
			return err
		}
	}

	if err := clients.Realm.AllowedIPUpdate(
		app.GroupID,
		app.ID,
		allowedIP.ID,
		cmd.inputs.NewIPAddress,
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
	return nil
}

func (i *updateInputs) resolveAllowedIP(ui terminal.UI, accessList realm.AccessList) (realm.AllowedIP, error) {
	if i.IPAddress != "" {
		for _, allowedIP := range accessList.AllowedIPs {
			if allowedIP.IPAddress == i.IPAddress {
				return allowedIP, nil
			}
		}
		return realm.AllowedIP{}, fmt.Errorf("unable to find allowed IP: %s", i.IPAddress)
	}

	selectableAllowedIPs := map[string]realm.AllowedIP{}
	selectableOptions := make([]string, len(accessList.AllowedIPs))
	for i, allowedIP := range accessList.AllowedIPs {
		option := allowedIP.ID
		selectableOptions[i] = option
		selectableAllowedIPs[option] = allowedIP
	}

	var selected string
	if err := ui.AskOne(
		&selected,
		&survey.Select{
			Message: "Which IP address would you like to update?",
			Options: selectableOptions,
		},
	); err != nil {
		return realm.AllowedIP{}, err
	}

	return selectableAllowedIPs[selected], nil
}
