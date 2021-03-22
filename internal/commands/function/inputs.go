package function

import (
	"errors"
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
)

const (
	flagFunctionName      = "function"
	flagFunctionNameUsage = "specify the function to run"

	flagFunctionArgs      = "args"
	flagFunctionArgsUsage = "specify the args to pass to your function"

	flagAsUser      = "user"
	flagAsUserUsage = "specify the user to run the function as; defaults to system"
)

type inputs struct {
	cli.ProjectInputs
	Name string
	Args []string
	User string
}

func (i *inputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	if i.Name == "" {
		if err := ui.AskOne(&i.Name, &survey.Input{Message: "Function Name"}); err != nil {
			return err
		}
	}
	return nil
}

func (i *inputs) ResolveFunction(ui terminal.UI, client realm.Client, groupID, appID string) (realm.Function, error) {
	functions, err := client.Functions(groupID, appID)
	if err != nil {
		return realm.Function{}, err
	}

	switch len(functions) {
	case 0:
		return realm.Function{}, errors.New("failed to find function")
	case 1:
		return functions[0], nil
	}

	for _, function := range functions {
		if function.Name == i.Name {
			return function, nil
		}
	}

	functionsByOption := make(map[string]realm.Function, len(functions))
	functionOptions := make([]string, len(functions))
	for i, function := range functions {
		functionsByOption[function.Name] = function
		functionOptions[i] = function.Name
	}

	var selection string
	if err := ui.AskOne(&selection, &survey.Select{
		Message: "Select Function",
		Options: functionOptions,
	}); err != nil {
		return realm.Function{}, fmt.Errorf("failed to select function: %s", err)
	}
	return functionsByOption[selection], nil
}
