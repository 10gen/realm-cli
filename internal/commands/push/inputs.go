package push

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

const (
	flagLocalPath      = "local"
	flagLocalPathUsage = "provide the path to a Realm app containing the changes to push"

	flagProject      = "project"
	flagProjectUsage = "the MongoDB cloud project id"

	flagRemote      = "remote"
	flagRemoteUsage = "choose a Realm app to push changes towards"

	flagDryRun      = "dry-run"
	flagDryRunShort = "x"
	flagDryRunUsage = "include to run without pushing any changes to the Realm server"

	flagIncludeDependencies      = "include-dependencies"
	flagIncludeDependenciesShort = "d"
	flagIncludeDependenciesUsage = "include to push Realm app dependencies changes as well"

	flagIncludeHosting      = "include-hosting"
	flagIncludeHostingShort = "s"
	flagIncludeHostingUsage = "include to push Realm app hosting changes as well"

	flagResetCDNCache      = "reset-cdn-cache"
	flagResetCDNCacheShort = "c"
	flagResetCDNCacheUsage = "include to reset the Realm app hosting CDN cache"
)

type appRemote struct {
	GroupID string
	AppID   string
}

type inputs struct {
	LocalPath           string
	Project             string
	Remote              string
	IncludeDependencies bool
	IncludeHosting      bool
	ResetCDNCache       bool
	DryRun              bool
}

func (i *inputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
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

	if i.Remote == "" {
		i.Remote = app.Option()
	}

	return nil
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

func (i inputs) args(omitDryRun bool) []flags.Arg {
	args := make([]flags.Arg, 0, 7)
	if i.Project != "" {
		args = append(args, flags.Arg{flagProject, i.Project})
	}
	if i.LocalPath != "" {
		args = append(args, flags.Arg{flagLocalPath, i.LocalPath})
	}
	if i.Remote != "" {
		args = append(args, flags.Arg{flagRemote, i.Remote})
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
