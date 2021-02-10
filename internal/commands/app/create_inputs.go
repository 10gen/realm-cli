package app

import (
	"fmt"
	"os"
	"path"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
)

var (
	flagDirectory      = "app-dir"
	flagDirectoryShort = "c"
	flagDirectoryUsage = "the directory to create your new Realm app, defaults to Realm app name"

	// TODO(REALMC-8135): Implement data-source flag for app create command
	// flagDataSource      = "data-source"
	// flagDataSourceShort = "s"
	// flagDataSourceUsage = "atlas cluster to back your Realm app, defaults to first available"

	// TODO(REALMC-8134): Implement dry-run for app create command
	// flagDryRun      = "dry-run"
	// flagDryRunShort = "x"
	// flagDryRunUsage = "include to run without writing any changes to the file system or import/export the new Realm app"
)

type createInputs struct {
	newAppInputs
	Directory string
	// TODO(REALMC-8135): Implement data-source flag for app create command
	// DataSource string
	// TODO(REALMC-8134): Implement dry-run for app create command
	// DryRun bool
}

func (i *createInputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	if i.From == "" {
		if i.Name == "" {
			if err := ui.AskOne(&i.Name, &survey.Input{Message: "App Name"}); err != nil {
				return err
			}
		}
		if i.DeploymentModel == realm.DeploymentModelEmpty {
			i.DeploymentModel = flagDeploymentModelDefault
		}
		if i.Location == realm.LocationEmpty {
			i.Location = flagLocationDefault
		}
	}

	return nil
}

func (i *createInputs) resolveAppName(ui terminal.UI, client realm.Client, f from) (string, error) {
	if i.Name == "" {
		app, err := cli.ResolveApp(ui, client, realm.AppFilter{GroupID: f.GroupID, App: f.AppID})
		if err != nil {
			return "", err
		}
		return app.Name, nil
	}
	return i.Name, nil
}

func (i *createInputs) resolveProject(ui terminal.UI, client atlas.Client) (string, error) {
	if i.Project == "" {
		groupID, groupErr := cli.ResolveGroupID(ui, client)
		if groupErr != nil {
			return "", groupErr
		}
		i.Project = groupID
	}
	return i.Project, nil
}

func (i *createInputs) resolveDirectory(wd, appName string) (string, error) {
	dir := i.Directory
	if dir == "" {
		dir = appName
	}
	fullPath := path.Join(wd, dir)
	fi, statErr := os.Stat(fullPath)
	if statErr != nil && !os.IsNotExist(statErr) {
		return "", statErr
	}
	if fi != nil {
		switch mode := fi.Mode(); {
		case mode.IsDir():
			_, appOK, appErr := local.FindApp(fullPath)
			if appErr != nil {
				return "", appErr
			}
			if appOK {
				return "", fmt.Errorf("A Realm app already exists at %s", fullPath)
			}
			return fullPath, nil
		}
	}
	return fullPath, nil
}

// TODO(REALMC-8135): Implement data-source flag for app create command
// func (i *createInputs) resolveDataSource(client realm.Client, groupID, appID string) (string, error) {
// 	clusters, err := client.ListClusters(groupID, appID)
// 	if err != nil {
// 		return "", err
// 	}
// 	var dsCluster string
// 	for _, cluster := range clusters {
// 		if (i.DataSource == "" && cluster.State == "IDLE") || i.DataSource == cluster.Name {
// 			dsCluster = cluster.Name
// 			break
// 		}
// 	}
// 	if dsCluster == "" {
// 		if i.DataSource != "" {
// 			return "", fmt.Errorf("Unable to find the %s cluster", i.DataSource)
// 		}
// 		return "", fmt.Errorf("Unable to find any available cluster for Group ID %s", groupID)
// 	}
// 	return dsCluster, nil
// }
