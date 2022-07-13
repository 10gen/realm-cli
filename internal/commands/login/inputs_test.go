package login

import (
	"bytes"
	"errors"
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
		expectedURL    string
	}{
		{
			description: "should prompt for public api key when not provided for cloud type",
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
			expectedURL: apiKeysPage,
		},
		{
			description: "should prompt for private api key when not provided for cloud type",
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
			expectedURL: apiKeysPage,
		},
		{
			description:    "should prompt for both api keys when not provided for cloud type",
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
			expectedURL: apiKeysPage,
		},
		{
			description: "should not prompt for inputs when flags provide the data for cloud type",
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
			description: "should not prompt for inputs when profile provides the data for cloud type",
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
			description: "should prompt for username when not provided for local type",
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
			expectedURL: apiKeysPage,
		},
		{
			description: "should prompt for password when not provided for local type",
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
			expectedURL: apiKeysPage,
		},
		{
			description:    "should prompt for both username and password when not provided for local type",
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
			expectedURL: apiKeysPage,
		},
		{
			description: "should not prompt for inputs when flags provide the data for local type",
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
			description: "should not prompt for inputs when profile provides the data for local type",
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
		{
			description: "should open browser when flag is set even if a profile exists",
			inputs:      inputs{AuthType: authTypeCloud, Browser: true},
			prepareProfile: func(p *user.Profile) {
				p.SetCredentials(user.Credentials{PublicAPIKey: "username", PrivateAPIKey: "password"})
			},
			procedure: func(c *expect.Console) {
				c.ExpectString("API Key")
				c.SendLine("username")
				c.ExpectString("Private API Key")
				c.SendLine("password")
				c.ExpectEOF()
			},
			test:        func(t *testing.T, i inputs) {},
			expectedURL: apiKeysPage,
		},
		{
			description: "should not open browser when api keys are provided as inputs",
			inputs: inputs{
				AuthType:      authTypeCloud,
				PublicAPIKey:  "public",
				PrivateAPIKey: "private",
			},
			prepareProfile: func(p *user.Profile) {},
			procedure:      func(c *expect.Console) {},
			test:           func(t *testing.T, i inputs) {},
		},
		{
			description: "should not open browser when username and password are provided as inputs",
			inputs: inputs{
				AuthType: authTypeLocal,
				Username: "username",
				Password: "password",
			},
			prepareProfile: func(p *user.Profile) {},
			procedure:      func(c *expect.Console) {},
			test:           func(t *testing.T, i inputs) {},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			var expectedURL string
			openBrowserFunc := func(url string) error {
				expectedURL = url
				return nil
			}

			console, _, ui, consoleErr := mock.NewVT10XConsoleWithOptions(
				mock.UIOptions{OpenBrowserFn: openBrowserFunc},
				new(bytes.Buffer),
			)

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

			assert.Equal(t, tc.expectedURL, expectedURL)
			tc.test(t, tc.inputs)
		})
	}

	t.Run("should output error message and prompt for credentials if browser cannot be opened", func(t *testing.T) {
		openBrowserFunc := func(url string) error {
			return errors.New("there was an issue opening your browser")
		}

		out := new(bytes.Buffer)
		console, _, ui, consoleErr := mock.NewVT10XConsoleWithOptions(
			mock.UIOptions{OpenBrowserFn: openBrowserFunc},
			out,
		)

		assert.Nil(t, consoleErr)
		defer console.Close()

		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)
			console.ExpectString("there was an issue opening your browser")
			console.ExpectString("Username")
			console.SendLine("username")
			console.ExpectString("Password")
			console.SendLine("password")
			console.ExpectEOF()
		}()

		i := inputs{AuthType: authTypeLocal}

		profile := mock.NewProfile(t)
		assert.Nil(t, i.Resolve(profile, ui))

		console.Tty().Close() // flush the writers
		<-doneCh              // wait for procedure to complete

		assert.Equal(t, "username", i.Username)
		assert.Equal(t, "password", i.Password)
	})

	t.Run("should error if browser flag is enabled and credentials are input", func(t *testing.T) {
		profile := mock.NewProfile(t)

		i := inputs{
			AuthType: authTypeLocal,
			Username: "username",
			Password: "password",
			Browser:  true,
		}

		err := i.Resolve(profile, nil)
		assert.Equal(t, errors.New("credentials will not be authenticated while using browser flag, please login with one or the other"), err)
	})
}
