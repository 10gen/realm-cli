package app

import (
	"fmt"
	"os"
	"path"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
)

var (
	flagDirectory      = "app-dir"
	flagDirectoryShort = "c"
	flagDirectoryUsage = "the directory to create your new Realm app, defaults to Realm app name"

	flagDataSource      = "data-source"
	flagDataSourceShort = "s"
	flagDataSourceUsage = "atlas cluster to back your Realm app, defaults to first available"

	// TODO(REALMC-8134): Implement dry-run for app create command
	// flagDryRun      = "dry-run"
	// flagDryRunShort = "x"
	// flagDryRunUsage = "include to run without writing any changes to the file system or import/export the new Realm app"
)

type createInputs struct {
	newAppInputs
	Directory  string
	DataSource string
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

func (i *createInputs) resolveName(ui terminal.UI, client realm.Client, f from) error {
	if i.Name == "" {
		app, err := cli.ResolveApp(ui, client, realm.AppFilter{GroupID: f.GroupID, App: f.AppID})
		if err != nil {
			return err
		}
		i.Name = app.Name
	}
	return nil
}

func (i *createInputs) resolveDirectory(wd string) (string, error) {
	if i.Directory == "" {
		i.Directory = i.Name
	}
	fullPath := path.Join(wd, i.Directory)
	fi, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fullPath, nil
		}
		return "", err
	}
	if !fi.Mode().IsDir() {
		return fullPath, nil
	}
	_, appOK, appErr := local.FindApp(fullPath)
	if appErr != nil {
		return "", appErr
	}
	if appOK {
		return "", fmt.Errorf("%s is inside or is a Realm app directory", fullPath)
	}
	return fullPath, nil
}

func (i *createInputs) resolveDataSource(client realm.Client, groupID, appID string) (string, error) {
	clusters, err := client.ListClusters(groupID, appID)
	if err != nil {
		return "", err
	}
	var dsCluster string
	for _, cluster := range clusters {
		if (i.DataSource == "" && cluster.State == "IDLE") || i.DataSource == cluster.Name {
			dsCluster = cluster.Name
			break
		}
	}
	if dsCluster == "" {
		if i.DataSource != "" {
			return "", fmt.Errorf("unable to find the %s cluster", i.DataSource)
		}
		return "", fmt.Errorf("unable to find any available cluster for Project %s", groupID)
	}
	return dsCluster, nil
}
