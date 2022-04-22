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

		console.Tty().Close() // flush the writers
		<-doneCh              // wait for procedure to complete

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
		rc.AllTemplatesFn = func() ([]realm.Template, error) {
			return []realm.Template{}, nil
		}

		inputs := createInputs{newAppInputs: newAppInputs{
			Name:          existingApp.Name,
			Project:       existingApp.GroupID,
			ConfigVersion: realm.DefaultAppConfigVersion,
		}}
		cmd := &CommandCreate{inputs}
		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: rc}))

		existingDir := filepath.Join(profile.WorkingDirectory, existingApp.Name)
		dir, err := inputs.resolveLocalPath(ui, existingDir)

		assert.Equal(t, errProjectExists(existingDir), err)
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
		rc.AllTemplatesFn = func() ([]realm.Template, error) {
			return []realm.Template{}, nil
		}

		inputs := createInputs{newAppInputs: newAppInputs{
			Name:          existingApp.Name,
			Project:       existingApp.GroupID,
			ConfigVersion: realm.DefaultAppConfigVersion,
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

		assert.Equal(t, errProjectExists(existingApp.Name), err)
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

		console.Tty().Close() // flush the writers
		<-doneCh              // wait for procedure to complete

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

		inputs := createInputs{
			newAppInputs:        newAppInputs{Name: "test-app"},
			Clusters:            []string{"test-cluster"},
			ClusterServiceNames: []string{"mongodb-atlas"},
		}

		ds, _, err := inputs.resolveClusters(ui, ac, "123")
		assert.Nil(t, err)

		assert.Equal(t, []dataSourceCluster{
			{
				Name: "mongodb-atlas",
				Type: realm.ServiceTypeCluster,
				Config: configCluster{
					ClusterName:         "test-cluster",
					ReadPreference:      "primary",
					WireProtocolEnabled: false,
				},
				Version: 1,
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

		inputs := createInputs{
			newAppInputs:        newAppInputs{Name: "test-app"},
			Clusters:            []string{"test-cluster-1", "test-cluster-2"},
			ClusterServiceNames: []string{"mongodb-atlas", "another-data-source"},
		}

		ds, _, err := inputs.resolveClusters(ui, ac, "123")
		assert.Nil(t, err)

		assert.Equal(t, []dataSourceCluster{
			{
				Name: "mongodb-atlas",
				Type: realm.ServiceTypeCluster,
				Config: configCluster{
					ClusterName:         "test-cluster-1",
					ReadPreference:      "primary",
					WireProtocolEnabled: false,
				},
				Version: 1,
			},
			{
				Name: "another-data-source",
				Type: realm.ServiceTypeCluster,
				Config: configCluster{
					ClusterName:         "test-cluster-2",
					ReadPreference:      "primary",
					WireProtocolEnabled: false,
				},
				Version: 1,
			},
		}, ds)
		assert.Equal(t, "123", expectedGroupID)
	})

	t.Run("should warn user if clusters are not found and auto confirm is set", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()

		console, ui, err := mock.NewConsoleWithOptions(mock.UIOptions{AutoConfirm: true})
		assert.Nil(t, err)
		defer console.Close()

		testApp := realm.App{
			ID:          primitive.NewObjectID().Hex(),
			GroupID:     primitive.NewObjectID().Hex(),
			ClientAppID: "test-app-abcde",
			Name:        "test-app",
		}

		rc := mock.RealmClient{}
		rc.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return testApp, nil
		}
		rc.ImportFn = func(groupID, appID string, appData interface{}) error {
			return nil
		}
		rc.AllTemplatesFn = func() ([]realm.Template, error) {
			return []realm.Template{}, nil
		}

		ac := mock.AtlasClient{}
		ac.ClustersFn = func(groupID string) ([]atlas.Cluster, error) {
			return []atlas.Cluster{{ID: "789", Name: "test-cluster-1"}}, nil
		}

		dummyClusters := []string{"test-cluster-dummy-1", "test-cluster-dummy-2"}
		inputs := createInputs{
			newAppInputs:        newAppInputs{Name: testApp.Name, Project: testApp.GroupID},
			Clusters:            []string{"test-cluster-1", dummyClusters[0], dummyClusters[1]},
			ClusterServiceNames: []string{"mongodb-atlas"},
		}

		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)
			console.ExpectString("Note: The following data sources were not linked because they could not be found: 'test-cluster-dummy-1', 'test-cluster-dummy-2'")
			console.ExpectEOF()
		}()

		cmd := &CommandCreate{inputs}
		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: rc, Atlas: ac}))

		console.Tty().Close() // flush the writers
		<-doneCh              // wait for procedure to complete

		ds, _, err := inputs.resolveClusters(ui, ac, "123")
		assert.Nil(t, err)

		assert.Equal(t, []dataSourceCluster{
			{
				Name: "mongodb-atlas",
				Type: realm.ServiceTypeCluster,
				Config: configCluster{
					ClusterName:         "test-cluster-1",
					ReadPreference:      "primary",
					WireProtocolEnabled: false,
				},
				Version: 1,
			},
		}, ds)
	})

	t.Run("should prompt user for confirmation if clusters are not found", func(t *testing.T) {
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
			return []atlas.Cluster{{ID: "789", Name: "test-cluster-1"}}, nil
		}
		dummyClusters := []string{"test-cluster-dummy-1", "test-cluster-dummy-2"}
		inputs := createInputs{
			newAppInputs:        newAppInputs{Name: testApp.Name, Project: testApp.GroupID},
			Clusters:            []string{"test-cluster-1", dummyClusters[0], dummyClusters[1]},
			ClusterServiceNames: []string{"mongodb-atlas"},
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
					console.ExpectString("Note: The following data sources were not linked because they could not be found: 'test-cluster-dummy-1', 'test-cluster-dummy-2'")
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

				console.Tty().Close() // flush the writers
				<-doneCh              // wait for procedure to complete

				assert.Equal(t, tc.appCreated, appCreated)
			})
		}
	})

	t.Run("should return data source configs of clusters with cluster service names", func(t *testing.T) {
		clusterNames := []string{"Cluster0", "Cluster1"}
		clusterServiceNames := []string{"mongodb-atlas", "another-data-source"}

		ac := mock.AtlasClient{}
		ac.ClustersFn = func(groupID string) ([]atlas.Cluster, error) {
			return []atlas.Cluster{
				{ID: "456", Name: clusterNames[0]},
				{ID: "789", Name: clusterNames[1]},
			}, nil
		}

		for _, tc := range []struct {
			description                 string
			clusterNames                []string
			clusterServiceNames         []string
			procedure                   func(c *expect.Console)
			autoConfirm                 bool
			expectedClusterServiceNames []string
		}{
			{
				description:                 "use cluster service names provided",
				clusterNames:                clusterNames,
				clusterServiceNames:         clusterServiceNames,
				procedure:                   func(c *expect.Console) {},
				autoConfirm:                 false,
				expectedClusterServiceNames: clusterServiceNames,
			},
			{
				description:  "prompt user if no cluster service names are provided",
				clusterNames: clusterNames,
				procedure: func(c *expect.Console) {
					c.ExpectString("Enter a Service Name for Cluster 'Cluster0'")
					c.SendLine(clusterServiceNames[0])
					c.ExpectString("Enter a Service Name for Cluster 'Cluster1'")
					c.SendLine(clusterServiceNames[1])
					c.ExpectEOF()
				},
				autoConfirm:                 false,
				expectedClusterServiceNames: clusterServiceNames,
			},
			{
				description:         "prompt user if any cluster service name is not provided",
				clusterNames:        clusterNames,
				clusterServiceNames: []string{clusterServiceNames[0]},
				procedure: func(c *expect.Console) {
					c.ExpectString("Enter a Service Name for Cluster 'Cluster1'")
					c.SendLine(clusterServiceNames[1])
					c.ExpectEOF()
				},
				autoConfirm:                 false,
				expectedClusterServiceNames: clusterServiceNames,
			},
			{
				description:                 "default cluster service names to cluster names if not provided and auto confirm is set",
				clusterNames:                clusterNames,
				procedure:                   func(c *expect.Console) {},
				autoConfirm:                 true,
				expectedClusterServiceNames: clusterNames,
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				console, _, ui, err := mock.NewVT10XConsoleWithOptions(mock.UIOptions{AutoConfirm: tc.autoConfirm})
				assert.Nil(t, err)
				defer console.Close()

				doneCh := make(chan (struct{}))
				go func() {
					defer close(doneCh)
					tc.procedure(console)
				}()

				inputs := createInputs{newAppInputs: newAppInputs{Name: "test-app"},
					Clusters:            tc.clusterNames,
					ClusterServiceNames: tc.clusterServiceNames,
				}

				ds, _, err := inputs.resolveClusters(ui, ac, "123")
				assert.Nil(t, err)

				console.Tty().Close() // flush the writers
				<-doneCh              // wait for procedure to complete

				assert.Equal(t, []dataSourceCluster{
					{
						Name: tc.expectedClusterServiceNames[0],
						Type: realm.ServiceTypeCluster,
						Config: configCluster{
							ClusterName:         clusterNames[0],
							ReadPreference:      "primary",
							WireProtocolEnabled: false,
						},
						Version: 1,
					},
					{
						Name: tc.expectedClusterServiceNames[1],
						Type: realm.ServiceTypeCluster,
						Config: configCluster{
							ClusterName:         clusterNames[1],
							ReadPreference:      "primary",
							WireProtocolEnabled: false,
						},
						Version: 1,
					},
				}, ds)
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

		_, _, err := inputs.resolveClusters(ui, ac, "123")
		assert.Equal(t, errors.New("client error"), err)
		assert.Equal(t, "123", expectedGroupID)
	})

	t.Run("when creating template apps", func(t *testing.T) {
		t.Run("should error when more than one cluster passed in", func(t *testing.T) {
			_, ui := mock.NewUI()
			clusterNames := []string{"Cluster0", "Cluster1"}

			ac := mock.AtlasClient{}
			ac.ClustersFn = func(groupID string) ([]atlas.Cluster, error) {
				return []atlas.Cluster{
					{ID: "456", Name: clusterNames[0]},
				}, nil
			}
			inputs := createInputs{
				newAppInputs: newAppInputs{
					Template: "ios.template.todo",
				},
				Clusters: clusterNames,
			}
			_, _, err := inputs.resolveClusters(ui, ac, "123")
			assert.Equal(t, errors.New("template apps can only be created with one cluster"), err)
		})

		t.Run("should error if no atlas clusters exist", func(t *testing.T) {
			_, ui := mock.NewUI()
			clusterNames := []string{"Cluster0"}

			ac := mock.AtlasClient{}
			ac.ClustersFn = func(groupID string) ([]atlas.Cluster, error) {
				return []atlas.Cluster{}, nil
			}
			inputs := createInputs{
				newAppInputs: newAppInputs{
					Template: "ios.template.todo",
				},
				Clusters: clusterNames,
			}
			_, _, err := inputs.resolveClusters(ui, ac, "123")
			assert.Equal(t, errors.New("please create an Atlas cluster before creating a template app"), err)
		})

		t.Run("should error when provided cluster name does not exist", func(t *testing.T) {
			_, ui := mock.NewUI()
			clusterNames := []string{"nonExistentCluster"}

			ac := mock.AtlasClient{}
			ac.ClustersFn = func(groupID string) ([]atlas.Cluster, error) {
				return []atlas.Cluster{
					{ID: "456", Name: "Cluster0"},
				}, nil
			}
			inputs := createInputs{
				newAppInputs: newAppInputs{
					Template: "ios.template.todo",
				},
				Clusters: clusterNames,
			}

			_, _, err := inputs.resolveClusters(ui, ac, "123")
			assert.Equal(t, errors.New("could not find Atlas cluster 'nonExistentCluster'"), err)
		})

		t.Run("should resolve a single data source when cluster name is passed in", func(t *testing.T) {
			_, ui := mock.NewUI()
			clusterNames := []string{"Cluster0"}

			ac := mock.AtlasClient{}
			ac.ClustersFn = func(groupID string) ([]atlas.Cluster, error) {
				return []atlas.Cluster{
					{ID: "1", Name: clusterNames[0]},
					{ID: "2", Name: "Cluster1"},
				}, nil
			}
			inputs := createInputs{
				newAppInputs: newAppInputs{
					Template: "ios.template.todo",
				},
				Clusters: clusterNames,
			}

			clusters, _, err := inputs.resolveClusters(ui, ac, "123")
			assert.Equal(t, nil, err)
			assert.Equal(t, 1, len(clusters))
			assert.Equal(t, "Cluster0", clusters[0].Config.ClusterName)
			assert.Equal(t, "mongodb-atlas", clusters[0].Name)
		})

		t.Run("should automatically select cluster if there's only one available to choose", func(t *testing.T) {
			_, ui := mock.NewUI()
			ac := mock.AtlasClient{}
			ac.ClustersFn = func(groupID string) ([]atlas.Cluster, error) {
				return []atlas.Cluster{
					{ID: "1", Name: "Cluster1"},
				}, nil
			}
			inputs := createInputs{
				newAppInputs: newAppInputs{
					Template: "ios.template.todo",
				},
			}

			clusters, _, err := inputs.resolveClusters(ui, ac, "123")
			assert.Equal(t, nil, err)
			assert.Equal(t, 1, len(clusters))
			assert.Equal(t, "Cluster1", clusters[0].Config.ClusterName)
			assert.Equal(t, "mongodb-atlas", clusters[0].Name)
		})

		t.Run("should prompt for cluster name if multiple cluster options exist", func(t *testing.T) {
			_, console, _, ui, consoleErr := mock.NewVT10XConsole()
			assert.Nil(t, consoleErr)
			defer console.Close()

			doneCh := make(chan struct{})
			go func() {
				defer close(doneCh)
				console.ExpectString("Select a cluster to link to your Realm application:")
				console.SendLine("Cluster1")
				console.ExpectEOF()
			}()

			ac := mock.AtlasClient{}
			ac.ClustersFn = func(groupID string) ([]atlas.Cluster, error) {
				return []atlas.Cluster{
					{ID: "1", Name: "Cluster0"},
					{ID: "2", Name: "Cluster1"},
				}, nil
			}
			inputs := createInputs{
				newAppInputs: newAppInputs{
					Template: "ios.template.todo",
				},
			}

			clusters, _, err := inputs.resolveClusters(ui, ac, "123")
			assert.Equal(t, nil, err)
			assert.Equal(t, 1, len(clusters))
			assert.Equal(t, "Cluster1", clusters[0].Config.ClusterName)
			assert.Equal(t, "mongodb-atlas", clusters[0].Name)
		})
		t.Run("should force mongodb-atlas as template data source name", func(t *testing.T) {
			_, ui := mock.NewUI()
			clusterNames := []string{"Cluster0"}

			ac := mock.AtlasClient{}
			ac.ClustersFn = func(groupID string) ([]atlas.Cluster, error) {
				return []atlas.Cluster{
					{ID: "456", Name: clusterNames[0]},
				}, nil
			}
			inputs := createInputs{
				newAppInputs: newAppInputs{
					Template: "ios.template.todo",
				},
				Clusters:            clusterNames,
				ClusterServiceNames: []string{"overridden_name"},
			}
			clusters, _, err := inputs.resolveClusters(ui, ac, "123")
			assert.Equal(t, nil, err)
			assert.Equal(t, 1, len(clusters))
			assert.Equal(t, "mongodb-atlas", clusters[0].Name)
		})
	})
}

func TestAppCreateInputsResolveServerlessInstance(t *testing.T) {
	t.Run("should return data source config of a provided serverless instance", func(t *testing.T) {
		_, ui := mock.NewUI()
		var expectedGroupID string
		ac := mock.AtlasClient{}
		ac.ServerlessInstancesFn = func(groupID string) ([]atlas.ServerlessInstance, error) {
			expectedGroupID = groupID
			return []atlas.ServerlessInstance{{Name: "test-serverless-instance"}}, nil
		}

		inputs := createInputs{
			newAppInputs:                   newAppInputs{Name: "test-app"},
			ServerlessInstances:            []string{"test-serverless-instance"},
			ServerlessInstanceServiceNames: []string{"mongodb-atlas"},
		}

		ds, _, err := inputs.resolveServerlessInstances(ui, ac, "123")
		assert.Nil(t, err)

		assert.Equal(t, []dataSourceCluster{
			{
				Name: "mongodb-atlas",
				Type: realm.ServiceTypeCluster,
				Config: configCluster{
					ClusterName:         "test-serverless-instance",
					ReadPreference:      "primary",
					WireProtocolEnabled: false,
				},
				Version: 1,
			},
		}, ds)
		assert.Equal(t, "123", expectedGroupID)
	})

	t.Run("should return data source config of multiple provided serverless instances", func(t *testing.T) {
		_, ui := mock.NewUI()
		var expectedGroupID string
		ac := mock.AtlasClient{}
		ac.ServerlessInstancesFn = func(groupID string) ([]atlas.ServerlessInstance, error) {
			expectedGroupID = groupID
			return []atlas.ServerlessInstance{
				{Name: "test-serverless-instance-1"},
				{Name: "test-serverless-instance-2"},
			}, nil
		}

		inputs := createInputs{
			newAppInputs:                   newAppInputs{Name: "test-app"},
			ServerlessInstances:            []string{"test-serverless-instance-1", "test-serverless-instance-2"},
			ServerlessInstanceServiceNames: []string{"mongodb-atlas", "another-data-source"},
		}

		ds, _, err := inputs.resolveServerlessInstances(ui, ac, "123")
		assert.Nil(t, err)

		assert.Equal(t, []dataSourceCluster{
			{
				Name: "mongodb-atlas",
				Type: realm.ServiceTypeCluster,
				Config: configCluster{
					ClusterName:         "test-serverless-instance-1",
					ReadPreference:      "primary",
					WireProtocolEnabled: false,
				},
				Version: 1,
			},
			{
				Name: "another-data-source",
				Type: realm.ServiceTypeCluster,
				Config: configCluster{
					ClusterName:         "test-serverless-instance-2",
					ReadPreference:      "primary",
					WireProtocolEnabled: false,
				},
				Version: 1,
			},
		}, ds)
		assert.Equal(t, "123", expectedGroupID)
	})

	t.Run("if serverless instances are not found and auto confirm is set", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()

		console, ui, err := mock.NewConsoleWithOptions(mock.UIOptions{AutoConfirm: true})
		assert.Nil(t, err)
		defer console.Close()

		testApp := realm.App{
			ID:          primitive.NewObjectID().Hex(),
			GroupID:     primitive.NewObjectID().Hex(),
			ClientAppID: "test-app-abcde",
			Name:        "test-app",
		}

		rc := mock.RealmClient{}
		rc.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return testApp, nil
		}
		rc.ImportFn = func(groupID, appID string, appData interface{}) error {
			return nil
		}

		ac := mock.AtlasClient{}
		ac.ServerlessInstancesFn = func(groupID string) ([]atlas.ServerlessInstance, error) {
			return []atlas.ServerlessInstance{
				{Name: "test-serverless-instance-1"},
			}, nil
		}

		dummyServerlessInstances := []string{"test-dummy-serverless-instance-1", "test-dummy-serverless-instance-2"}
		inputs := createInputs{
			newAppInputs:                   newAppInputs{Name: testApp.Name, Project: testApp.GroupID},
			ServerlessInstances:            []string{"test-serverless-instance-1", dummyServerlessInstances[0], dummyServerlessInstances[1]},
			ServerlessInstanceServiceNames: []string{"mongodb-atlas"},
		}

		t.Run("should warn user that data sources were not linked because they could not be found", func(t *testing.T) {
			doneCh := make(chan (struct{}))
			go func() {
				defer close(doneCh)
				console.ExpectString("Note: The following data sources were not linked because they could not be found: 'test-dummy-serverless-instance-1', 'test-dummy-serverless-instance-2'")
				console.ExpectEOF()
			}()

			cmd := &CommandCreate{inputs}
			assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: rc, Atlas: ac}))

			console.Tty().Close() // flush the writers
			<-doneCh              // wait for procedure to complete
		})

		t.Run("should resolve to found serverless instances", func(t *testing.T) {
			ds, _, err := inputs.resolveServerlessInstances(ui, ac, "123")
			assert.Nil(t, err)

			assert.Equal(t, []dataSourceCluster{
				{
					Name: "mongodb-atlas",
					Type: realm.ServiceTypeCluster,
					Config: configCluster{
						ClusterName:         "test-serverless-instance-1",
						ReadPreference:      "primary",
						WireProtocolEnabled: false,
					},
					Version: 1,
				},
			}, ds)
		})
	})

	t.Run("should prompt user for confirmation if serverless instances are not found", func(t *testing.T) {
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
		ac.ServerlessInstancesFn = func(groupID string) ([]atlas.ServerlessInstance, error) {
			return []atlas.ServerlessInstance{
				{Name: "test-serverless-instance-1"},
			}, nil
		}

		dummyServerlessInstances := []string{"test-dummy-serverless-instance-1", "test-dummy-serverless-instance-2"}
		inputs := createInputs{
			newAppInputs:                   newAppInputs{Name: testApp.Name, Project: testApp.GroupID},
			ServerlessInstances:            []string{"test-serverless-instance-1", dummyServerlessInstances[0], dummyServerlessInstances[1]},
			ServerlessInstanceServiceNames: []string{"mongodb-atlas"},
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
					console.ExpectString("Note: The following data sources were not linked because they could not be found: 'test-dummy-serverless-instance-1', 'test-dummy-serverless-instance-2'")
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

				console.Tty().Close() // flush the writers
				<-doneCh              // wait for procedure to complete

				assert.Equal(t, tc.appCreated, appCreated)
			})
		}
	})

	t.Run("should return data source configs of serverless instances with serverless instance service names", func(t *testing.T) {
		serverlessInstanceNames := []string{"ServerlessInstance0", "ServerlessInstance1"}
		serverlessInstanceServiceNames := []string{"mongodb-atlas", "another-data-source"}

		ac := mock.AtlasClient{}
		ac.ServerlessInstancesFn = func(groupID string) ([]atlas.ServerlessInstance, error) {
			return []atlas.ServerlessInstance{
				{Name: serverlessInstanceNames[0]},
				{Name: serverlessInstanceNames[1]},
			}, nil
		}

		for _, tc := range []struct {
			description                            string
			serverlessInstanceNames                []string
			serverlessInstanceServiceNames         []string
			procedure                              func(c *expect.Console)
			autoConfirm                            bool
			expectedServerlessInstanceServiceNames []string
		}{
			{
				description:                            "use serverless instance names provided",
				serverlessInstanceNames:                serverlessInstanceNames,
				serverlessInstanceServiceNames:         serverlessInstanceServiceNames,
				procedure:                              func(c *expect.Console) {},
				autoConfirm:                            false,
				expectedServerlessInstanceServiceNames: serverlessInstanceServiceNames,
			},
			{
				description:             "prompt user if no serverless instance service names are provided",
				serverlessInstanceNames: serverlessInstanceNames,
				procedure: func(c *expect.Console) {
					c.ExpectString("Enter a Service Name for Serverless instance 'ServerlessInstance0'")
					c.SendLine(serverlessInstanceServiceNames[0])
					c.ExpectString("Enter a Service Name for Serverless instance 'ServerlessInstance1'")
					c.SendLine(serverlessInstanceServiceNames[1])
					c.ExpectEOF()
				},
				autoConfirm:                            false,
				expectedServerlessInstanceServiceNames: serverlessInstanceServiceNames,
			},
			{
				description:                    "prompt user if any serverless instance service name is not provided",
				serverlessInstanceNames:        serverlessInstanceNames,
				serverlessInstanceServiceNames: []string{serverlessInstanceServiceNames[0]},
				procedure: func(c *expect.Console) {
					c.ExpectString("Enter a Service Name for Serverless instance 'ServerlessInstance1'")
					c.SendLine(serverlessInstanceServiceNames[1])
					c.ExpectEOF()
				},
				autoConfirm:                            false,
				expectedServerlessInstanceServiceNames: serverlessInstanceServiceNames,
			},
			{
				description:                            "default serverless instance service names to serverless instance names if not provided and auto confirm is set",
				serverlessInstanceNames:                serverlessInstanceNames,
				procedure:                              func(c *expect.Console) {},
				autoConfirm:                            true,
				expectedServerlessInstanceServiceNames: serverlessInstanceNames,
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				console, _, ui, err := mock.NewVT10XConsoleWithOptions(mock.UIOptions{AutoConfirm: tc.autoConfirm})
				assert.Nil(t, err)
				defer console.Close()

				doneCh := make(chan (struct{}))
				go func() {
					defer close(doneCh)
					tc.procedure(console)
				}()

				inputs := createInputs{
					newAppInputs:                   newAppInputs{Name: "test-app"},
					ServerlessInstances:            tc.serverlessInstanceNames,
					ServerlessInstanceServiceNames: tc.serverlessInstanceServiceNames,
				}

				ds, _, err := inputs.resolveServerlessInstances(ui, ac, "123")
				assert.Nil(t, err)

				console.Tty().Close() // flush the writers
				<-doneCh              // wait for procedure to complete

				assert.Equal(t, []dataSourceCluster{
					{
						Name: tc.expectedServerlessInstanceServiceNames[0],
						Type: realm.ServiceTypeCluster,
						Config: configCluster{
							ClusterName:         tc.serverlessInstanceNames[0],
							ReadPreference:      "primary",
							WireProtocolEnabled: false,
						},
						Version: 1,
					},
					{
						Name: tc.expectedServerlessInstanceServiceNames[1],
						Type: realm.ServiceTypeCluster,
						Config: configCluster{
							ClusterName:         tc.serverlessInstanceNames[1],
							ReadPreference:      "primary",
							WireProtocolEnabled: false,
						},
						Version: 1,
					},
				}, ds)
			})
		}
	})

	t.Run("should error from a client error", func(t *testing.T) {
		_, ui := mock.NewUI()
		var expectedGroupID string
		ac := mock.AtlasClient{}
		ac.ServerlessInstancesFn = func(groupID string) ([]atlas.ServerlessInstance, error) {
			expectedGroupID = groupID
			return nil, errors.New("client error")
		}

		inputs := createInputs{ServerlessInstances: []string{"test-serverless-instance"}}

		_, _, err := inputs.resolveServerlessInstances(ui, ac, "123")
		assert.Equal(t, errors.New("client error"), err)
		assert.Equal(t, "123", expectedGroupID)
	})

	t.Run("should error if creating template with a serverless instance", func(t *testing.T) {
		_, ui := mock.NewUI()
		ac := mock.AtlasClient{}
		inputs := createInputs{
			newAppInputs: newAppInputs{
				Template: "ios.template.todo",
			},
			ServerlessInstances: []string{"test-serverless-instance"},
		}

		_, _, err := inputs.resolveServerlessInstances(ui, ac, "123")
		assert.Equal(t, errors.New("cannot create a template app with Serverless instances"), err)
	})
}

func TestAppCreateInputsResolveDatalake(t *testing.T) {
	t.Run("should return data source config of a provided data lake", func(t *testing.T) {
		_, ui := mock.NewUI()
		var expectedGroupID string
		ac := mock.AtlasClient{}
		ac.DatalakesFn = func(groupID string) ([]atlas.Datalake, error) {
			expectedGroupID = groupID
			return []atlas.Datalake{{Name: "test-datalake"}}, nil
		}

		inputs := createInputs{
			newAppInputs:         newAppInputs{Name: "test-app"},
			Datalakes:            []string{"test-datalake"},
			DatalakeServiceNames: []string{"mongodb-datalake"},
		}

		ds, _, err := inputs.resolveDatalakes(ui, ac, "123")
		assert.Nil(t, err)

		assert.Equal(t, []dataSourceDatalake{
			{
				Name: "mongodb-datalake",
				Type: realm.ServiceTypeDatalake,
				Config: configDatalake{
					DatalakeName: "test-datalake",
				},
			},
		}, ds)
		assert.Equal(t, "123", expectedGroupID)
	})

	t.Run("should return data source config of multiple provided data lakes", func(t *testing.T) {
		_, ui := mock.NewUI()
		var expectedGroupID string
		ac := mock.AtlasClient{}
		ac.DatalakesFn = func(groupID string) ([]atlas.Datalake, error) {
			expectedGroupID = groupID
			return []atlas.Datalake{
				{Name: "test-datalake-1"},
				{Name: "test-datalake-2"},
			}, nil
		}

		inputs := createInputs{
			newAppInputs:         newAppInputs{Name: "test-app"},
			Datalakes:            []string{"test-datalake-1", "test-datalake-2"},
			DatalakeServiceNames: []string{"mongodb-datalake", "another-data-source"},
		}

		ds, _, err := inputs.resolveDatalakes(ui, ac, "123")
		assert.Nil(t, err)

		assert.Equal(t, []dataSourceDatalake{
			{
				Name: "mongodb-datalake",
				Type: realm.ServiceTypeDatalake,
				Config: configDatalake{
					DatalakeName: "test-datalake-1",
				},
			},
			{
				Name: "another-data-source",
				Type: realm.ServiceTypeDatalake,
				Config: configDatalake{
					DatalakeName: "test-datalake-2",
				},
			},
		}, ds)
		assert.Equal(t, "123", expectedGroupID)
	})

	t.Run("should warn user if data lakes are not found and auto confirm is set", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()

		console, ui, err := mock.NewConsoleWithOptions(mock.UIOptions{AutoConfirm: true})
		assert.Nil(t, err)
		defer console.Close()

		testApp := realm.App{
			ID:          primitive.NewObjectID().Hex(),
			GroupID:     primitive.NewObjectID().Hex(),
			ClientAppID: "test-app-abcde",
			Name:        "test-app",
		}

		rc := mock.RealmClient{}
		rc.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return testApp, nil
		}
		rc.ImportFn = func(groupID, appID string, appData interface{}) error {
			return nil
		}
		rc.AllTemplatesFn = func() ([]realm.Template, error) {
			return []realm.Template{}, nil
		}

		ac := mock.AtlasClient{}
		ac.DatalakesFn = func(groupID string) ([]atlas.Datalake, error) {
			return []atlas.Datalake{
				{Name: "test-datalake-1"},
			}, nil
		}

		dummyDatalakes := []string{"test-dummy-lake-1", "test-dummy-lake-2"}
		inputs := createInputs{
			newAppInputs:         newAppInputs{Name: testApp.Name, Project: testApp.GroupID},
			Datalakes:            []string{"test-datalake-1", dummyDatalakes[0], dummyDatalakes[1]},
			DatalakeServiceNames: []string{"mongodb-datalake"},
		}

		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)
			console.ExpectString("Note: The following data sources were not linked because they could not be found: 'test-dummy-lake-1', 'test-dummy-lake-2'")
			console.ExpectEOF()
		}()

		cmd := &CommandCreate{inputs}
		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: rc, Atlas: ac}))

		console.Tty().Close() // flush the writers
		<-doneCh              // wait for procedure to complete

		ds, _, err := inputs.resolveDatalakes(ui, ac, "123")
		assert.Nil(t, err)

		assert.Equal(t, []dataSourceDatalake{
			{
				Name: "mongodb-datalake",
				Type: realm.ServiceTypeDatalake,
				Config: configDatalake{
					DatalakeName: "test-datalake-1",
				},
			},
		}, ds)
	})

	t.Run("should prompt user for confirmation if data lakes are not found", func(t *testing.T) {
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
		ac.DatalakesFn = func(groupID string) ([]atlas.Datalake, error) {
			return []atlas.Datalake{
				{Name: "test-datalake-1"},
			}, nil
		}

		dummyDatalakes := []string{"test-dummy-lake-1", "test-dummy-lake-2"}
		inputs := createInputs{
			newAppInputs:         newAppInputs{Name: testApp.Name, Project: testApp.GroupID},
			Datalakes:            []string{"test-datalake-1", dummyDatalakes[0], dummyDatalakes[1]},
			DatalakeServiceNames: []string{"mongodb-datalake"},
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
					console.ExpectString("Note: The following data sources were not linked because they could not be found: 'test-dummy-lake-1', 'test-dummy-lake-2'")
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

				console.Tty().Close() // flush the writers
				<-doneCh              // wait for procedure to complete

				assert.Equal(t, tc.appCreated, appCreated)
			})
		}
	})

	t.Run("should return data source configs of data lakes with data lake service names", func(t *testing.T) {
		datalakeNames := []string{"Datalake0", "Datalake1"}
		datalakeServiceNames := []string{"mongodb-datalake", "another-data-source"}

		ac := mock.AtlasClient{}
		ac.DatalakesFn = func(groupID string) ([]atlas.Datalake, error) {
			return []atlas.Datalake{
				{Name: datalakeNames[0]},
				{Name: datalakeNames[1]},
			}, nil
		}

		for _, tc := range []struct {
			description                  string
			datalakeNames                []string
			datalakeServiceNames         []string
			procedure                    func(c *expect.Console)
			autoConfirm                  bool
			expectedDatalakeServiceNames []string
		}{
			{
				description:                  "use data lake names provided",
				datalakeNames:                datalakeNames,
				datalakeServiceNames:         datalakeServiceNames,
				procedure:                    func(c *expect.Console) {},
				autoConfirm:                  false,
				expectedDatalakeServiceNames: datalakeServiceNames,
			},
			{
				description:   "prompt user if no data lake service names are provided",
				datalakeNames: datalakeNames,
				procedure: func(c *expect.Console) {
					c.ExpectString("Enter a Service Name for Data Lake 'Datalake0'")
					c.SendLine(datalakeServiceNames[0])
					c.ExpectString("Enter a Service Name for Data Lake 'Datalake1'")
					c.SendLine(datalakeServiceNames[1])
					c.ExpectEOF()
				},
				autoConfirm:                  false,
				expectedDatalakeServiceNames: datalakeServiceNames,
			},
			{
				description:          "prompt user if any data lake service name is not provided",
				datalakeNames:        datalakeNames,
				datalakeServiceNames: []string{datalakeServiceNames[0]},
				procedure: func(c *expect.Console) {
					c.ExpectString("Enter a Service Name for Data Lake 'Datalake1'")
					c.SendLine(datalakeServiceNames[1])
					c.ExpectEOF()
				},
				autoConfirm:                  false,
				expectedDatalakeServiceNames: datalakeServiceNames,
			},
			{
				description:                  "default data lake service names to data lake names if not provided and auto confirm is set",
				datalakeNames:                datalakeNames,
				procedure:                    func(c *expect.Console) {},
				autoConfirm:                  true,
				expectedDatalakeServiceNames: datalakeNames,
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				console, _, ui, err := mock.NewVT10XConsoleWithOptions(mock.UIOptions{AutoConfirm: tc.autoConfirm})
				assert.Nil(t, err)
				defer console.Close()

				doneCh := make(chan (struct{}))
				go func() {
					defer close(doneCh)
					tc.procedure(console)
				}()

				inputs := createInputs{
					newAppInputs:         newAppInputs{Name: "test-app"},
					Datalakes:            tc.datalakeNames,
					DatalakeServiceNames: tc.datalakeServiceNames,
				}

				ds, _, err := inputs.resolveDatalakes(ui, ac, "123")
				assert.Nil(t, err)

				console.Tty().Close() // flush the writers
				<-doneCh              // wait for procedure to complete

				assert.Equal(t, []dataSourceDatalake{
					{
						Name: tc.expectedDatalakeServiceNames[0],
						Type: realm.ServiceTypeDatalake,
						Config: configDatalake{
							DatalakeName: tc.datalakeNames[0],
						},
					},
					{
						Name: tc.expectedDatalakeServiceNames[1],
						Type: realm.ServiceTypeDatalake,
						Config: configDatalake{
							DatalakeName: tc.datalakeNames[1],
						},
					},
				}, ds)
			})
		}
	})

	t.Run("should error from client", func(t *testing.T) {
		_, ui := mock.NewUI()
		var expectedGroupID string
		ac := mock.AtlasClient{}
		ac.DatalakesFn = func(groupID string) ([]atlas.Datalake, error) {
			expectedGroupID = groupID
			return nil, errors.New("client error")
		}

		inputs := createInputs{Datalakes: []string{"test-datalake"}}

		_, _, err := inputs.resolveDatalakes(ui, ac, "123")
		assert.Equal(t, errors.New("client error"), err)
		assert.Equal(t, "123", expectedGroupID)
	})

	t.Run("should error if creating template with a datalake", func(t *testing.T) {
		_, ui := mock.NewUI()
		ac := mock.AtlasClient{}
		inputs := createInputs{
			newAppInputs: newAppInputs{
				Template: "ios.template.todo",
			},
			Datalakes: []string{"test-datalake"},
		}

		_, _, err := inputs.resolveDatalakes(ui, ac, "123")
		assert.Equal(t, errors.New("cannot create a template app with data lakes"), err)
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
		directoryID := strings.TrimLeft(defaultPath, testAppName+"-")
		assert.True(t, primitive.IsValidObjectID(directoryID), fmt.Sprintf("should be primitive object id, but found %s", directoryID))
	})
}
