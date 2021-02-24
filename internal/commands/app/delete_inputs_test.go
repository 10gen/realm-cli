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
				description: "should return an empty apps slice with no found apps or input apps without prompting",
			},
			{
				description:  "should prompt user to select apps if no input provided",
				apps:         testApps,
				expectedApps: []realm.App{app2},
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				realmClient := mock.RealmClient{}
				realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
					return tc.apps, nil
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

				apps, err := tc.inputs.resolveApps(ui, realmClient)
				assert.Equal(t, tc.expectedErr, err)

				console.Tty().Close() // flush the writers
				<-doneCh              // wait for procedure to complete

				assert.Equal(t, tc.expectedApps, apps)
			})
		}
	})

}
