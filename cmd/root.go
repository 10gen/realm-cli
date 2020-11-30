package cmd

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"

	"github.com/spf13/cobra"
	"honnef.co/go/tools/version"
)

const (
	cliName = "realm-cli"
)

// Run runs the CLI
func Run() {
	cmd := &cobra.Command{
		Version:       version.Version,
		Use:           cliName,
		Short:         "CLI tool to manage your MongoDB Realm application",
		Long:          fmt.Sprintf("Use %s command help for information on a specific command", cliName),
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	factory := cli.NewCommandFactory()
	cobra.OnInitialize(factory.Setup)
	defer factory.Close()

	factory.SetGlobalFlags(cmd.PersistentFlags())

	cmd.AddCommand(factory.Build(cli.LoginCommand))
	cmd.AddCommand(factory.Build(cli.LogoutCommand))
	cmd.AddCommand(factory.Build(cli.WhoamiCommand))

	factory.Run(cmd)
}
