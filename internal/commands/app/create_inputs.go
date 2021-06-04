package app

import (
	"errors"
	"os"
	"path"
	"strconv"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/AlecAivazis/survey/v2"
)

var (
	flagLocalPathCreate      = "local"
	flagLocalPathCreateUsage = "the local path to create your new Realm app in, defaults to the Realm app name"

	flagCluster      = "cluster"
	flagClusterUsage = "include to link an Atlas cluster to your Realm app"

	flagDataLake      = "data-lake"
	flagDataLakeUsage = "include to link an Atlas data lake to your Realm app"

	flagDryRun      = "dry-run"
	flagDryRunShort = "x"
	flagDryRunUsage = "include to run without writing any changes to the file system nor deploying any changes to the Realm server"
)

type createInputs struct {
	newAppInputs
	LocalPath string
	Cluster   string
	DataLake  string
	DryRun    bool
}

type dataSourceCluster struct {
	Name   string        `json:"name"`
	Type   string        `json:"type"`
	Config configCluster `json:"config"`
}

type configCluster struct {
	ClusterName         string `json:"clusterName"`
	ReadPreference      string `json:"readPreference"`
	WireProtocolEnabled bool   `json:"wireProtocolEnabled"`
}

type dataSourceDataLake struct {
	Name   string         `json:"name"`
	Type   string         `json:"type"`
	Config configDataLake `json:"config"`
}

type configDataLake struct {
	DataLakeName string `json:"dataLakeName"`
}

func (i *createInputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	if i.RemoteApp == "" {
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
		if i.ConfigVersion == realm.AppConfigVersionZero {
			i.ConfigVersion = realm.DefaultAppConfigVersion
		}
	}

	return nil
}

func (i *createInputs) resolveName(ui terminal.UI, client realm.Client, groupID string, appNameOrClientID string) error {
	if i.Name == "" {
		app, err := cli.ResolveApp(ui, client, realm.AppFilter{
			GroupID: groupID,
			App:     appNameOrClientID,
		})
		if err != nil {
			return err
		}
		i.Name = app.Name
	}
	return nil
}

func (i *createInputs) resolveLocalPath(ui terminal.UI, wd string) (string, error) {
	if i.LocalPath == "" {
		i.LocalPath = i.Name
	}
	fullPath := path.Join(wd, i.LocalPath)
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

	//check if we are in an app directory already
	_, appOK, err := local.FindApp(wd)
	if err != nil {
		return "", err
	}
	if appOK {
		return "", errProjectExists{wd}
	}

	defaultLocalPath := getDefaultPath(wd, i.LocalPath)
	if ui.AutoConfirm() {
		fullPath = path.Join(wd, defaultLocalPath)
		return fullPath, nil
	}

	ui.Print(terminal.NewWarningLog("Local path './%s' already exists, writing app contents to that destination may result in file conflicts.", i.LocalPath))
	proceed, err := ui.Confirm("Would you still like to write app contents to './%s'? ('No' will prompt you to provide another destination)", i.LocalPath)
	if err != nil {
		return "", err
	}
	if !proceed {
		var newDir string
		if err := ui.AskOne(&newDir, &survey.Input{Message: "Local Path", Default: defaultLocalPath}); err != nil {
			return "", err
		}

		_, appOK, err := local.FindApp(newDir)
		if err != nil {
			return "", err
		}
		if appOK {
			return "", errProjectExists{newDir}
		}

		i.LocalPath = newDir
		fullPath = path.Join(wd, i.LocalPath)
	}
	return fullPath, nil
}

func (i *createInputs) resolveCluster(client atlas.Client, groupID string) (dataSourceCluster, error) {
	clusters, err := client.Clusters(groupID)
	if err != nil {
		return dataSourceCluster{}, err
	}
	var clusterName string
	for _, cluster := range clusters {
		if i.Cluster == cluster.Name {
			clusterName = cluster.Name
			break
		}
	}
	if clusterName == "" {
		return dataSourceCluster{}, errors.New("failed to find Atlas cluster")
	}
	dsCluster := dataSourceCluster{
		Name: "mongodb-atlas",
		Type: "mongodb-atlas",
		Config: configCluster{
			ClusterName:         clusterName,
			ReadPreference:      "primary",
			WireProtocolEnabled: false,
		},
	}
	return dsCluster, nil
}

func (i *createInputs) resolveDataLake(client atlas.Client, groupID string) (dataSourceDataLake, error) {
	dataLakes, err := client.DataLakes(groupID)
	if err != nil {
		return dataSourceDataLake{}, err
	}
	var dataLakeName string
	for _, dataLake := range dataLakes {
		if i.DataLake == dataLake.Name {
			dataLakeName = dataLake.Name
			break
		}
	}
	if dataLakeName == "" {
		return dataSourceDataLake{}, errors.New("failed to find Atlas data lake")
	}
	dsDataLake := dataSourceDataLake{
		Name: "mongodb-datalake",
		Type: "datalake",
		Config: configDataLake{
			DataLakeName: dataLakeName,
		},
	}
	return dsDataLake, nil
}

func (i createInputs) args(omitDryRun bool) []flags.Arg {
	args := make([]flags.Arg, 0, 8)
	if i.Project != "" {
		args = append(args, flags.Arg{flagProject, i.Project})
	}
	if i.Name != "" {
		args = append(args, flags.Arg{flagName, i.Name})
	}
	if i.RemoteApp != "" {
		args = append(args, flags.Arg{flagRemote, i.RemoteApp})
	}
	if i.LocalPath != "" {
		args = append(args, flags.Arg{flagLocalPathCreate, i.LocalPath})
	}
	if i.Location != flagLocationDefault {
		args = append(args, flags.Arg{flagLocation, i.Location.String()})
	}
	if i.DeploymentModel != flagDeploymentModelDefault {
		args = append(args, flags.Arg{flagDeploymentModel, i.DeploymentModel.String()})
	}
	if i.Environment != realm.EnvironmentNone {
		args = append(args, flags.Arg{flagEnvironment, i.Environment.String()})
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

func getDefaultPath(wd string, localPath string) string {
	i := 1
	for {
		appPath := path.Join(wd, localPath) + "-" + strconv.Itoa(i)
		_, found, err := local.FindApp(appPath)
		if err != nil || !found {
			return localPath + "-" + strconv.Itoa(i)
		}
		i++
	}
}
