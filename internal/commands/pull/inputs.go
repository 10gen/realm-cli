package pull

import (
	"errors"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/mitchellh/go-homedir"
)

const (
	flagLocalPath      = "local"
	flagLocalPathUsage = "specify the local path to export a Realm app to"

	flagRemote      = "remote"
	flagRemoteUsage = "specify the remote app to pull changes down from"

	flagIncludeDependencies      = "include-dependencies"
	flagIncludeDependenciesShort = "d"
	flagIncludeDependenciesUsage = "include to to export Realm app dependencies changes as well"

	flagIncludeHosting      = "include-hosting"
	flagIncludeHostingShort = "s"
	flagIncludeHostingUsage = "include to export Realm app hosting changes as well"

	flagDryRun      = "dry-run"
	flagDryRunShort = "x"
	flagDryRunUsage = "include to run without writing any changes to the file system"

	flagProject      = "project"
	flagProjectUsage = "the MongoDB cloud project id"

	flagConfigVersion      = "config-version"
	flagConfigVersionUsage = "specify the app config version to export as"
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
