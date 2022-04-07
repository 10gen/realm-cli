package function

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

// CommandMetaRun is the command meta for the `function run` command
var CommandMetaRun = cli.CommandMeta{
	Use:         "run",
	Description: "Run a Function from your Realm app",
	HelpText: `Realm Functions allow you to define and execute server-side logic for your Realm
app. Once you select and run a Function for your Realm app, the following will
be displayed:
  - A list of logs, if present
  - The function result as a document
  - A list of error logs, if present`,
}

// CommandRun is the `function run` command
type CommandRun struct {
	inputs runInputs
}

// Flags is the command flags
func (cmd *CommandRun) Flags() []flags.Flag {
	return []flags.Flag{
		cli.AppFlagWithContext(&cmd.inputs.App, "to run its function"),
		cli.ProjectFlag(&cmd.inputs.Project),
		cli.ProductFlag(&cmd.inputs.Products),
		flags.StringFlag{
			Value: &cmd.inputs.Name,
			Meta: flags.Meta{
				Name:  "name",
				Usage: flags.Usage{Description: "Specify the name of the function to run"},
			},
		},
		flags.StringArrayFlag{
			Value: &cmd.inputs.Args,
			Meta: flags.Meta{
				Name: "args",
				Usage: flags.Usage{
					Description: "Specify the args to pass to your function",
					DocsLink:    "https://docs.mongodb.com/realm/functions/call-a-function/#call-from-realm-cli",
				},
			},
		},
		flags.StringFlag{
			Value: &cmd.inputs.User,
			Meta: flags.Meta{
				Name: "user",
				Usage: flags.Usage{
					Description:   "Specify which user to run the function as",
					DefaultValue:  "<none>",
					AllowedValues: []string{"<none>", "<userID>"},
					Note:          "Using <none> will run as the System user",
				},
			},
		},
	}
}

// Inputs is the command inputs
func (cmd *CommandRun) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *CommandRun) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	app, err := cli.ResolveApp(ui, clients.Realm, cli.AppOptions{
		AppMeta: cmd.inputs.AppMeta,
		Filter:  cmd.inputs.Filter(),
	})
	if err != nil {
		return err
	}

	function, err := cmd.inputs.resolveFunction(ui, clients.Realm, app.GroupID, app.ID)
	if err != nil {
		return err
	}

	args := make([]interface{}, 0, len(cmd.inputs.Args))
	if cmd.inputs.Args != nil {
		for _, arg := range cmd.inputs.Args {
			if isJSON(arg) {
				var argNew interface{}
				if err := json.Unmarshal([]byte(arg), &argNew); err != nil {
					return err
				}
				args = append(args, argNew)
				continue
			}
			if isInt(arg) {
				num, err := strconv.Atoi(arg)
				if err != nil {
					return err
				}
				args = append(args, num)
				continue
			}
			if isFloat(arg) {
				num, err := strconv.ParseFloat(arg, 64)
				if err != nil {
					return err
				}
				args = append(args, num)
				continue
			}
			args = append(args, arg)
		}
	}

	s := ui.Spinner(fmt.Sprintf("Running function %s with args %s...", cmd.inputs.Name, cmd.inputs.Args), terminal.SpinnerOptions{})

	runFunction := func() (realm.ExecutionResults, error) {
		s.Start()
		defer s.Stop()

		return clients.Realm.AppDebugExecuteFunction(app.GroupID, app.ID, cmd.inputs.User, function.Name, args)
	}

	response, err := runFunction()
	if err != nil {
		return err
	}
	if response.Logs != nil {
		ui.Print(terminal.NewListLog("Logs", response.Logs))
	}
	if response.ErrorLogs != nil {
		ui.Print(terminal.NewJSONLog("Error Logs", response.ErrorLogs))
	}
	ui.Print(terminal.NewJSONLog("Result", response.Result))

	return nil
}

func isJSON(data string) bool {
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(data), &obj); err == nil {
		return true
	}
	var list []interface{}
	if err := json.Unmarshal([]byte(data), &list); err == nil {
		return true
	}
	return false
}

func isInt(data string) bool {
	if _, err := strconv.Atoi(data); err != nil {
		return false
	}
	return true
}

func isFloat(data string) bool {
	if _, err := strconv.ParseFloat(data, 64); err != nil {
		return false
	}
	return true
}
