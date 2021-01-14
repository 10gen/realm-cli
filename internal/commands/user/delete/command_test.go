package delete

import (
	"errors"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/commands/user/shared"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestUserDeleteSetup(t *testing.T) {
	t.Run("Should construct a Realm client with the configured base url", func(t *testing.T) {
		profile := mock.NewProfile(t)
		profile.SetRealmBaseURL("http://localhost:8080")

		cmd := &command{inputs: inputs{}}
		assert.Nil(t, cmd.realmClient)

		assert.Nil(t, cmd.Setup(profile, nil))
		assert.NotNil(t, cmd.realmClient)
	})
}

func TestUserDeleteHandler(t *testing.T) {
	projectID := "projectID"
	appID := "appID"
	testApp := realm.App{
		ID:          appID,
		GroupID:     projectID,
		ClientAppID: "eggcorn-abcde",
		Name:        "eggcorn",
	}
	testUsers := []realm.User{
		{
			ID: "user-1",
			Identities: []realm.UserIdentity{
				{ProviderType: shared.ProviderTypeAnonymous},
			},
			Disabled: false,
		},
	}

	t.Run("Should delete a user when a user id is provided", func(t *testing.T) {
		var capturedAppFilter realm.AppFilter
		var capturedProjectID, capturedAppID string

		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			capturedAppFilter = filter
			return []realm.App{testApp}, nil
		}

		realmClient.DeleteUserFn = func(groupID, appID, userID string) error {
			capturedProjectID = groupID
			capturedAppID = appID
			return nil
		}

		cmd := &command{
			inputs: inputs{
				ProjectAppInputs: cli.ProjectAppInputs{
					Project: projectID,
					App:     appID,
				},
				Users: []string{testUsers[0].ID},
			},
			realmClient: realmClient,
		}

		assert.Nil(t, cmd.Handler(nil, nil))
		assert.Equal(t, realm.AppFilter{App: appID, GroupID: projectID}, capturedAppFilter)
		assert.Equal(t, projectID, capturedProjectID)
		assert.Equal(t, appID, capturedAppID)
		assert.Equal(t, testUsers[0].ID, cmd.inputs.Users[0])
		assert.Nil(t, cmd.outputs.failed)
	})

	t.Run("Should save failed deletion errors", func(t *testing.T) {
		var capturedAppFilter realm.AppFilter
		var capturedProjectID, capturedAppID string

		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			capturedAppFilter = filter
			return []realm.App{testApp}, nil
		}

		realmClient.DeleteUserFn = func(groupID, appID, userID string) error {
			capturedProjectID = groupID
			capturedAppID = appID
			return errors.New("client error")
		}

		cmd := &command{
			inputs: inputs{
				ProjectAppInputs: cli.ProjectAppInputs{
					Project: projectID,
					App:     appID,
				},
				Users: []string{testUsers[0].ID},
			},
			realmClient: realmClient,
		}

		assert.Nil(t, cmd.Handler(nil, nil))
		assert.Equal(t, realm.AppFilter{App: appID, GroupID: projectID}, capturedAppFilter)
		assert.Equal(t, projectID, capturedProjectID)
		assert.Equal(t, appID, capturedAppID)
		assert.Equal(t, testUsers[0].ID, cmd.inputs.Users[0])
		assert.Equal(t, cmd.outputs.failed[0], errors.New("failed to delete user (user-1): client error"))
	})

	t.Run("Should return an error", func(t *testing.T) {
		for _, tc := range []struct {
			description string
			setupClient func() realm.Client
			expectedErr error
		}{
			{
				description: "When resolving the app fails",
				setupClient: func() realm.Client {
					realmClient := mock.RealmClient{}
					realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
						return nil, errors.New("something bad happened")
					}
					return realmClient
				},
				expectedErr: errors.New("something bad happened"),
			},
			{
				description: "When finding the users fails",
				setupClient: func() realm.Client {
					realmClient := mock.RealmClient{}
					realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
						return []realm.App{testApp}, nil
					}
					realmClient.FindUsersFn = func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
						return nil, errors.New("something bad happened")
					}
					return realmClient
				},
				expectedErr: errors.New("something bad happened"),
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				realmClient := tc.setupClient()

				cmd := &command{
					realmClient: realmClient,
				}

				err := cmd.Handler(nil, nil)
				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})
}

func TestUserDeleteFeedback(t *testing.T) {
	for _, tc := range []struct {
		description    string
		failed         []error
		expectedOutput string
	}{
		{
			description:    "Should show success message",
			failed:         []error{},
			expectedOutput: "01:23:45 UTC INFO  Successfully deleted all selected users!\n",
		},
		{
			description: "Should show 1 failed user",
			failed: []error{
				errors.New("failed to delete user (user-1): client error"),
			},
			expectedOutput: strings.Join(
				[]string{
					"01:23:45 UTC INFO  Unable to delete the following users:",
					"[failed to delete user (user-1): client error]\n",
				},
				" ",
			),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			out, ui := mock.NewUI()

			cmd := &command{
				outputs: outputs{
					failed: tc.failed,
				},
			}

			assert.Nil(t, cmd.Feedback(nil, ui))
			assert.Equal(t, tc.expectedOutput, out.String())
		})
	}
}
