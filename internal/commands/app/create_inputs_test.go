package app

import (
	"errors"
	"os"
	"path"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"github.com/Netflix/go-expect"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAppCreateInputsResolve(t *testing.T) {
	t.Run("with no flags set should prompt for just name and set location and deployment model to defaults", func(t *testing.T) {
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
	})
	t.Run("with a name flag set should prompt for nothing else and set location and deployment model to defaults", func(t *testing.T) {
		profile := mock.NewProfile(t)

		inputs := createInputs{newAppInputs: newAppInputs{Name: "test-app"}}
		assert.Nil(t, inputs.Resolve(profile, nil))

		assert.Equal(t, "test-app", inputs.Name)
		assert.Equal(t, flagDeploymentModelDefault, inputs.DeploymentModel)
		assert.Equal(t, flagLocationDefault, inputs.Location)
	})
	t.Run("with name location and deployment model flags set should prompt for nothing else", func(t *testing.T) {
		profile := mock.NewProfile(t)

		inputs := createInputs{newAppInputs: newAppInputs{
			Name:            "test-app",
			DeploymentModel: realm.DeploymentModelLocal,
			Location:        realm.LocationOregon,
		}}
		assert.Nil(t, inputs.Resolve(profile, nil))

		assert.Equal(t, "test-app", inputs.Name)
		assert.Equal(t, realm.DeploymentModelLocal, inputs.DeploymentModel)
		assert.Equal(t, realm.LocationOregon, inputs.Location)
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
		from           from
		expectedName   string
		expectedFilter realm.AppFilter
	}{
		{
			description:  "should return name if name is set",
			inputs:       createInputs{newAppInputs: newAppInputs{Name: testApp.Name}},
			expectedName: testApp.Name,
		},
		{
			description:    "should use from app for name if name is not set",
			from:           from{testApp.GroupID, testApp.ID},
			expectedName:   testApp.Name,
			expectedFilter: realm.AppFilter{GroupID: testApp.GroupID, App: testApp.ID},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			var appFilter realm.AppFilter
			rc := mock.RealmClient{}
			rc.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				appFilter = filter
				return []realm.App{testApp}, nil
			}

			err := tc.inputs.resolveName(nil, rc, tc.from)

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
		err := inputs.resolveName(nil, rc, from{testApp.GroupID, testApp.ID})

		assert.Equal(t, errors.New("realm client error"), err)
		assert.Equal(t, "", inputs.Name)
		assert.Equal(t, realm.AppFilter{GroupID: testApp.GroupID, App: testApp.ID}, appFilter)
	})
}

func TestAppCreateInputsResolveDirectory(t *testing.T) {
	t.Run("should return path of wd with app name appended", func(t *testing.T) {
		profile := mock.NewProfileFromWd(t)

		appName := "test-app"
		inputs := createInputs{newAppInputs: newAppInputs{Name: appName}}

		dir, err := inputs.resolveDirectory(profile.WorkingDirectory)

		assert.Nil(t, err)
		assert.Equal(t, path.Join(profile.WorkingDirectory, appName), dir)
	})

	t.Run("should return path of wd with directory appended when directory is set", func(t *testing.T) {
		profile := mock.NewProfileFromWd(t)

		specifiedDir := "test-dir"
		inputs := createInputs{Directory: specifiedDir}

		dir, err := inputs.resolveDirectory(profile.WorkingDirectory)

		assert.Nil(t, err)
		assert.Equal(t, path.Join(profile.WorkingDirectory, specifiedDir), dir)
	})

	t.Run("should return path of wd with app name appended even with file of app name in wd", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()

		appName := "test-app"
		inputs := createInputs{newAppInputs: newAppInputs{Name: appName}}

		testFile, err := os.Create(appName)
		assert.Nil(t, err)
		assert.Nil(t, testFile.Close())

		dir, err := inputs.resolveDirectory(profile.WorkingDirectory)

		assert.Nil(t, err)
		assert.Equal(t, path.Join(profile.WorkingDirectory, appName), dir)
		assert.Nil(t, os.Remove(appName))
	})

	t.Run("should error when path specified is another realm app", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()

		specifiedDir := "test-dir"
		inputs := createInputs{Directory: specifiedDir}
		fullDir := path.Join(profile.WorkingDirectory, specifiedDir)

		localApp := local.NewApp(
			fullDir,
			"test-app",
			flagLocationDefault,
			flagDeploymentModelDefault,
		)
		assert.Nil(t, localApp.WriteConfig())

		dir, err := inputs.resolveDirectory(profile.WorkingDirectory)

		assert.Equal(t, "", dir)
		assert.Equal(t, errProjectExists{fullDir}, err)
	})
}

func TestAppCreateInputsResolveDataSource(t *testing.T) {
	t.Run("should return data source config of a provided cluster", func(t *testing.T) {
		var expectedGroupID, expectedAppID string
		rc := mock.RealmClient{}
		rc.ListClustersFn = func(groupID, appID string) ([]realm.PartialAtlasCluster, error) {
			expectedGroupID = groupID
			expectedAppID = appID
			return []realm.PartialAtlasCluster{{ID: "789", Name: "test-cluster"}}, nil
		}

		inputs := createInputs{newAppInputs: newAppInputs{Name: "test-app"}, DataSource: "test-cluster"}

		ds, err := inputs.resolveDataSource(rc, "123", "456")
		assert.Nil(t, err)

		assert.Equal(t, dataSource{
			Name: "test-app_cluster",
			Type: "mongodb-atlas",
			Config: dataSourceConfig{
				ClusterName:         "test-cluster",
				ReadPreference:      "primary",
				WireProtocolEnabled: false,
			},
		}, ds)
		assert.Equal(t, "123", expectedGroupID)
		assert.Equal(t, "456", expectedAppID)
	})

	t.Run("should not be able to find specified cluster", func(t *testing.T) {
		var expectedGroupID, expectedAppID string
		rc := mock.RealmClient{}
		rc.ListClustersFn = func(groupID, appID string) ([]realm.PartialAtlasCluster, error) {
			expectedGroupID = groupID
			expectedAppID = appID
			return nil, nil
		}

		inputs := createInputs{DataSource: "test-cluster"}

		_, err := inputs.resolveDataSource(rc, "123", "456")
		assert.Equal(t, errors.New("failed to find Atlas cluster"), err)
		assert.Equal(t, "123", expectedGroupID)
		assert.Equal(t, "456", expectedAppID)
	})

	t.Run("should error from client", func(t *testing.T) {
		var expectedGroupID, expectedAppID string
		rc := mock.RealmClient{}
		rc.ListClustersFn = func(groupID, appID string) ([]realm.PartialAtlasCluster, error) {
			expectedGroupID = groupID
			expectedAppID = appID
			return nil, errors.New("client error")
		}

		inputs := createInputs{DataSource: "test-cluster"}

		_, err := inputs.resolveDataSource(rc, "123", "456")
		assert.Equal(t, errors.New("client error"), err)
		assert.Equal(t, "123", expectedGroupID)
		assert.Equal(t, "456", expectedAppID)
	})
}
