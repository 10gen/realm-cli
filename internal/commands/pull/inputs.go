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

	"github.com/AlecAivazis/survey/v2"
	"github.com/mitchellh/go-homedir"
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

func (i *inputs) resolveClientTemplates(ui terminal.UI, realmClient realm.Client, groupID, appID string) ([]clientTemplate, error) {
	compatibleTemplates, err := realmClient.CompatibleTemplates(groupID, appID)
	if err != nil {
		return nil, err
	}

	if i.TemplateID != "" {
		for _, template := range compatibleTemplates {
			if template.ID == i.TemplateID {
				templateZip, ok, err := realmClient.ClientTemplate(groupID, appID, template.ID)
				if err != nil {
					return nil, err
				} else if !ok {
					return nil, fmt.Errorf("template '%s' does not have a frontend to pull", template.ID)
				}
				return []clientTemplate{{
					id:     template.ID,
					zipPkg: templateZip,
				}}, nil
			}
		}
		return nil, fmt.Errorf("template '%s' is not compatible with this app", i.TemplateID)
	}

	if len(compatibleTemplates) == 0 {
		return nil, nil
	}

	if proceed, err := ui.Confirm("Would you like to export with a client template?"); err != nil {
		return nil, err
	} else if !proceed {
		return nil, nil
	}

	nameOptions := make([]string, len(compatibleTemplates))
	for idx, compatibleTemplate := range compatibleTemplates {
		nameOptions[idx] = fmt.Sprintf("[%s]: %s", compatibleTemplate.ID, compatibleTemplate.Name)
	}

	var selectedTemplateIdxs []int
	if err := ui.AskOne(
		&selectedTemplateIdxs,
		&survey.MultiSelect{Message: "Which template(s) would you like to export this app with", Options: nameOptions}); err != nil {
		return nil, err
	}

	result := make([]clientTemplate, 0, len(selectedTemplateIdxs))
	for _, selectedTemplateIdx := range selectedTemplateIdxs {
		templateID := compatibleTemplates[selectedTemplateIdx].ID
		templateZip, ok, err := realmClient.ClientTemplate(groupID, appID, templateID)
		if err != nil {
			return nil, err
		} else if !ok {
			return nil, fmt.Errorf("template '%s' does not have a frontend to pull", templateID)
		}
		result = append(result, clientTemplate{
			id:     templateID,
			zipPkg: templateZip,
		})
	}

	return result, nil
}
