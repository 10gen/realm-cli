package app

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/spf13/pflag"
)

// CommandMetaList is the command meta for the `app list` command
var CommandMetaList = cli.CommandMeta{
	Use:         "list",
	Aliases:     []string{"ls"},
	Display:     "apps list",
	Description: "List the Realm apps you have access to",
	HelpText:    `Lists and filters your Realm apps.`,
}

// TODO(REALMC-9256): this should not be duplicated (with cli.ProjectInputs)
const (
	flagListApp      = "app"
	flagListAppShort = "a"
	flagListAppUsage = "Filter the list of Realm apps by name"

	flagListProject      = "project"
	flagListProjectUsage = "Specify the ID of a MongoDB Atlas project"

	flagListProduct      = "product"
	flagListProductUsage = `Specify the Realm app product(s) (Allowed values: "standard", "atlas")`
)

// CommandList is the `app list` command
type CommandList struct {
	inputs cli.ProjectInputs
}

// Flags is the command flags
func (cmd *CommandList) Flags(fs *pflag.FlagSet) {
	fs.StringVarP(&cmd.inputs.App, flagListApp, flagListAppShort, "", flagListAppUsage)

	fs.StringVar(&cmd.inputs.Project, flagListProject, "", flagListProjectUsage)
	flags.MarkHidden(fs, flagProject)

	fs.StringSliceVar(&cmd.inputs.Products, flagListProduct, []string{}, flagListProductUsage)
	flags.MarkHidden(fs, flagListProduct)
}

// Handler is the command handler
func (cmd *CommandList) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	apps, err := clients.Realm.FindApps(cmd.inputs.Filter())
	if err != nil {
		return err
	}

	if len(apps) == 0 {
		ui.Print(terminal.NewTextLog("No available apps to show"))
		return nil
	}

	rows := make([]interface{}, 0, len(apps))
	for _, app := range apps {
		rows = append(rows, app.Option())
	}
	ui.Print(terminal.NewListLog(fmt.Sprintf("Found %d apps", len(rows)), rows...))
	return nil
}
