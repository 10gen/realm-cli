package app

import (
	"archive/zip"
	"errors"
	"fmt"
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

		cmd := &CommandCreate{createInputs{newAppInputs: newAppInputs{
			Name:            "test-app",
			Project:         "123",
			Location:        realm.LocationVirginia,
			DeploymentModel: realm.DeploymentModelGlobal,
		}}}

		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: client}))

		appLocal, err := local.LoadApp(filepath.Join(profile.WorkingDirectory, cmd.inputs.Name))
		assert.Nil(t, err)

		assert.Equal(t, &local.AppRealmConfigJSON{local.AppDataV2{local.AppStructureV2{
			ConfigVersion:   realm.DefaultAppConfigVersion,
			Name:            "test-app",
			Location:        realm.LocationVirginia,
			DeploymentModel: realm.DeploymentModelGlobal,
		}}}, appLocal.AppData)

		// TODO(REALMC-8262): Investigate file path display options
		dirLength := len(appLocal.RootDir)
		fmtStr := fmt.Sprintf("%%-%ds", dirLength)

		assert.Equal(t, strings.Join([]string{
			"01:23:45 UTC INFO  Successfully created app",
			fmt.Sprintf("  Info             "+fmtStr, "Details"),
			"  ---------------  " + strings.Repeat("-", dirLength),
			fmt.Sprintf("  Client App ID    "+fmtStr, "test-app-abcde"),
			"  Realm Directory  " + appLocal.RootDir,
			fmt.Sprintf("  Realm UI         "+fmtStr, "http://localhost:8080/groups/123/apps/456/dashboard"),
			"01:23:45 UTC DEBUG Check out your app: cd ./test-app && realm-cli app describe",
			"",
		}, "\n"), out.String())
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
		ac.GroupsFn = func() ([]atlas.Group, error) {
			return []atlas.Group{{ID: "123"}}, nil
		}

		cmd := &CommandCreate{createInputs{newAppInputs: newAppInputs{
			Name:            "test-app",
			Project:         "123",
			Location:        realm.LocationVirginia,
			DeploymentModel: realm.DeploymentModelGlobal,
		}}}

		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: rc, Atlas: ac}))

		console.Tty().Close() // flush the writers
		<-doneCh              // wait for procedure to complete

		appLocal, err := local.LoadApp(filepath.Join(profile.WorkingDirectory, cmd.inputs.Name))
		assert.Nil(t, err)

		expectedAppData := local.AppRealmConfigJSON{local.AppDataV2{local.AppStructureV2{
			ConfigVersion:   realm.DefaultAppConfigVersion,
			Name:            "test-app",
			Location:        realm.LocationVirginia,
			DeploymentModel: realm.DeploymentModelGlobal,
		}}}

		assert.Equal(t, &expectedAppData, appLocal.AppData)
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

	t.Run("should create a new app with a structure based on the specified remote app", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()
		profile.SetRealmBaseURL("http://localhost:8080")

		out, ui := mock.NewUI()

		testApp := realm.App{
			ID:          "789",
			GroupID:     "123",
			ClientAppID: "remote-app-abcde",
			Name:        "remote-app",
		}

		zipPkg, err := zip.OpenReader("testdata/project.zip")
		assert.Nil(t, err)

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

		cmd := &CommandCreate{createInputs{newAppInputs: newAppInputs{
			RemoteApp:       testApp.Name,
			Project:         testApp.GroupID,
			Location:        realm.LocationIreland,
			DeploymentModel: realm.DeploymentModelGlobal,
		}}}

		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: client}))

		appLocal, err := local.LoadApp(filepath.Join(profile.WorkingDirectory, cmd.inputs.RemoteApp))
		assert.Nil(t, err)

		assert.Equal(t, &local.AppRealmConfigJSON{local.AppDataV2{local.AppStructureV2{
			ConfigVersion:   realm.DefaultAppConfigVersion,
			Name:            testApp.Name,
			Location:        realm.LocationIreland,
			DeploymentModel: realm.DeploymentModelGlobal,
			Auth: &local.AuthStructure{
				CustomUserData: map[string]interface{}{"enabled": false},
				Providers:      map[string]interface{}{},
			},
			Sync: &local.SyncStructure{Config: map[string]interface{}{"development_mode_enabled": false}},
		}}}, appLocal.AppData)

		// TODO(REALMC-8262): Investigate file path display options
		dirLength := len(appLocal.RootDir)
		fmtStr := fmt.Sprintf("%%-%ds", dirLength)

		assert.Equal(t, strings.Join([]string{
			"01:23:45 UTC INFO  Successfully created app",
			fmt.Sprintf("  Info             "+fmtStr, "Details"),
			"  ---------------  " + strings.Repeat("-", dirLength),
			fmt.Sprintf("  Client App ID    "+fmtStr, "remote-app-abcde"),
			"  Realm Directory  " + appLocal.RootDir,
			fmt.Sprintf("  Realm UI         "+fmtStr, "http://localhost:8080/groups/123/apps/456/dashboard"),
			"01:23:45 UTC DEBUG Check out your app: cd ./remote-app && realm-cli app describe",
			"",
		}, "\n"), out.String())
	})

	for _, tc := range []struct {
		description string
		cluster     string
		dataLake    string
		atlasClient atlas.Client
	}{
		{
			description: "should create minimal project with a cluster data source when cluster is set",
			cluster:     "test-cluster",
			atlasClient: mock.AtlasClient{
				ClustersFn: func(groupID string) ([]atlas.Cluster, error) {
					return []atlas.Cluster{{Name: "test-cluster"}}, nil
				},
			},
		},
		{
			description: "should create minimal project with a data lake data source when data lake is set",
			dataLake:    "test-datalake",
			atlasClient: mock.AtlasClient{
				DataLakesFn: func(groupID string) ([]atlas.DataLake, error) {
					return []atlas.DataLake{{Name: "test-datalake"}}, nil
				},
			},
		},
		{
			description: "should create minimal project with a data lake and cluster data source when data lake and cluster is set",
			cluster:     "test-cluster",
			dataLake:    "test-datalake",
			atlasClient: mock.AtlasClient{
				ClustersFn: func(groupID string) ([]atlas.Cluster, error) {
					return []atlas.Cluster{{Name: "test-cluster"}}, nil
				},
				DataLakesFn: func(groupID string) ([]atlas.DataLake, error) {
					return []atlas.DataLake{{Name: "test-datalake"}}, nil
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

			cmd := &CommandCreate{
				inputs: createInputs{
					newAppInputs: newAppInputs{
						Name:            "test-app",
						Project:         "123",
						Location:        realm.LocationVirginia,
						DeploymentModel: realm.DeploymentModelGlobal,
					},
					Cluster:  tc.cluster,
					DataLake: tc.dataLake,
				},
			}

			assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: rc, Atlas: tc.atlasClient}))

			localApp, err := local.LoadApp(filepath.Join(profile.WorkingDirectory, cmd.inputs.Name))
			assert.Nil(t, err)

			assert.Equal(t, importAppData, localApp.AppData)
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
			dirLength := len(localApp.RootDir)
			fmtStr := fmt.Sprintf("%%-%ds", dirLength)

			var spaceBuffer, dashBuffer string
			if tc.dataLake != "" {
				spaceBuffer = "  "
				dashBuffer = "--"
			}

			display := make([]string, 0, 10)
			display = append(display, "01:23:45 UTC INFO  Successfully created app",
				fmt.Sprintf("  Info                   "+spaceBuffer+fmtStr, "Details"),
				"  ---------------------"+dashBuffer+"  "+strings.Repeat("-", dirLength),
				fmt.Sprintf("  Client App ID          "+spaceBuffer+fmtStr, "test-app-abcde"),
				"  Realm Directory        "+spaceBuffer+localApp.RootDir,
				fmt.Sprintf("  Realm UI               "+spaceBuffer+fmtStr, "http://localhost:8080/groups/123/apps/456/dashboard"),
			)
			if tc.cluster != "" {
				display = append(display, fmt.Sprintf("  Data Source (Cluster)  "+spaceBuffer+fmtStr, "mongodb-atlas"))
			}
			if tc.dataLake != "" {
				display = append(display, fmt.Sprintf("  Data Source (Data Lake)  "+fmtStr, "mongodb-datalake"))
			}
			display = append(display, "01:23:45 UTC DEBUG Check out your app: cd ./test-app && realm-cli app describe", "")
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
		description     string
		appRemote       string
		cluster         string
		dataLake        string
		datalake        string
		clients         cli.Clients
		displayExpected func(dir string, cmd *CommandCreate) string
	}{
		{
			description: "should create a minimal project dry run",
			displayExpected: func(dir string, cmd *CommandCreate) string {
				return strings.Join([]string{
					fmt.Sprintf("01:23:45 UTC INFO  A minimal Realm app would be created at %s", dir),
					"01:23:45 UTC DEBUG To create this app run: " + cmd.display(true),
					"",
				}, "\n")
			},
		},
		{
			description: "should create a dry run for the specified remote app",
			appRemote:   "remote-app",
			clients: cli.Clients{
				Realm: mock.RealmClient{
					FindAppsFn: func(filter realm.AppFilter) ([]realm.App, error) {
						return []realm.App{testApp}, nil
					},
				},
			},
			displayExpected: func(dir string, cmd *CommandCreate) string {
				return strings.Join([]string{
					fmt.Sprintf("01:23:45 UTC INFO  A Realm app based on the Realm app 'remote-app' would be created at %s", dir),
					"01:23:45 UTC DEBUG To create this app run: " + cmd.display(true),
					"",
				}, "\n")
			},
		},
		{
			description: "should create a minimal project dry run with cluster set",
			cluster:     "test-cluster",
			clients: cli.Clients{
				Atlas: mock.AtlasClient{
					ClustersFn: func(groupID string) ([]atlas.Cluster, error) {
						return []atlas.Cluster{{Name: "test-cluster"}}, nil
					},
				},
			},
			displayExpected: func(dir string, cmd *CommandCreate) string {
				return strings.Join([]string{
					fmt.Sprintf("01:23:45 UTC INFO  A minimal Realm app would be created at %s", dir),
					"01:23:45 UTC INFO  The cluster 'test-cluster' would be linked as data source 'mongodb-atlas'",
					"01:23:45 UTC DEBUG To create this app run: " + cmd.display(true),
					"",
				}, "\n")
			},
		},
		{
			description: "should create a minimal project dry run with data lake set",
			dataLake:    "test-datalake",
			clients: cli.Clients{
				Atlas: mock.AtlasClient{
					DataLakesFn: func(groupID string) ([]atlas.DataLake, error) {
						return []atlas.DataLake{{Name: "test-datalake"}}, nil
					},
				},
			},
			displayExpected: func(dir string, cmd *CommandCreate) string {
				return strings.Join([]string{
					fmt.Sprintf("01:23:45 UTC INFO  A minimal Realm app would be created at %s", dir),
					"01:23:45 UTC INFO  The data lake 'test-datalake' would be linked as data source 'mongodb-datalake'",
					"01:23:45 UTC DEBUG To create this app run: " + cmd.display(true),
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
						Location:        realm.LocationVirginia,
						DeploymentModel: realm.DeploymentModelGlobal,
					},
					Cluster:  tc.cluster,
					DataLake: tc.dataLake,
					DryRun:   true,
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
		cluster     string
		dataLake    string
		clients     cli.Clients
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
			cluster:     "test-cluster",
			clients: cli.Clients{
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
			dataLake:    "test-datalake",
			clients: cli.Clients{
				Atlas: mock.AtlasClient{
					DataLakesFn: func(groupID string) ([]atlas.DataLake, error) {
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
	} {
		t.Run(tc.description, func(t *testing.T) {
			profile := mock.NewProfileFromWd(t)

			cmd := &CommandCreate{createInputs{
				newAppInputs: newAppInputs{
					RemoteApp:       tc.appRemote,
					Project:         tc.groupID,
					Name:            "test-app",
					Location:        realm.LocationVirginia,
					DeploymentModel: realm.DeploymentModelGlobal,
				},
				Cluster:  tc.cluster,
				DataLake: tc.dataLake,
			}}

			assert.Equal(t, tc.expectedErr, cmd.Handler(profile, nil, tc.clients))
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
					Location:        realm.LocationIreland,
					DeploymentModel: realm.DeploymentModelLocal,
				},
				Directory: "realm-app",
				Cluster:   "Cluster0",
				DataLake:  "DataLake0",
				DryRun:    true,
			},
		}
		assert.Equal(t,
			cli.Name+" app create --project 123 --name test-app --remote remote-app --app-dir realm-app --location IE --deployment-model LOCAL --cluster Cluster0 --data-lake DataLake0 --dry-run",
			cmd.display(false),
		)
	})
}
