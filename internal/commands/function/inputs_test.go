package function

import (
	"errors"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
	"github.com/Netflix/go-expect"
)

func TestFunctionInputsResolve(t *testing.T) {
	t.Run("should pass without interaction when function name is set", func(t *testing.T) {
		profile := mock.NewProfile(t)

		i := inputs{Name: "test"}
		assert.Nil(t, i.Resolve(profile, nil))
	})

	t.Run("should prompt for function name", func(t *testing.T) {
		profile := mock.NewProfile(t)

		procedure := func(c *expect.Console) {
			c.ExpectString("Function Name")
			c.Send("test")
			c.SendLine(" ")
			c.ExpectEOF()
		}

		_, console, _, ui, err := mock.NewVT10XConsole()
		assert.Nil(t, err)
		defer console.Close()

		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)
			procedure(console)
		}()

		i := inputs{Name: "test"}
		assert.Nil(t, i.Resolve(profile, ui))

		console.Tty().Close() // flush the writers
		<-doneCh              // wait for procedure to complete
	})
}

func TestFunctionInputsResolveFunction(t *testing.T) {
	t.Run("should confirm function name exists when function name is set", func(t *testing.T) {
		var group, app string
		rc := mock.RealmClient{}
		rc.FunctionsFn = func(groupID, appID string) ([]realm.Function, error) {
			group = groupID
			app = appID
			return []realm.Function{{Name: "test"}}, nil
		}

		i := inputs{Name: "test"}
		function, err := i.ResolveFunction(nil, rc, "test-project", "test-app")
		assert.Nil(t, err)

		assert.Equal(t, realm.Function{Name: "test"}, function)
		assert.Equal(t, "test-project", group)
		assert.Equal(t, "test-app", app)
	})

	t.Run("should select function name when multiple matches", func(t *testing.T) {
		procedure := func(c *expect.Console) {
			c.ExpectString("Select Function")
			c.SendLine("foo")
			c.ExpectEOF()
		}

		_, console, _, ui, err := mock.NewVT10XConsole()
		assert.Nil(t, err)
		defer console.Close()

		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)
			procedure(console)
		}()

		var group, app string
		rc := mock.RealmClient{}
		rc.FunctionsFn = func(groupID, appID string) ([]realm.Function, error) {
			group = groupID
			app = appID
			return []realm.Function{{Name: "foo"}, {Name: "bar"}}, nil
		}

		i := inputs{Name: "foo bar"}
		function, err := i.ResolveFunction(ui, rc, "test-project", "test-app")
		assert.Nil(t, err)

		console.Tty().Close() // flush the writers
		<-doneCh              // wait for procedure to complete

		assert.Equal(t, realm.Function{Name: "foo"}, function)
		assert.Equal(t, "test-project", group)
		assert.Equal(t, "test-app", app)
	})

	t.Run("should error with function name set and no found functions", func(t *testing.T) {
		var group, app string
		rc := mock.RealmClient{}
		rc.FunctionsFn = func(groupID, appID string) ([]realm.Function, error) {
			group = groupID
			app = appID
			return []realm.Function{}, errors.New("realm client error")
		}

		i := inputs{Name: "test"}
		function, err := i.ResolveFunction(nil, rc, "test-project", "test-app")
		assert.Equal(t, errors.New("realm client error"), err)
		assert.Equal(t, realm.Function{}, function)
		assert.Equal(t, "test-project", group)
		assert.Equal(t, "test-app", app)
	})
}
