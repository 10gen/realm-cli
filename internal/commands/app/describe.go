package app

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/spf13/pflag"
)

// set of supported `app describe` command strings
const (
	CommandDescribeUse = "app describe"
)

// set of supported `app describe` command strings
var (
	CommandDescribeAliases = []string{}
)

type describeInputs struct {
	cli.ProjectInputs
}

func (i *describeInputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	return nil
}

// CommandDescribe is the `app describe` command
type CommandDescribe struct {
	inputs describeInputs
}

// Flags is the command flags
func (cmd *CommandDescribe) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)
}

// Inputs is the command inputs
func (cmd *CommandDescribe) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *CommandDescribe) Handler(profile *cli.Profile, ui terminal.UI, clients cli.Clients) error {
	if cmd.inputs.Project == "" && cmd.inputs.App == "" {
		projectID, err := cli.ResolveGroupID(ui, clients.Atlas)
		if err != nil {
			return err
		}
		cmd.inputs.Project = projectID
	}
	app, err := cli.ResolveApp(ui, clients.Realm, cmd.inputs.Filter())
	if err != nil {
		return err
	}

	appDesc, err := clients.Realm.AppDescription(app.GroupID, app.ID)
	if err != nil {
		return err
	}

	ui.Print(terminal.NewJSONLog("App Description", appDesc))

	return nil
}
