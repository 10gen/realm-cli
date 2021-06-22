package function

import (
	"errors"
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
)

type runInputs struct {
	cli.ProjectInputs
	Name string
	Args []string
	User string
}

func (i *runInputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	return i.ProjectInputs.Resolve(ui, profile.WorkingDirectory, true)
}

func (i *runInputs) resolveFunction(ui terminal.UI, client realm.Client, groupID, appID string) (realm.Function, error) {
	functions, err := client.Functions(groupID, appID)
	if err != nil {
		return realm.Function{}, err
	}

	if len(functions) == 0 {
		return realm.Function{}, errors.New("no functions available to run")
	}

	if i.Name != "" {
		for _, function := range functions {
			if function.Name == i.Name {
				return function, nil
			}
		}
		return realm.Function{}, fmt.Errorf("failed to find function '%s'", i.Name)
	}

	if len(functions) == 1 {
		return functions[0], nil
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
