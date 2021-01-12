package delete

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
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

	t.Run("Should delete a user when a user id is provided", func(t *testing.T) {
		testUsers := []realm.User{{ID: "user1"}}

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
}

func TestUserDeleteFeedback(t *testing.T) {
}
