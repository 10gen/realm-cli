package push

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

const (
	flagLocalPath      = "local"
	flagLocalPathUsage = "specify the local path to a Realm app to import"

	flagRemote      = "remote"
	flagRemoteUsage = "specify a remote Realm app (id or name) to push changes to"

	flagIncludeDependencies      = "include-dependencies"
	flagIncludeDependenciesShort = "d"
	flagIncludeDependenciesUsage = "include to import Realm app dependencies changes as well"

	flagIncludeHosting      = "include-hosting"
	flagIncludeHostingShort = "s"
	flagIncludeHostingUsage = "include to import Realm app hosting changes as well"

	flagResetCDNCache      = "reset-cdn-cache"
	flagResetCDNCacheShort = "c"
	flagResetCDNCacheUsage = "include to reset the Realm app hosting CDN cache"

	flagDryRun      = "dry-run"
	flagDryRunShort = "x"
	flagDryRunUsage = "include to run without pushing any changes to the Realm server"

	flagProject      = "project"
	flagProjectUsage = "the MongoDB cloud project id"
)

type appRemote struct {
	GroupID string
	AppID   string
}

type inputs struct {
	LocalPath           string
	Project             string
	RemoteApp           string
	IncludeDependencies bool
	IncludeHosting      bool
	ResetCDNCache       bool
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

	if i.LocalPath == "" {
		if app.RootDir == "" {
			return errProjectNotFound{}
		}
		i.LocalPath = app.RootDir
	}

	if i.RemoteApp == "" {
		i.RemoteApp = app.ID()
	}

	return nil
}

func (i inputs) resolveRemoteApp(ui terminal.UI, client realm.Client) (appRemote, error) {
	r := appRemote{GroupID: i.Project}

	if i.RemoteApp == "" {
		return r, nil
	}

	app, err := cli.ResolveApp(ui, client, realm.AppFilter{GroupID: i.Project, App: i.RemoteApp})
	if err != nil {
		if _, ok := err.(cli.ErrAppNotFound); !ok {
			return appRemote{}, err
		}
	}

	r.GroupID = app.GroupID
	r.AppID = app.ID
	return r, nil
}

func (i inputs) args(omitDryRun bool) []flags.Arg {
	args := make([]flags.Arg, 0, 7)
	if i.Project != "" {
		args = append(args, flags.Arg{flagProject, i.Project})
	}
	if i.LocalPath != "" {
		args = append(args, flags.Arg{flagLocalPath, i.LocalPath})
	}
	if i.RemoteApp != "" {
		args = append(args, flags.Arg{flagRemote, i.RemoteApp})
	}
	if i.IncludeDependencies {
		args = append(args, flags.Arg{Name: flagIncludeDependencies})
	}
	if i.IncludeHosting {
		args = append(args, flags.Arg{Name: flagIncludeHosting})
	}
	if i.ResetCDNCache {
		args = append(args, flags.Arg{Name: flagResetCDNCache})
	}
	if i.DryRun && !omitDryRun {
		args = append(args, flags.Arg{Name: flagDryRun})
	}
	return args
}
