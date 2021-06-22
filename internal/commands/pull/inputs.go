package pull

import (
	"archive/zip"
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey/v2"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/mitchellh/go-homedir"
)

const (
	flagLocalPath      = "local"
	flagLocalPathUsage = "Specify a local filepath to export a Realm app to"

	flagRemote      = "remote"
	flagRemoteUsage = "Specify the name or ID of a remote Realm app to export"

	flagIncludeDependencies      = "include-dependencies"
	flagIncludeDependenciesShort = "d"
	flagIncludeDependenciesUsage = "Export and include Realm app dependencies"

	flagIncludeHosting      = "include-hosting"
	flagIncludeHostingShort = "s"
	flagIncludeHostingUsage = "Export and include Realm app hosting files"

	flagDryRun      = "dry-run"
	flagDryRunShort = "x"
	flagDryRunUsage = "Run without writing any changes to the local filepath"

	flagProject      = "project"
	flagProjectUsage = "Specify the MongoDB Cloud project ID"

	flagConfigVersion      = "config-version"
	flagConfigVersionUsage = "Specify the app config version to export as"

	flagTemplate      = "template"
	flagTemplateShort = "t"
	flagTemplateUsage = "specify the template ID that is used for this app"
)

var (
	errConfigVersionMismatch = errors.New("must export an app with the same config version as found in the current project directory")
)

type inputs struct {
	Project             string
	RemoteApp           string
	LocalPath           string
	AppVersion          realm.AppConfigVersion
	IncludeDependencies bool
	IncludeHosting      bool
	DryRun              bool
	TemplateID          string
}

func (i *inputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	wd := i.LocalPath
	if wd == "" {
		wd = profile.WorkingDirectory
	}

	app, appErr := local.LoadAppConfig(wd)
	if appErr != nil {
		return appErr
	}

	var pathLocal string
	if i.LocalPath == "" {
		pathLocal = app.RootDir
	} else {
		l, err := homedir.Expand(i.LocalPath)
		if err != nil {
			return err
		}
		pathLocal = l
	}
	i.LocalPath = pathLocal

	if app.RootDir != "" {
		if i.AppVersion == realm.AppConfigVersionZero {
			i.AppVersion = app.ConfigVersion()
		} else if i.AppVersion != app.ConfigVersion() && app.ConfigVersion() != realm.AppConfigVersionZero {
			return errConfigVersionMismatch
		}

		if i.RemoteApp == "" {
			i.RemoteApp = app.Option()
		}
	}

	return nil
}

func (i *inputs) resolveRemoteApp(ui terminal.UI, clients cli.Clients) (realm.App, error) {
	if i.Project == "" {
		groupID, err := cli.ResolveGroupID(ui, clients.Atlas)
		if err != nil {
			return realm.App{}, err
		}
		i.Project = groupID
	}

	app, err := cli.ResolveApp(ui, clients.Realm, realm.AppFilter{GroupID: i.Project, App: i.RemoteApp})
	if err != nil {
		if _, ok := err.(cli.ErrAppNotFound); ok {
			return realm.App{}, errProjectNotFound{}
		}
		return realm.App{}, err
	}

	return app, nil
}

func (i *inputs) resolveTemplate(ui terminal.UI, realmClient realm.Client, groupID, appID string) (*zip.Reader, []string,  error) {
	if i.TemplateID == "" {
		return nil, nil, nil
	}

	var paths []string
	if !ui.AutoConfirm() {
		selectablePaths := []string{
			frontendPath,
			backendPath,
		}
		if err := ui.AskOne(
			&paths,
			&survey.MultiSelect{
				Message: "Where would you like to export the template?",
				Options: selectablePaths,
			}); err != nil {
			return nil, nil, err
		}
	} else {
		paths = append(paths, frontendPath)
	}

	compatibleTemplates, err := realmClient.CompatibleTemplates(groupID, appID)
	if err != nil {
		return nil, nil, err
	}

	for _, template := range compatibleTemplates {
		// This is the ID that they selected
		if template.ID == i.TemplateID {
			// Fetch the template
			templateZip, err := realmClient.ClientTemplate(groupID, appID, template.ID)
			if err != nil {
				return nil, nil, err
			}
			return templateZip, paths, nil
		}
	}

	return nil, nil, fmt.Errorf("templateID %s is not compatible with this app", i.TemplateID)
}
