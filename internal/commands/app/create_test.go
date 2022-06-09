package app

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"github.com/Netflix/go-expect"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAppCreateHandler(t *testing.T) {
	t.Run("should create minimal project when no remote type is specified", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()
		profile.SetRealmBaseURL("http://localhost:8080")

		out, ui := mock.NewUI()

		client := mock.RealmClient{}
		client.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return realm.App{
				GroupID:     groupID,
				ID:          "456",
				ClientAppID: name + "-abcde",
				Name:        name,
				AppMeta:     meta,
			}, nil
		}
		client.ImportFn = func(groupID, appID string, appData interface{}) error {
			return nil
		}
		client.AllTemplatesFn = func() ([]realm.Template, error) {
			return []realm.Template{}, nil
		}

		cmd := &CommandCreate{createInputs{newAppInputs: newAppInputs{
			Name:            "test-app",
			Project:         "123",
			Location:        realm.LocationVirginia,
			DeploymentModel: realm.DeploymentModelGlobal,
			ConfigVersion:   realm.DefaultAppConfigVersion,
		}}}

		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: client}))

		fullDir := filepath.Join(profile.WorkingDirectory, cmd.inputs.Name)

		appLocal, err := local.LoadApp(fullDir)
		assert.Nil(t, err)

		assert.Equal(t, &local.AppRealmConfigJSON{local.AppDataV2{local.AppStructureV2{
			ConfigVersion:   realm.DefaultAppConfigVersion,
			ID:              "test-app-abcde",
			Name:            "test-app",
			Location:        realm.LocationVirginia,
			DeploymentModel: realm.DeploymentModelGlobal,
			Environments: map[string]map[string]interface{}{
				"development.json": {
					"values": map[string]interface{}{},
				},
				"no-environment.json": {
					"values": map[string]interface{}{},
				},
				"production.json": {
					"values": map[string]interface{}{},
				},
				"qa.json": {
					"values": map[string]interface{}{},
				},
				"testing.json": {
					"values": map[string]interface{}{},
				},
			},
			Auth: local.AuthStructure{
				CustomUserData: map[string]interface{}{"enabled": false},
				Providers: map[string]interface{}{
					"api-key": map[string]interface{}{
						"name":     "api-key",
						"type":     "api-key",
						"disabled": true,
					},
				},
			},
			Sync: local.SyncStructure{Config: map[string]interface{}{"development_mode_enabled": false}},
			Functions: local.FunctionsStructure{
				Configs: []map[string]interface{}{},
				Sources: map[string]string{},
			},
			GraphQL: local.GraphQLStructure{
				Config: map[string]interface{}{
					"use_natural_pluralization": true,
				},
				CustomResolvers: []map[string]interface{}{},
			},
			Values: []map[string]interface{}{},
		}}}, appLocal.AppData)

		assert.Equal(t, fmt.Sprintf(`Successfully created app
{
  "client_app_id": "test-app-abcde",
  "filepath": %q,
  "url": "http://localhost:8080/groups/123/apps/456/dashboard"
}
Check out your app: cd ./test-app && realm-cli app describe
`, appLocal.RootDir), out.String())

		t.Run("should have the expected contents in the auth custom user data file", func(t *testing.T) {
			config, err := ioutil.ReadFile(filepath.Join(fullDir, local.NameAuth, local.FileCustomUserData.String()))
			assert.Nil(t, err)
			assert.Equal(t, `{
    "enabled": false
}
`, string(config))
		})

		t.Run("should have the expected contents in the auth providers file", func(t *testing.T) {
			config, err := ioutil.ReadFile(filepath.Join(fullDir, local.NameAuth, local.FileProviders.String()))
			assert.Nil(t, err)
			assert.Equal(t, `{
    "api-key": {
        "disabled": true,
        "name": "api-key",
        "type": "api-key"
    }
}
`, string(config))
		})

		t.Run("should have data sources directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(fullDir, local.NameDataSources))
			assert.Nil(t, err)
		})

		t.Run("should have the expected contents in the functions config file", func(t *testing.T) {
			config, err := ioutil.ReadFile(filepath.Join(fullDir, local.NameFunctions, local.FileConfig.String()))
			assert.Nil(t, err)
			assert.Equal(t, "[]\n", string(config))
		})

		t.Run("should have graphql custom resolvers directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(fullDir, local.NameGraphQL, local.NameCustomResolvers))
			assert.Nil(t, err)
		})

		t.Run("should have the expected contents in the graphql config file", func(t *testing.T) {
			config, err := ioutil.ReadFile(filepath.Join(fullDir, local.NameGraphQL, local.FileConfig.String()))
			assert.Nil(t, err)
			assert.Equal(t, `{
    "use_natural_pluralization": true
}
`, string(config))
		})

		t.Run("should have http endpoints directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(fullDir, local.NameHTTPEndpoints))
			assert.Nil(t, err)
		})

		t.Run("should have services directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(fullDir, local.NameServices))
			assert.Nil(t, err)
		})

		t.Run("should have the expected contents in the sync config file", func(t *testing.T) {
			config, err := ioutil.ReadFile(filepath.Join(fullDir, local.NameSync, local.FileConfig.String()))
			assert.Nil(t, err)
			assert.Equal(t, `{
    "development_mode_enabled": false
}
`, string(config))
		})

		t.Run("should have values directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(fullDir, local.NameValues))
			assert.Nil(t, err)
		})
	})

	t.Run("when remote and project are not set should create minimal project and prompt for project", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()
		profile.SetRealmBaseURL("http://localhost:8080")

		procedure := func(c *expect.Console) {
			c.ExpectString("Atlas Project")
			c.Send("123")
			c.SendLine(" ")
			c.ExpectEOF()
		}

		// TODO(REALMC-8264): Mock console in tests does not behave as initially expected
		_, console, _, ui, consoleErr := mock.NewVT10XConsole()
		assert.Nil(t, consoleErr)
		defer console.Close()

		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)
			procedure(console)
		}()

		var createdApp realm.App
		rc := mock.RealmClient{}
		rc.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			createdApp = realm.App{
				GroupID:     groupID,
				ID:          "456",
				ClientAppID: name + "-abcde",
				Name:        name,
				AppMeta:     meta,
			}
			return createdApp, nil
		}
		rc.ImportFn = func(groupID, appID string, appData interface{}) error {
			return nil
		}
		ac := mock.AtlasClient{}
		ac.GroupsFn = func(url string, useBaseURL bool) (atlas.Groups, error) {
			return atlas.Groups{Results: []atlas.Group{{ID: "123"}}}, nil
		}

		cmd := &CommandCreate{createInputs{newAppInputs: newAppInputs{
			Name:            "test-app",
			Project:         "123",
			Location:        realm.LocationVirginia,
			DeploymentModel: realm.DeploymentModelGlobal,
			ConfigVersion:   realm.DefaultAppConfigVersion,
		}}}

		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: rc, Atlas: ac}))

		console.Tty().Close() // flush the writers
		<-doneCh              // wait for procedure to complete

		appLocal, err := local.LoadApp(filepath.Join(profile.WorkingDirectory, cmd.inputs.Name))
		assert.Nil(t, err)

		assert.Equal(t, &local.AppRealmConfigJSON{local.AppDataV2{local.AppStructureV2{
			ConfigVersion:   realm.DefaultAppConfigVersion,
			ID:              "test-app-abcde",
			Name:            "test-app",
			Location:        realm.LocationVirginia,
			DeploymentModel: realm.DeploymentModelGlobal,
			Environments: map[string]map[string]interface{}{
				"development.json": {
					"values": map[string]interface{}{},
				},
				"no-environment.json": {
					"values": map[string]interface{}{},
				},
				"production.json": {
					"values": map[string]interface{}{},
				},
				"qa.json": {
					"values": map[string]interface{}{},
				},
				"testing.json": {
					"values": map[string]interface{}{},
				},
			},
			Auth: local.AuthStructure{
				CustomUserData: map[string]interface{}{"enabled": false},
				Providers: map[string]interface{}{
					"api-key": map[string]interface{}{
						"name":     "api-key",
						"type":     "api-key",
						"disabled": true,
					},
				},
			},
			Sync: local.SyncStructure{Config: map[string]interface{}{"development_mode_enabled": false}},
			Functions: local.FunctionsStructure{
				Configs: []map[string]interface{}{},
				Sources: map[string]string{},
			},
			GraphQL: local.GraphQLStructure{
				Config: map[string]interface{}{
					"use_natural_pluralization": true,
				},
				CustomResolvers: []map[string]interface{}{},
			},
			Values: []map[string]interface{}{},
		}}}, appLocal.AppData)

		assert.Equal(t, realm.App{
			ID:          "456",
			GroupID:     "123",
			Name:        "test-app",
			ClientAppID: "test-app-abcde",
			AppMeta: realm.AppMeta{
				Location:        realm.LocationVirginia,
				DeploymentModel: realm.DeploymentModelGlobal,
			},
		}, createdApp)
	})

	t.Run("should not prompt for a template if template flag is not declared", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()
		profile.SetRealmBaseURL("http://localhost:8080")

		procedure := func(c *expect.Console) {
			c.ExpectString("Enter a Service Name for Cluster 'test-cluster")
			c.SendLine("test-cluster")
			c.ExpectEOF()
		}

		// TODO(REALMC-8264): Mock console in tests does not behave as initially expected
		_, console, _, ui, consoleErr := mock.NewVT10XConsole()
		assert.Nil(t, consoleErr)
		defer console.Close()

		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)
			procedure(console)
		}()

		testApp := realm.App{
			ID:          "456",
			GroupID:     "123",
			ClientAppID: "bitcoin-miner-abcde",
			Name:        "bitcoin-miner",
		}

		defaultBackendZipPkg, err := zip.OpenReader("testdata/project.zip")
		assert.Nil(t, err)
		defer defaultBackendZipPkg.Close()

		realmClient := mock.RealmClient{
			FindAppsFn: func(filter realm.AppFilter) ([]realm.App, error) {
				return []realm.App{}, nil
			},
			ExportFn: func(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error) {
				return "", &defaultBackendZipPkg.Reader, nil
			},
			CreateAppFn: func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
				return realm.App{
					GroupID:     groupID,
					ID:          "456",
					ClientAppID: name + "-abcde",
					Name:        name,
					AppMeta:     meta,
				}, nil
			},
			ImportFn: func(groupID, appID string, appData interface{}) error {
				return nil
			},
		}
		atlasClient := mock.AtlasClient{
			ClustersFn: func(groupID string) ([]atlas.Cluster, error) {
				return []atlas.Cluster{{Name: "test-cluster"}}, nil
			},
		}

		cmd := &CommandCreate{createInputs{
			newAppInputs: newAppInputs{
				Name:            "bitcoin-miner",
				Project:         testApp.GroupID,
				Location:        realm.LocationIreland,
				DeploymentModel: realm.DeploymentModelGlobal,
			},
			Clusters: []string{"test-cluster"},
		}}

		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: realmClient, Atlas: atlasClient}))

		appPath := filepath.Join(profile.WorkingDirectory, cmd.inputs.Name)
		_, err = local.LoadApp(appPath)
		assert.Equal(t, err, errors.New("failed to find app at "+appPath))

		// we expect for the command to have created a default app. we will assert that realm_config.json exists and that
		// backend/ and frontend/ do not exist in the directory.
		_, err = os.Stat(filepath.Join(profile.WorkingDirectory, cmd.inputs.Name, "realm_config.json"))
		assert.False(t, os.IsNotExist(err), "expected realm_config.json to exist")

		_, err = os.Stat(filepath.Join(profile.WorkingDirectory, cmd.inputs.Name, local.BackendPath))
		assert.True(t, os.IsNotExist(err), "expected for backend path to not exist")

		_, err = os.Stat(filepath.Join(profile.WorkingDirectory, cmd.inputs.Name, local.FrontendPath))
		assert.True(t, os.IsNotExist(err), "expected for frontend path to not exist")
	})

	t.Run("should create a new app with a structure based on the specified remote app", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()
		profile.SetRealmBaseURL("http://localhost:8080")

		out := new(bytes.Buffer)
		ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

		testApp := realm.App{
			ID:          "789",
			GroupID:     "123",
			ClientAppID: "remote-app-abcde",
			Name:        "remote-app",
		}

		zipPkg, err := zip.OpenReader("testdata/project.zip")
		assert.Nil(t, err)
		defer zipPkg.Close()

		client := mock.RealmClient{}
		client.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{testApp}, nil
		}
		client.ExportFn = func(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error) {
			return "", &zipPkg.Reader, err
		}
		client.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return realm.App{
				GroupID:     groupID,
				ID:          "456",
				ClientAppID: name + "-abcde",
				Name:        name,
				AppMeta:     meta,
			}, nil
		}
		client.ImportFn = func(groupID, appID string, appData interface{}) error {
			return nil
		}
		client.AllTemplatesFn = func() ([]realm.Template, error) {
			return []realm.Template{}, nil
		}

		for _, tc := range []struct {
			description     string
			groupID         string
			localPath       string
			expectedGroupID string
		}{
			{
				description:     "with default remote project",
				localPath:       testApp.Name,
				expectedGroupID: testApp.GroupID,
			},
			{
				description:     "with provided project",
				groupID:         "1111111",
				localPath:       testApp.Name + "-1",
				expectedGroupID: "1111111",
			},
		} {
			t.Run(tc.description, func(t *testing.T) {

				cmd := &CommandCreate{createInputs{newAppInputs: newAppInputs{
					RemoteApp:       testApp.Name,
					Project:         tc.groupID,
					Location:        realm.LocationIreland,
					DeploymentModel: realm.DeploymentModelGlobal,
				}}}

				assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: client}))

				appLocal, err := local.LoadApp(filepath.Join(profile.WorkingDirectory, tc.localPath))
				assert.Nil(t, err)

				assert.Equal(t, &local.AppRealmConfigJSON{local.AppDataV2{local.AppStructureV2{
					ConfigVersion:   realm.DefaultAppConfigVersion,
					Name:            testApp.Name,
					Location:        realm.LocationIreland,
					DeploymentModel: realm.DeploymentModelGlobal,
					Auth: local.AuthStructure{
						CustomUserData: map[string]interface{}{"enabled": false},
						Providers:      map[string]interface{}{},
					},
					Sync: local.SyncStructure{Config: map[string]interface{}{"development_mode_enabled": false}},
					Functions: local.FunctionsStructure{
						Sources: map[string]string{},
					},
					GraphQL: local.GraphQLStructure{
						CustomResolvers: []map[string]interface{}{},
					},
					Values: []map[string]interface{}{},
				}}}, appLocal.AppData)

				assert.Equal(t, fmt.Sprintf(`Successfully created app
{
  "client_app_id": "remote-app-abcde",
  "filepath": %q,
  "url": "http://localhost:8080/groups/%s/apps/456/dashboard"
}
Check out your app: cd ./remote-app && realm-cli app describe
`, appLocal.RootDir, tc.expectedGroupID), out.String())

				out.Reset()
			})
		}
	})

	t.Run("should create a new app from a template", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()
		profile.SetRealmBaseURL("http://localhost:8080")

		procedure := func(c *expect.Console) {
			c.ExpectString("Please select a template from the available options")
			c.SendLine("palm-pilot.bitcoin-miner")
			c.ExpectString("Enter a Service Name for Cluster 'test-cluster")
			c.SendLine("test-cluster")
			c.ExpectEOF()
		}

		// TODO(REALMC-8264): Mock console in tests does not behave as initially expected
		out := new(bytes.Buffer)
		console, _, ui, consoleErr := mock.NewVT10XConsoleWithOptions(mock.UIOptions{AutoConfirm: true}, out)
		assert.Nil(t, consoleErr)
		defer console.Close()

		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)
			procedure(console)
		}()

		testApp := realm.App{
			ID:          "456",
			GroupID:     "123",
			ClientAppID: "bitcoin-miner-abcde",
			Name:        "bitcoin-miner",
		}

		backendZipPkg, err := zip.OpenReader("testdata/project.zip")
		assert.Nil(t, err)
		defer backendZipPkg.Close()

		templateID := "palm-pilot.bitcoin-miner"
		client := mock.RealmClient{
			FindAppsFn: func(filter realm.AppFilter) ([]realm.App, error) {
				return []realm.App{}, nil
			},
			ExportFn: func(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error) {
				return "", &backendZipPkg.Reader, nil
			},
			CreateAppFn: func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
				return realm.App{
					GroupID:     groupID,
					ID:          "456",
					ClientAppID: name + "-abcde",
					Name:        name,
					AppMeta:     meta,
				}, nil
			},
			ImportFn: func(groupID, appID string, appData interface{}) error {
				return nil
			},
			AllTemplatesFn: func() ([]realm.Template, error) {
				return []realm.Template{
					{
						ID:   templateID,
						Name: "Mine bitcoin on your Palm Pilot from the comfort of your home, electricity not included",
					},
				}, nil
			},
		}

		frontendZipPkg, err := zip.OpenReader("testdata/react-native.zip")
		assert.Nil(t, err)
		defer frontendZipPkg.Close()

		client.ClientTemplateFn = func(groupID, appID, templateID string) (*zip.Reader, bool, error) {
			return &frontendZipPkg.Reader, true, err
		}

		cmd := &CommandCreate{createInputs{
			newAppInputs: newAppInputs{
				Name:            "bitcoin-miner",
				Project:         testApp.GroupID,
				Template:        templateID,
				Location:        realm.LocationIreland,
				DeploymentModel: realm.DeploymentModelGlobal,
			},
			Clusters: []string{"test-cluster"},
		}}

		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: client, Atlas: mock.AtlasClient{
			ClustersFn: func(groupID string) ([]atlas.Cluster, error) {
				return []atlas.Cluster{{Name: "test-cluster"}}, nil
			},
		}}))

		console.Tty().Close() // flush the writers
		<-doneCh              // wait for procedure to complete

		_, err = local.LoadApp(filepath.Join(profile.WorkingDirectory, cmd.inputs.Name, local.BackendPath))
		assert.Nil(t, err)

		backendFileInfo, err := ioutil.ReadDir(filepath.Join(profile.WorkingDirectory, cmd.inputs.Name, local.BackendPath))
		assert.Nil(t, err)
		assert.Equal(t, len(backendFileInfo), 6)

		frontendFileInfo, err := ioutil.ReadDir(filepath.Join(profile.WorkingDirectory, cmd.inputs.Name, local.FrontendPath, templateID))
		assert.Nil(t, err)
		assert.Equal(t, len(frontendFileInfo), 1)
		assert.Equal(t, frontendFileInfo[0].Name(), "react-native")
	})

	for _, tc := range []struct {
		description                    string
		clusters                       []string
		clusterServiceNames            []string
		serverlessInstances            []string
		serverlessInstanceServiceNames []string
		datalakes                      []string
		datalakeServiceNames           []string
		atlasClient                    atlas.Client
		dataSourceOutput               string
	}{
		{
			description:         "should create minimal project with a cluster data source when cluster is set",
			clusters:            []string{"test-cluster"},
			clusterServiceNames: []string{"mongodb-atlas"},
			atlasClient: mock.AtlasClient{
				ClustersFn: func(groupID string) ([]atlas.Cluster, error) {
					return []atlas.Cluster{{Name: "test-cluster"}}, nil
				},
			},
			dataSourceOutput: `"clusters": [
    {
      "name": "mongodb-atlas"
    }
  ]`,
		},
		{
			description:         "should create minimal project with multiple cluster data sources when clusters are set",
			clusters:            []string{"test-cluster", "test-cluster-2"},
			clusterServiceNames: []string{"mongodb-atlas-1", "mongodb-atlas-2"},
			atlasClient: mock.AtlasClient{
				ClustersFn: func(groupID string) ([]atlas.Cluster, error) {
					return []atlas.Cluster{{Name: "test-cluster"}, {Name: "test-cluster-2"}}, nil
				},
			},
			dataSourceOutput: `"clusters": [
    {
      "name": "mongodb-atlas-1"
    },
    {
      "name": "mongodb-atlas-2"
    }
  ]`,
		},
		{
			description:                    "should create minimal project with a serverless instance data source when serverless instance is set",
			serverlessInstances:            []string{"test-serverless-instance"},
			serverlessInstanceServiceNames: []string{"mongodb-atlas"},
			atlasClient: mock.AtlasClient{
				ServerlessInstancesFn: func(groupID string) ([]atlas.ServerlessInstance, error) {
					return []atlas.ServerlessInstance{{Name: "test-serverless-instance"}}, nil
				},
			},
			dataSourceOutput: `"serverless_instances": [
    {
      "name": "mongodb-atlas"
    }
  ]`,
		},
		{
			description:          "should create minimal project with a data lake data source when data lake is set",
			datalakes:            []string{"test-datalake"},
			datalakeServiceNames: []string{"mongodb-datalake"},
			atlasClient: mock.AtlasClient{
				DatalakesFn: func(groupID string) ([]atlas.Datalake, error) {
					return []atlas.Datalake{{Name: "test-datalake"}}, nil
				},
			},
			dataSourceOutput: `"datalakes": [
    {
      "name": "mongodb-datalake"
    }
  ]`,
		},
		{
			description:                    "should create minimal project with a cluster and serverless instance and data lake data source when cluster and serverless instance and data lake is set",
			clusters:                       []string{"test-cluster"},
			clusterServiceNames:            []string{"mongodb-atlas"},
			serverlessInstances:            []string{"test-serverless-instance"},
			serverlessInstanceServiceNames: []string{"mongodb-atlas-1"},
			datalakes:                      []string{"test-datalake"},
			datalakeServiceNames:           []string{"mongodb-datalake"},
			atlasClient: mock.AtlasClient{
				ClustersFn: func(groupID string) ([]atlas.Cluster, error) {
					return []atlas.Cluster{{Name: "test-cluster"}}, nil
				},
				ServerlessInstancesFn: func(groupID string) ([]atlas.ServerlessInstance, error) {
					return []atlas.ServerlessInstance{{Name: "test-serverless-instance"}}, nil
				},
				DatalakesFn: func(groupID string) ([]atlas.Datalake, error) {
					return []atlas.Datalake{{Name: "test-datalake"}}, nil
				},
			},
			dataSourceOutput: `"clusters": [
    {
      "name": "mongodb-atlas"
    }
  ],
  "serverless_instances": [
    {
      "name": "mongodb-atlas-1"
    }
  ],
  "datalakes": [
    {
      "name": "mongodb-datalake"
    }
  ]`,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
			defer teardown()
			profile.SetRealmBaseURL("http://localhost:8080")

			out, ui := mock.NewUI()

			var createdApp realm.App
			rc := mock.RealmClient{}
			rc.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
				createdApp = realm.App{
					GroupID:     groupID,
					ID:          "456",
					ClientAppID: name + "-abcde",
					Name:        name,
					AppMeta:     meta,
				}
				return createdApp, nil
			}

			var importAppData interface{}
			rc.ImportFn = func(groupID, appID string, appData interface{}) error {
				importAppData = appData
				return nil
			}
			rc.AllTemplatesFn = func() ([]realm.Template, error) {
				return []realm.Template{}, nil
			}

			cmd := &CommandCreate{
				inputs: createInputs{
					newAppInputs: newAppInputs{
						Name:            "test-app",
						Project:         "123",
						Location:        realm.LocationVirginia,
						DeploymentModel: realm.DeploymentModelGlobal,
						ConfigVersion:   realm.DefaultAppConfigVersion,
					},
					Clusters:                       tc.clusters,
					ClusterServiceNames:            tc.clusterServiceNames,
					ServerlessInstances:            tc.serverlessInstances,
					ServerlessInstanceServiceNames: tc.serverlessInstanceServiceNames,
					Datalakes:                      tc.datalakes,
					DatalakeServiceNames:           tc.datalakeServiceNames,
				},
			}

			assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: rc, Atlas: tc.atlasClient}))

			appLocal, err := local.LoadApp(filepath.Join(profile.WorkingDirectory, cmd.inputs.Name))
			assert.Nil(t, err)

			assert.Equal(t, importAppData, appLocal.AppData)
			assert.Equal(t, realm.App{
				ID:          "456",
				GroupID:     "123",
				Name:        "test-app",
				ClientAppID: "test-app-abcde",
				AppMeta: realm.AppMeta{
					Location:        realm.LocationVirginia,
					DeploymentModel: realm.DeploymentModelGlobal,
				},
			}, createdApp)

			assert.Equal(t, fmt.Sprintf(`Successfully created app
{
  "client_app_id": "test-app-abcde",
  "filepath": %q,
  "url": "http://localhost:8080/groups/123/apps/456/dashboard",
  %s
}
Check out your app: cd ./test-app && realm-cli app describe
`, appLocal.RootDir, tc.dataSourceOutput), out.String())
		})
	}

	testApp := realm.App{
		ID:          "789",
		GroupID:     "123",
		ClientAppID: "remote-app-abcde",
		Name:        "remote-app",
	}

	for _, tc := range []struct {
		description                    string
		appRemote                      string
		clusters                       []string
		clusterServiceNames            []string
		serverlessInstances            []string
		serverlessInstanceServiceNames []string
		datalakes                      []string
		datalakeServiceNames           []string
		clients                        cli.Clients
		template                       string
		displayExpected                func(dir string, cmd *CommandCreate) string
	}{
		{
			description: "should create a minimal project dry run",
			clients: cli.Clients{
				Realm: mock.RealmClient{
					AllTemplatesFn: func() ([]realm.Template, error) {
						return []realm.Template{}, nil
					},
				},
			},
			displayExpected: func(dir string, cmd *CommandCreate) string {
				return strings.Join([]string{
					fmt.Sprintf("A minimal Realm app would be created at %s", dir),
					"To create this app run: " + cmd.display(true),
					"",
				}, "\n")
			},
		},
		{
			description: "should create a dry run for the specified remote app",
			appRemote:   "remote-app",
			clients: cli.Clients{
				Realm: mock.RealmClient{
					AllTemplatesFn: func() ([]realm.Template, error) {
						return []realm.Template{}, nil
					},
					FindAppsFn: func(filter realm.AppFilter) ([]realm.App, error) {
						return []realm.App{testApp}, nil
					},
				},
			},
			displayExpected: func(dir string, cmd *CommandCreate) string {
				return strings.Join([]string{
					fmt.Sprintf("A Realm app based on the Realm app 'remote-app' would be created at %s", dir),
					"To create this app run: " + cmd.display(true),
					"",
				}, "\n")
			},
		},
		{
			description:         "should create a minimal project dry run with cluster set",
			clusters:            []string{"test-cluster"},
			clusterServiceNames: []string{"mongodb-atlas"},
			clients: cli.Clients{
				Realm: mock.RealmClient{
					AllTemplatesFn: func() ([]realm.Template, error) {
						return []realm.Template{}, nil
					},
				},
				Atlas: mock.AtlasClient{
					ClustersFn: func(groupID string) ([]atlas.Cluster, error) {
						return []atlas.Cluster{{Name: "test-cluster"}}, nil
					},
				},
			},
			displayExpected: func(dir string, cmd *CommandCreate) string {
				return strings.Join([]string{
					fmt.Sprintf("A minimal Realm app would be created at %s", dir),
					"The cluster 'test-cluster' would be linked as data source 'mongodb-atlas'",
					"To create this app run: " + cmd.display(true),
					"",
				}, "\n")
			},
		},
		{
			description:                    "should create a minimal project dry run with serverless instance set",
			serverlessInstances:            []string{"test-serverless-instance"},
			serverlessInstanceServiceNames: []string{"mongodb-atlas"},
			clients: cli.Clients{
				Realm: mock.RealmClient{
					AllTemplatesFn: func() ([]realm.Template, error) {
						return []realm.Template{}, nil
					},
				},
				Atlas: mock.AtlasClient{
					ServerlessInstancesFn: func(groupID string) ([]atlas.ServerlessInstance, error) {
						return []atlas.ServerlessInstance{{Name: "test-serverless-instance"}}, nil
					},
				},
			},
			displayExpected: func(dir string, cmd *CommandCreate) string {
				return strings.Join([]string{
					fmt.Sprintf("A minimal Realm app would be created at %s", dir),
					"The serverless instance 'test-serverless-instance' would be linked as data source 'mongodb-atlas'",
					"To create this app run: " + cmd.display(true),
					"",
				}, "\n")
			},
		},
		{
			description:          "should create a minimal project dry run with data lake set",
			datalakes:            []string{"test-datalake"},
			datalakeServiceNames: []string{"mongodb-datalake"},
			clients: cli.Clients{
				Realm: mock.RealmClient{
					AllTemplatesFn: func() ([]realm.Template, error) {
						return []realm.Template{}, nil
					},
				},
				Atlas: mock.AtlasClient{
					DatalakesFn: func(groupID string) ([]atlas.Datalake, error) {
						return []atlas.Datalake{{Name: "test-datalake"}}, nil
					},
				},
			},
			displayExpected: func(dir string, cmd *CommandCreate) string {
				return strings.Join([]string{
					fmt.Sprintf("A minimal Realm app would be created at %s", dir),
					"The data lake 'test-datalake' would be linked as data source 'mongodb-datalake'",
					"To create this app run: " + cmd.display(true),
					"",
				}, "\n")
			},
		},
		{
			description: "should create a minimal project dry run when using a template",
			template:    "palm-pilot.bitcoin-miner",
			clients: cli.Clients{
				Realm: mock.RealmClient{
					AllTemplatesFn: func() ([]realm.Template, error) {
						return []realm.Template{
							{
								ID:   "palm-pilot.bitcoin-miner",
								Name: "Mine bitcoin on your Palm Pilot from the comfort of your home, electricity not included",
							},
						}, nil
					},
				},
				Atlas: mock.AtlasClient{
					ClustersFn: func(groupID string) ([]atlas.Cluster, error) {
						return []atlas.Cluster{
							{Name: "cluster0"},
						}, nil
					},
				},
			},
			clusters: []string{"cluster0"},
			displayExpected: func(dir string, cmd *CommandCreate) string {
				return strings.Join([]string{
					fmt.Sprintf("A Realm app would be created at %s using the 'palm-pilot.bitcoin-miner' template", dir),
					"The cluster 'cluster0' would be linked as data source 'mongodb-atlas'",
					"To create this app run: " + cmd.display(true),
					"",
				}, "\n")
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
			defer teardown()
			profile.SetRealmBaseURL("http://localhost:8080")

			out, ui := mock.NewUI()

			cmd := &CommandCreate{
				inputs: createInputs{
					newAppInputs: newAppInputs{
						RemoteApp:       tc.appRemote,
						Name:            "test-app",
						Project:         "123",
						Template:        tc.template,
						Location:        realm.LocationVirginia,
						DeploymentModel: realm.DeploymentModelGlobal,
					},
					Clusters:                       tc.clusters,
					ClusterServiceNames:            tc.clusterServiceNames,
					ServerlessInstances:            tc.serverlessInstances,
					ServerlessInstanceServiceNames: tc.serverlessInstanceServiceNames,
					Datalakes:                      tc.datalakes,
					DatalakeServiceNames:           tc.datalakeServiceNames,
					DryRun:                         true,
				},
			}

			assert.Nil(t, cmd.Handler(profile, ui, tc.clients))

			expectedDir := filepath.Join(profile.WorkingDirectory, "test-app")
			assert.Equal(t, tc.displayExpected(expectedDir, cmd), out.String())
		})
	}

	for _, tc := range []struct {
		description         string
		appRemote           string
		groupID             string
		clusters            []string
		serverlessInstances []string
		datalakes           []string
		template            string
		clients             cli.Clients
		uiOptions           mock.UIOptions
		expectedErr         error
	}{
		{
			description: "should error when resolving groupID when project is not set",
			clients: cli.Clients{
				Atlas: mock.AtlasClient{
					GroupsFn: func(url string, useBaseURL bool) (atlas.Groups, error) {
						return atlas.Groups{}, errors.New("atlas client error")
					},
				},
			},
			expectedErr: errors.New("atlas client error"),
		},
		{
			description: "should error when resolving clusters when cluster is set",
			groupID:     "123",
			clusters:    []string{"test-cluster"},
			clients: cli.Clients{
				Realm: mock.RealmClient{
					AllTemplatesFn: func() ([]realm.Template, error) {
						return []realm.Template{}, nil
					},
				},
				Atlas: mock.AtlasClient{
					ClustersFn: func(groupID string) ([]atlas.Cluster, error) {
						return nil, errors.New("atlas client error")
					},
				},
			},
			expectedErr: errors.New("atlas client error"),
		},
		{
			description:         "should error when resolving serverless instances when serverless instance is set",
			groupID:             "123",
			serverlessInstances: []string{"test-serverless-instance"},
			clients: cli.Clients{
				Realm: mock.RealmClient{
					AllTemplatesFn: func() ([]realm.Template, error) {
						return []realm.Template{}, nil
					},
				},
				Atlas: mock.AtlasClient{
					ServerlessInstancesFn: func(groupID string) ([]atlas.ServerlessInstance, error) {
						return nil, errors.New("atlas client error")
					},
				},
			},
			expectedErr: errors.New("atlas client error"),
		},
		{
			description: "should error when resolving data lakes when data lake is set",
			groupID:     "123",
			datalakes:   []string{"test-datalake"},
			clients: cli.Clients{
				Realm: mock.RealmClient{
					AllTemplatesFn: func() ([]realm.Template, error) {
						return []realm.Template{}, nil
					},
				},
				Atlas: mock.AtlasClient{
					DatalakesFn: func(groupID string) ([]atlas.Datalake, error) {
						return nil, errors.New("atlas client error")
					},
				},
			},
			expectedErr: errors.New("atlas client error"),
		},
		{
			description: "should error when resolving app when remote is set",
			appRemote:   "remote-app",
			clients: cli.Clients{
				Realm: mock.RealmClient{
					FindAppsFn: func(filter realm.AppFilter) ([]realm.App, error) {
						return nil, errors.New("realm client error")
					},
				},
			},
			expectedErr: errors.New("realm client error"),
		},
		{
			description: "should error when fetching templates fails",
			groupID:     "123",
			clients: cli.Clients{
				Realm: mock.RealmClient{
					AllTemplatesFn: func() ([]realm.Template, error) {
						return nil, errors.New("unable to find available templates")
					},
				},
				Atlas: mock.AtlasClient{
					GroupsFn: func(url string, useBaseURL bool) (atlas.Groups, error) {
						return atlas.Groups{Results: []atlas.Group{{ID: "123"}}}, nil
					},
				},
			},
			template:    "ios.template.todo",
			expectedErr: errors.New("unable to find available templates"),
		},
		{
			description: "should error when the requested template is not available",
			template:    "palm-pilot.bitcoin-miner",
			groupID:     "123",
			clients: cli.Clients{
				Realm: mock.RealmClient{
					AllTemplatesFn: func() ([]realm.Template, error) {
						return []realm.Template{}, nil
					},
				},
			},
			expectedErr: errors.New("unable to find template 'palm-pilot.bitcoin-miner'"),
		},
		{
			description: "should not error when no templates are available and a template id has not been provided",
			groupID:     "123",
			clients: cli.Clients{
				Realm: mock.RealmClient{
					AllTemplatesFn: func() ([]realm.Template, error) {
						return []realm.Template{}, nil
					},
					CreateAppFn: func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
						return realm.App{
							GroupID:     "123",
							ID:          "456",
							ClientAppID: "app-abcde",
							Name:        "app",
						}, nil
					},
					ImportFn: func(groupID, appID string, appData interface{}) error {
						return nil
					},
				},
			},
			expectedErr: nil,
		},
		{
			description: "should not error when using auto confirm and no template id has been provided",
			groupID:     "123",
			uiOptions: mock.UIOptions{
				AutoConfirm: true,
			},
			clients: cli.Clients{
				Realm: mock.RealmClient{
					AllTemplatesFn: func() ([]realm.Template, error) {
						return []realm.Template{
							{
								ID:   "palm-pilot.bitcoin-miner",
								Name: "Mine bitcoin on your Palm Pilot from the comfort of your home, electricity not included",
							},
						}, nil
					},
					CreateAppFn: func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
						return realm.App{
							GroupID:     "123",
							ID:          "456",
							ClientAppID: "app-abcde",
							Name:        "app",
						}, nil
					},
					ImportFn: func(groupID, appID string, appData interface{}) error {
						return nil
					},
				},
			},
			expectedErr: nil,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			profile, teardown := mock.NewProfileFromTmpDir(t, "test-app-errors")
			defer teardown()

			cmd := &CommandCreate{createInputs{
				newAppInputs: newAppInputs{
					RemoteApp:       tc.appRemote,
					Project:         tc.groupID,
					Name:            "test-app",
					Template:        tc.template,
					Location:        realm.LocationVirginia,
					DeploymentModel: realm.DeploymentModelGlobal,
				},
				Clusters:            tc.clusters,
				ServerlessInstances: tc.serverlessInstances,
				Datalakes:           tc.datalakes,
			}}

			out := new(bytes.Buffer)
			ui := mock.NewUIWithOptions(tc.uiOptions, out)

			assert.Equal(t, tc.expectedErr, cmd.Handler(profile, ui, tc.clients))
		})
	}

	t.Run("should prompt user for confirmation if input data sources are not found", func(t *testing.T) {
		testApp := realm.App{
			ID:          primitive.NewObjectID().Hex(),
			GroupID:     primitive.NewObjectID().Hex(),
			ClientAppID: "test-app-abcde",
			Name:        "test-app",
		}

		rc := mock.RealmClient{}
		rc.ImportFn = func(groupID, appID string, appData interface{}) error {
			return nil
		}
		rc.AllTemplatesFn = func() ([]realm.Template, error) {
			return []realm.Template{}, nil
		}

		ac := mock.AtlasClient{}
		ac.ClustersFn = func(groupID string) ([]atlas.Cluster, error) {
			return []atlas.Cluster{
				{ID: "789", Name: "test-cluster-1"},
			}, nil
		}
		ac.ServerlessInstancesFn = func(groupID string) ([]atlas.ServerlessInstance, error) {
			return []atlas.ServerlessInstance{
				{ID: "1011", Name: "test-serverless-instance-1"},
			}, nil
		}
		ac.DatalakesFn = func(groupID string) ([]atlas.Datalake, error) {
			return []atlas.Datalake{
				{Name: "test-datalake-1"},
			}, nil
		}

		dummyDatalakes := []string{"test-dummy-lake-1", "test-dummy-lake-2"}
		dummyServerlessInstances := []string{"test-serverless-instance-dummy-1", "test-serverless-instance-dummy-2"}
		dummyClusters := []string{"test-cluster-dummy-1", "test-cluster-dummy-2"}

		inputs := createInputs{
			newAppInputs:                   newAppInputs{Name: testApp.Name, Project: testApp.GroupID},
			Clusters:                       []string{"test-cluster-1", dummyClusters[0], dummyClusters[1]},
			ClusterServiceNames:            []string{"mongodb-atlas"},
			ServerlessInstances:            []string{"test-serverless-instance-1", dummyServerlessInstances[0], dummyServerlessInstances[1]},
			ServerlessInstanceServiceNames: []string{"mongodb-atlas-1"},
			Datalakes:                      []string{"test-datalake-1", dummyDatalakes[0], dummyDatalakes[1]},
			DatalakeServiceNames:           []string{"mongodb-datalake"},
		}

		for _, tc := range []struct {
			description string
			response    string
			appCreated  bool
		}{
			{
				description: "and quit if not confirmed",
				response:    "no",
			},
			{
				description: "and continue to create app if confirmed",
				response:    "yes",
				appCreated:  true,
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
				defer teardown()

				_, console, _, ui, err := mock.NewVT10XConsole()
				assert.Nil(t, err)
				defer console.Close()

				doneCh := make(chan (struct{}))
				go func() {
					defer close(doneCh)
					console.ExpectString("Note: The following data sources were not linked because they could not be found: 'test-cluster-dummy-1', 'test-cluster-dummy-2', 'test-serverless-instance-dummy-1', 'test-serverless-instance-dummy-2', 'test-dummy-lake-1', 'test-dummy-lake-2'")
					console.ExpectString("Would you still like to create the app?")
					console.SendLine(tc.response)
					console.ExpectEOF()
				}()

				var appCreated bool
				rc.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
					appCreated = true
					return testApp, nil
				}

				cmd := &CommandCreate{inputs}
				assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: rc, Atlas: ac}))
				assert.Equal(t, tc.appCreated, appCreated)
			})
		}
	})
}

func TestAppCreateCommandDisplay(t *testing.T) {
	t.Run("should create a minimal command", func(t *testing.T) {
		cmd := &CommandCreate{
			inputs: createInputs{
				newAppInputs: newAppInputs{
					Name:            "test-app",
					Project:         "123",
					Location:        realm.LocationVirginia,
					DeploymentModel: realm.DeploymentModelGlobal,
				},
			},
		}
		assert.Equal(t, cli.Name+" app create --project 123 --name test-app", cmd.display(false))
	})

	t.Run("should create a command with all inputs", func(t *testing.T) {
		cmd := &CommandCreate{
			inputs: createInputs{
				newAppInputs: newAppInputs{
					Name:            "test-app",
					Project:         "123",
					RemoteApp:       "remote-app",
					Template:        "palm-pilot.bitcoin-miner",
					Location:        realm.LocationIreland,
					DeploymentModel: realm.DeploymentModelLocal,
				},
				LocalPath:                      "realm-app",
				Clusters:                       []string{"Cluster0"},
				ClusterServiceNames:            []string{"mongodb-atlas"},
				ServerlessInstances:            []string{"ServerlessInstance0"},
				ServerlessInstanceServiceNames: []string{"mongodb-atlas-1"},
				Datalakes:                      []string{"Datalake0"},
				DatalakeServiceNames:           []string{"mongodb-datalake"},
				DryRun:                         true,
			},
		}
		assert.Equal(t,
			cli.Name+" app create --project 123 --name test-app --remote remote-app --local realm-app --template palm-pilot.bitcoin-miner --location IE --deployment-model LOCAL --cluster Cluster0 --cluster-service-name mongodb-atlas --serverless-instance ServerlessInstance0 --serverless-instance-service-name mongodb-atlas-1 --datalake Datalake0 --datalake-service-name mongodb-datalake --dry-run",
			cmd.display(false),
		)
	})

	t.Run("should create a command with multiple input clusters and serverless instrances and data lakes", func(t *testing.T) {
		cmd := &CommandCreate{
			inputs: createInputs{
				newAppInputs: newAppInputs{
					Name:            "test-app",
					Project:         "123",
					RemoteApp:       "remote-app",
					Template:        "palm-pilot.bitcoin-miner",
					Location:        realm.LocationIreland,
					DeploymentModel: realm.DeploymentModelLocal,
				},
				LocalPath:                      "realm-app",
				Clusters:                       []string{"Cluster0", "Cluster1", "Cluster2"},
				ClusterServiceNames:            []string{"mongodb-atlas-0", "mongodb-atlas-1", "mongodb-atlas-2"},
				ServerlessInstances:            []string{"ServerlessInstance0", "ServerlessInstance1", "ServerlessInstance2"},
				ServerlessInstanceServiceNames: []string{"mongodb-atlas-3", "mongodb-atlas-4", "mongodb-atlas-5"},
				Datalakes:                      []string{"Datalake0", "Datalake1", "Datalake2"},
				DatalakeServiceNames:           []string{"mongodb-datalake-0", "mongodb-datalake-1", "mongodb-datalake-2"},
				DryRun:                         true,
			},
		}
		assert.Equal(t,
			cli.Name+" app create --project 123 --name test-app --remote remote-app --local realm-app --template palm-pilot.bitcoin-miner --location IE --deployment-model LOCAL --cluster Cluster0 --cluster-service-name mongodb-atlas-0 --cluster Cluster1 --cluster-service-name mongodb-atlas-1 --cluster Cluster2 --cluster-service-name mongodb-atlas-2 --serverless-instance ServerlessInstance0 --serverless-instance-service-name mongodb-atlas-3 --serverless-instance ServerlessInstance1 --serverless-instance-service-name mongodb-atlas-4 --serverless-instance ServerlessInstance2 --serverless-instance-service-name mongodb-atlas-5 --datalake Datalake0 --datalake-service-name mongodb-datalake-0 --datalake Datalake1 --datalake-service-name mongodb-datalake-1 --datalake Datalake2 --datalake-service-name mongodb-datalake-2 --dry-run",
			cmd.display(false),
		)
	})
}
