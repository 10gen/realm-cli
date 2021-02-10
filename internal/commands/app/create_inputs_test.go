package app

import (
	"errors"
	"fmt"
	"path"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"github.com/Netflix/go-expect"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAppCreateInputsResolve(t *testing.T) {
	for _, tc := range []struct {
		description string
		inputs      createInputs
		procedure   func(c *expect.Console)
		test        func(t *testing.T, i createInputs)
	}{
		{
			description: "with no flags set should prompt for just name and set realm.Location and deployment model to defaults",
			procedure: func(c *expect.Console) {
				c.ExpectString("App Name")
				c.SendLine("test-app")
				c.ExpectEOF()
			},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, "test-app", i.Name)
				assert.Equal(t, flagDeploymentModelDefault, i.DeploymentModel)
				assert.Equal(t, flagLocationDefault, i.Location)
			},
		},
		{
			description: "with a name flag set should prompt for nothing else and set realm.Location and deployment model to defaults",
			inputs:      createInputs{newAppInputs: newAppInputs{Name: "test-app"}},
			procedure:   func(c *expect.Console) {},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, "test-app", i.Name)
				assert.Equal(t, flagDeploymentModelDefault, i.DeploymentModel)
				assert.Equal(t, flagLocationDefault, i.Location)
			},
		},
		{
			description: "with name realm.Location and deployment model flags set should prompt for nothing else",
			inputs: createInputs{newAppInputs: newAppInputs{
				Name:            "test-app",
				DeploymentModel: realm.DeploymentModelLocal,
				Location:        realm.LocationOregon,
			}},
			procedure: func(c *expect.Console) {},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, "test-app", i.Name)
				assert.Equal(t, realm.DeploymentModelLocal, i.DeploymentModel)
				assert.Equal(t, realm.LocationOregon, i.Location)
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			profile := mock.NewProfile(t)

			_, console, _, ui, consoleErr := mock.NewVT10XConsole()
			assert.Nil(t, consoleErr)
			defer console.Close()

			doneCh := make(chan (struct{}))
			go func() {
				defer close(doneCh)
				tc.procedure(console)
			}()

			assert.Nil(t, tc.inputs.Resolve(profile, ui))

			console.Tty().Close() // flush the writers
			<-doneCh              // wait for procedure to complete

			tc.test(t, tc.inputs)
		})
	}
}

func TestAppCreateInputsResolveAppName(t *testing.T) {
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
		procedure      func(c *expect.Console)
		findAppErr     error
		expectedName   string
		expectedFilter realm.AppFilter
		expectedErr    error
	}{
		{
			description:  "should return name if name is set",
			inputs:       createInputs{newAppInputs: newAppInputs{Name: testApp.Name}},
			procedure:    func(c *expect.Console) {},
			expectedName: testApp.Name,
		},
		{
			description:    "should use from app for name if name is not set",
			from:           from{testApp.GroupID, testApp.ID},
			procedure:      func(c *expect.Console) {},
			expectedName:   testApp.Name,
			expectedFilter: realm.AppFilter{GroupID: testApp.GroupID, App: testApp.ID},
		},
		{
			description:    "should error when finding app",
			from:           from{testApp.GroupID, testApp.ID},
			procedure:      func(c *expect.Console) {},
			findAppErr:     errors.New("realm client error"),
			expectedFilter: realm.AppFilter{GroupID: testApp.GroupID, App: testApp.ID},
			expectedErr:    errors.New("realm client error"),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			var appFilter realm.AppFilter
			rc := mock.RealmClient{}
			rc.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				appFilter = filter
				return []realm.App{testApp}, tc.findAppErr
			}

			_, console, _, ui, consoleErr := mock.NewVT10XConsole()
			assert.Nil(t, consoleErr)
			defer console.Close()

			doneCh := make(chan (struct{}))
			go func() {
				defer close(doneCh)
				tc.procedure(console)
			}()

			name, err := tc.inputs.resolveAppName(ui, rc, tc.from)

			console.Tty().Close() // flush the writers
			<-doneCh              // wait for procedure to complete

			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedName, name)
			assert.Equal(t, tc.expectedFilter, appFilter)
		})
	}
}

func TestAppCreateInputsResolveProject(t *testing.T) {
	testApp := realm.App{
		ID:          primitive.NewObjectID().Hex(),
		GroupID:     primitive.NewObjectID().Hex(),
		ClientAppID: "test-app-abcde",
		Name:        "test-app",
	}

	for _, tc := range []struct {
		description     string
		inputs          createInputs
		procedure       func(c *expect.Console)
		groupsErr       error
		expectedProject string
		expectedErr     error
	}{
		{
			description:     "should return project if project is set",
			inputs:          createInputs{newAppInputs: newAppInputs{Project: testApp.GroupID}},
			procedure:       func(c *expect.Console) {},
			expectedProject: testApp.GroupID,
		},
		{
			description: "should prompt for project if project is not set",
			procedure: func(c *expect.Console) {
				c.ExpectString("Atlas Project")
				c.Send(testApp.GroupID)
				c.SendLine(" ")
				c.ExpectEOF()
			},
			expectedProject: testApp.GroupID,
		},
		{
			description: "should error when finding group",
			procedure:   func(c *expect.Console) {},
			groupsErr:   errors.New("atlas client error"),
			expectedErr: errors.New("atlas client error"),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			ac := mock.AtlasClient{}
			ac.GroupsFn = func() ([]atlas.Group, error) {
				return []atlas.Group{{ID: testApp.GroupID}}, tc.groupsErr
			}

			_, console, _, ui, consoleErr := mock.NewVT10XConsole()
			assert.Nil(t, consoleErr)
			defer console.Close()

			doneCh := make(chan (struct{}))
			go func() {
				defer close(doneCh)
				tc.procedure(console)
			}()

			projectID, err := tc.inputs.resolveProject(ui, ac)

			console.Tty().Close() // flush the writers
			<-doneCh              // wait for procedure to complete

			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedProject, projectID)
		})
	}
}

func TestAppCreateInputsResolveDirectory(t *testing.T) {
	t.Run("should return path of wd with app name appended", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()

		inputs := createInputs{}
		appName := "test-app"

		dir, err := inputs.resolveDirectory(profile.WorkingDirectory, appName)

		assert.Nil(t, err)
		assert.Equal(t, path.Join(profile.WorkingDirectory, appName), dir)
	})

	t.Run("should return path of wd with directory appended when directory is set", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()

		specifiedDir := "test-dir"
		inputs := createInputs{Directory: specifiedDir}

		dir, err := inputs.resolveDirectory(profile.WorkingDirectory, "")

		assert.Nil(t, err)
		assert.Equal(t, path.Join(profile.WorkingDirectory, specifiedDir), dir)
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
		configErr := localApp.WriteConfig()
		assert.Nil(t, configErr)

		dir, err := inputs.resolveDirectory(profile.WorkingDirectory, "")

		assert.Equal(t, "", dir)
		assert.Equal(t, fmt.Errorf("A Realm app already exists at %s", fullDir), err)
	})
}
