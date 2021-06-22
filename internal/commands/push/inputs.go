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
	flagLocalPath           = "local"
	flagRemote              = "remote"
	flagIncludeDependencies = "include-dependencies"
	flagIncludeHosting      = "include-hosting"
	flagResetCDNCache       = "reset-cdn-cache"
	flagDryRun              = "dry-run"
)

type appRemote struct {
	GroupID string
	AppID   string
}

type inputs struct {
	LocalPath           string
	RemoteApp           string
	Project             string
	IncludeDependencies bool
	IncludeHosting      bool
	ResetCDNCache       bool
	DryRun              bool
}

func (i *inputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	searchPath := i.LocalPath
	if searchPath == "" {
		searchPath = profile.WorkingDirectory
	}

	app, err := local.LoadAppConfig(searchPath)
	if err != nil {
		return err
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
		return r, nil
	}

	r.GroupID = app.GroupID
	r.AppID = app.ID
	return r, nil
}

func (i inputs) args(omitDryRun bool) []flags.Arg {
	args := make([]flags.Arg, 0, 7)
	if i.Project != "" {
		args = append(args, flags.Arg{cli.ProjectFlagName, i.Project})
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
