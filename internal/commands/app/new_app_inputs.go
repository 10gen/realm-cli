package app

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
)

const (
	flagDeploymentModelDefault = realm.DeploymentModelGlobal
	flagLocationDefault        = realm.LocationVirginia
)

type newAppInputs struct {
	Project         string
	RemoteApp       string
	Name            string
	DeploymentModel realm.DeploymentModel
	Location        realm.Location
	Environment     realm.Environment
	Template        string
	ConfigVersion   realm.AppConfigVersion
}

func (i *newAppInputs) resolveRemoteApp(ui terminal.UI, rc realm.Client) (realm.App, error) {
	var ra realm.App
	if i.RemoteApp != "" {
		app, err := cli.ResolveApp(ui, rc, cli.AppOptions{Filter: realm.AppFilter{App: i.RemoteApp}})
		if err != nil {
			return realm.App{}, err
		}
		ra = app
	}
	return ra, nil
}

// resolveTemplateID is responsible for resolving a template id from a user cli request.
// If the --template flag is not set, the CLI will create a default app without any template,
// otherwise the CLI will attempt to create an app based on the specified template
func (i *newAppInputs) resolveTemplateID(client realm.Client) error {
	if i.Template == "" {
		return nil
	}

	templates, err := client.AllTemplates()
	if err != nil {
		return err
	}
	if len(templates) == 0 {
		return fmt.Errorf("unable to find template '%s'", i.Template)
	}

	// Check if supplied template id is a valid template
	for _, template := range templates {
		if template.ID == i.Template {
			return nil
		}
	}
	return fmt.Errorf("template '%s' not found", i.Template)
}
