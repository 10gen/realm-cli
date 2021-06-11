package app

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
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

func TestAppCreateInputsResolve(t *testing.T) {
	t.Run("with no flags set should prompt for just name and set location deployment model and environment to defaults", func(t *testing.T) {
		profile := mock.NewProfile(t)

		_, console, _, ui, consoleErr := mock.NewVT10XConsole()
		assert.Nil(t, consoleErr)
		defer console.Close()

		procedure := func(c *expect.Console) {
			c.ExpectString("App Name")
			c.SendLine("test-app")
			c.ExpectEOF()
		}

		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)
			procedure(console)
		}()

		inputs := createInputs{}
		assert.Nil(t, inputs.Resolve(profile, ui))

		console.Tty().Close() // flush the writers
		<-doneCh              // wait for procedure to complete

		assert.Equal(t, "test-app", inputs.Name)
		assert.Equal(t, flagDeploymentModelDefault, inputs.DeploymentModel)
		assert.Equal(t, flagLocationDefault, inputs.Location)
		assert.Equal(t, realm.EnvironmentNone, inputs.Environment)
	})
	t.Run("with a name flag set should prompt for nothing else and set location deployment model and environment to defaults", func(t *testing.T) {
		profile := mock.NewProfile(t)

		inputs := createInputs{newAppInputs: newAppInputs{Name: "test-app"}}
		assert.Nil(t, inputs.Resolve(profile, nil))

		assert.Equal(t, "test-app", inputs.Name)
		assert.Equal(t, flagDeploymentModelDefault, inputs.DeploymentModel)
		assert.Equal(t, flagLocationDefault, inputs.Location)
		assert.Equal(t, realm.EnvironmentNone, inputs.Environment)
	})
	t.Run("with name location deployment model and environment flags set should prompt for nothing else", func(t *testing.T) {
		profile := mock.NewProfile(t)

		inputs := createInputs{newAppInputs: newAppInputs{
			Name:            "test-app",
			DeploymentModel: realm.DeploymentModelLocal,
			Location:        realm.LocationOregon,
			Environment:     realm.EnvironmentDevelopment,
		}}
		assert.Nil(t, inputs.Resolve(profile, nil))

		assert.Equal(t, "test-app", inputs.Name)
		assert.Equal(t, realm.DeploymentModelLocal, inputs.DeploymentModel)
		assert.Equal(t, realm.LocationOregon, inputs.Location)
		assert.Equal(t, realm.EnvironmentDevelopment, inputs.Environment)
	})
}

func TestAppCreateInputsResolveName(t *testing.T) {
	testApp := realm.App{
		ID:          primitive.NewObjectID().Hex(),
		GroupID:     primitive.NewObjectID().Hex(),
		ClientAppID: "test-app-abcde",
		Name:        "test-app",
	}

	for _, tc := range []struct {
		description    string
		inputs         createInputs
		appRemote      realm.App
		expectedName   string
		expectedFilter realm.AppFilter
	}{
		{
			description:  "should return name if name is set",
			inputs:       createInputs{newAppInputs: newAppInputs{Name: testApp.Name}},
			expectedName: testApp.Name,
		},
		{
			description:    "should use remote app for name if name is not set",
			appRemote:      realm.App{GroupID: testApp.GroupID, ClientAppID: testApp.ClientAppID},
			expectedName:   testApp.Name,
			expectedFilter: realm.AppFilter{GroupID: testApp.GroupID, App: testApp.ClientAppID},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			var appFilter realm.AppFilter
			rc := mock.RealmClient{}
			rc.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				appFilter = filter
				return []realm.App{testApp}, nil
			}

			err := tc.inputs.resolveName(nil, rc, tc.appRemote.GroupID, tc.appRemote.ClientAppID)

			assert.Nil(t, err)
			assert.Equal(t, tc.expectedName, tc.inputs.Name)
			assert.Equal(t, tc.expectedFilter, appFilter)
		})
	}

	t.Run("should error when finding app", func(t *testing.T) {
		var appFilter realm.AppFilter
		rc := mock.RealmClient{}
		rc.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			appFilter = filter
			return nil, errors.New("realm client error")
		}
		inputs := createInputs{}
		err := inputs.resolveName(nil, rc, testApp.GroupID, testApp.ClientAppID)

		assert.Equal(t, errors.New("realm client error"), err)
		assert.Equal(t, "", inputs.Name)
		assert.Equal(t, realm.AppFilter{GroupID: testApp.GroupID, App: testApp.ClientAppID}, appFilter)
	})
}

func TestAppCreateInputsResolveDirectory(t *testing.T) {
	t.Run("should return path of wd with app name appended", func(t *testing.T) {
		profile := mock.NewProfileFromWd(t)

		appName := "test-app"
		inputs := createInputs{newAppInputs: newAppInputs{Name: appName}}

		dir, err := inputs.resolveLocalPath(nil, profile.WorkingDirectory)

		assert.Nil(t, err)
		assert.Equal(t, path.Join(profile.WorkingDirectory, appName), dir)
	})

	t.Run("should return path of wd with directory appended when local path is set", func(t *testing.T) {
		profile := mock.NewProfileFromWd(t)

		specifiedPath := "test-dir"
		inputs := createInputs{LocalPath: specifiedPath}

		dir, err := inputs.resolveLocalPath(nil, profile.WorkingDirectory)

		assert.Nil(t, err)
		assert.Equal(t, path.Join(profile.WorkingDirectory, specifiedPath), dir)
	})

	t.Run("should return path of wd with app name appended even with file of app name in wd", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()

		appName := "test-app"
		inputs := createInputs{newAppInputs: newAppInputs{Name: appName}}

		testFile, err := os.Create(appName)
		assert.Nil(t, err)
		assert.Nil(t, testFile.Close())

		dir, err := inputs.resolveLocalPath(nil, profile.WorkingDirectory)

		assert.Nil(t, err)
		assert.Equal(t, path.Join(profile.WorkingDirectory, appName), dir)
		assert.Nil(t, os.Remove(appName))
	})

	t.Run("should return path of wd with a new app name appended trying to write to a local directory", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()

		_, console, _, ui, err := mock.NewVT10XConsole()
		assert.Nil(t, err)
		defer console.Close()

		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)

			console.ExpectString("Local path './test-app' already exists, writing app contents to that destination may result in file conflicts.")
			console.ExpectString("Would you still like to write app contents to './test-app'? ('No' will prompt you to provide another destination)")
			console.SendLine("no")
			console.ExpectString("Local Path")
			console.SendLine("new-app")
			console.ExpectEOF()
		}()

		inputs := createInputs{newAppInputs: newAppInputs{Name: "test-app"}}

		err = os.Mkdir(path.Join(profile.WorkingDirectory, "test-app"), os.ModePerm)
		assert.Nil(t, err)

		dir, err := inputs.resolveLocalPath(ui, profile.WorkingDirectory)
		assert.Nil(t, err)
		assert.Equal(t, path.Join(profile.WorkingDirectory, "new-app"), dir)
		assert.Equal(t, "new-app", inputs.LocalPath)
	})

	t.Run("should fail when realm app already exists in current wd", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()

		_, ui := mock.NewUI()

		existingApp := realm.App{
			ID:          primitive.NewObjectID().Hex(),
			GroupID:     primitive.NewObjectID().Hex(),
			ClientAppID: "existing-app-abcde",
			Name:        "existing-app",
		}

		rc := mock.RealmClient{}
		rc.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return existingApp, nil
		}
		rc.ImportFn = func(groupID, appID string, appData interface{}) error {
			return nil
		}

		inputs := createInputs{newAppInputs: newAppInputs{
			Name:    existingApp.Name,
			Project: existingApp.GroupID,
		}}
		cmd := &CommandCreate{inputs}
		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: rc}))

		existingDir := filepath.Join(profile.WorkingDirectory, existingApp.Name)
		dir, err := inputs.resolveLocalPath(ui, existingDir)

		assert.Equal(t, errProjectExists{existingDir}, err)
		assert.Equal(t, dir, "")
	})

	t.Run("should fail when realm app already exists in newly specified path", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()

		_, console, _, ui, err := mock.NewVT10XConsole()
		assert.Nil(t, err)
		defer console.Close()

		existingApp := realm.App{
			ID:          primitive.NewObjectID().Hex(),
			GroupID:     primitive.NewObjectID().Hex(),
			ClientAppID: "existing-app-abcde",
			Name:        "existing-app",
		}

		rc := mock.RealmClient{}
		rc.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return existingApp, nil
		}
		rc.ImportFn = func(groupID, appID string, appData interface{}) error {
			return nil
		}

		inputs := createInputs{newAppInputs: newAppInputs{
			Name:    existingApp.Name,
			Project: existingApp.GroupID,
		}}
		cmd := &CommandCreate{inputs}
		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: rc}))

		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)
			console.ExpectString("Local path './existing-app' already exists, writing app contents to that destination may result in file conflicts.")
			console.ExpectString("Would you still like to write app contents to './existing-app'? ('No' will prompt you to provide another destination)")
			console.SendLine("no")
			console.ExpectString("Local Path")
			console.SendLine("existing-app")
			console.ExpectEOF()
		}()

		dir, err := inputs.resolveLocalPath(ui, profile.WorkingDirectory)

		console.Tty().Close() // flush the writers
		<-doneCh              // wait for procedure to complete

		assert.Equal(t, errProjectExists{existingApp.Name}, err)
		assert.Equal(t, dir, "")

	})

	t.Run("should create default local directory name when ui is set to auto confirm", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()

		_, ui, err := mock.NewConsoleWithOptions(
			mock.UIOptions{AutoConfirm: true},
			new(bytes.Buffer),
		)
		assert.Nil(t, err)

		testAppName := "test-app"
		newDefaultName := "test-app-1"
		inputs := createInputs{newAppInputs: newAppInputs{Name: testAppName}}

		err = os.Mkdir(path.Join(profile.WorkingDirectory, testAppName), os.ModePerm)
		assert.Nil(t, err)

		dir, err := inputs.resolveLocalPath(ui, profile.WorkingDirectory)
		assert.Nil(t, err)
		assert.Equal(t, path.Join(profile.WorkingDirectory, newDefaultName), dir)
	})

	t.Run("should request new path when path specified is another realm app", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()

		_, console, _, ui, err := mock.NewVT10XConsole()
		assert.Nil(t, err)
		defer console.Close()

		specifiedDir := "test-app-dir"
		inputs := createInputs{newAppInputs: newAppInputs{Name: "test-app"}, LocalPath: specifiedDir}
		fullDir := path.Join(profile.WorkingDirectory, specifiedDir)

		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)

			console.ExpectString("Local path './test-app-dir' already exists, writing app contents to that destination may result in file conflicts.")
			console.ExpectString("Would you still like to write app contents to './test-app-dir'? ('No' will prompt you to provide another destination)")
			console.SendLine("no")
			console.ExpectString("Local Path")
			console.SendLine("new-app")
			console.ExpectEOF()
		}()

		appLocal := local.NewApp(
			fullDir,
			"test-app-abcde",
			"test-app",
			flagLocationDefault,
			flagDeploymentModelDefault,
			realm.EnvironmentNone,
			realm.DefaultAppConfigVersion,
		)
		assert.Nil(t, appLocal.WriteConfig())

		dir, err := inputs.resolveLocalPath(ui, profile.WorkingDirectory)
		assert.Nil(t, err)
		assert.Equal(t, path.Join(profile.WorkingDirectory, "new-app"), dir)
	})
}

func TestAppCreateInputsResolveCluster(t *testing.T) {
	t.Run("should return data source config of a provided cluster", func(t *testing.T) {
		_, ui := mock.NewUI()
		var expectedGroupID string
		ac := mock.AtlasClient{}
		ac.ClustersFn = func(groupID string) ([]atlas.Cluster, error) {
			expectedGroupID = groupID
			return []atlas.Cluster{{ID: "789", Name: "test-cluster"}}, nil
		}

		inputs := createInputs{newAppInputs: newAppInputs{Name: "test-app"}, Clusters: []string{"test-cluster"}}

		ds, err := inputs.resolveClusters(ui, ac, "123")
		assert.Nil(t, err)

		assert.Equal(t, []dataSourceCluster{
			{
				Name: "mongodb-atlas",
				Type: realm.ClusterType,
				Config: configCluster{
					ClusterName:         "test-cluster",
					ReadPreference:      "primary",
					WireProtocolEnabled: false,
				},
			},
		}, ds)
		assert.Equal(t, "123", expectedGroupID)
	})

	t.Run("should return data source configs of multiple provided clusters", func(t *testing.T) {
		_, ui := mock.NewUI()
		var expectedGroupID string
		ac := mock.AtlasClient{}
		ac.ClustersFn = func(groupID string) ([]atlas.Cluster, error) {
			expectedGroupID = groupID
			return []atlas.Cluster{
				{ID: "789", Name: "test-cluster-1"},
				{ID: "1011", Name: "test-cluster-2"},
			}, nil
		}

		inputs := createInputs{newAppInputs: newAppInputs{Name: "test-app"}, Clusters: []string{"test-cluster-1", "test-cluster-2"}}

		ds, err := inputs.resolveClusters(ui, ac, "123")
		assert.Nil(t, err)

		assert.Equal(t, []dataSourceCluster{
			{
				Name: "mongodb-atlas",
				Type: realm.ClusterType,
				Config: configCluster{
					ClusterName:         "test-cluster-1",
					ReadPreference:      "primary",
					WireProtocolEnabled: false,
				},
			},
			{
				Name: "mongodb-atlas",
				Type: realm.ClusterType,
				Config: configCluster{
					ClusterName:         "test-cluster-2",
					ReadPreference:      "primary",
					WireProtocolEnabled: false,
				},
			},
		}, ds)
		assert.Equal(t, "123", expectedGroupID)
	})

	t.Run("should not be able to find specified cluster", func(t *testing.T) {
		_, ui := mock.NewUI()
		var expectedGroupID string
		ac := mock.AtlasClient{}
		ac.ClustersFn = func(groupID string) ([]atlas.Cluster, error) {
			expectedGroupID = groupID
			return nil, nil
		}

		inputs := createInputs{Clusters: []string{"test-cluster"}}

		_, err := inputs.resolveClusters(ui, ac, "123")
		assert.Equal(t, errors.New("failed to find Atlas cluster"), err)
		assert.Equal(t, "123", expectedGroupID)
	})

	t.Run("should warn user if clusters are not found and auto confirm is set", func(t *testing.T) {
		out := new(bytes.Buffer)
		ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

		ac := mock.AtlasClient{}
		ac.ClustersFn = func(groupID string) ([]atlas.Cluster, error) {
			return []atlas.Cluster{{ID: "789", Name: "test-cluster-1"}}, nil
		}
		dummyClusters := []string{"test-cluster-dummy-1", "test-cluster-dummy-2"}
		inputs := createInputs{newAppInputs: newAppInputs{Name: "test-app"}, Clusters: []string{"test-cluster-1", dummyClusters[0], dummyClusters[1]}}

		ds, err := inputs.resolveClusters(ui, ac, "123")
		assert.Nil(t, err)

		assert.Equal(t, []dataSourceCluster{
			{
				Name: "mongodb-atlas",
				Type: realm.ClusterType,
				Config: configCluster{
					ClusterName:         "test-cluster-1",
					ReadPreference:      "primary",
					WireProtocolEnabled: false,
				},
			},
		}, ds)
		assert.Equal(t, fmt.Sprintf("Please note, the data sources '%s' were not linked because Atlas clusters were not found\n", strings.Join(dummyClusters[:], ", ")), out.String())
	})

	t.Run("should prompt user for confirmation if clusters are not found", func(t *testing.T) {
		ac := mock.AtlasClient{}
		ac.ClustersFn = func(groupID string) ([]atlas.Cluster, error) {
			return []atlas.Cluster{{ID: "789", Name: "test-cluster-1"}}, nil
		}
		dummyClusters := []string{"test-cluster-dummy-1", "test-cluster-dummy-2"}
		inputs := createInputs{newAppInputs: newAppInputs{Name: "test-app"}, Clusters: []string{"test-cluster-1", dummyClusters[0], dummyClusters[1]}}

		for _, tc := range []struct {
			description      string
			response         string
			expectedErr      error
			expectedName     string
			expectedClusters []dataSourceCluster
		}{
			{
				description:      "and error if not confirmed",
				response:         "no",
				expectedErr:      errors.New("failed to find Atlas cluster"),
				expectedClusters: nil,
			},
			{
				description: "and continue to create app if confirmed",
				response:    "yes",
				expectedErr: nil,
				expectedClusters: []dataSourceCluster{
					{
						Name: "mongodb-atlas",
						Type: realm.ClusterType,
						Config: configCluster{
							ClusterName:         "test-cluster-1",
							ReadPreference:      "primary",
							WireProtocolEnabled: false,
						},
					},
				},
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				_, console, _, ui, err := mock.NewVT10XConsole()
				assert.Nil(t, err)
				defer console.Close()

				doneCh := make(chan (struct{}))
				go func() {
					defer close(doneCh)
					console.ExpectString(fmt.Sprintf("Please note, the data sources '%s' were not linked because Atlas clusters were not found", strings.Join(dummyClusters[:], ", ")))
					console.ExpectString("Would you still like to create the app?")
					console.SendLine(tc.response)
					console.ExpectEOF()
				}()

				ds, err := inputs.resolveClusters(ui, ac, "123")
				assert.Equal(t, tc.expectedErr, err)
				assert.Equal(t, tc.expectedClusters, ds)
			})
		}
	})

	t.Run("should error from client", func(t *testing.T) {
		_, ui := mock.NewUI()
		var expectedGroupID string
		ac := mock.AtlasClient{}
		ac.ClustersFn = func(groupID string) ([]atlas.Cluster, error) {
			expectedGroupID = groupID
			return nil, errors.New("client error")
		}

		inputs := createInputs{Clusters: []string{"test-cluster"}}

		_, err := inputs.resolveClusters(ui, ac, "123")
		assert.Equal(t, errors.New("client error"), err)
		assert.Equal(t, "123", expectedGroupID)
	})
}

func TestAppCreateInputsResolveDataLake(t *testing.T) {
	t.Run("should return data source config of a provided data lake", func(t *testing.T) {
		_, ui := mock.NewUI()
		var expectedGroupID string
		ac := mock.AtlasClient{}
		ac.DataLakesFn = func(groupID string) ([]atlas.DataLake, error) {
			expectedGroupID = groupID
			return []atlas.DataLake{{Name: "test-datalake"}}, nil
		}

		inputs := createInputs{newAppInputs: newAppInputs{Name: "test-app"}, DataLakes: []string{"test-datalake"}}

		ds, err := inputs.resolveDataLakes(ui, ac, "123")
		assert.Nil(t, err)

		assert.Equal(t, []dataSourceDataLake{
			{
				Name: "mongodb-datalake",
				Type: realm.DataLakeType,
				Config: configDataLake{
					DataLakeName: "test-datalake",
				},
			},
		}, ds)
		assert.Equal(t, "123", expectedGroupID)
	})

	t.Run("should return data source config of multiple provided data lakes", func(t *testing.T) {
		_, ui := mock.NewUI()
		var expectedGroupID string
		ac := mock.AtlasClient{}
		ac.DataLakesFn = func(groupID string) ([]atlas.DataLake, error) {
			expectedGroupID = groupID
			return []atlas.DataLake{
				{Name: "test-datalake-1"},
				{Name: "test-datalake-2"},
			}, nil
		}

		inputs := createInputs{newAppInputs: newAppInputs{Name: "test-app"}, DataLakes: []string{"test-datalake-1", "test-datalake-2"}}

		ds, err := inputs.resolveDataLakes(ui, ac, "123")
		assert.Nil(t, err)

		assert.Equal(t, []dataSourceDataLake{
			{
				Name: "mongodb-datalake",
				Type: realm.DataLakeType,
				Config: configDataLake{
					DataLakeName: "test-datalake-1",
				},
			},
			{
				Name: "mongodb-datalake",
				Type: realm.DataLakeType,
				Config: configDataLake{
					DataLakeName: "test-datalake-2",
				},
			},
		}, ds)
		assert.Equal(t, "123", expectedGroupID)
	})

	t.Run("should warn user if datalakes are not found and auto confirm is set", func(t *testing.T) {
		out := new(bytes.Buffer)
		ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

		ac := mock.AtlasClient{}
		ac.DataLakesFn = func(groupID string) ([]atlas.DataLake, error) {
			return []atlas.DataLake{
				{Name: "test-datalake-1"},
			}, nil
		}

		dummyDataLakes := []string{"test-dummy-lake-1", "test-dummy-lake-2"}
		inputs := createInputs{newAppInputs: newAppInputs{Name: "test-app"}, DataLakes: []string{"test-datalake-1", dummyDataLakes[0], dummyDataLakes[1]}}

		ds, err := inputs.resolveDataLakes(ui, ac, "123")
		assert.Nil(t, err)

		assert.Equal(t, []dataSourceDataLake{
			{
				Name: "mongodb-datalake",
				Type: realm.DataLakeType,
				Config: configDataLake{
					DataLakeName: "test-datalake-1",
				},
			},
		}, ds)

		assert.Equal(t, fmt.Sprintf("Please note, the data sources '%s' were not linked because Atlas data lakes were not found\n", strings.Join(dummyDataLakes[:], ", ")), out.String())
	})

	t.Run("should prompt user for confirmation if datalakes are not found", func(t *testing.T) {
		ac := mock.AtlasClient{}
		ac.DataLakesFn = func(groupID string) ([]atlas.DataLake, error) {
			return []atlas.DataLake{
				{Name: "test-datalake-1"},
			}, nil
		}

		dummyDataLakes := []string{"test-dummy-lake-1", "test-dummy-lake-2"}
		inputs := createInputs{newAppInputs: newAppInputs{Name: "test-app"}, DataLakes: []string{"test-datalake-1", dummyDataLakes[0], dummyDataLakes[1]}}

		for _, tc := range []struct {
			description       string
			response          string
			expectedErr       error
			expectedName      string
			expectedDataLakes []dataSourceDataLake
		}{
			{
				description:       "and error if not confirmed",
				response:          "no",
				expectedErr:       errors.New("failed to find Atlas data lake"),
				expectedDataLakes: nil,
			},
			{
				description: "and continue to create app if confirmed",
				response:    "yes",
				expectedErr: nil,
				expectedDataLakes: []dataSourceDataLake{
					{
						Name: "mongodb-datalake",
						Type: realm.DataLakeType,
						Config: configDataLake{
							DataLakeName: "test-datalake-1",
						},
					},
				},
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				_, console, _, ui, err := mock.NewVT10XConsole()
				assert.Nil(t, err)
				defer console.Close()

				doneCh := make(chan (struct{}))
				go func() {
					defer close(doneCh)
					console.ExpectString(fmt.Sprintf("Please note, the data sources '%s' were not linked because Atlas data lakes were not found", strings.Join(dummyDataLakes[:], ", ")))
					console.ExpectString("Would you still like to create the app?")
					console.SendLine(tc.response)
					console.ExpectEOF()
				}()

				ds, err := inputs.resolveDataLakes(ui, ac, "123")
				assert.Equal(t, tc.expectedErr, err)
				assert.Equal(t, tc.expectedDataLakes, ds)
			})
		}
	})

	t.Run("should not be able to find specified data lake", func(t *testing.T) {
		_, ui := mock.NewUI()
		var expectedGroupID string
		ac := mock.AtlasClient{}
		ac.DataLakesFn = func(groupID string) ([]atlas.DataLake, error) {
			expectedGroupID = groupID
			return nil, nil
		}

		inputs := createInputs{DataLakes: []string{"test-datalake"}}

		_, err := inputs.resolveDataLakes(ui, ac, "123")
		assert.Equal(t, errors.New("failed to find Atlas data lake"), err)
		assert.Equal(t, "123", expectedGroupID)
	})

	t.Run("should error from client", func(t *testing.T) {
		_, ui := mock.NewUI()
		var expectedGroupID string
		ac := mock.AtlasClient{}
		ac.DataLakesFn = func(groupID string) ([]atlas.DataLake, error) {
			expectedGroupID = groupID
			return nil, errors.New("client error")
		}

		inputs := createInputs{DataLakes: []string{"test-datalake"}}

		_, err := inputs.resolveDataLakes(ui, ac, "123")
		assert.Equal(t, errors.New("client error"), err)
		assert.Equal(t, "123", expectedGroupID)
	})
}

func TestFindDefaultPath(t *testing.T) {
	t.Run("should return new incremented directory if provided directory already exists", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()

		testAppName := "test-app"
		for i := 1; i < 10; i++ {
			defaultPath := findDefaultPath(profile.WorkingDirectory, testAppName)
			localPath := testAppName + "-" + strconv.Itoa(i)

			assert.Equal(t, localPath, defaultPath)

			assert.Nil(t, ioutil.WriteFile(
				filepath.Join(profile.WorkingDirectory, defaultPath),
				[]byte(`{"config_version":20210101,"app_id":"test-app-abcde","name":"test-app"}`),
				0666,
			))
		}

		//if file options 1-9 are exhausted, use hex
		defaultPath := findDefaultPath(profile.WorkingDirectory, testAppName)
		directoryID := strings.Trim(defaultPath, testAppName+"-")
		assert.True(t, primitive.IsValidObjectID(directoryID), "should be primitive object id")
	})
}
