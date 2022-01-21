package whoami

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestWhoamiFeedback(t *testing.T) {
	t.Run("should print the auth details", func(t *testing.T) {
		for _, tc := range []struct {
			description string
			setup       func(t *testing.T, profile *user.Profile)
			test        func(t *testing.T, output string)
		}{
			{
				description: "with no user logged in",
				test: func(t *testing.T, output string) {
					assert.Equal(t, "No user is currently logged in\n", output)
				},
			},
			{
				description: "with a user that has no active session with cloud credentials",
				setup: func(t *testing.T, profile *user.Profile) {
					profile.SetCredentials(user.Credentials{PublicAPIKey: "apiKey", PrivateAPIKey: "my-super-secret-key"})
				},
				test: func(t *testing.T, output string) {
					assert.Equal(t, "The user, apiKey, is not currently logged in\n", output)
				},
			},
			{
				description: "with a user fully logged in with cloud credentials",
				setup: func(t *testing.T, profile *user.Profile) {
					profile.SetCredentials(user.Credentials{PublicAPIKey: "apiKey", PrivateAPIKey: "my-super-secret-key"})
					profile.SetSession(user.Session{"accessToken", "refreshToken"})
				},
				test: func(t *testing.T, output string) {
					assert.Equal(t, "Currently logged in user: apiKey (**-*****-******-key)\n", output)
				},
			},
			{
				description: "with a user that has no active session with local credentials",
				setup: func(t *testing.T, profile *user.Profile) {
					profile.SetCredentials(user.Credentials{Username: "username", Password: "my-super-secret-pwd"})
				},
				test: func(t *testing.T, output string) {
					assert.Equal(t, "The user, username, is not currently logged in\n", output)
				},
			},
			{
				description: "with a user fully logged in with local credentials",
				setup: func(t *testing.T, profile *user.Profile) {
					profile.SetCredentials(user.Credentials{Username: "username", Password: "my-super-secret-pwd"})
					profile.SetSession(user.Session{"accessToken", "refreshToken"})
				},
				test: func(t *testing.T, output string) {
					assert.Equal(t, "Currently logged in user: username (*******************)\n", output)
				},
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				profile := mock.NewProfile(t)

				if tc.setup != nil {
					tc.setup(t, profile)
				}

				out, ui := mock.NewUI()

				cmd := &Command{}
				err := cmd.Handler(profile, ui, cli.Clients{})
				assert.Nil(t, err)

				tc.test(t, out.String())
			})
		}
	})
}
