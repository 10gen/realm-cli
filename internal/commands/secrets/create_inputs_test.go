package secrets

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
	"github.com/Netflix/go-expect"
)

func TestCreateInputs(t *testing.T) {
	for _, tc := range []struct {
		description    string
		createInputs   createInputs
		prepareProfile func(p *cli.Profile)
		procedure      func(c *expect.Console)
		test           func(t *testing.T, i createInputs)
	}{
		{
			description: "should prompt for secret name when not provided",
			createInputs: createInputs{
				Value: "value",
			},
			prepareProfile: func(p *cli.Profile) {},
			procedure: func(c *expect.Console) {
				c.ExpectString("Secret Name")
				c.SendLine("name")
			},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, "name", i.Name)
			},
		},
		{
			description: "Should prompt for secret value when not provided",
			createInputs: createInputs{
				Name: "name",
			},
			prepareProfile: func(p *cli.Profile) {},
			procedure: func(c *expect.Console) {
				c.ExpectString("Secret Value")
				c.SendLine("value")
				c.ExpectEOF()
			},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, "value", i.Value)
			},
		},
		{
			description:    "Should prompt for both secret name and secret value when not provided",
			prepareProfile: func(p *cli.Profile) {},
			procedure: func(c *expect.Console) {
				c.ExpectString("Secret Name")
				c.SendLine("name")
				c.ExpectString("Secret Value")
				c.SendLine("value")
				c.ExpectEOF()
			},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, "name", i.Name)
				assert.Equal(t, "value", i.Value)
			},
		},
		{
			description: "Should not prompt for inputs when flags provide the data",
			createInputs: createInputs{
				Name:  "name",
				Value: "value",
			},
			prepareProfile: func(p *cli.Profile) {},
			procedure: func(c *expect.Console) {
				// c.ExpectEOF()
			},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, "name", i.Name)
				assert.Equal(t, "value", i.Value)
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			_, console, _, ui, consoleErr := mock.NewVT10XConsole()
			assert.Nil(t, consoleErr)
			defer console.Close()

			profile := mock.NewProfile(t)
			tc.prepareProfile(profile)

			doneCh := make(chan (struct{}))
			go func() {
				defer close(doneCh)
				tc.procedure(console)
			}()

			assert.Nil(t, tc.createInputs.Resolve(profile, ui))

			console.Tty().Close() // flush the writers
			<-doneCh              // wait for procedure to complete

			tc.test(t, tc.createInputs)
		})
	}
}
