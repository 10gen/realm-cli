package whoami

import (
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestWhoamiFeedback(t *testing.T) {
	t.Run("should print the auth details and associated projects", func(t *testing.T) {
		for _, tc := range []struct {
			description    string
			setup          func(t *testing.T, profile *user.Profile)
			expectedOutput string
			showProjects   bool
		}{
			{
				description:    "with no user logged in",
				expectedOutput: "No user is currently logged in\n",
			},
			{
				description: "with a user that has no active session with cloud credentials",
				setup: func(t *testing.T, profile *user.Profile) {
					profile.SetCredentials(user.Credentials{PublicAPIKey: "apiKey", PrivateAPIKey: "my-super-secret-key"})
				},
				expectedOutput: "The user, apiKey, is not currently logged in\n",
			},
			{
				description: "with a user fully logged in with cloud credentials",
				setup: func(t *testing.T, profile *user.Profile) {
					profile.SetCredentials(user.Credentials{PublicAPIKey: "apiKey", PrivateAPIKey: "my-super-secret-key"})
					profile.SetSession(user.Session{"accessToken", "refreshToken"})
				},
				expectedOutput: "Currently logged in user: apiKey (**-*****-******-key)\n",
			},
			{
				description: "with a user that has no active session with local credentials",
				setup: func(t *testing.T, profile *user.Profile) {
					profile.SetCredentials(user.Credentials{Username: "username", Password: "my-super-secret-pwd"})
				},
				expectedOutput: "The user, username, is not currently logged in\n",
			},
			{
				description: "with a user fully logged in with local credentials",
				setup: func(t *testing.T, profile *user.Profile) {
					profile.SetCredentials(user.Credentials{Username: "username", Password: "my-super-secret-pwd"})
					profile.SetSession(user.Session{"accessToken", "refreshToken"})
				},
				expectedOutput: "Currently logged in user: username (*******************)\n",
			},
			{
				description: "with show-projects flag enabled should output list of projects",
				setup: func(t *testing.T, profile *user.Profile) {
					profile.SetCredentials(user.Credentials{PublicAPIKey: "apiKey", PrivateAPIKey: "my-super-secret-key"})
					profile.SetSession(user.Session{"accessToken", "refreshToken"})
				},
				expectedOutput: strings.Join([]string{
					"Currently logged in user: apiKey (**-*****-******-key)",
					"Projects available (1)",
					"  Project ID  Project Name",
					"  ----------  ------------",
					"  groupID     groupName   ",
					"",
				}, "\n"),
				showProjects: true,
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				profile := mock.NewProfile(t)

				var groupsCalled bool
				var atlasClient mock.AtlasClient
				atlasClient.GroupsFn = func(url string, useBaseURL bool) (atlas.Groups, error) {
					groupsCalled = true
					return atlas.Groups{Results: []atlas.Group{{ID: "groupID", Name: "groupName"}}}, nil
				}

				if tc.setup != nil {
					tc.setup(t, profile)
				}

				out, ui := mock.NewUI()

				cmd := &Command{inputs{showProjects: tc.showProjects}}
				err := cmd.Handler(profile, ui, cli.Clients{Atlas: atlasClient})
				assert.Nil(t, err)

				assert.Equal(t, tc.showProjects, groupsCalled)
				assert.Equal(t, tc.expectedOutput, out.String())
			})
		}
	})
}
