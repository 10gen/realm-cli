package e2e

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"

	"github.com/AlecAivazis/survey/v2/core"
	"github.com/Netflix/go-expect"
)

// TestCLI is responsible for ensuring each command can successfully compile and execute
func TestCLI(t *testing.T) {
	core.DisableColor = true
	defer func() { core.DisableColor = false }()

	// run the command once to make sure all dependencies are installed
	cmd := exec.Command("go", "run", "../main.go", "--help")
	assert.Nil(t, cmd.Start())
	assert.Nil(t, cmd.Wait())

	for _, tc := range []struct {
		description string
		args        []string
		firstLine   string
	}{
		{
			description: "the root command",
			firstLine:   `Use "realm-cli [command] --help" for information on a specific command`,
		},
		{
			description: "the whoami command",
			args:        []string{"whoami"},
			firstLine:   "Display information about the current user",
		},
		{
			description: "the login command",
			args:        []string{"login"},
			firstLine:   "Log the CLI into Realm using a MongoDB Cloud API key",
		},
		{
			description: "the logout command",
			args:        []string{"logout"},
			firstLine:   "Log the CLI out of Realm",
		},
		{
			description: "the push command",
			args:        []string{"push"},
			firstLine:   "Imports and deploys changes from your local directory to your Realm app",
		},
		{
			description: "the pull command",
			args:        []string{"pull"},
			firstLine:   "Exports the latest version of your Realm app into your local directory",
		},
		{
			description: "the app init command",
			args:        []string{"app", "init"},
			firstLine:   "Initialize a Realm app in your current working directory",
		},
		{
			description: "the app create command",
			args:        []string{"app", "create"},
			firstLine:   "Create a new app from your current working directory and deploy it to the Realm server",
		},
		{
			description: "the app list command",
			args:        []string{"app", "list"},
			firstLine:   "List the Realm apps you have access to",
		},
		{
			description: "the app delete command",
			args:        []string{"app", "delete"},
			firstLine:   "Delete a Realm app",
		},
		{
			description: "the app diff command",
			args:        []string{"app", "diff"},
			firstLine:   "Show differences between your local directory and your Realm app",
		},
		{
			description: "the app describe command",
			args:        []string{"app", "describe"},
			firstLine:   "Displays information about your Realm app",
		},
		{
			description: "the user create command",
			args:        []string{"user", "create"},
			firstLine:   "Create an application user for your Realm app",
		},
		{
			description: "the user list command",
			args:        []string{"user", "list"},
			firstLine:   "List the application users of your Realm app",
		},
		{
			description: "the user disable command",
			args:        []string{"user", "disable"},
			firstLine:   "Disable an application User of your Realm app",
		},
		{
			description: "the user enable command",
			args:        []string{"user", "enable"},
			firstLine:   "Enable an application User of your Realm app",
		},
		{
			description: "the user revoke command",
			args:        []string{"user", "revoke"},
			firstLine:   "Revoke an application User’s sessions from your Realm app",
		},
		{
			description: "the user delete command",
			args:        []string{"user", "delete"},
			firstLine:   "Delete an application user from your Realm app",
		},
		{
			description: "the secrets create command",
			args:        []string{"secrets", "create"},
			firstLine:   "Create a Secret for your Realm app",
		},
		{
			description: "the secrets list command",
			args:        []string{"secrets", "list"},
			firstLine:   "List the Secrets in your Realm app",
		},
		{
			description: "the secrets update command",
			args:        []string{"secrets", "update"},
			firstLine:   "Update a Secret in your Realm app",
		},
		{
			description: "the secrets delete command",
			args:        []string{"secrets", "delete"},
			firstLine:   "Delete a Secret from your Realm app",
		},
		{
			description: "the function run command",
			args:        []string{"function", "run"},
			firstLine:   "Run a Function from your Realm app",
		},
		{
			description: "the logs list command",
			args:        []string{"logs", "list"},
			firstLine:   "Lists the Logs in your Realm app",
		},
		{
			description: "the schema datamodels command",
			args:        []string{"schema", "datamodels"},
			firstLine:   "Generate data models based on your Schema",
		},
		{
			description: "the accesslist create command",
			args:        []string{"accesslist", "create"},
			firstLine:   "Create an IP address or CIDR block in the Access List for your Realm app",
		},
		{
			description: "the accesslist list command",
			args:        []string{"accesslist", "list"},
			firstLine:   "List the allowed entries in the Access List of your Realm app",
		},
		{
			description: "the accesslist update command",
			args:        []string{"accesslist", "update"},
			firstLine:   "Modify an IP address or CIDR block in the Access List of your Realm app",
		},
		{
			description: "the accesslist delete command",
			args:        []string{"accesslist", "delete"},
			firstLine:   "Delete an IP address or CIDR block from the Access List of your Realm app",
		},
	} {
		t.Run("should display help text for "+tc.description, func(t *testing.T) {
			out := new(bytes.Buffer)

			console, err := expect.NewConsole(expect.WithStdout(out))
			assert.Nil(t, err)

			go func() {
				console.ExpectEOF()
			}()

			args := make([]string, 3+len(tc.args))
			args[0] = "run"
			args[1] = "../main.go"
			for i, a := range tc.args {
				args[i+2] = a
			}
			args[len(args)-1] = "--help"

			cmd := exec.Command("go", args...)
			cmd.Stdin = console.Tty()
			cmd.Stdout = console.Tty()
			cmd.Stderr = console.Tty()

			assert.Nil(t, cmd.Start())
			assert.Nil(t, cmd.Wait())

			console.Close() // flush the writers

			assert.True(t,
				strings.HasPrefix(out.String(), tc.firstLine),
				"expected %s help text to begin with:\n%s\ninstead it was:\n%s",
				tc.description, tc.firstLine, out.String(),
			)
		})
	}

}
