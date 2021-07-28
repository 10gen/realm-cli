package app

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
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
// 1. --template: This will populate the current value of i.Template to be `noArgsDefaultValueTemplate`. This tells us
//    that the user wants to create a template but wants to be prompted for the template id
// 2. --template=<template-id>: This populates the value of i.Template to be <template-id>. This skips the template id
//    prompt and checks whether the supplied template id exists.
// 3. no flag defined: This populates the value of i.Template to be "". This tells us to skip the template-app-related
//    parts of app creation
func (i *newAppInputs) resolveTemplateID(ui terminal.UI, client realm.Client) error {
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

	if i.Template != noArgsDefaultValueTemplate {
		for _, template := range templates {
			if template.ID == i.Template {
				i.Template = template.ID
				return nil
			}
		}

		return fmt.Errorf("template '%s' not found", i.Template)
	}

	options := make([]string, 0, len(templates)+1)
	templateIDs := make([]string, 0, len(templates)+1)
	for _, template := range templates {
		options = append(options, fmt.Sprintf("[%s]: %s", template.ID, template.Name))
		templateIDs = append(templateIDs, template.ID)
	}

	var selectedIndex int
	if err := ui.AskOne(
		&selectedIndex,
		&survey.Select{
			Message: "Please select a template from the available options",
			Options: options,
		},
	); err != nil {
		return err
	}

	i.Template = templateIDs[selectedIndex]

	return nil
}
