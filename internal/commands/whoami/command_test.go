package whoami

import (
	"testing"

	"github.com/10gen/realm-cli/internal/auth"
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestWhoamiHandler(t *testing.T) {
	t.Run("Handler should run as a noop", func(t *testing.T) {
		cmd := &Command{}

		err := cmd.Handler(nil, nil)
		assert.Nil(t, err)
	})
}

func TestWhoamiFeedback(t *testing.T) {
	t.Run("Feedback should print the auth details", func(t *testing.T) {
		for _, tc := range []struct {
			description string
			setup       func(t *testing.T, profile *cli.Profile)
			test        func(t *testing.T, output string)
		}{
			{
				description: "with no user logged in",
				test: func(t *testing.T, output string) {
					assert.Equal(t, "01:23:45 UTC INFO  No user is currently logged in\n", output)
				},
			},
			{
				description: "with a user that has no active session",
				setup: func(t *testing.T, profile *cli.Profile) {
					profile.SetUser(auth.User{"username", "my-super-secret-key"})
				},
				test: func(t *testing.T, output string) {
					assert.Equal(t, "01:23:45 UTC INFO  The user, username, is not currently logged in\n", output)
				},
			},
			{
				description: "with a user fully logged in",
				setup: func(t *testing.T, profile *cli.Profile) {
					profile.SetUser(auth.User{"username", "my-super-secret-key"})
					profile.SetSession(auth.Session{"accessToken", "refreshToken"})
				},
				test: func(t *testing.T, output string) {
					assert.Equal(t, "01:23:45 UTC INFO  Currently logged in user: username (**-*****-******-key)\n", output)
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
				err := cmd.Feedback(profile, ui)
				assert.Nil(t, err)

				tc.test(t, out.String())
			})
		}
	})
}
