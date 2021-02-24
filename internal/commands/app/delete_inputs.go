package app

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
)

var (
	flagApps      = "apps"
	flagAppsUsage = "the Realm app names or ids to manage"
)

type deleteInputs struct {
	Apps    []string
	Project string
}

func (inputs *deleteInputs) resolveApps(ui terminal.UI, client realm.Client) ([]realm.App, error) {
	apps, err := client.FindApps(realm.AppFilter{GroupID: inputs.Project})
	if err != nil {
		return nil, err
	}

	if len(inputs.Apps) > 0 && len(apps) == 0 {
		return nil, cli.ErrAppNotFound{}
	}

	if len(apps) == 0 {
		return apps, nil
	}

	foundAppNames := map[string]bool{}
	for _, app := range apps {
		foundAppNames[app.Name] = true
	}

	inputAppsNotFound := make([]string, 0)
	for _, inputApp := range inputs.Apps {
		if !foundAppNames[inputApp] {
			inputAppsNotFound = append(inputAppsNotFound, inputApp)
		}
	}

	if len(inputAppsNotFound) > 0 {
		return nil, fmt.Errorf("Failed to find the following apps: %v", inputAppsNotFound)
	}

	if len(inputs.Apps) > 0 {
		return apps, nil
	}

	appsByOption := make(map[string]realm.App, len(apps))
	appOptions := make([]string, len(apps))
	for i, app := range apps {
		appsByOption[app.Option()] = app
		appOptions[i] = app.Option()
	}

	var selectedApps []string
	if err := ui.AskOne(&selectedApps, &survey.MultiSelect{
		Message: "Select App(s)",
		Options: appOptions,
	}); err != nil {
		return nil, fmt.Errorf("failed to select app(s): %s", err)
	}
	selected := make([]realm.App, len(selectedApps))
	for idx, app := range selectedApps {
		selected[idx] = appsByOption[app]
	}
	return selected, nil
}
