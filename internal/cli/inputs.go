package cli

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/pflag"
)

const (
	// project app inputs field names, per survey
	appFieldName = "app"

	flagApp      = "app"
	flagAppUsage = "the Realm app name or id to manage"

	flagProject      = "project"
	flagProjectUsage = "the MongoDB cloud project id"
)

// ProjectAppInputs are the project/app inputs for a command
type ProjectAppInputs struct {
	Project string
	App     string
}

// Filter returns a realm.AppFlter based on the inputs
func (i ProjectAppInputs) Filter() realm.AppFilter { return realm.AppFilter{i.Project, i.App} }

// Flags registers the project app input flags to the provided flag set
func (i *ProjectAppInputs) Flags(fs *pflag.FlagSet) {
	fs.StringVar(&i.Project, flagProject, "", flagProjectUsage)
	fs.StringVar(&i.App, flagApp, "", flagAppUsage)
}

// Resolve resolves the necessary inputs that remain unset after flags have been parsed
func (i *ProjectAppInputs) Resolve(ui terminal.UI, wd string) error {
	appData, appDataErr := ResolveAppData(wd)
	if appDataErr != nil {
		return appDataErr
	}

	var questions []*survey.Question
	if i.App == "" {
		questions = append(questions, &survey.Question{
			Name: appFieldName,
			Prompt: &survey.Input{
				Message: "App Filter",
				Default: getAppString(appData),
			},
		})
	}

	if len(questions) > 0 {
		if err := ui.Ask(i, questions...); err != nil {
			return err
		}
	}

	return nil
}

// ResolveApp will use the provided Realm client to resolve the app specified by the filter
func ResolveApp(ui terminal.UI, client realm.Client, filter realm.AppFilter) (realm.App, error) {
	apps, err := client.FindApps(filter)
	if err != nil {
		return realm.App{}, err
	}

	switch len(apps) {
	case 0:
		return realm.App{}, fmt.Errorf("failed to find app '%s'", filter.App)
	case 1:
		return apps[0], nil
	}

	appsByOption := make(map[string]realm.App, len(apps))
	appOptions := make([]string, len(apps))
	for i, app := range apps {
		appsByOption[app.String()] = app
		appOptions[i] = app.String()
	}

	var selection string
	if err := ui.AskOne(&selection, &survey.Select{
		Message: "Select App",
		Options: appOptions,
	}); err != nil {
		return realm.App{}, fmt.Errorf("failed to select app: %s", err)
	}
	return appsByOption[selection], nil
}

func getAppString(appData AppData) string {
	if appData.ID == "" {
		return appData.Name
	}
	return appData.ID
}
