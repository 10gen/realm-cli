package delete

import (
	"errors"
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/commands/user/shared"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
	"github.com/AlecAivazis/survey/v2"
	"github.com/Netflix/go-expect"
)

// TODO: Test on ui error
func TestUserDeleteResolveInputs(t *testing.T) {
	for _, tc := range []struct {
		description string
		inputs      inputs
		procedure   func(c *expect.Console)
		test        func(t *testing.T, i inputs)
	}{
		{
			description: "With no input set",
			inputs:      inputs{},
			procedure:   func(c *expect.Console) {},
			test: func(t *testing.T, i inputs) {
				assert.Nil(t, i.Users)
			},
		},
		{
			description: "With providers set to interactive",
			inputs: inputs{
				cli.ProjectAppInputs{},
				shared.UsersInputs{ProviderTypes: []string{shared.ProviderTypeInteractive}},
				[]string{},
			},
			procedure: func(c *expect.Console) {
				c.ExpectString("Which provider(s) would you like to filter confirmed users by? Selecting none is equivalent to selecting all.")
				c.Send(shared.ProviderTypeAPIKey)
				c.SendLine(" ")
				c.ExpectEOF()
			},
			test: func(t *testing.T, i inputs) {
				assert.Equal(t, i.ProviderTypes, []string{shared.ProviderTypeAPIKey})
			},
		},
		{
			description: "With state set to interactive",
			inputs: inputs{
				cli.ProjectAppInputs{},
				shared.UsersInputs{State: shared.UserStateTypeInteractive},
				[]string{},
			},
			procedure: func(c *expect.Console) {
				c.ExpectString("Which user state would you like to filter confirmed users by? Selecting none is equivalent to selecting all.")
				c.Send(shared.UserStateTypeEnabled.String())
				c.SendLine(" ")
				c.ExpectEOF()
			},
			test: func(t *testing.T, i inputs) {
				assert.Equal(t, i.State, shared.UserStateTypeEnabled)
			},
		},
		{
			description: "With status set to interactive",
			inputs: inputs{
				cli.ProjectAppInputs{},
				shared.UsersInputs{Status: shared.StatusTypeInteractive},
				[]string{},
			},
			procedure: func(c *expect.Console) {
				c.ExpectString("Which user status would you like to filter users by? Selecting none is equivalent to selecting all.")
				c.Send(shared.StatusTypePending.String())
				c.SendLine(" ")
				c.ExpectEOF()
			},
			test: func(t *testing.T, i inputs) {
				assert.Equal(t, i.Status, shared.StatusTypePending)
			},
		},
	} {
		t.Run(fmt.Sprintf("%s Setup should prompt for the missing inputs, except Users", tc.description), func(t *testing.T) {
			profile := mock.NewProfile(t)

			_, console, _, ui, consoleErr := mock.NewVT10XConsole()
			assert.Nil(t, consoleErr)
			defer console.Close()

			doneCh := make(chan (struct{}))
			go func() {
				defer close(doneCh)
				tc.procedure(console)
			}()

			tc.inputs.App = "some-app" // avoid app resolution
			assert.Nil(t, tc.inputs.Resolve(profile, ui))

			console.Tty().Close() // flush the writers
			<-doneCh              // wait for procedure to complete

			tc.test(t, tc.inputs)
		})
	}
	t.Run("Setup should Error from UI", func(t *testing.T) {
		profile := mock.NewProfile(t)
		inputs := inputs{
			cli.ProjectAppInputs{},
			shared.UsersInputs{Status: shared.StatusTypeInteractive},
			[]string{},
		}
		procedure := func(c *expect.Console) {}
		expectedErr := errors.New("ui error")

		_, console, _, ui, consoleErr := mock.NewVT10XConsole()
		assert.Nil(t, consoleErr)
		defer console.Close()

		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)
			procedure(console)
		}()

		ui.AskOneFn = func(answer interface{}, prompt survey.Prompt) error {
			return errors.New("ui error")
		}

		inputs.App = "some-app" // avoid app resolution
		assert.Equal(t, inputs.Resolve(profile, ui), expectedErr)

		console.Tty().Close() // flush the writers
		<-doneCh              // wait for procedure to complete
	})
}
