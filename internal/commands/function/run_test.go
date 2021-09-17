package function

import (
	"bytes"
	"errors"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestFunctionHandler(t *testing.T) {
	t.Run("should return results of a single arg when project app function name and args are set", func(t *testing.T) {
		profile := mock.NewProfile(t)

		rc := mock.RealmClient{}
		rc.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", Name: "test-app"}}, nil
		}
		rc.FunctionsFn = func(groupID, appID string) ([]realm.Function, error) {
			return []realm.Function{{Name: "test"}}, nil
		}
		rc.AppDebugExecuteFunctionFn = func(groupID, appID, userID, name string, args []interface{}) (realm.ExecutionResults, error) {
			return realm.ExecutionResults{Result: map[string]interface{}{
				"arg": []interface{}{
					"hello",
					map[string]interface{}{
						"$numberInt": "1",
					},
					map[string]interface{}{
						"$numberDouble": "2.1",
					},
					map[string]interface{}{
						"foo": "bar",
					},
				},
			}}, nil
		}

		clients := cli.Clients{Realm: rc}

		out := new(bytes.Buffer)
		ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

		cmd := &CommandRun{runInputs{
			ProjectInputs: cli.ProjectInputs{
				Project: "test-project",
				App:     "test-app",
			},
			Name: "test",
			Args: []string{"[\"hello\",1,2.1,{\"foo\": \"bar\"}]"},
		}}
		assert.Nil(t, cmd.Handler(profile, ui, clients))

		display := `Result
{
  "arg": [
    "hello",
    {
      "$numberInt": "1"
    },
    {
      "$numberDouble": "2.1"
    },
    {
      "foo": "bar"
    }
  ]
}
`
		assert.Equal(t, display, out.String())
	})

	t.Run("should return results of two args when project app function name and args are set", func(t *testing.T) {
		profile := mock.NewProfile(t)

		rc := mock.RealmClient{}
		rc.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", Name: "test-app"}}, nil
		}
		rc.FunctionsFn = func(groupID, appID string) ([]realm.Function, error) {
			return []realm.Function{{Name: "test"}}, nil
		}
		rc.AppDebugExecuteFunctionFn = func(groupID, appID, userID, name string, args []interface{}) (realm.ExecutionResults, error) {
			return realm.ExecutionResults{Result: map[string]interface{}{
				"arg1": map[string]interface{}{
					"value1": map[string]interface{}{"$numberInt": "1"},
					"abcs":   []interface{}{"x", "y", "z"},
				},
				"arg2": []interface{}{
					map[string]interface{}{
						"$numberInt": "1",
					},
					map[string]interface{}{
						"$numberInt": "2",
					},
				},
			}}, nil
		}

		clients := cli.Clients{Realm: rc}

		out := new(bytes.Buffer)
		ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

		cmd := &CommandRun{runInputs{
			ProjectInputs: cli.ProjectInputs{
				Project: "test-project",
				App:     "test-app",
			},
			Name: "test",
			Args: []string{"{\"value1\": 1,\"abcs\": [\"x\", \"y\", \"z\"]}", "[1,2]"},
		}}
		assert.Nil(t, cmd.Handler(profile, ui, clients))

		display := `Result
{
  "arg1": {
    "abcs": [
      "x",
      "y",
      "z"
    ],
    "value1": {
      "$numberInt": "1"
    }
  },
  "arg2": [
    {
      "$numberInt": "1"
    },
    {
      "$numberInt": "2"
    }
  ]
}
`
		assert.Equal(t, display, out.String())
	})

	for _, tc := range []struct {
		description   string
		realmClient   realm.Client
		errorExpected error
	}{
		{
			description: "should error on find apps",
			realmClient: mock.RealmClient{
				FindAppsFn: func(filter realm.AppFilter) ([]realm.App, error) {
					return []realm.App{{ID: "appID", Name: "test-app"}}, errors.New("find apps error")
				},
			},
			errorExpected: errors.New("find apps error"),
		},
		{
			description: "should error on find functions",
			realmClient: mock.RealmClient{
				FindAppsFn: func(filter realm.AppFilter) ([]realm.App, error) {
					return []realm.App{{ID: "appID", Name: "test-app"}}, nil
				},
				FunctionsFn: func(groupID, appID string) ([]realm.Function, error) {
					return []realm.Function{{Name: "test"}}, errors.New("find functions error")
				},
			},
			errorExpected: errors.New("find functions error"),
		},
		{
			description: "should error on function execution",
			realmClient: mock.RealmClient{
				FindAppsFn: func(filter realm.AppFilter) ([]realm.App, error) {
					return []realm.App{{ID: "appID", Name: "test-app"}}, nil
				},
				FunctionsFn: func(groupID, appID string) ([]realm.Function, error) {
					return []realm.Function{{Name: "test"}}, nil
				},
				AppDebugExecuteFunctionFn: func(groupID, appID, userID, name string, args []interface{}) (realm.ExecutionResults, error) {
					return realm.ExecutionResults{}, errors.New("function execution error")
				},
			},
			errorExpected: errors.New("function execution error"),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			profile := mock.NewProfile(t)

			_, ui := mock.NewUI()

			clients := cli.Clients{Realm: tc.realmClient}

			cmd := &CommandRun{runInputs{
				ProjectInputs: cli.ProjectInputs{
					Project: "test-project",
					App:     "test-app",
				},
				Name: "test",
				Args: []string{"Hello world"},
			}}
			assert.Equal(t, tc.errorExpected, cmd.Handler(profile, ui, clients))
		})
	}
}
