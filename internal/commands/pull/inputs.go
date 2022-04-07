package pull

import (
	"archive/zip"
	"errors"
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/mitchellh/go-homedir"
)

const (
	errDependencyFlagConflictTemplate = `cannot use both "%s" and "%s" at the same time`
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
	IncludeNodeModules  bool
	IncludePackageJSON  bool
	IncludeHosting      bool
	DryRun              bool
	TemplateIDs         []string

	// derived inputs
	appMeta local.AppMeta
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

	wd := i.LocalPath
	if wd == "" {
		wd = profile.WorkingDirectory
	}

	app, _, appErr := local.FindApp(wd)
	if appErr != nil {
		return appErr
	}
	i.appMeta = app.Meta

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

		if i.RemoteApp == "" && i.appMeta.AppID == "" {
			i.RemoteApp = app.Option()
		}
	}

	return nil
}

func (i *inputs) resolveRemoteApp(ui terminal.UI, clients cli.Clients) (realm.App, error) {
	if i.Project == "" && i.appMeta.GroupID == "" {
		groupID, err := cli.ResolveGroupID(ui, clients.Atlas)
		if err != nil {
			return realm.App{}, err
		}
		i.Project = groupID
	}

	app, err := cli.ResolveApp(ui, clients.Realm, cli.AppOptions{
		Filter:  realm.AppFilter{GroupID: i.Project, App: i.RemoteApp},
		AppMeta: i.appMeta,
	})
	if err != nil {
		if _, ok := err.(cli.ErrAppNotFound); ok {
			return realm.App{}, errProjectNotFound
		}
		return realm.App{}, err
	}

	// TODO(REALMC-9462): remove this once /apps has "template_id" in the payload
	app, err = clients.Realm.FindApp(app.GroupID, app.ID)
	if err != nil {
		return realm.App{}, err
	}

	return app, nil
}

type clientTemplate struct {
	id     string
	zipPkg *zip.Reader
}

func (i *inputs) resolveClientTemplates(realmClient realm.Client, groupID, appID string) ([]clientTemplate, error) {
	if len(i.TemplateIDs) == 0 {
		return nil, nil
	}

	compatibleTemplates, err := realmClient.CompatibleTemplates(groupID, appID)
	if err != nil {
		return nil, err
	}
	compatibleTemplatesByID := compatibleTemplates.MapByID()

	templateFrontends := make([]clientTemplate, 0, len(i.TemplateIDs))
	for _, requestedTemplate := range i.TemplateIDs {
		if template, ok := compatibleTemplatesByID[requestedTemplate]; ok {
			templateZip, ok, err := realmClient.ClientTemplate(groupID, appID, template.ID)
			if err != nil {
				return nil, err
			} else if !ok {
				return nil, fmt.Errorf("template '%s' does not have a frontend to pull", template.ID)
			}
			templateFrontends = append(templateFrontends, clientTemplate{
				id:     template.ID,
				zipPkg: templateZip,
			})
		} else {
			return nil, fmt.Errorf("frontend template '%s' is not compatible with this app", requestedTemplate)
		}
	}
	return templateFrontends, nil
}
