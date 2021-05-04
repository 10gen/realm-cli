package function

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/briandowns/spinner"
	"github.com/spf13/pflag"
)

// CommandRun is the `function run` command
type CommandRun struct {
	inputs inputs
}

// Flags is the command flags
func (cmd *CommandRun) Flags(fs *pflag.FlagSet) {
	cmd.inputs.Flags(fs)

	fs.StringVar(&cmd.inputs.Name, flagFunctionName, "", flagFunctionNameUsage)
	fs.StringArrayVar(&cmd.inputs.Args, flagFunctionArgs, nil, flagFunctionArgsUsage)
	fs.StringVar(&cmd.inputs.User, flagAsUser, "", flagAsUserUsage)
}

// Inputs is the command inputs
func (cmd *CommandRun) Inputs() cli.InputResolver {
	return &cmd.inputs
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

// Handler is the command handler
func (cmd *CommandRun) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	app, err := cli.ResolveApp(ui, clients.Realm, cmd.inputs.Filter())
	if err != nil {
		return err
	}

	function, err := cmd.inputs.ResolveFunction(ui, clients.Realm, app.GroupID, app.ID)
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

	s := spinner.New(terminal.SpinnerCircles, 250*time.Millisecond)
	s.Suffix = fmt.Sprintf(" Running function %s with args %s...", cmd.inputs.Name, cmd.inputs.Args)

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
