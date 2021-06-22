package cli

import (
	"errors"
	"fmt"

	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
)

// ProjectInputs are the project/app inputs for a command
type ProjectInputs struct {
	Project  string
	App      string
	Products []string
}

// Filter returns a realm.AppFlter based on the inputs
func (i ProjectInputs) Filter() realm.AppFilter {
	return realm.AppFilter{i.Project, i.App, i.Products}
}

// Resolve resolves the necessary inputs that remain unset after flags have been parsed
func (i *ProjectInputs) Resolve(ui terminal.UI, wd string, skipAppPrompt bool) error {
	app, appErr := local.LoadAppConfig(wd)
	if appErr != nil {
		return appErr
	}

	if i.App == "" {
		var appOption string

		if app.RootDir != "" {
			appOption = app.Option()
		} else {
			if !skipAppPrompt {
				if err := ui.AskOne(&appOption, &survey.Input{Message: "App ID or Name"}); err != nil {
					return err
				}
			}
		}
		i.App = appOption
	}

	return nil
}

// ErrAppNotFound is an app not found error
type ErrAppNotFound struct {
	App string
}

func (err ErrAppNotFound) Error() string {
	errMsg := "failed to find app"
	if err.App != "" {
		errMsg += fmt.Sprintf(" '%s'", err.App)
	}

	return errMsg
}

// set of known app input errors
var (
	ErrGroupNotFound = errors.New("failed to find group")
)

// ResolveApp will use the provided Realm client to resolve the app specified by the filter
func ResolveApp(ui terminal.UI, client realm.Client, filter realm.AppFilter) (realm.App, error) {
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
		appsByOption[app.Option()] = app
		appOptions[i] = app.Option()
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

func getGroupString(group atlas.Group) string {
	return fmt.Sprintf("%s - %s", group.Name, group.ID)
}
