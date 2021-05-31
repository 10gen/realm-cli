package function

import (
	"errors"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
	"github.com/AlecAivazis/survey/v2/terminal"
)

func TestFunctionRunInputsResolveFunction(t *testing.T) {
	t.Run("should return an error when finding functions fails", func(t *testing.T) {
		var realmClient mock.RealmClient

		var capturedGroupID, capturedAppID string
		realmClient.FunctionsFn = func(groupID, appID string) ([]realm.Function, error) {
			capturedGroupID = groupID
			capturedAppID = appID
			return nil, errors.New("something bad happened")
		}

		var i runInputs

		_, err := i.resolveFunction(nil, realmClient, "groupID", "appID")
		assert.Equal(t, errors.New("something bad happened"), err)

		t.Log("and receive the expected args to the client functions call")
		assert.Equal(t, "groupID", capturedGroupID)
		assert.Equal(t, "appID", capturedAppID)
	})

	t.Run("should return an error when no functions are found", func(t *testing.T) {
		var realmClient mock.RealmClient

		realmClient.FunctionsFn = func(groupID, appID string) ([]realm.Function, error) {
			return nil, nil
		}

		var i runInputs

		_, err := i.resolveFunction(nil, realmClient, "groupID", "appID")
		assert.Equal(t, errors.New("no functions available to run"), err)
	})

	t.Run("should return an error when no functions of the specified name are found", func(t *testing.T) {
		var realmClient mock.RealmClient

		realmClient.FunctionsFn = func(groupID, appID string) ([]realm.Function, error) {
			return []realm.Function{{Name: "fredFunc"}}, nil
		}

		i := runInputs{Name: "uptownFunc"}

		_, err := i.resolveFunction(nil, realmClient, "groupID", "appID")
		assert.Equal(t, errors.New("failed to find function 'uptownFunc'"), err)
	})

	for _, tc := range []struct {
		description  string
		appFunctions []realm.Function
		inputs       runInputs
		expectedFn   realm.Function
	}{
		{
			description:  "should find a function by name",
			appFunctions: []realm.Function{{Name: "func1"}, {Name: "func2"}, {Name: "func3"}},
			inputs:       runInputs{Name: "func2"},
			expectedFn:   realm.Function{Name: "func2"},
		},
		{
			description:  "should return a function if only one found and name input is not set",
			appFunctions: []realm.Function{{Name: "func1"}},
			expectedFn:   realm.Function{Name: "func1"},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			var realmClient mock.RealmClient
			realmClient.FunctionsFn = func(groupID, appID string) ([]realm.Function, error) {
				return tc.appFunctions, nil
			}

			fn, err := tc.inputs.resolveFunction(nil, realmClient, "groupID", "appID")
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedFn, fn)
		})
	}

	t.Run("should prompt the user to select from a list of available apps if no name input is set", func(t *testing.T) {
		var realmClient mock.RealmClient
		realmClient.FunctionsFn = func(groupID, appID string) ([]realm.Function, error) {
			return []realm.Function{{Name: "func1"}, {Name: "func2"}, {Name: "func3"}}, nil
		}

		_, console, _, ui, err := mock.NewVT10XConsole()
		assert.Nil(t, err)
		defer console.Close()

		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)

			console.ExpectString("Select Function")
			console.Send(string(terminal.KeyArrowDown))
			console.SendLine("")
			console.ExpectEOF()
		}()

		var i runInputs

		fn, err := i.resolveFunction(ui, realmClient, "groupID", "appID")
		assert.Nil(t, err)

		console.Tty().Close() // flush the writers
		<-doneCh              // wait for procedure to complete

		assert.Equal(t, realm.Function{Name: "func2"}, fn)
	})
}
