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
	AppMeta  local.AppMeta
}

// Filter returns a realm.AppFlter based on the inputs
func (i ProjectInputs) Filter() realm.AppFilter {
	return realm.AppFilter{i.Project, i.App, i.Products}
}

// Resolve resolves the necessary inputs that remain unset after flags have been parsed
func (i *ProjectInputs) Resolve(ui terminal.UI, wd string, skipAppPrompt bool) error {
	app, _, err := local.FindApp(wd)
	if err != nil {
		return err
	}

	if i.App == "" {
		var appOption string

		if app.RootDir != "" {
			i.AppMeta = app.Meta
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

// AppOptions are the parameters for resolving app
type AppOptions struct {
	AppMeta      local.AppMeta
	Filter       realm.AppFilter
	FetchDetails bool
}

// ResolveApp will use the provided Realm client to resolve the app specified by the options
func ResolveApp(ui terminal.UI, client realm.Client, opts AppOptions) (realm.App, error) {
	// if no flags were set and the optional app meta file is present
	if opts.Filter.GroupID == "" && opts.Filter.App == "" && opts.AppMeta.ConfigVersion != 0 {
		if opts.FetchDetails {
			return client.FindApp(opts.AppMeta.GroupID, opts.AppMeta.AppID)
		}
		return realm.App{ID: opts.AppMeta.AppID, GroupID: opts.AppMeta.GroupID}, nil
	}

	apps, err := client.FindApps(opts.Filter)
	if err != nil {
		return realm.App{}, err
	}

	switch len(apps) {
	case 0:
		return realm.App{}, ErrAppNotFound{opts.Filter.App}
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
	groups, groupsErr := atlas.AllGroups(client)
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
