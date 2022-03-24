package app

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"github.com/Netflix/go-expect"
)

func TestAppInitInputsResolve(t *testing.T) {
	t.Run("should return an error if ran from a directory that already has a project", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_init_input_test")
		defer teardown()

		assert.Nil(t, ioutil.WriteFile(
			filepath.Join(profile.WorkingDirectory, local.FileRealmConfig.String()),
			[]byte(fmt.Sprintf(`{"config_version": %d, "name":"eggcorn"}`, realm.DefaultAppConfigVersion)),
			0666,
		))

		var i initInputs
		assert.Equal(t, errProjectExists(""), i.Resolve(profile, nil))
	})

	for _, tc := range []struct {
		description string
		inputs      initInputs
		procedure   func(c *expect.Console)
		test        func(t *testing.T, i initInputs)
	}{
		{
			description: "with no flags set should prompt for just name and set location deployment model and environment to defaults",
			procedure: func(c *expect.Console) {
				c.ExpectString("App Name")
				c.SendLine("test-app")
				c.ExpectEOF()
			},
			test: func(t *testing.T, i initInputs) {
				assert.Equal(t, "test-app", i.Name)
				assert.Equal(t, flagDeploymentModelDefault, i.DeploymentModel)
				assert.Equal(t, flagLocationDefault, i.Location)
				assert.Equal(t, realm.EnvironmentNone, i.Environment)
			},
		},
		{
			description: "with a name flag set should prompt for nothing else and set location deployment model and environment to defaults",
			inputs:      initInputs{newAppInputs: newAppInputs{Name: "test-app"}},
			procedure:   func(c *expect.Console) {},
			test: func(t *testing.T, i initInputs) {
				assert.Equal(t, "test-app", i.Name)
				assert.Equal(t, flagDeploymentModelDefault, i.DeploymentModel)
				assert.Equal(t, flagLocationDefault, i.Location)
				assert.Equal(t, realm.EnvironmentNone, i.Environment)
			},
		},
		{
			description: "with name location deployment model and environment flags set should prompt for nothing else",
			inputs: initInputs{newAppInputs: newAppInputs{
				Name:            "test-app",
				DeploymentModel: realm.DeploymentModelLocal,
				Location:        realm.LocationOregon,
				Environment:     realm.EnvironmentDevelopment,
			}},
			procedure: func(c *expect.Console) {},
			test: func(t *testing.T, i initInputs) {
				assert.Equal(t, "test-app", i.Name)
				assert.Equal(t, realm.DeploymentModelLocal, i.DeploymentModel)
				assert.Equal(t, realm.LocationOregon, i.Location)
				assert.Equal(t, realm.EnvironmentDevelopment, i.Environment)
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
