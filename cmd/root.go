package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/commands"

	"github.com/spf13/cobra"
)

// Run runs the CLI
func Run() {
	// print commands in help/usage text in the order they are declared
	cobra.EnableCommandSorting = false

	cmd := &cobra.Command{
		Version:       cli.Version,
		Use:           cli.Name,
		Short:         "CLI tool to manage your MongoDB Realm application",
		Long:          fmt.Sprintf(`Use "%s [command] --help" for information on a specific command`, cli.Name),
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	factory, err := cli.NewCommandFactory()
	if err != nil {
		log.Fatal(err)
	}

	cobra.OnInitialize(factory.Setup)

	cmd.Flags().SortFlags = false // ensures CLI help text displays global flags unsorted
	factory.SetGlobalFlags(cmd.PersistentFlags())

	cmd.AddCommand(factory.Build(commands.Whoami))
	cmd.AddCommand(factory.Build(commands.Login))
	cmd.AddCommand(factory.Build(commands.Logout))
	cmd.AddCommand(factory.Build(commands.Push))
	cmd.AddCommand(factory.Build(commands.Pull))
	cmd.AddCommand(factory.Build(commands.App))
	cmd.AddCommand(factory.Build(commands.User))
	cmd.AddCommand(factory.Build(commands.Secrets))
	cmd.AddCommand(factory.Build(commands.Logs))
	cmd.AddCommand(factory.Build(commands.Function))
	cmd.AddCommand(factory.Build(commands.Schema))
	cmd.AddCommand(factory.Build(commands.AccessList))
	cmd.AddCommand(factory.Build(commands.Profiles))

	os.Exit(factory.Run(cmd))
}
