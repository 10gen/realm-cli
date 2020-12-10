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
	flagAppUsage = "this is the --app usage"

	flagProject      = "project"
	flagProjectUsage = "this is the --project usage"
)

// ProjectAppInputs are the project/app inputs for a command
type ProjectAppInputs struct {
	Project string
	App     string
}

// Flags registers the project app input flags to the provided flag set
func (i *ProjectAppInputs) Flags(fs *pflag.FlagSet) {
	fs.StringVar(&i.Project, flagProject, "", flagProjectUsage)
	fs.StringVar(&i.App, flagApp, "", flagAppUsage)
}

// Resolve resolves the necessary inputs that remain unset after flags have been parsed
func (i *ProjectAppInputs) Resolve(ui terminal.UI, appData AppData) error {
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

// ResolveApp will use the provided Realm client to resolve the app specified by the inputs
func (i ProjectAppInputs) ResolveApp(ui terminal.UI, client realm.Client) (realm.App, error) {
	apps, err := client.FindApps(realm.AppFilter{i.Project, i.App})
	if err != nil {
		return realm.App{}, err
	}

	switch len(apps) {
	case 0:
		return realm.App{}, fmt.Errorf("failed to find app '%s'", i.App)
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
