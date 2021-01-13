package delete

import (
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
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
			// TODO: Test on valid inputs
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
}
