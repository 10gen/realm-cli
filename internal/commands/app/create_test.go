package app

import (
	"archive/zip"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
	"github.com/Netflix/go-expect"
)

func TestAppCreateHandler(t *testing.T) {
	t.Run("should create minimal project when no from type is specified", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()

		profile.SetRealmBaseURL("http://localhost:8080")

		out, ui := mock.NewUI()

		client := mock.RealmClient{}
		client.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return realm.App{
				GroupID:     groupID,
				ID:          "456",
				ClientAppID: name + "-abcde",
				Name:        name,
				AppMeta:     meta,
			}, nil
		}
		client.ImportFn = func(groupID, appID string, appData interface{}) error {
			return nil
		}

		cmd := &CommandCreate{createInputs{newAppInputs: newAppInputs{
			Name:            "test-app",
			Project:         "123",
			Location:        realm.LocationVirginia,
			DeploymentModel: realm.DeploymentModelGlobal,
		}}}

		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: client}))

		localApp, err := local.LoadApp(filepath.Join(profile.WorkingDirectory, cmd.inputs.Name))
		assert.Nil(t, err)

		// FIXME: lets figure out to display a more deterministic output (filepaths can vary wildly in length here)
		dirLength := len(localApp.RootDir)
		fmtStr := fmt.Sprintf("%%-%ds", dirLength)
		fmtStrLast := fmt.Sprintf("%%-%ds", dirLength+20)

		assert.Equal(t, strings.Join([]string{
			"01:23:45 UTC INFO  Successfully created app",
			fmt.Sprintf("  Info                "+fmtStr, "Details"),
			"  ------------------  " + strings.Repeat("-", dirLength),
			fmt.Sprintf("  Client App ID       "+fmtStr, "test-app-abcde"),
			"  Realm Directory     " + localApp.RootDir,
			fmt.Sprintf("  Realm UI            "+fmtStr, "http://localhost:8080/groups/123/apps/456/dashboard"),
			fmt.Sprintf("  "+fmtStrLast, "Check out your app  cd ./test-app && realm-cli app describe"),
			"",
		}, "\n"), out.String())

		assert.Equal(t, &local.AppRealmConfigJSON{local.AppDataV2{local.AppStructureV2{
			ConfigVersion:   realm.DefaultAppConfigVersion,
			Name:            "test-app",
			Location:        realm.LocationVirginia,
			DeploymentModel: realm.DeploymentModelGlobal,
		}}}, localApp.AppData)
	})

	t.Run("when from and project is not set should create minimal project and prompt for project", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()

		var createdApp realm.App
		rc := mock.RealmClient{}
		rc.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			createdApp = realm.App{
				GroupID: groupID,
				Name:    name,
				AppMeta: meta,
			}
			return createdApp, nil
		}
		rc.ImportFn = func(groupID, appID string, appData interface{}) error {
			return nil
		}
		ac := mock.AtlasClient{}
		ac.GroupsFn = func() ([]atlas.Group, error) {
			return []atlas.Group{{ID: "test-project"}}, nil
		}

		procedure := func(c *expect.Console) {
			c.ExpectString("Atlas Project")
			c.Send("test-project")
			c.SendLine(" ")
			c.ExpectEOF()
		}

		_, console, _, ui, consoleErr := mock.NewVT10XConsole()
		assert.Nil(t, consoleErr)
		defer console.Close()

		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)
			procedure(console)
		}()

		cmd := &CommandCreate{createInputs{newAppInputs: newAppInputs{
			Name:            "test-app",
			Project:         "test-project",
			Location:        realm.LocationVirginia,
			DeploymentModel: realm.DeploymentModelGlobal,
		}}}

		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: rc, Atlas: ac}))

		console.Tty().Close() // flush the writers
		<-doneCh              // wait for procedure to complete

		localApp, err := local.LoadApp(filepath.Join(profile.WorkingDirectory, cmd.inputs.Name))
		assert.Nil(t, err)

		expectedAppData := local.AppRealmConfigJSON{local.AppDataV2{local.AppStructureV2{
			ConfigVersion:   realm.DefaultAppConfigVersion,
			Name:            "test-app",
			Location:        realm.LocationVirginia,
			DeploymentModel: realm.DeploymentModelGlobal,
		}}}

		assert.Equal(t, &expectedAppData, localApp.AppData)
		assert.Equal(t, realm.App{
			GroupID: "test-project",
			Name:    "test-app",
			AppMeta: realm.AppMeta{
				Location:        realm.LocationVirginia,
				DeploymentModel: realm.DeploymentModelGlobal,
			},
		}, createdApp)
	})

	t.Run("should create a new app with a structure based on the specified from app", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()

		profile.SetRealmBaseURL("http://localhost:8080")

		out, ui := mock.NewUI()

		testApp := realm.App{
			ID:          "789",
			GroupID:     "123",
			ClientAppID: "from-app-abcde",
			Name:        "from-app",
		}

		zipPkg, err := zip.OpenReader("testdata/project.zip")
		assert.Nil(t, err)

		client := mock.RealmClient{}
		client.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{testApp}, nil
		}
		client.ExportFn = func(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error) {
			return "", &zipPkg.Reader, err
		}
		client.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return realm.App{
				GroupID:     groupID,
				ID:          "456",
				ClientAppID: name + "-abcde",
				Name:        name,
				AppMeta:     meta,
			}, nil
		}
		client.ImportFn = func(groupID, appID string, appData interface{}) error {
			return nil
		}

		cmd := &CommandCreate{createInputs{newAppInputs: newAppInputs{
			From:            testApp.Name,
			Project:         testApp.GroupID,
			Location:        realm.LocationIreland,
			DeploymentModel: realm.DeploymentModelGlobal,
		}}}

		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: client}))

		localApp, err := local.LoadApp(filepath.Join(profile.WorkingDirectory, cmd.inputs.From))
		assert.Nil(t, err)

		// FIXME: lets figure out to display a more deterministic output (filepaths can vary wildly in length here)
		dirLength := len(localApp.RootDir)
		fmtStr := fmt.Sprintf("%%-%ds", dirLength)
		fmtStrLast := fmt.Sprintf("%%-%ds", dirLength+20)

		assert.Equal(t, strings.Join([]string{
			"01:23:45 UTC INFO  Successfully created app",
			fmt.Sprintf("  Info                "+fmtStr, "Details"),
			"  ------------------  " + strings.Repeat("-", dirLength),
			fmt.Sprintf("  Client App ID       "+fmtStr, "from-app-abcde"),
			"  Realm Directory     " + localApp.RootDir,
			fmt.Sprintf("  Realm UI            "+fmtStr, "http://localhost:8080/groups/123/apps/456/dashboard"),
			fmt.Sprintf("  "+fmtStrLast, "Check out your app  cd ./from-app && realm-cli app describe"),
			"",
		}, "\n"), out.String())

		assert.Equal(t, &local.AppRealmConfigJSON{local.AppDataV2{local.AppStructureV2{
			ConfigVersion:   realm.DefaultAppConfigVersion,
			Name:            testApp.Name,
			Location:        realm.LocationIreland,
			DeploymentModel: realm.DeploymentModelGlobal,
			Auth: &local.AuthStructure{
				CustomUserData: map[string]interface{}{"enabled": false},
				Providers:      map[string]interface{}{},
			},
			Sync: &local.SyncStructure{Config: map[string]interface{}{"development_mode_enabled": false}},
		}}}, localApp.AppData)
	})

	t.Run("should create minimal project with data source when data source is set", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_create_test")
		defer teardown()

		var createdApp realm.App
		client := mock.RealmClient{}
		client.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			createdApp = realm.App{
				GroupID: groupID,
				Name:    name,
				AppMeta: meta,
			}
			return createdApp, nil
		}
		client.ListClustersFn = func(groupID, appID string) ([]realm.PartialAtlasCluster, error) {
			return []realm.PartialAtlasCluster{{Name: "test-cluster", State: "IDLE"}}, nil
		}
		var importAppData interface{}
		client.ImportFn = func(groupID, appID string, appData interface{}) error {
			importAppData = appData
			return nil
		}

		cmd := &CommandCreate{
			inputs: createInputs{
				newAppInputs: newAppInputs{
					Name:            "test-app",
					Project:         "test-project",
					Location:        realm.LocationVirginia,
					DeploymentModel: realm.DeploymentModelGlobal,
				},
				DataSource: "test-cluster"},
			realmClient: client,
		}

		assert.Nil(t, cmd.Handler(profile, nil))

		localApp, err := local.LoadApp(filepath.Join(profile.WorkingDirectory, cmd.inputs.Name))
		assert.Nil(t, err)

		assert.Equal(t, importAppData, localApp.AppData)
		assert.Equal(t, realm.App{
			GroupID: "test-project",
			Name:    "test-app",
			AppMeta: realm.AppMeta{
				Location:        realm.LocationVirginia,
				DeploymentModel: realm.DeploymentModelGlobal,
			},
		}, createdApp)
	})

	t.Run("should error when resolving groupID when project is not set", func(t *testing.T) {
		profile := mock.NewProfileFromWD(t)

		client := mock.AtlasClient{}
		client.GroupsFn = func() ([]atlas.Group, error) {
			return nil, errors.New("atlas client error")
		}

		cmd := &CommandCreate{createInputs{newAppInputs: newAppInputs{
			Name:            "test-app",
			Location:        realm.LocationVirginia,
			DeploymentModel: realm.DeploymentModelGlobal,
		}}}

		assert.Equal(t, errors.New("atlas client error"), cmd.Handler(profile, nil, cli.Clients{Atlas: client}))
	})

	t.Run("should error when resolving app when from is set", func(t *testing.T) {
		profile := mock.NewProfileFromWD(t)

		client := mock.RealmClient{}
		client.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return nil, errors.New("realm client error")
		}

		cmd := &CommandCreate{createInputs{newAppInputs: newAppInputs{
			From:            "test-app",
			Location:        realm.LocationVirginia,
			DeploymentModel: realm.DeploymentModelGlobal,
		}}}

		assert.Equal(t, errors.New("realm client error"), cmd.Handler(profile, nil, cli.Clients{Realm: client}))
	})
}

// func TestAppCreateFeedback(t *testing.T) {
// 	t.Run("feedback should print a message that app creation was successful", func(t *testing.T) {
// 		out, ui := mock.NewUI()

// 		cmd := &CommandCreate{
// 			outputs: createOutputs{
// 				clientAppID: "test-client-id",
// 				dir:         "/file/path/to/test-app",
// 				uiURL:       "https://realm.mongodb.com/groups/123/apps/123/dashboard",
// 				followUpCmd: "cd ./test-app && realm-cli app describe",
// 			},
// 		}

// 		err := cmd.Feedback(nil, ui)
// 		assert.Nil(t, err)

// 		expectedContent := strings.Join(
// 			[]string{
// 				"01:23:45 UTC INFO  Successfully created app",
// 				"  Info                Details                                                ",
// 				"  ------------------  -------------------------------------------------------",
// 				"  Client App ID       test-client-id                                         ",
// 				"  Realm Directory     /file/path/to/test-app                                 ",
// 				"  Realm UI            https://realm.mongodb.com/groups/123/apps/123/dashboard",
// 				"  Check out your app  cd ./test-app && realm-cli app describe                ",
// 				"",
// 			},
// 			"\n",
// 		)

// 		assert.Equal(t, expectedContent, out.String())
// 	})
// }
