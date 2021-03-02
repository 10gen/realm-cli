package app

import (
	"errors"
	"os"
	"path"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/AlecAivazis/survey/v2"
)

var (
	flagDirectory      = "app-dir"
	flagDirectoryShort = "p"
	flagDirectoryUsage = "the directory to create your new Realm app, defaults to Realm app name"

	flagCluster      = "cluster"
	flagClusterUsage = "include to link an Atlas cluster to your Realm app"

	flagDataLake      = "data-lake"
	flagDataLakeUsage = "include to link an Atlas data lake to your Realm app"

	flagDryRun      = "dry-run"
	flagDryRunShort = "x"
	flagDryRunUsage = "include to run without writing any changes to the file system or import/export the new Realm app"
)

type createInputs struct {
	newAppInputs
	Directory string
	Cluster   string
	DataLake  string
	DryRun    bool
}

type clusterService struct {
	Name   string        `json:"name"`
	Type   string        `json:"type"`
	Config clusterConfig `json:"config"`
}

type clusterConfig struct {
	ClusterName         string `json:"clusterName"`
	ReadPreference      string `json:"readPreference"`
	WireProtocolEnabled bool   `json:"wireProtocolEnabled"`
}

type dataLakeService struct {
	Name   string         `json:"name"`
	Type   string         `json:"type"`
	Config dataLakeConfig `json:"config"`
}

type dataLakeConfig struct {
	DataLakeName string `json:"dataLakeName"`
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
	_, appOK, err := local.FindApp(fullPath)
	if err != nil {
		return "", err
	}
	if appOK {
		return "", errProjectExists{fullPath}
	}
	return fullPath, nil
}

func (i *createInputs) resolveCluster(client atlas.Client, groupID string) (clusterService, error) {
	clusters, err := client.Clusters(groupID)
	if err != nil {
		return clusterService{}, err
	}
	var clusterName string
	for _, cluster := range clusters {
		if i.Cluster == cluster.Name {
			clusterName = cluster.Name
			break
		}
	}
	if clusterName == "" {
		return clusterService{}, errors.New("failed to find Atlas cluster")
	}
	clusterService := clusterService{
		Name: "mongodb-atlas",
		Type: "mongodb-atlas",
		Config: clusterConfig{
			ClusterName:         clusterName,
			ReadPreference:      "primary",
			WireProtocolEnabled: false,
		},
	}
	return clusterService, nil
}

func (i *createInputs) resolveDataLake(client atlas.Client, groupID string) (dataLakeService, error) {
	dataLakes, err := client.DataLakes(groupID)
	if err != nil {
		return dataLakeService{}, err
	}
	var dataLakeName string
	for _, dataLake := range dataLakes {
		if i.DataLake == dataLake.Name {
			dataLakeName = dataLake.Name
			break
		}
	}
	if dataLakeName == "" {
		return dataLakeService{}, errors.New("failed to find Atlas data lake")
	}
	dataLakeService := dataLakeService{
		Name: "mongodb-datalake",
		Type: "datalake",
		Config: dataLakeConfig{
			DataLakeName: dataLakeName,
		},
	}
	return dataLakeService, nil
}

func (i createInputs) args(omitDryRun bool) []flags.Arg {
	args := make([]flags.Arg, 0, 8)
	if i.Project != "" {
		args = append(args, flags.Arg{flagProject, i.Project})
	}
	if i.Name != "" {
		args = append(args, flags.Arg{flagName, i.Name})
	}
	if i.From != "" {
		args = append(args, flags.Arg{flagFrom, i.From})
	}
	if i.Directory != "" {
		args = append(args, flags.Arg{flagDirectory, i.Directory})
	}
	if i.Location != flagLocationDefault {
		args = append(args, flags.Arg{flagLocation, i.Location.String()})
	}
	if i.DeploymentModel != flagDeploymentModelDefault {
		args = append(args, flags.Arg{flagDeploymentModel, i.DeploymentModel.String()})
	}
	if i.Cluster != "" {
		args = append(args, flags.Arg{flagCluster, i.Cluster})
	}
	if i.DataLake != "" {
		args = append(args, flags.Arg{flagDataLake, i.DataLake})
	}
	if i.DryRun && !omitDryRun {
		args = append(args, flags.Arg{Name: flagDryRun})
	}
	return args
}
