package app

import (
	"errors"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestResolveApps(t *testing.T) {
	app1 := realm.App{
		ID:          "60344735b37e3733de2adf40",
		GroupID:     "groupID1",
		ClientAppID: "app1-abcde",
		Name:        "app1",
	}
	app2 := realm.App{
		ID:          "60344735b37e3733de2adf41",
		GroupID:     "groupID1",
		ClientAppID: "app2-hijkl",
		Name:        "app2",
	}
	app3 := realm.App{
		ID:          "60344735b37e3733de2adf42",
		GroupID:     "groupID1",
		ClientAppID: "app3-wxyza",
		Name:        "app3",
	}

	testApps := []realm.App{app1, app2, app3}

	t.Run("when finding apps", func(t *testing.T) {
		t.Run("should return a client error if present", func(t *testing.T) {
			realmClient := mock.RealmClient{}
			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				return nil, errors.New("client error")
			}
			inputs := deleteInputs{}
			_, err := inputs.resolveApps(nil, realmClient)

			assert.Equal(t, errors.New("client error"), err)
		})

		for _, tc := range []struct {
			description  string
			inputs       deleteInputs
			apps         []realm.App
			expectedApps []realm.App
			expectedErr  error
		}{
			{
				description: "should error with input apps and no found apps",
				inputs:      deleteInputs{Apps: []string{"app1"}},
				expectedErr: cli.ErrAppNotFound{},
			},
			{
				description: "should return no apps with no found apps nor input apps without prompting",
			},
			{
				description:  "should return found apps without prompting based on app name inputs provided",
				inputs:       deleteInputs{Apps: []string{"app2", "app3"}},
				apps:         testApps,
				expectedApps: []realm.App{app2, app3},
			},
			{
				description:  "should return found apps without prompting based on app client ID inputs provided",
				inputs:       deleteInputs{Apps: []string{"app1-abcde", "app2-hijkl"}},
				apps:         testApps,
				expectedApps: []realm.App{app1, app2},
			},
			{
				description:  "should return found apps without prompting based on app client ID or name inputs provided",
				inputs:       deleteInputs{Apps: []string{"app1-abcde", "app3"}},
				apps:         testApps,
				expectedApps: []realm.App{app1, app3},
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				realmClient := mock.RealmClient{}
				realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
					return tc.apps, nil
				}

				apps, err := tc.inputs.resolveApps(nil, realmClient)
				assert.Equal(t, tc.expectedErr, err)
				assert.Equal(t, tc.expectedApps, apps)
			})
		}

		t.Run("should prompt user to select apps if no input provided", func(t *testing.T) {
			realmClient := mock.RealmClient{}
			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				return testApps, nil
			}

			_, console, _, ui, consoleErr := mock.NewVT10XConsole()
			assert.Nil(t, consoleErr)
			defer console.Close()

			doneCh := make(chan (struct{}))
			go func() {
				defer close(doneCh)

				console.ExpectString("Select App(s)")
				console.Send("app2")
				console.SendLine(" ")
				console.ExpectEOF()
			}()
			inputs := deleteInputs{}
			apps, err := inputs.resolveApps(ui, realmClient)
			assert.Nil(t, err)

			console.Tty().Close() // flush the writers
			<-doneCh              // wait for procedure to complete

			assert.Equal(t, []realm.App{app2}, apps)
		})

		t.Run("should print a warning if any input apps are not found", func(t *testing.T) {
			realmClient := mock.RealmClient{}
			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				return testApps, nil
			}

			out, ui := mock.NewUI()

			inputs := deleteInputs{Apps: []string{"nonexistent", "app1"}}
			apps, err := inputs.resolveApps(ui, realmClient)
			assert.Nil(t, err)
			assert.Equal(t, "Unable to delete certain apps because they were not found: nonexistent\n", out.String())
			assert.Equal(t, []realm.App{app1}, apps)
		})

		t.Run("should print a warning if no input apps found", func(t *testing.T) {
			realmClient := mock.RealmClient{}
			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				return testApps, nil
			}

			out, ui := mock.NewUI()

			inputs := deleteInputs{Apps: []string{"nonexistent", "missing"}}
			apps, err := inputs.resolveApps(ui, realmClient)
			assert.Nil(t, err)
			assert.Equal(t, "Unable to delete certain apps because they were not found: nonexistent, missing\n", out.String())
			assert.Equal(t, []realm.App{}, apps)
		})
	})

}
