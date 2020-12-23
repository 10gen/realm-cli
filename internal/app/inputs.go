package app

import (
	"errors"
	"fmt"

	"github.com/10gen/realm-cli/internal/cloud/atlas"
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

// ProjectInputs are the project/app inputs for a command
type ProjectInputs struct {
	Project string
	App     string
}

// Filter returns a realm.AppFlter based on the inputs
func (i ProjectInputs) Filter() realm.AppFilter { return realm.AppFilter{i.Project, i.App} }

// Flags registers the project app input flags to the provided flag set
func (i *ProjectInputs) Flags(fs *pflag.FlagSet) {
	fs.StringVar(&i.Project, flagProject, "", flagProjectUsage)
	fs.StringVar(&i.App, flagApp, "", flagAppUsage)
}

// Resolve resolves the necessary inputs that remain unset after flags have been parsed
func (i *ProjectInputs) Resolve(ui terminal.UI, wd string) error {
	config, configErr := ResolveConfig(wd)
	if configErr != nil {
		return configErr
	}

	var questions []*survey.Question
	if i.App == "" {
		questions = append(questions, &survey.Question{
			Name: appFieldName,
			Prompt: &survey.Input{
				Message: "App Filter",
				Default: getAppString(config),
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

// ErrAppNotFound is an app not found error
type ErrAppNotFound struct {
	app string
}

func (err ErrAppNotFound) Error() string {
	errMsg := "failed to find app"
	if err.app != "" {
		errMsg += fmt.Sprintf(" '%s'", err.app)
	}

	return errMsg
}

// set of known app input errors
var (
	ErrGroupNotFound = errors.New("failed to find group")
)

// Resolve will use the provided Realm client to resolve the app specified by the filter
func Resolve(ui terminal.UI, client realm.Client, filter realm.AppFilter) (realm.App, error) {
	apps, err := client.FindApps(filter)
	if err != nil {
		return realm.App{}, err
	}

	switch len(apps) {
	case 0:
		return realm.App{}, ErrAppNotFound{filter.App}
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

// ResolveGroupID will use the provided MongoDB Cloud Atlas client to resolve the user's group id
func ResolveGroupID(ui terminal.UI, client atlas.Client) (string, error) {
	groups, groupsErr := client.Groups()
	if groupsErr != nil {
		return "", groupsErr
	}

	switch len(groups) {
	case 0:
		return "", ErrGroupNotFound
	case 1:
		return groups[0].ID, nil
	}

	groupIDsByOption := make(map[string]string, len(groups))
	groupIDOptions := make([]string, len(groups))
	for i, group := range groups {
		option := getGroupString(group)

		groupIDsByOption[option] = group.ID
		groupIDOptions[i] = option
	}

	var selection string
	if err := ui.AskOne(&selection, &survey.Select{
		Message: "Atlas Project",
		Options: groupIDOptions,
	}); err != nil {
		return "", fmt.Errorf("failed to select group id: %s", err)
	}

	return groupIDsByOption[selection], nil
}

func getAppString(config Config) string {
	if config.ID == "" {
		return config.Name
	}
	return config.ID
}

func getGroupString(group atlas.Group) string {
	return fmt.Sprintf("%s - %s", group.Name, group.ID)
}
