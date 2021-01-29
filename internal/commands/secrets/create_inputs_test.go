package secrets

import (
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"github.com/Netflix/go-expect"
)

func TestCreateInputs(t *testing.T) {
	for _, tc := range []struct {
		description string
		inputs      createInputs
		procedure   func(c *expect.Console)
		test        func(t *testing.T, i createInputs)
	}{
		{
			description: "should prompt for secret name when not provided",
			inputs: createInputs{
				Value: "value",
			},
			procedure: func(c *expect.Console) {
				c.ExpectString("Secret Name")
				c.SendLine("name")
				c.ExpectEOF()
			},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, createInputs{Name: "name", Value: "value"}, i)
			},
		},
		{
			description: "Should prompt for secret value when not provided",
			inputs: createInputs{
				Name: "name",
			},
			procedure: func(c *expect.Console) {
				c.ExpectString("Secret Value")
				c.SendLine("value")
				c.ExpectEOF()
			},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, createInputs{Name: "name", Value: "value"}, i)
			},
		},
		{
			description: "Should prompt for both secret name and secret value when not provided",
			procedure: func(c *expect.Console) {
				c.ExpectString("Secret Name")
				c.SendLine("name")
				c.ExpectString("Secret Value")
				c.SendLine("value")
				c.ExpectEOF()
			},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, createInputs{Name: "name", Value: "value"}, i)
			},
		},
		{
			description: "Should not prompt for inputs when flags provide the data",
			inputs: createInputs{
				Name:  "name",
				Value: "value",
			},
			procedure: func(c *expect.Console) {},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, createInputs{Name: "name", Value: "value"}, i)
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			_, console, _, ui, consoleErr := mock.NewVT10XConsole()
			assert.Nil(t, consoleErr)
			defer console.Close()

			profile := mock.NewProfile(t)

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
