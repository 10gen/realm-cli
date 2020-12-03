package list

import (
	"errors"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAppListSetup(t *testing.T) {
	t.Run("Setup creates a realm client with a session", func(t *testing.T) {
		ctx := cli.Context{RealmBaseURL: "http://localhost:8080"}
		cmd := &command{}

		profile, profileErr := cli.NewProfile(primitive.NewObjectID().Hex())
		assert.Nil(t, profileErr)

		profile.SetSession("authToken", "refreshToken")

		err := cmd.Setup(profile, nil, ctx)
		assert.Nil(t, err)
		assert.NotNil(t, &cmd.realmClient)
	})
}

func TestAppListHandler(t *testing.T) {
	project1 := "project1"
	project2 := "project2"
	roles := []realm.Role{realm.Role{project1}, realm.Role{project2}, realm.Role{project2}}

	realmClient := mock.RealmClient{}
	realmClient.GetAuthProfileFn = func() (realm.AuthProfile, error) {
		return realm.AuthProfile{
			Roles: roles,
		}, nil
	}

	testApp1 := realm.App{
		GroupID: project1,
		Name:    "name1",
	}

	testApp2 := realm.App{
		GroupID: project2,
		Name:    "name2",
	}

	testApp3 := realm.App{
		GroupID: project2,
		Name:    "name3",
	}

	realmClient.GetAppsForUserFn = func() ([]realm.App, error) {
		return []realm.App{
			testApp1,
			testApp2,
			testApp3,
		}, nil
	}

	realmClient.GetAppsFn = func(groupID string) ([]realm.App, error) {
		switch groupID {
		case project1:
			return []realm.App{testApp1}, nil
		case project2:
			return []realm.App{
				testApp2,
				testApp3,
			}, nil
		default:
			return nil, errors.New("test error")
		}
	}

	t.Run("Returns all apps for all projects if no app or project flags present", func(t *testing.T) {
		cmd := &command{
			realmClient: realmClient,
		}
		cmd.Handler(nil, nil, nil)
		assert.Equal(t, []realm.App{testApp1, testApp2, testApp3}, cmd.appListResult)
	})

	t.Run("Returns all apps for specified projects if project flag present", func(t *testing.T) {
		cmd := &command{
			project:     project2,
			realmClient: realmClient,
		}
		cmd.Handler(nil, nil, nil)
		assert.Equal(t, []realm.App{testApp2, testApp3}, cmd.appListResult)
	})

	//TODO REALMC-7547 uncomment and reimplement these tests once app flag is supported

	// t.Run("Returns single app if app flag present", func(t *testing.T) {
	// 	cmd := &command{
	// 		app:         appName3,
	// 		realmClient: realmClient,
	// 	}
	// 	cmd.Handler(nil, nil, nil)
	// 	assert.Equal(t, 1, len(cmd.appListResult))
	// 	assert.Equal(t, []realm.App{testApp3}, cmd.appListResult)
	// })

	// t.Run("Returns correct app if app and project flags present", func(t *testing.T) {
	// 	cmd := &command{
	// 		app:         appName3,
	// 		project:     project2,
	// 		realmClient: realmClient,
	// 	}
	// 	cmd.Handler(nil, nil, nil)
	// 	assert.Equal(t, 1, len(cmd.appListResult))
	// 	assert.Equal(t, []realm.App{testApp3}, cmd.appListResult)
	// })

	t.Run("Returns no apps if user has no apps and no app flag provided", func(t *testing.T) {
		realmClient.GetAppsForUserFn = func() ([]realm.App, error) {
			return nil, nil
		}
		cmd := &command{
			realmClient: realmClient,
		}
		err := cmd.Handler(nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(cmd.appListResult))
	})
}

// TODO: REALMC-7156 pretty print into table
// func TestAppListFeedback(t *testing.T) {
// 	t.Run("Feedback should print apps in a table", func(t *testing.T) {
// 		out := new(bytes.Buffer)
// 		ui := mock.NewUI(mock.UIOptions{}, out)

// 		testApp1 := realm.App{
// 			ID:          "id1",
// 			GroupID:     "project1",
// 			ClientAppID: "client1",
// 			Name:        "app1",
// 		}

// 		testApp2 := realm.App{
// 			ID:          "id2",
// 			GroupID:     "project2",
// 			ClientAppID: "client2",
// 			Name:        "app2",
// 		}

// 		cmd := &command{
// 			appListResult: []realm.App{testApp1, testApp2},
// 		}

// 		err := cmd.Feedback(nil, ui)
// 		assert.Nil(t, err)

// 		assert.Equal(t, "INFO  01:23:45: Successfully logged in.\n", out.String())
// 	})
// }
