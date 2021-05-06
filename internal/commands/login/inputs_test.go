package login

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
	"github.com/Netflix/go-expect"
)

func TestLoginInputs(t *testing.T) {
	for _, tc := range []struct {
		description    string
		inputs         inputs
		prepareProfile func(p *user.Profile)
		procedure      func(c *expect.Console)
		test           func(t *testing.T, i inputs)
	}{
		{
			description: "Should prompt for public api key when not provided",
			inputs: inputs{
				PrivateAPIKey: "password",
			},
			prepareProfile: func(p *user.Profile) {},
			procedure: func(c *expect.Console) {
				c.ExpectString("API Key")
				c.SendLine("username")
			},
			test: func(t *testing.T, i inputs) {
				assert.Equal(t, "username", i.PublicAPIKey)
			},
		},
		{
			description: "Should prompt for private api key when not provided",
			inputs: inputs{
				PublicAPIKey: "username",
			},
			prepareProfile: func(p *user.Profile) {},
			procedure: func(c *expect.Console) {
				c.ExpectString("Private API Key")
				c.SendLine("password")
				c.ExpectEOF()
			},
			test: func(t *testing.T, i inputs) {
				assert.Equal(t, "password", i.PrivateAPIKey)
			},
		},
		{
			description:    "Should prompt for both api keys when not provided",
			prepareProfile: func(p *user.Profile) {},
			procedure: func(c *expect.Console) {
				c.ExpectString("API Key")
				c.SendLine("username")
				c.ExpectString("Private API Key")
				c.SendLine("password")
				c.ExpectEOF()
			},
			test: func(t *testing.T, i inputs) {
				assert.Equal(t, "username", i.PublicAPIKey)
				assert.Equal(t, "password", i.PrivateAPIKey)
			},
		},
		{
			description: "Should not prompt for inputs when flags provide the data",
			inputs: inputs{
				PublicAPIKey:  "username",
				PrivateAPIKey: "password",
			},
			prepareProfile: func(p *user.Profile) {},
			procedure: func(c *expect.Console) {
				// c.ExpectEOF()
			},
			test: func(t *testing.T, i inputs) {
				assert.Equal(t, "username", i.PublicAPIKey)
				assert.Equal(t, "password", i.PrivateAPIKey)
			},
		},
		{
			description: "Should not prompt for inputs when profile provides the data",
			prepareProfile: func(p *user.Profile) {
				p.SetCredentials(user.Credentials{"username", "password"})
			},
			procedure: func(c *expect.Console) {
				c.ExpectEOF()
			},
			test: func(t *testing.T, i inputs) {
				assert.Equal(t, "username", i.PublicAPIKey)
				assert.Equal(t, "password", i.PrivateAPIKey)
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

			assert.Nil(t, tc.inputs.Resolve(profile, ui))

			console.Tty().Close() // flush the writers
			<-doneCh              // wait for procedure to complete

			tc.test(t, tc.inputs)
		})
	}
}
