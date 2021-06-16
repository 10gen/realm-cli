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
		client.TemplatesFn = func() ([]realm.Template, error) {
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

		// TODO(REALMC-8262): Investigate file path display options
		dirLength := len(appLocal.RootDir)
		fmtStr := fmt.Sprintf("%%-%ds", dirLength)

		assert.Equal(t, strings.Join([]string{
			"Successfully created app",
			fmt.Sprintf("  Info             "+fmtStr, "Details"),
			"  ---------------  " + strings.Repeat("-", dirLength),
			fmt.Sprintf("  Client App ID    "+fmtStr, "test-app-abcde"),
			"  Realm Directory  " + appLocal.RootDir,
			fmt.Sprintf("  Realm UI         "+fmtStr, "http://localhost:8080/groups/123/apps/456/dashboard"),
			"Check out your app: cd ./test-app && realm-cli app describe",
			"",
		}, "\n"), out.String())

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
		rc.TemplatesFn = func() ([]realm.Template, error) {
			return []realm.Template{}, nil
		}
		ac := mock.AtlasClient{}
		ac.GroupsFn = func() ([]atlas.Group, error) {
			return []atlas.Group{{ID: "123"}}, nil
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

	// TODO REALMC-9228 Re-enable prompting for template selection
	//	t.Run("when a template is not provided should prompt for template selection", func(t *testing.T) {
	//		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
	//		defer teardown()
	//		profile.SetRealmBaseURL("http://localhost:8080")
	//
	//		procedure := func(c *expect.Console) {
	//			c.ExpectString("Please select a template from the available options")
	//			c.SendLine("palm-pilot.bitcoin-miner")
	//			c.ExpectEOF()
	//		}
	//
	//		// TODO(REALMC-8264): Mock console in tests does not behave as initially expected
	//		_, console, _, ui, consoleErr := mock.NewVT10XConsole()
	//		assert.Nil(t, consoleErr)
	//		defer console.Close()
	//
	//		doneCh := make(chan (struct{}))
	//		go func() {
	//			defer close(doneCh)
	//			procedure(console)
	//		}()
	//
	//		var createdApp realm.App
	//		rc := mock.RealmClient{}
	//		rc.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
	//			createdApp = realm.App{
	//				GroupID:     groupID,
	//				ID:          "456",
	//				ClientAppID: name + "-abcde",
	//				Name:        name,
	//				AppMeta:     meta,
	//			}
	//			return createdApp, nil
	//		}
	//		rc.ImportFn = func(groupID, appID string, appData interface{}) error {
	//			return nil
	//		}
	//		rc.TemplatesFn = func() ([]realm.Template, error) {
	//			return []realm.Template{
	//				{
	//					ID:   "palm-pilot.bitcoin-miner",
	//					Name: "Mine bitcoin on your Palm Pilot from the comfort of your home, electricity not included",
	//				},
	//				{
	//					ID:   "blackberry.important-business-app",
	//					Name: "Oh wow, a Blackberry... you must a very powerful, extravagant man.",
	//				},
	//			}, nil
	//		}
	//
	//		clientZipPkg, err := zip.OpenReader("testdata/react-native.zip")
	//		assert.Nil(t, err)
	//		rc.ClientTemplateFn = func(groupID, appID, templateID string) (*zip.Reader, error) {
	//			return &clientZipPkg.Reader, nil
	//		}
	//		ac := mock.AtlasClient{}
	//		ac.GroupsFn = func() ([]atlas.Group, error) {
	//			return []atlas.Group{{ID: "123"}}, nil
	//		}
	//
	//		cmd := &CommandCreate{createInputs{newAppInputs: newAppInputs{
	//			Name:            "template-app",
	//			Location:        realm.LocationVirginia,
	//			DeploymentModel: realm.DeploymentModelGlobal,
	//			ConfigVersion:   realm.DefaultAppConfigVersion,
	//		}}}
	//
	//		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: rc, Atlas: ac}))
	//
	//		console.Tty().Close() // flush the writers
	//		<-doneCh              // wait for procedure to complete
	//
	//		path := filepath.Join(profile.WorkingDirectory, cmd.inputs.Name, backendPath)
	//		appLocal, err := local.LoadApp(path)
	//		assert.Nil(t, err)
	//
	//		assert.Equal(t, appLocal.RootDir, path)
	//	})

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
		client.TemplatesFn = func() ([]realm.Template, error) {
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

				// TODO(REALMC-8262): Investigate file path display options
				dirLength := len(appLocal.RootDir)
				fmtStr := fmt.Sprintf("%%-%ds", dirLength)

				assert.Equal(t, strings.Join([]string{
					"Successfully created app",
					fmt.Sprintf("  Info             "+fmtStr, "Details"),
					"  ---------------  " + strings.Repeat("-", dirLength),
					fmt.Sprintf("  Client App ID    "+fmtStr, "remote-app-abcde"),
					"  Realm Directory  " + appLocal.RootDir,
					fmt.Sprintf("  Realm UI         "+fmtStr, "http://localhost:8080/groups/"+tc.expectedGroupID+"/apps/456/dashboard"),
					"Check out your app: cd ./remote-app && realm-cli app describe",
					"",
				}, "\n"), out.String())

				out.Reset()
			})
		}
	})

	t.Run("should create a new app from a template", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()
		profile.SetRealmBaseURL("http://localhost:8080")

		out, ui := mock.NewUI()

		testApp := realm.App{
			ID:          "789",
			GroupID:     "123",
			ClientAppID: "bitcoin-miner-abcde",
			Name:        "bitcoin-miner",
		}

		backendZipPkg, err := zip.OpenReader("testdata/bitcoin-miner.zip")
		assert.Nil(t, err)

		client := mock.RealmClient{
			FindAppsFn: func(filter realm.AppFilter) ([]realm.App, error) {
				return []realm.App{}, nil
			},
			ExportFn: func(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error) {
				return "", &backendZipPkg.Reader, err
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
			TemplatesFn: func() ([]realm.Template, error) {
				return []realm.Template{
					{
						ID:   "palm-pilot.bitcoin-miner",
						Name: "Mine bitcoin on your Palm Pilot from the comfort of your home, electricity not included",
					},
				}, nil
			},
		}

		frontendZipPkg, err := zip.OpenReader("testdata/react-native.zip")
		assert.Nil(t, err)
		client.ClientTemplateFn = func(groupID, appID, templateID string) (*zip.Reader, error) {
			return &frontendZipPkg.Reader, err
		}

		cmd := &CommandCreate{createInputs{newAppInputs: newAppInputs{
			Name:            "bitcoin-miner",
			Project:         testApp.GroupID,
			Template:        "palm-pilot.bitcoin-miner",
			Location:        realm.LocationIreland,
			DeploymentModel: realm.DeploymentModelGlobal,
		}}}

		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: client}))

		appLocal, err := local.LoadApp(filepath.Join(profile.WorkingDirectory, cmd.inputs.Name, backendPath))
		assert.Nil(t, err)

		backendFileInfo, err := ioutil.ReadDir(filepath.Join(profile.WorkingDirectory, cmd.inputs.Name, backendPath))
		assert.Nil(t, err)
		assert.Equal(t, len(backendFileInfo), 10)

		frontendFileInfo, err := ioutil.ReadDir(filepath.Join(profile.WorkingDirectory, cmd.inputs.Name, frontendPath))
		assert.Nil(t, err)
		assert.Equal(t, len(frontendFileInfo), 1)
		assert.Equal(t, frontendFileInfo[0].Name(), "react-native")

		dirLength := len(appLocal.RootDir)
		fmtStr := fmt.Sprintf("%%-%ds", dirLength)

		assert.Equal(t, strings.Join([]string{
			"Successfully created app",
			fmt.Sprintf("  Info             "+fmtStr, "Details"),
			"  ---------------  " + strings.Repeat("-", dirLength),
			fmt.Sprintf("  Client App ID    "+fmtStr, "bitcoin-miner-abcde"),
			"  Realm Directory  " + appLocal.RootDir,
			fmt.Sprintf("  Realm UI         "+fmtStr, "http://localhost:8080/groups/123/apps/456/dashboard"),
			"Check out your app: cd ./bitcoin-miner && realm-cli app describe",
			"",
		}, "\n"), out.String())
	})

	for _, tc := range []struct {
		description          string
		clusters             []string
		clusterServiceNames  []string
		datalakes            []string
		datalakeServiceNames []string
		atlasClient          atlas.Client
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
		},
		{
			description:         "should create minimal project with multiple cluster data sources when clusters are set",
			clusters:            []string{"test-cluster", "test-cluster-2"},
			clusterServiceNames: []string{"mongodb-atlas", "mongodb-atlas"},
			atlasClient: mock.AtlasClient{
				ClustersFn: func(groupID string) ([]atlas.Cluster, error) {
					return []atlas.Cluster{{Name: "test-cluster"}, {Name: "test-cluster-2"}}, nil
				},
			},
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
		},
		{
			description:          "should create minimal project with a data lake and cluster data source when data lake and cluster is set",
			clusters:             []string{"test-cluster"},
			clusterServiceNames:  []string{"mongodb-atlas"},
			datalakes:            []string{"test-datalake"},
			datalakeServiceNames: []string{"mongodb-datalake"},
			atlasClient: mock.AtlasClient{
				ClustersFn: func(groupID string) ([]atlas.Cluster, error) {
					return []atlas.Cluster{{Name: "test-cluster"}}, nil
				},
				DatalakesFn: func(groupID string) ([]atlas.Datalake, error) {
					return []atlas.Datalake{{Name: "test-datalake"}}, nil
				},
			},
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
			rc.TemplatesFn = func() ([]realm.Template, error) {
				return []realm.Template{}, nil
			}

			cmd := &CommandCreate{
				inputs: createInputs{
					newAppInputs: newAppInputs{
						Name:            "test-app",
						Project:         "123",
						Location:        realm.LocationVirginia,
						DeploymentModel: realm.DeploymentModelGlobal,
					},
					Clusters:             tc.clusters,
					ClusterServiceNames:  tc.clusterServiceNames,
					Datalakes:            tc.datalakes,
					DatalakeServiceNames: tc.datalakeServiceNames,
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

			// TODO(REALMC-8262): Investigate file path display options
			dirLength := len(appLocal.RootDir)
			fmtStr := fmt.Sprintf(" %%-%ds", dirLength)

			var spaceBuffer, dashBuffer string
			if len(tc.datalakes) > 0 {
				spaceBuffer = "  "
				dashBuffer = "--"
			}

			display := make([]string, 0, 10)
			display = append(display, "Successfully created app",
				fmt.Sprintf("  Info                   "+spaceBuffer+fmtStr, "Details"),
				"  ----------------------"+dashBuffer+"  "+strings.Repeat("-", dirLength),
				fmt.Sprintf("  Client App ID          "+spaceBuffer+fmtStr, "test-app-abcde"),
				"  Realm Directory         "+spaceBuffer+appLocal.RootDir,
				fmt.Sprintf("  Realm UI               "+spaceBuffer+fmtStr, "http://localhost:8080/groups/123/apps/456/dashboard"),
			)
			if len(tc.clusters) > 0 {
				display = append(display, fmt.Sprintf("  Data Source (Clusters) "+spaceBuffer+fmtStr, strings.Join(tc.clusters, ", ")))
			}
			if len(tc.datalakes) > 0 {
				display = append(display, fmt.Sprintf("  Data Source (Data Lakes) "+fmtStr, strings.Join(tc.datalakes, ", ")))
			}
			display = append(display, "Check out your app: cd ./test-app && realm-cli app describe", "")
			assert.Equal(t, strings.Join(display, "\n"), out.String())
		})
	}

	testApp := realm.App{
		ID:          "789",
		GroupID:     "123",
		ClientAppID: "remote-app-abcde",
		Name:        "remote-app",
	}

	for _, tc := range []struct {
		description          string
		appRemote            string
		clusters             []string
		clusterServiceNames  []string
		datalakes            []string
		datalakeServiceNames []string
		datalake             string
		clients              cli.Clients
		template             string
		displayExpected      func(dir string, cmd *CommandCreate) string
	}{
		{
			description: "should create a minimal project dry run",
			clients: cli.Clients{
				Realm: mock.RealmClient{
					TemplatesFn: func() ([]realm.Template, error) {
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
					TemplatesFn: func() ([]realm.Template, error) {
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
					TemplatesFn: func() ([]realm.Template, error) {
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
			description:          "should create a minimal project dry run with data lake set",
			datalakes:            []string{"test-datalake"},
			datalakeServiceNames: []string{"mongodb-datalake"},
			clients: cli.Clients{
				Realm: mock.RealmClient{
					TemplatesFn: func() ([]realm.Template, error) {
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
					TemplatesFn: func() ([]realm.Template, error) {
						return []realm.Template{
							{
								ID:   "palm-pilot.bitcoin-miner",
								Name: "Mine bitcoin on your Palm Pilot from the comfort of your home, electricity not included",
							},
						}, nil
					},
				},
			},
			displayExpected: func(dir string, cmd *CommandCreate) string {
				return strings.Join([]string{
					fmt.Sprintf("A minimal Realm app would be created at %s/backend using the 'palm-pilot.bitcoin-miner' template", dir),
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
					Clusters:             tc.clusters,
					ClusterServiceNames:  tc.clusterServiceNames,
					Datalakes:            tc.datalakes,
					DatalakeServiceNames: tc.datalakeServiceNames,
					DryRun:               true,
				},
			}

			assert.Nil(t, cmd.Handler(profile, ui, tc.clients))

			expectedDir := filepath.Join(profile.WorkingDirectory, "test-app")
			assert.Equal(t, tc.displayExpected(expectedDir, cmd), out.String())
		})
	}

	for _, tc := range []struct {
		description string
		appRemote   string
		groupID     string
		clusters    []string
		datalakes   []string
		template    string
		clients     cli.Clients
		uiOptions   mock.UIOptions
		expectedErr error
	}{
		{
			description: "should error when resolving groupID when project is not set",
			clients: cli.Clients{
				Atlas: mock.AtlasClient{
					GroupsFn: func() ([]atlas.Group, error) {
						return nil, errors.New("atlas client error")
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
					TemplatesFn: func() ([]realm.Template, error) {
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
			description: "should error when resolving data lakes when data lake is set",
			groupID:     "123",
			datalakes:   []string{"test-datalake"},
			clients: cli.Clients{
				Realm: mock.RealmClient{
					TemplatesFn: func() ([]realm.Template, error) {
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
					TemplatesFn: func() ([]realm.Template, error) {
						return nil, errors.New("unable to find available templates")
					},
				},
				Atlas: mock.AtlasClient{
					GroupsFn: func() ([]atlas.Group, error) {
						return []atlas.Group{{ID: "123"}}, nil
					},
				},
			},
			expectedErr: errors.New("unable to find available templates"),
		},
		{
			description: "should error when the requested template is not available",
			template:    "palm-pilot.bitcoin-miner",
			groupID:     "123",
			clients: cli.Clients{
				Realm: mock.RealmClient{
					TemplatesFn: func() ([]realm.Template, error) {
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
					TemplatesFn: func() ([]realm.Template, error) {
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
					TemplatesFn: func() ([]realm.Template, error) {
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
				Clusters:  tc.clusters,
				Datalakes: tc.datalakes,
			}}

			out := new(bytes.Buffer)
			ui := mock.NewUIWithOptions(tc.uiOptions, out)

			assert.Equal(t, tc.expectedErr, cmd.Handler(profile, ui, tc.clients))
		})
	}
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
				LocalPath:            "realm-app",
				Clusters:             []string{"Cluster0"},
				ClusterServiceNames:  []string{"mongodb-atlas"},
				Datalakes:            []string{"Datalake0"},
				DatalakeServiceNames: []string{"mongodb-datalake"},
				DryRun:               true,
			},
		}
		assert.Equal(t,
			cli.Name+" app create --project 123 --name test-app --remote remote-app --local realm-app --template palm-pilot.bitcoin-miner --location IE --deployment-model LOCAL --cluster Cluster0 --cluster-service-name mongodb-atlas --datalake Datalake0 --datalake-service-name mongodb-datalake --dry-run",
			cmd.display(false),
		)
	})

	t.Run("should create a command with multiple input clusters and data lakes", func(t *testing.T) {
		cmd := &CommandCreate{
			inputs: createInputs{
				newAppInputs: newAppInputs{
					Name:            "test-app",
					Project:         "123",
					RemoteApp:       "remote-app",
					Location:        realm.LocationIreland,
					DeploymentModel: realm.DeploymentModelLocal,
				},
				LocalPath:            "realm-app",
				Clusters:             []string{"Cluster0", "Cluster1", "Cluster2"},
				ClusterServiceNames:  []string{"mongodb-atlas-0", "mongodb-atlas-1", "mongodb-atlas-2"},
				Datalakes:            []string{"Datalake0", "Datalake1", "Datalake2"},
				DatalakeServiceNames: []string{"mongodb-datalake-0", "mongodb-datalake-1", "mongodb-datalake-2"},
				DryRun:               true,
			},
		}
		assert.Equal(t,
			cli.Name+" app create --project 123 --name test-app --remote remote-app --local realm-app --location IE --deployment-model LOCAL --cluster Cluster0 --cluster-service-name mongodb-atlas-0 --cluster Cluster1 --cluster-service-name mongodb-atlas-1 --cluster Cluster2 --cluster-service-name mongodb-atlas-2 --datalake Datalake0 --datalake-service-name mongodb-datalake-0 --datalake Datalake1 --datalake-service-name mongodb-datalake-1 --datalake Datalake2 --datalake-service-name mongodb-datalake-2 --dry-run",
			cmd.display(false),
		)
	})
}
