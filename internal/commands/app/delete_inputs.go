package app

import (
	"strings"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
)

var (
	flagApp      = "app"
	flagAppShort = "a"
	flagAppUsage = "the remote Realm app name or id"
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

	if len(apps) == 0 {
		if len(inputs.Apps) == 0 {
			return nil, nil
		}
		return nil, cli.ErrAppNotFound{}
	}

	if len(inputs.Apps) > 0 {
		appsByName := map[string]realm.App{}
		for _, app := range apps {
			appsByName[app.Name] = app
		}

		missingApps := make([]string, 0)
		appsFiltered := make([]realm.App, 0, len(inputs.Apps))
		for _, inputApp := range inputs.Apps {
			app, ok := appsByName[inputApp]
			if !ok {
				missingApps = append(missingApps, inputApp)
				continue
			}
			appsFiltered = append(appsFiltered, app)
		}

		if len(missingApps) > 0 {
			ui.Print(terminal.NewWarningLog(
				"Unable to delete certain apps because they were not found: %s",
				strings.Join(missingApps, ", "),
			))
		}
		return appsFiltered, nil
	}

	appsByOption := make(map[string]realm.App, len(apps))
	appOptions := make([]string, len(apps))
	for i, app := range apps {
		appsByOption[app.Option()] = app
		appOptions[i] = app.Option()
	}

	var selected []string
	if err := ui.AskOne(&selected, &survey.MultiSelect{
		Message: "Select App(s)",
		Options: appOptions,
	}); err != nil {
		return nil, err
	}
	selectedApps := make([]realm.App, len(selected))
	for idx, app := range selected {
		selectedApps[idx] = appsByOption[app]
	}
	return selectedApps, nil
}
