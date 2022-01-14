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
			description: "Should prompt for public api key when not provided for cloud type",
			inputs: inputs{
				AuthType:      authTypeCloud,
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
			description: "Should prompt for private api key when not provided for cloud type",
			inputs: inputs{
				AuthType:     authTypeCloud,
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
			description:    "Should prompt for both api keys when not provided for cloud type",
			inputs:         inputs{AuthType: authTypeCloud},
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
			description: "Should not prompt for inputs when flags provide the data for cloud type",
			inputs: inputs{
				AuthType:      authTypeCloud,
				PublicAPIKey:  "username",
				PrivateAPIKey: "password",
			},
			prepareProfile: func(p *user.Profile) {},
			procedure:      func(c *expect.Console) {},
			test: func(t *testing.T, i inputs) {
				assert.Equal(t, "username", i.PublicAPIKey)
				assert.Equal(t, "password", i.PrivateAPIKey)
			},
		},
		{
			description: "Should not prompt for inputs when profile provides the data for cloud type",
			inputs:      inputs{AuthType: authTypeCloud},
			prepareProfile: func(p *user.Profile) {
				p.SetCredentials(user.Credentials{PublicAPIKey: "username", PrivateAPIKey: "password"})
			},
			procedure: func(c *expect.Console) {
				c.ExpectEOF()
			},
			test: func(t *testing.T, i inputs) {
				assert.Equal(t, "username", i.PublicAPIKey)
				assert.Equal(t, "password", i.PrivateAPIKey)
			},
		},
		{
			description: "Should prompt for username when not provided for local type",
			inputs: inputs{
				AuthType: authTypeLocal,
				Password: "password",
			},
			prepareProfile: func(p *user.Profile) {},
			procedure: func(c *expect.Console) {
				c.ExpectString("Username")
				c.SendLine("username")
			},
			test: func(t *testing.T, i inputs) {
				assert.Equal(t, "username", i.Username)
			},
		},
		{
			description: "Should prompt for password when not provided for local type",
			inputs: inputs{
				AuthType: authTypeLocal,
				Username: "username",
			},
			prepareProfile: func(p *user.Profile) {},
			procedure: func(c *expect.Console) {
				c.ExpectString("Password")
				c.SendLine("password")
				c.ExpectEOF()
			},
			test: func(t *testing.T, i inputs) {
				assert.Equal(t, "password", i.Password)
			},
		},
		{
			description:    "Should prompt for both username and password when not provided for local type",
			inputs:         inputs{AuthType: authTypeLocal},
			prepareProfile: func(p *user.Profile) {},
			procedure: func(c *expect.Console) {
				c.ExpectString("Username")
				c.SendLine("username")
				c.ExpectString("Password")
				c.SendLine("password")
				c.ExpectEOF()
			},
			test: func(t *testing.T, i inputs) {
				assert.Equal(t, "username", i.Username)
				assert.Equal(t, "password", i.Password)
			},
		},
		{
			description: "Should not prompt for inputs when flags provide the data for local type",
			inputs: inputs{
				AuthType: authTypeLocal,
				Username: "username",
				Password: "password",
			},
			prepareProfile: func(p *user.Profile) {},
			procedure:      func(c *expect.Console) {},
			test: func(t *testing.T, i inputs) {
				assert.Equal(t, "username", i.Username)
				assert.Equal(t, "password", i.Password)
			},
		},
		{
			description: "Should not prompt for inputs when profile provides the data for local type",
			inputs:      inputs{AuthType: authTypeLocal},
			prepareProfile: func(p *user.Profile) {
				p.SetCredentials(user.Credentials{Username: "username", Password: "password"})
			},
			procedure: func(c *expect.Console) {
				c.ExpectEOF()
			},
			test: func(t *testing.T, i inputs) {
				assert.Equal(t, "username", i.Username)
				assert.Equal(t, "password", i.Password)
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
