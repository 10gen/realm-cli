package pull

import (
	"errors"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/mitchellh/go-homedir"
)

const (
	flagRemote      = "remote"
	flagRemoteUsage = "specify the app to pull changes down from"

	flagProject      = "project"
	flagProjectUsage = "the MongoDB cloud project id"

	flagAppVersion      = "app-version"
	flagAppVersionUsage = "specify the app config version to pull changes down as"

	flagLocal      = "local"
	flagLocalUsage = "provide the path to export a Realm app to"

	flagIncludeDependencies      = "include-dependencies"
	flagIncludeDependenciesShort = "d"
	flagIncludeDependenciesUsage = "include to to push Realm app dependencies changes as well"

	flagIncludeHosting      = "include-hosting"
	flagIncludeHostingShort = "s"
	flagIncludeHostingUsage = "include to push Realm app hosting changes as well"

	flagDryRun      = "dry-run"
	flagDryRunShort = "x"
	flagDryRunUsage = "include to run without writing any changes to the file system"
)

var (
	errConfigVersionMismatch = errors.New("must export an app with the same config version as found in the current project directory")
)

type inputs struct {
	Project             string
	Remote              string
	Local               string
	AppVersion          realm.AppConfigVersion
	IncludeDependencies bool
	IncludeHosting      bool
	DryRun              bool
}

func (i *inputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	wd := i.Local
	if wd == "" {
		wd = profile.WorkingDirectory
	}

	app, appErr := local.LoadAppConfig(wd)
	if appErr != nil {
		return appErr
	}

	var local string
	if i.Local == "" {
		local = app.RootDir
	} else {
		l, err := homedir.Expand(i.Local)
		if err != nil {
			return err
		}
		local = l
	}
	i.Local = local

	if app.RootDir != "" {
		if i.AppVersion == realm.AppConfigVersionZero {
			i.AppVersion = app.ConfigVersion()
		} else if i.AppVersion != app.ConfigVersion() && app.ConfigVersion() != realm.AppConfigVersionZero {
			return errConfigVersionMismatch
		}

		if i.Remote == "" {
			i.Remote = app.Option()
		}
	}

	return nil
}

type appRemote struct {
	GroupID string
	AppID   string
}

func (i inputs) resolveRemoteApp(ui terminal.UI, client realm.Client) (appRemote, error) {
	r := appRemote{GroupID: i.Project}

	if i.Remote == "" {
		return r, nil
	}

	app, err := cli.ResolveApp(ui, client, realm.AppFilter{GroupID: i.Project, App: i.Remote})
	if err != nil {
		if _, ok := err.(cli.ErrAppNotFound); !ok {
			return appRemote{}, err
		}
	}

	r.GroupID = app.GroupID
	r.AppID = app.ID
	return r, nil
}
