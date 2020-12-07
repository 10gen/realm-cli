package app

import (
	"errors"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestAppListSetup(t *testing.T) {
	t.Run("Setup creates a realm client with a session", func(t *testing.T) {
		config := cli.CommandConfig{RealmBaseURL: "http://localhost:8080"}
		cmd := &appListCommand{}
		authToken := "authToken"
		refreshToken := "refreshToken"
		profile := &cli.Profile{}
		profile.SetSession(authToken, refreshToken)
		err := cmd.Setup(profile, nil, config)
		assert.Nil(t, err)
		assert.NotNil(t, &cmd.realmClient)
	})
}

func TestAppListHandler(t *testing.T) {
	project1 := "project1"
	project2 := "project2"

	appName1 := "name1"
	appName2 := "name2"
	appName3 := "name3"

	role1 := realm.Role{project1}
	role2 := realm.Role{project2}
	roleWithDuplicateProject := realm.Role{project2}
	roles := []realm.Role{role1, role2, roleWithDuplicateProject}
	realmClient := mock.RealmClient{}
	realmClient.GetUserProfileFn = func() (realm.UserProfile, error) {
		return realm.UserProfile{
			Roles: roles,
		}, nil
	}

	testApp1 := realm.App{
		GroupID: project1,
		Name:    appName1,
	}

	testApp2 := realm.App{
		GroupID: project2,
		Name:    appName2,
	}

	testApp3 := realm.App{
		GroupID: project2,
		Name:    appName3,
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
		cmd := &appListCommand{
			realmClient: realmClient,
		}
		cmd.Handler(nil, nil, nil)
		assert.Equal(t, 3, len(cmd.appListResult))
		assert.Equal(t, []realm.App{testApp1, testApp2, testApp3}, cmd.appListResult)
	})

	t.Run("Returns all apps for specified projects if project flag present", func(t *testing.T) {
		cmd := &appListCommand{
			project:     project2,
			realmClient: realmClient,
		}
		cmd.Handler(nil, nil, nil)
		assert.Equal(t, 2, len(cmd.appListResult))
		assert.Equal(t, []realm.App{testApp2, testApp3}, cmd.appListResult)
	})

	//TODO REALMC-7156 uncomment and reimplement these tests once app flag is supported

	// t.Run("Returns single app if app flag present", func(t *testing.T) {
	// 	cmd := &appListCommand{
	// 		app:         appName3,
	// 		realmClient: realmClient,
	// 	}
	// 	cmd.Handler(nil, nil, nil)
	// 	assert.Equal(t, 1, len(cmd.appListResult))
	// 	assert.Equal(t, []realm.App{testApp3}, cmd.appListResult)
	// })

	// t.Run("Returns correct app if app and project flags present", func(t *testing.T) {
	// 	cmd := &appListCommand{
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
		cmd := &appListCommand{
			realmClient: realmClient,
		}
		err := cmd.Handler(nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(cmd.appListResult))
	})

	t.Run("Returns error if user has no apps and app flag provided", func(t *testing.T) {
		realmClient.GetAppsForUserFn = func() ([]realm.App, error) {
			return nil, nil
		}
		cmd := &appListCommand{
			app:         "random",
			realmClient: realmClient,
		}
		err := cmd.Handler(nil, nil, nil)
		assert.NotNil(t, err)
		assert.Equal(t, "Found no matches, try changing the --app input", err.Error())
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

// 		cmd := &appListCommand{
// 			appListResult: []realm.App{testApp1, testApp2},
// 		}

// 		err := cmd.Feedback(nil, ui)
// 		assert.Nil(t, err)

// 		assert.Equal(t, "INFO  01:23:45: Successfully logged in.\n", out.String())
// 	})
// }
