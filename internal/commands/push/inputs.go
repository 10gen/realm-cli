package push

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

const (
	errDependencyFlagConflictTemplate = `cannot use both "%s" and "%s" at the same time`
)

type appRemote struct {
	GroupID     string
	AppID       string
	ClientAppID string
}

type inputs struct {
	LocalPath           string
	RemoteApp           string
	Project             string
	IncludeNodeModules  bool
	IncludePackageJSON  bool
	IncludeDependencies bool
	IncludeHosting      bool
	ResetCDNCache       bool
	DryRun              bool
}

func (i *inputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	if i.IncludePackageJSON {
		if i.IncludeNodeModules {
			return fmt.Errorf(errDependencyFlagConflictTemplate, flagIncludeNodeModules, flagIncludePackageJSON)
		}
		if i.IncludeDependencies {
			return fmt.Errorf(errDependencyFlagConflictTemplate, flagIncludeDependencies, flagIncludePackageJSON)
		}
	}

	searchPath := i.LocalPath
	if searchPath == "" {
		searchPath = profile.WorkingDirectory
	}

	searchPathAbs, err := filepath.Abs(searchPath)
	if err != nil {
		return err
	}

	if _, err = os.Stat(searchPathAbs); os.IsNotExist(err) {
		return errProjectInvalid(searchPath, false)
	}

	app, _, err := local.FindApp(searchPath)
	if err != nil {
		return err
	}

	if app.RootDir == "" {
		return errProjectInvalid(searchPath, true)
	}

	if i.LocalPath == "" {
		i.LocalPath = app.RootDir
	}

	if i.RemoteApp == "" && app.Meta.AppID == "" {
		i.RemoteApp = app.Option()
	}

	return nil
}

func (i inputs) resolveRemoteApp(ui terminal.UI, client realm.Client, appMeta local.AppMeta) (appRemote, error) {
	r := appRemote{GroupID: i.Project}
	app, err := cli.ResolveApp(ui, client, cli.AppOptions{
		Filter:  realm.AppFilter{GroupID: i.Project, App: i.RemoteApp},
		AppMeta: appMeta,
	})
	if err != nil {
		if _, ok := err.(cli.ErrAppNotFound); !ok {
			return appRemote{}, err
		}
		return r, nil
	}

	r.GroupID = app.GroupID
	r.AppID = app.ID
	r.ClientAppID = app.ClientAppID
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
	if i.IncludeNodeModules {
		args = append(args, flags.Arg{Name: flagIncludeNodeModules})
	}
	if i.IncludePackageJSON {
		args = append(args, flags.Arg{Name: flagIncludePackageJSON})
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
