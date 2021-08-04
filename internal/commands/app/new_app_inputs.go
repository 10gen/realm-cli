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
		app, err := cli.ResolveApp(ui, rc, realm.AppFilter{App: i.RemoteApp})
		if err != nil {
			return realm.App{}, err
		}
		ra = app
	}
	return ra, nil
}

// resolveTemplateID is responsible for resolving a template id from a user cli request.
// a user can use the --template flag in 1 of the following ways:
// 1. --template=<template-id>: This populates the value of i.Template to be <template-id> and attempts to create a
//    template with id = <template-id>
// 2. no flag defined: This populates the value of i.Template to be "". This tells us to skip the template-app-related
//    parts of app creation and create a default app
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

	if i.Template != "" {
		for _, template := range templates {
			if template.ID == i.Template {
				i.Template = template.ID
				return nil
			}
		}

		return fmt.Errorf("template '%s' not found", i.Template)
	}
	return nil
}
