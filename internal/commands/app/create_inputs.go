package app

import (
	"errors"
	"fmt"
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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	flagLocalPathCreate               = "local"
	flagCluster                       = "cluster"
	flagClusterServiceName            = "cluster-service-name"
	flagServerlessInstance            = "serverless-instance"
	flagServerlessInstanceServiceName = "serverless-instance-service-name"
	flagDatalake                      = "datalake"
	flagDatalakeServiceName           = "datalake-service-name"
	flagTemplate                      = "template"
	flagDryRun                        = "dry-run"
)

type createInputs struct {
	newAppInputs
	LocalPath                      string
	Clusters                       []string
	ClusterServiceNames            []string
	ServerlessInstances            []string
	ServerlessInstanceServiceNames []string
	Datalakes                      []string
	DatalakeServiceNames           []string
	DryRun                         bool
}

type dataSourceCluster struct {
	Name    string        `json:"name"`
	Type    string        `json:"type"`
	Config  configCluster `json:"config"`
	Version int           `json:"version"`
}

type configCluster struct {
	ClusterName         string `json:"clusterName"`
	ReadPreference      string `json:"readPreference"`
	WireProtocolEnabled bool   `json:"wireProtocolEnabled"`
}

type dataSourceDatalake struct {
	Name    string         `json:"name"`
	Type    string         `json:"type"`
	Config  configDatalake `json:"config"`
	Version int            `json:"version"`
}

type configDatalake struct {
	DatalakeName string `json:"dataLakeName"`
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
		app, err := cli.ResolveApp(ui, client, cli.AppOptions{Filter: realm.AppFilter{
			GroupID: groupID,
			App:     appNameOrClientID,
		}})
		if err != nil {
			return err
		}
		i.Name = app.Name
	}
	return nil
}

func (i *createInputs) resolveLocalPath(ui terminal.UI, wd string) (string, error) {
	// Check if current working directory is an app directory
	_, appOK, err := local.FindApp(wd)
	if err != nil {
		return "", err
	}
	if appOK {
		return "", errProjectExists(wd)
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
			return "", errProjectExists(newDir)
		}

		i.LocalPath = newDir
		fullPath = path.Join(wd, i.LocalPath)
	}
	return fullPath, nil
}

func (i *createInputs) resolveClusters(ui terminal.UI, client atlas.Client, groupID string) ([]dataSourceCluster, []string, error) {
	if i.Template != "" {
		clusters, err := client.Clusters(groupID)
		if err != nil {
			return nil, nil, err
		}

		// If creating template app, only allow a 0 or 1 passed-in cluster names
		switch len(i.Clusters) {
		case 0, 1:
		default:
			return nil, nil, errors.New("template apps can only be created with one cluster")
		}

		if len(clusters) == 0 {
			// TODO(REALMC-9713): Enable the ability to create a new cluster upon template app creation
			return nil, nil, errors.New("please create an Atlas cluster before creating a template app")
		}

		var clusterName string
		if len(i.Clusters) == 1 {
			for _, c := range clusters {
				if c.Name == i.Clusters[0] {
					clusterName = i.Clusters[0]
					break
				}
			}
			if clusterName == "" {
				return nil, nil, fmt.Errorf("could not find Atlas cluster '%s'", i.Clusters[0])
			}
		} else if len(clusters) == 1 {
			clusterName = clusters[0].Name
		} else {
			clusterOptions := make([]string, 0, len(clusters))
			for _, c := range clusters {
				clusterOptions = append(clusterOptions, c.Name)
			}

			if err := ui.AskOne(&clusterName, &survey.Select{
				Message: "Select a cluster to link to your Realm application:",
				Options: clusterOptions,
			}); err != nil {
				return nil, nil, err
			}
		}

		if len(i.ClusterServiceNames) > 0 {
			ui.Print(terminal.NewTextLog("Overriding user-provided cluster service name(s). "+
				"The template app data source will be created with name '%s'", realm.DefaultServiceNameCluster))
		}

		i.Clusters = []string{clusterName}
		return []dataSourceCluster{{
			Name: realm.DefaultServiceNameCluster,
			Type: realm.ServiceTypeCluster,
			Config: configCluster{
				ClusterName:         clusterName,
				ReadPreference:      "primary",
				WireProtocolEnabled: false,
			},
			Version: 1,
		}}, nil, nil
	}

	// Non-template app creation cluster resolution
	dsClusters := make([]dataSourceCluster, 0, len(i.Clusters))
	nonExistingClusters := make([]string, 0, len(i.Clusters))
	if len(i.Clusters) > 0 {
		clusters, err := client.Clusters(groupID)
		if err != nil {
			return nil, nil, err
		}

		existingClusters := map[string]struct{}{}
		for _, c := range clusters {
			existingClusters[c.Name] = struct{}{}
		}

		for idx, clusterName := range i.Clusters {
			if _, ok := existingClusters[clusterName]; !ok {
				nonExistingClusters = append(nonExistingClusters, clusterName)
				continue
			}

			serviceName := clusterName
			if len(i.ClusterServiceNames) > idx {
				serviceName = i.ClusterServiceNames[idx]
			} else {
				if !ui.AutoConfirm() {
					if err := ui.AskOne(&serviceName, &survey.Input{
						Message: fmt.Sprintf("Enter a Service Name for Cluster '%s'", clusterName),
						Default: serviceName,
					}); err != nil {
						return nil, nil, err
					}
				}
			}

			dsClusters = append(dsClusters,
				dataSourceCluster{
					Name: serviceName,
					Type: realm.ServiceTypeCluster,
					Config: configCluster{
						ClusterName:         clusterName,
						ReadPreference:      "primary",
						WireProtocolEnabled: false,
					},
					Version: 1,
				})
		}
	}

	return dsClusters, nonExistingClusters, nil
}

func (i *createInputs) resolveServerlessInstances(ui terminal.UI, client atlas.Client, groupID string) ([]dataSourceCluster, []string, error) {
	if i.Template != "" && len(i.ServerlessInstances) > 0 {
		return nil, nil, errors.New("cannot create a template app with Serverless instances")
	}

	serverlessInstances, err := client.ServerlessInstances(groupID)
	if err != nil {
		return nil, nil, err
	}

	existingServerlessInstances := map[string]struct{}{}
	for _, d := range serverlessInstances {
		existingServerlessInstances[d.Name] = struct{}{}
	}
	nonExistingServerlessInstances := make([]string, 0, len(i.ServerlessInstances))

	dsServerlessInstances := make([]dataSourceCluster, 0, len(i.ServerlessInstances))
	for idx, serverlessInstanceName := range i.ServerlessInstances {
		if _, ok := existingServerlessInstances[serverlessInstanceName]; !ok {
			nonExistingServerlessInstances = append(nonExistingServerlessInstances, serverlessInstanceName)
			continue
		}

		serviceName := serverlessInstanceName
		if len(i.ServerlessInstanceServiceNames) > idx {
			serviceName = i.ServerlessInstanceServiceNames[idx]
		} else {
			if !ui.AutoConfirm() {
				if err := ui.AskOne(&serviceName, &survey.Input{
					Message: fmt.Sprintf("Enter a Service Name for Serverless instance '%s'", serverlessInstanceName),
					Default: serviceName,
				}); err != nil {
					return nil, nil, err
				}
			}
		}
		dsServerlessInstances = append(dsServerlessInstances,
			dataSourceCluster{
				Name: serviceName,
				Type: realm.ServiceTypeCluster,
				Config: configCluster{
					ClusterName:         serverlessInstanceName,
					ReadPreference:      "primary",
					WireProtocolEnabled: false,
				},
				Version: 1,
			})
	}

	return dsServerlessInstances, nonExistingServerlessInstances, nil
}

func (i *createInputs) resolveDatalakes(ui terminal.UI, client atlas.Client, groupID string) ([]dataSourceDatalake, []string, error) {
	if i.Template != "" && len(i.Datalakes) > 0 {
		return nil, nil, errors.New("cannot create a template app with data lakes")
	}

	datalakes, err := client.Datalakes(groupID)
	if err != nil {
		return nil, nil, err
	}

	existingDatalakes := map[string]struct{}{}
	for _, d := range datalakes {
		existingDatalakes[d.Name] = struct{}{}
	}
	nonExistingDatalakes := make([]string, 0, len(i.Datalakes))

	dsDatalakes := make([]dataSourceDatalake, 0, len(i.Datalakes))
	for idx, datalakeName := range i.Datalakes {
		if _, ok := existingDatalakes[datalakeName]; !ok {
			nonExistingDatalakes = append(nonExistingDatalakes, datalakeName)
			continue
		}

		serviceName := datalakeName
		if len(i.DatalakeServiceNames) > idx {
			serviceName = i.DatalakeServiceNames[idx]
		} else {
			if !ui.AutoConfirm() {
				if err := ui.AskOne(&serviceName, &survey.Input{
					Message: fmt.Sprintf("Enter a Service Name for Data Lake '%s'", datalakeName),
					Default: serviceName,
				}); err != nil {
					return nil, nil, err
				}
			}
		}
		dsDatalakes = append(dsDatalakes,
			dataSourceDatalake{
				Name: serviceName,
				Type: realm.ServiceTypeDatalake,
				Config: configDatalake{
					DatalakeName: datalakeName,
				},
			})
	}

	return dsDatalakes, nonExistingDatalakes, nil
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
		args = append(args, flags.Arg{flagRemoteApp, i.RemoteApp})
	}
	if i.LocalPath != "" {
		args = append(args, flags.Arg{flagLocalPathCreate, i.LocalPath})
	}
	if i.Template != "" {
		args = append(args, flags.Arg{flagTemplate, i.Template})
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
	for idx, clusterName := range i.Clusters {
		args = append(args, flags.Arg{flagCluster, clusterName})
		if len(i.ClusterServiceNames) > idx {
			args = append(args, flags.Arg{flagClusterServiceName, i.ClusterServiceNames[idx]})
		}
	}
	for idx, serverlessInstanceName := range i.ServerlessInstances {
		args = append(args, flags.Arg{flagServerlessInstance, serverlessInstanceName})
		if len(i.ServerlessInstanceServiceNames) > idx {
			args = append(args, flags.Arg{flagServerlessInstanceServiceName, i.ServerlessInstanceServiceNames[idx]})
		}
	}
	for idx, datalakeName := range i.Datalakes {
		args = append(args, flags.Arg{flagDatalake, datalakeName})
		if len(i.DatalakeServiceNames) > idx {
			args = append(args, flags.Arg{flagDatalakeServiceName, i.DatalakeServiceNames[idx]})
		}
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
