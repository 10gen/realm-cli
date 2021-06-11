package app

import (
	"errors"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/AlecAivazis/survey/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	Clusters  []string
	DataLakes []string
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

func (i *createInputs) resolveName(ui terminal.UI, client realm.Client, groupID, appNameOrClientID string) error {
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
	//check if we are in an app directory already
	_, appOK, err := local.FindApp(wd)
	if err != nil {
		return "", err
	}
	if appOK {
		return "", errProjectExists{wd}
	}

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

	defaultLocalPath := findDefaultPath(wd, i.LocalPath)
	if ui.AutoConfirm() {
		return path.Join(wd, defaultLocalPath), nil
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

		_, appOK, err := local.FindApp(path.Join(wd, newDir))
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

func (i *createInputs) resolveClusters(ui terminal.UI, client atlas.Client, groupID string) ([]dataSourceCluster, error) {
	clusters, err := client.Clusters(groupID)
	if err != nil {
		return nil, err
	}

	existingClusters := map[string]atlas.Cluster{}
	for _, c := range clusters {
		existingClusters[c.Name] = c
	}
	nonExistingClusters := make([]string, 0, len(i.Clusters))

	dsClusters := make([]dataSourceCluster, 0, len(i.Clusters))
	for _, clusterName := range i.Clusters {
		if _, ok := existingClusters[clusterName]; !ok {
			nonExistingClusters = append(nonExistingClusters, clusterName)
			continue
		}
		dsClusters = append(dsClusters,
			dataSourceCluster{
				Name: "mongodb-atlas",
				Type: realm.ClusterType,
				Config: configCluster{
					ClusterName:         clusterName,
					ReadPreference:      "primary",
					WireProtocolEnabled: false,
				},
			})
	}
	if len(dsClusters) == 0 {
		return nil, errors.New("failed to find Atlas cluster")
	}
	if len(nonExistingClusters) > 0 {
		ui.Print(terminal.NewWarningLog("Please note, the data sources '%s' were not linked because Atlas clusters were not found", strings.Join(nonExistingClusters[:], ", ")))

		if !ui.AutoConfirm() {
			proceed, err := ui.Confirm("Would you still like to create the app?")
			if err != nil {
				return nil, err
			}
			if !proceed {
				return nil, errors.New("failed to find Atlas cluster")
			}
		}
	}
	return dsClusters, nil
}

func (i *createInputs) resolveDataLakes(ui terminal.UI, client atlas.Client, groupID string) ([]dataSourceDataLake, error) {
	dataLakes, err := client.DataLakes(groupID)
	if err != nil {
		return nil, err
	}

	existingDataLakes := map[string]atlas.DataLake{}
	for _, d := range dataLakes {
		existingDataLakes[d.Name] = d
	}
	nonExistingDataLakes := make([]string, 0, len(i.DataLakes))

	dsDataLakes := make([]dataSourceDataLake, 0, len(i.DataLakes))
	for _, dataLakeName := range i.DataLakes {
		if _, ok := existingDataLakes[dataLakeName]; !ok {
			nonExistingDataLakes = append(nonExistingDataLakes, dataLakeName)
			continue
		}

		dsDataLakes = append(dsDataLakes,
			dataSourceDataLake{
				Name: "mongodb-datalake",
				Type: realm.DataLakeType,
				Config: configDataLake{
					DataLakeName: dataLakeName,
				},
			})
	}
	if len(dsDataLakes) == 0 {
		return nil, errors.New("failed to find Atlas data lake")
	}
	if len(nonExistingDataLakes) > 0 {
		ui.Print(terminal.NewWarningLog("Please note, the data sources '%s' were not linked because Atlas data lakes were not found", strings.Join(nonExistingDataLakes[:], ", ")))

		if !ui.AutoConfirm() {
			proceed, err := ui.Confirm("Would you still like to create the app?")
			if err != nil {
				return nil, err
			}
			if !proceed {
				return nil, errors.New("failed to find Atlas data lake")
			}
		}
	}
	return dsDataLakes, nil
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
		args = append(args, flags.Arg{flagRemoteAppNew, i.RemoteApp})
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
	if len(i.Clusters) > 0 {
		args = append(args, flags.Arg{flagCluster, strings.Join(i.Clusters, ",")})
	}
	if len(i.DataLakes) > 0 {
		args = append(args, flags.Arg{flagDataLake, strings.Join(i.DataLakes, ",")})
	}
	if i.DryRun && !omitDryRun {
		args = append(args, flags.Arg{Name: flagDryRun})
	}
	return args
}

func findDefaultPath(wd string, localPath string) string {
	for i := 1; i < 10; i++ {
		newPath := localPath + "-" + strconv.Itoa(i)
		_, found, err := local.FindApp(path.Join(wd, newPath))
		if err == nil && !found {
			return newPath
		}
	}
	return localPath + "-" + primitive.NewObjectID().Hex()
}
