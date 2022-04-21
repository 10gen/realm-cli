package local

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestNewApp(t *testing.T) {
	t.Run("new app should create an instance of the realm config json", func(t *testing.T) {
		expectedApp := App{
			RootDir: "/path/to/project",
			Config:  FileRealmConfig,
			AppData: &AppRealmConfigJSON{AppDataV2{AppStructureV2{
				ConfigVersion:   realm.AppConfigVersion20210101,
				ID:              "testID",
				Name:            "testName",
				Location:        realm.LocationOregon,
				DeploymentModel: realm.DeploymentModelGlobal,
				Environment:     realm.EnvironmentDevelopment,
				Environments: map[string]map[string]interface{}{
					"development.json": {
						"values": map[string]interface{}{},
					},
					"no-environment.json": {
						"values": map[string]interface{}{},
					},
					"production.json": {
						"values": map[string]interface{}{},
					},
					"qa.json": {
						"values": map[string]interface{}{},
					},
					"testing.json": {
						"values": map[string]interface{}{},
					},
				},
				Auth: AuthStructure{
					CustomUserData: map[string]interface{}{"enabled": false},
					Providers:      map[string]interface{}{},
				},
				Sync: SyncStructure{Config: map[string]interface{}{"development_mode_enabled": false}},
				Functions: FunctionsStructure{
					Configs: []map[string]interface{}{},
					Sources: map[string]string{},
				},
				GraphQL: GraphQLStructure{
					Config: map[string]interface{}{
						"use_natural_pluralization": true,
					},
					CustomResolvers: []map[string]interface{}{},
				},
			}}},
		}

		app := NewApp("/path/to/project", "testID", "testName", realm.LocationOregon, realm.DeploymentModelGlobal, realm.EnvironmentDevelopment, realm.DefaultAppConfigVersion)
		assert.Equal(t, expectedApp, app)
	})
}

func TestLoadApp(t *testing.T) {
	wd, wdErr := os.Getwd()
	assert.Nil(t, wdErr)
	testRoot := wd

	t.Run("loading an app with a valid path should succeed", func(t *testing.T) {
		projectRoot := filepath.Join(testRoot, "testdata", "full_project")

		expectedAppLocal := App{
			RootDir: projectRoot,
			Config:  FileConfig,
			AppData: fullProject,
			Meta:    appMeta20200603,
		}

		app, appErr := LoadApp(projectRoot)
		assert.Nil(t, appErr)
		assert.Equal(t, expectedAppLocal, app)
	})

	t.Run("loading an app from an invalid path should error", func(t *testing.T) {
		invalidAppPath := filepath.Join(testRoot, "testdata")

		_, appErr := LoadApp(invalidAppPath)
		assert.Equal(t, appErr, errors.New("failed to find app at "+invalidAppPath))
	})
}

func TestLoadAppMeta(t *testing.T) {
	wd, wdErr := os.Getwd()
	assert.Nil(t, wdErr)

	for _, tc := range []struct {
		configVersion realm.AppConfigVersion
		appMeta       AppMeta
	}{
		{realm.AppConfigVersion20180301, appMeta20180301},
		{realm.AppConfigVersion20200603, appMeta20200603},
		{realm.AppConfigVersion20210101, appMeta20210101},
	} {
		t.Run(fmt.Sprintf("should successfully load app meta config file with version %d", tc.configVersion), func(t *testing.T) {
			projectRoot := filepath.Join(wd, "testdata", tc.configVersion.String(), "app_meta")
			testApp := App{RootDir: projectRoot}

			assert.Nil(t, testApp.loadMeta())
			assert.Equal(t, testApp.AppData, nil)
			assert.Equal(t, testApp.Meta, tc.appMeta)
		})
	}
}

func TestLoadConfig(t *testing.T) {
	wd, wdErr := os.Getwd()
	assert.Nil(t, wdErr)

	for _, tc := range []struct {
		configVersion realm.AppConfigVersion
		configFile    File
		appData       AppData
	}{
		{realm.AppConfigVersion20180301, FileStitch, &AppStitchJSON{appData20180301Local}},
		{realm.AppConfigVersion20200603, FileConfig, &AppConfigJSON{appData20200603Local}},
		{realm.AppConfigVersion20210101, FileRealmConfig, &AppRealmConfigJSON{appData20210101Local}},
	} {
		t.Run(fmt.Sprintf("should successfully load config file with version %d", tc.configVersion), func(t *testing.T) {
			projectRoot := filepath.Join(wd, "testdata", tc.configVersion.String(), "local")
			testApp := App{RootDir: projectRoot, Config: tc.configFile}

			assert.Nil(t, testApp.loadConfig())
			assert.Equal(t, testApp.AppData, tc.appData)
		})
	}

	t.Run("should error with invalid config file", func(t *testing.T) {
		projectRoot := filepath.Join(wd, "testdata", "full_project")
		testApp := App{RootDir: projectRoot, Config: File{"bogus", ".json"}}

		assert.Equal(t, errors.New("invalid config file: bogus.json"), testApp.loadConfig())
	})
}

func TestFindReadMe(t *testing.T) {
	wd, wdErr := os.Getwd()
	assert.Nil(t, wdErr)

	relativePath := filepath.Join("testdata", "template_app")
	testRoot := filepath.Join(wd, relativePath)
	templateID := "xamarin.todo"

	readmePath, err := FindReadme(testRoot, relativePath, templateID)
	assert.Nil(t, err)

	readmeData, err := ioutil.ReadFile(readmePath)
	assert.Nil(t, err)

	expectedData, err := ioutil.ReadFile(filepath.Join(testRoot, "frontend", templateID, "README.md"))
	assert.Nil(t, err)

	assert.Equal(t, string(expectedData), string(readmeData))
}

func TestFindApp(t *testing.T) {
	wd, wdErr := os.Getwd()
	assert.Nil(t, wdErr)

	for _, config := range []struct {
		version       realm.AppConfigVersion
		file          File
		appDataLocal  AppData
		remoteAppData AppData
		nestedAppData AppData
		appMeta       AppMeta
	}{
		{
			version:       realm.AppConfigVersion20180301,
			file:          FileStitch,
			appDataLocal:  &AppStitchJSON{appData20180301Local},
			remoteAppData: &AppStitchJSON{appData20180301Remote},
			nestedAppData: &AppStitchJSON{appData20180301Nested},
			appMeta:       appMeta20180301,
		},
		{
			version:       realm.AppConfigVersion20200603,
			file:          FileConfig,
			appDataLocal:  &AppConfigJSON{appData20200603Local},
			remoteAppData: &AppConfigJSON{appData20200603Remote},
			nestedAppData: &AppConfigJSON{appData20200603Nested},
			appMeta:       appMeta20200603,
		},
		{
			version:       realm.AppConfigVersion20210101,
			file:          FileRealmConfig,
			appDataLocal:  &AppRealmConfigJSON{appData20210101Local},
			remoteAppData: &AppRealmConfigJSON{appData20210101Remote},
			nestedAppData: &AppRealmConfigJSON{appData20210101Nested},
			appMeta:       appMeta20210101,
		},
	} {
		t.Run(fmt.Sprintf("With a %d config version", config.version), func(t *testing.T) {
			testRoot := filepath.Join(wd, "testdata", config.version.String())

			t.Run("and a working directory outside of the project root should fail to find app", func(t *testing.T) {
				_, insideProject, err := FindApp(testRoot)
				assert.Nil(t, err)
				assert.False(t, insideProject, "should be outside project")
			})

			for _, tc := range []struct {
				description string
				name        string
				appData     AppData
				appMeta     AppMeta
			}{
				{"and a working directory at the root of a local project", "local", config.appDataLocal, AppMeta{}},
				{"and a working directory at the root of a remote project", "remote", config.remoteAppData, AppMeta{}},
				{"and a nested working directory containing another config", "nested/graphql", config.nestedAppData, AppMeta{}},
				{"and a nested working directory containing no config", "nested/values", config.nestedAppData, AppMeta{}},
				{"and a working directory with an app meta config", "app_meta", config.appDataLocal, config.appMeta},
			} {
				t.Run(tc.description, func(t *testing.T) {
					path := filepath.Join(testRoot, tc.name)

					_, insideProject, err := FindApp(path)
					assert.Nil(t, err)
					assert.True(t, insideProject, "should be inside project")

					app, err := LoadApp(path)
					assert.Nil(t, err)
					assert.Equal(t, tc.appData, app.AppData)
					assert.Equal(t, tc.appMeta, app.Meta)
				})
			}

			t.Run("and a config file with invalid json should return an error when finding app", func(t *testing.T) {
				projectRoot := filepath.Join(testRoot, "invalid_config")

				_, insideProject, err := FindApp(projectRoot)
				assert.Equal(t, errFailedToParseAppConfig(filepath.Join(projectRoot, config.file.String())), err)
				assert.False(t, insideProject, "should not be found")

				app, err := LoadApp(projectRoot)
				assert.Equal(t, errFailedToParseAppConfig(filepath.Join(projectRoot, config.file.String())), err)
				assert.Equal(t, App{}, app)
			})

			t.Run("and a meta file with invalid json should return an error when finding app", func(t *testing.T) {
				projectRoot := filepath.Join(testRoot, "invalid_meta")

				_, insideProject, err := FindApp(projectRoot)
				assert.Equal(t, errFailedToParseAppMeta(filepath.Join(projectRoot, NameDotMDB, FileAppMeta.String())), err)
				assert.False(t, insideProject, "should not be found")

				app, err := LoadApp(projectRoot)
				assert.Equal(t, errFailedToParseAppMeta(filepath.Join(projectRoot, NameDotMDB, FileAppMeta.String())), err)
				assert.Equal(t, App{}, app)
			})

			t.Run("and missing the version in the config file should fail to find app", func(t *testing.T) {
				path := filepath.Join(testRoot, "no_version")
				_, foundApp, err := FindApp(path)
				assert.Nil(t, err)
				assert.False(t, foundApp, "should not be found")
			})
		})
	}
}

func TestAppWriteConfig(t *testing.T) {
	t.Run("Should write the app config contents successfully", func(t *testing.T) {
		for _, tc := range []struct {
			description string
			appData     AppData
			config      File
			contents    string
		}{
			{
				description: "with an empty stitch json",
				appData:     &AppStitchJSON{},
				config:      FileStitch,
				contents: `{
    "config_version": 0,
    "name": "",
    "location": "",
    "deployment_model": "",
    "security": null,
    "custom_user_data_config": null,
    "sync": null
}
`,
			},
			{
				description: "with an empty config json",
				appData:     &AppConfigJSON{},
				config:      FileConfig,
				contents: `{
    "config_version": 0,
    "name": "",
    "location": "",
    "deployment_model": "",
    "security": null,
    "custom_user_data_config": null,
    "sync": null
}
`,
			},
			{
				description: "with an empty realm config json",
				appData:     &AppRealmConfigJSON{},
				config:      FileRealmConfig,
				contents: `{
    "config_version": 0
}
`,
			},
			{
				description: "With a full 20180301 config",
				appData: &AppStitchJSON{AppDataV1{AppStructureV1{
					ConfigVersion:        realm.AppConfigVersion20180301,
					ID:                   "test-abcde",
					Name:                 "test",
					Location:             realm.LocationVirginia,
					DeploymentModel:      realm.DeploymentModelGlobal,
					Security:             map[string]interface{}{"allowed_origins": []string{"http://localhost:8080"}},
					CustomUserDataConfig: map[string]interface{}{"enabled": true},
					Sync:                 map[string]interface{}{"development_mode_enabled": true},
				}}},
				config: FileStitch,
				contents: `{
    "config_version": 20180301,
    "app_id": "test-abcde",
    "name": "test",
    "location": "US-VA",
    "deployment_model": "GLOBAL",
    "security": {
        "allowed_origins": [
            "http://localhost:8080"
        ]
    },
    "custom_user_data_config": {
        "enabled": true
    },
    "sync": {
        "development_mode_enabled": true
    }
}
`,
			},
			{
				description: "With a full 20200603 config",
				appData: &AppConfigJSON{AppDataV1{AppStructureV1{
					ConfigVersion:        realm.AppConfigVersion20200603,
					ID:                   "test-abcde",
					Name:                 "test",
					Location:             realm.LocationVirginia,
					DeploymentModel:      realm.DeploymentModelGlobal,
					Security:             map[string]interface{}{"allowed_origins": []string{"http://localhost:8080"}},
					CustomUserDataConfig: map[string]interface{}{"enabled": true},
					Sync:                 map[string]interface{}{"development_mode_enabled": true},
				}}},
				config: FileConfig,
				contents: `{
    "config_version": 20200603,
    "app_id": "test-abcde",
    "name": "test",
    "location": "US-VA",
    "deployment_model": "GLOBAL",
    "security": {
        "allowed_origins": [
            "http://localhost:8080"
        ]
    },
    "custom_user_data_config": {
        "enabled": true
    },
    "sync": {
        "development_mode_enabled": true
    }
}
`,
			},
			{
				description: "With a full 20210101 config",
				appData: &AppRealmConfigJSON{AppDataV2{AppStructureV2{
					ConfigVersion:         realm.AppConfigVersion20210101,
					ID:                    "test-abcde",
					Name:                  "test",
					Location:              realm.LocationVirginia,
					DeploymentModel:       realm.DeploymentModelGlobal,
					AllowedRequestOrigins: []string{"http://localhost:8080"},
				}}},
				config: FileRealmConfig,
				contents: `{
    "config_version": 20210101,
    "app_id": "test-abcde",
    "name": "test",
    "location": "US-VA",
    "deployment_model": "GLOBAL",
    "allowed_request_origins": [
        "http://localhost:8080"
    ]
}
`,
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				tmpDir, cleanupTmpDir, err := u.NewTempDir("")
				assert.Nil(t, err)
				defer cleanupTmpDir()

				app := &App{
					RootDir: tmpDir,
					Config:  tc.config,
					AppData: tc.appData,
				}
				assert.Nil(t, app.WriteConfig())

				data, dataErr := ioutil.ReadFile(filepath.Join(tmpDir, tc.config.String()))
				assert.Nil(t, dataErr)
				assert.Equal(t, tc.contents, string(data))
			})
		}
	})
}

func TestAppWrite20180301(t *testing.T) {
	t.Run("should initialize an empty project", func(t *testing.T) {
		tmpDir, cleanupTmpDir, err := u.NewTempDir("")
		assert.Nil(t, err)
		defer cleanupTmpDir()

		app := NewApp(tmpDir, "test-app-abcde", "test-app", realm.LocationIreland, realm.DeploymentModelLocal, realm.EnvironmentDevelopment, realm.AppConfigVersion20180301)

		assert.Nil(t, app.Write())

		data, readErr := ioutil.ReadFile(filepath.Join(tmpDir, FileStitch.String()))
		assert.Nil(t, readErr)

		var config AppStitchJSON
		assert.Nil(t, json.Unmarshal(data, &config))
		assert.Equal(t, AppStitchJSON{AppDataV1{AppStructureV1{
			ConfigVersion:        realm.AppConfigVersion20180301,
			ID:                   "test-app-abcde",
			Name:                 "test-app",
			Location:             realm.LocationIreland,
			DeploymentModel:      realm.DeploymentModelLocal,
			Environment:          realm.EnvironmentDevelopment,
			CustomUserDataConfig: map[string]interface{}{"enabled": false},
			Sync:                 map[string]interface{}{"development_mode_enabled": false},
		}}}, config)

		t.Run("should have auth providers directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(tmpDir, NameAuthProviders))
			assert.Nil(t, err)
		})

		t.Run("should have functions directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(tmpDir, NameFunctions))
			assert.Nil(t, err)
		})

		t.Run("should have graphql custom resolvers directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(tmpDir, NameGraphQL, NameCustomResolvers))
			assert.Nil(t, err)
		})

		t.Run("should have the expected contents in the graphql config file", func(t *testing.T) {
			config, err := ioutil.ReadFile(filepath.Join(tmpDir, NameGraphQL, FileConfig.String()))
			assert.Nil(t, err)
			assert.Equal(t, `{
    "use_natural_pluralization": true
}
`, string(config))
		})

		t.Run("should have services directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(tmpDir, NameServices))
			assert.Nil(t, err)
		})

		t.Run("should have values directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(tmpDir, NameValues))
			assert.Nil(t, err)
		})
	})
}

func TestAppWrite20200603(t *testing.T) {
	t.Run("should initialize an empty project", func(t *testing.T) {
		tmpDir, cleanupTmpDir, err := u.NewTempDir("")
		assert.Nil(t, err)
		defer cleanupTmpDir()

		app := NewApp(tmpDir, "test-app-abcde", "test-app", realm.LocationIreland, realm.DeploymentModelLocal, realm.EnvironmentDevelopment, realm.AppConfigVersion20200603)

		assert.Nil(t, app.Write())

		data, readErr := ioutil.ReadFile(filepath.Join(tmpDir, FileConfig.String()))
		assert.Nil(t, readErr)

		var config AppConfigJSON
		assert.Nil(t, json.Unmarshal(data, &config))
		assert.Equal(t, AppConfigJSON{AppDataV1{AppStructureV1{
			ConfigVersion:        realm.AppConfigVersion20200603,
			ID:                   "test-app-abcde",
			Name:                 "test-app",
			Location:             realm.LocationIreland,
			DeploymentModel:      realm.DeploymentModelLocal,
			Environment:          realm.EnvironmentDevelopment,
			CustomUserDataConfig: map[string]interface{}{"enabled": false},
			Sync:                 map[string]interface{}{"development_mode_enabled": false},
		}}}, config)

		t.Run("should have auth providers directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(tmpDir, NameAuthProviders))
			assert.Nil(t, err)
		})

		t.Run("should have functions directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(tmpDir, NameFunctions))
			assert.Nil(t, err)
		})

		t.Run("should have graphql custom resolvers directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(tmpDir, NameGraphQL, NameCustomResolvers))
			assert.Nil(t, err)
		})

		t.Run("should have the expected contents in the graphql config file", func(t *testing.T) {
			config, err := ioutil.ReadFile(filepath.Join(tmpDir, NameGraphQL, FileConfig.String()))
			assert.Nil(t, err)
			assert.Equal(t, `{
    "use_natural_pluralization": true
}
`, string(config))
		})

		t.Run("should have services directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(tmpDir, NameServices))
			assert.Nil(t, err)
		})

		t.Run("should have values directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(tmpDir, NameValues))
			assert.Nil(t, err)
		})
	})
}

func TestAppWrite20210101(t *testing.T) {
	t.Run("should initialize an empty project", func(t *testing.T) {
		tmpDir, cleanupTmpDir, err := u.NewTempDir("")
		assert.Nil(t, err)
		defer cleanupTmpDir()

		app := NewApp(tmpDir, "test-app-abcde", "test-app", realm.LocationIreland, realm.DeploymentModelLocal, realm.EnvironmentDevelopment, realm.AppConfigVersion20210101)

		assert.Nil(t, app.Write())

		data, readErr := ioutil.ReadFile(filepath.Join(tmpDir, FileRealmConfig.String()))
		assert.Nil(t, readErr)

		var config AppRealmConfigJSON
		assert.Nil(t, json.Unmarshal(data, &config))
		assert.Equal(t, AppRealmConfigJSON{AppDataV2{AppStructureV2{
			ConfigVersion:   realm.AppConfigVersion20210101,
			ID:              "test-app-abcde",
			Name:            "test-app",
			Location:        realm.LocationIreland,
			DeploymentModel: realm.DeploymentModelLocal,
			Environment:     realm.EnvironmentDevelopment,
		}}}, config)

		t.Run("should have the expected contents in the auth custom user data file", func(t *testing.T) {
			config, err := ioutil.ReadFile(filepath.Join(tmpDir, NameAuth, FileCustomUserData.String()))
			assert.Nil(t, err)
			assert.Equal(t, "{\n    \"enabled\": false\n}\n", string(config))
		})

		t.Run("should have the expected contents in the auth providers file", func(t *testing.T) {
			config, err := ioutil.ReadFile(filepath.Join(tmpDir, NameAuth, FileProviders.String()))
			assert.Nil(t, err)
			assert.Equal(t, "{}\n", string(config))
		})

		t.Run("should have data sources directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(tmpDir, NameDataSources))
			assert.Nil(t, err)
		})

		t.Run("should have the expected contents in the functions config file", func(t *testing.T) {
			config, err := ioutil.ReadFile(filepath.Join(tmpDir, NameFunctions, FileConfig.String()))
			assert.Nil(t, err)
			assert.Equal(t, "[]\n", string(config))
		})

		t.Run("should have graphql custom resolvers directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(tmpDir, NameGraphQL, NameCustomResolvers))
			assert.Nil(t, err)
		})

		t.Run("should have the expected contents in the graphql config file", func(t *testing.T) {
			config, err := ioutil.ReadFile(filepath.Join(tmpDir, NameGraphQL, FileConfig.String()))
			assert.Nil(t, err)
			assert.Equal(t, "{\n    \"use_natural_pluralization\": true\n}\n", string(config))
		})

		t.Run("should have http endpoints directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(tmpDir, NameHTTPEndpoints))
			assert.Nil(t, err)
		})

		t.Run("should have services directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(tmpDir, NameServices))
			assert.Nil(t, err)
		})

		t.Run("should have the expected contents in the sync config file", func(t *testing.T) {
			config, err := ioutil.ReadFile(filepath.Join(tmpDir, NameSync, FileConfig.String()))
			assert.Nil(t, err)
			assert.Equal(t, "{\n    \"development_mode_enabled\": false\n}\n", string(config))
		})

		t.Run("should have values directory", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(tmpDir, NameValues))
			assert.Nil(t, err)
		})
	})
}

var fullProject = &AppConfigJSON{AppDataV1{AppStructureV1{
	ConfigVersion:   realm.AppConfigVersion20200603,
	ID:              "full-abcde",
	Name:            "full",
	Location:        "US-VA",
	DeploymentModel: "GLOBAL",
	Sync:            map[string]interface{}{"development_mode_enabled": false},
	AuthProviders: []map[string]interface{}{
		{
			"name":     "api-key",
			"type":     "api-key",
			"disabled": false,
		},
	},
	Functions: []map[string]interface{}{
		{
			NameConfig: map[string]interface{}{
				"name":    "test",
				"private": true,
			},
			NameSource: `exports = function(){
  console.log('got heem!');
};`,
		},
	},
	Services: []ServiceStructure{
		{
			Config: map[string]interface{}{
				"name":    "http",
				"type":    "http",
				"config":  map[string]interface{}{},
				"version": float64(1),
			},
			IncomingWebhooks: []map[string]interface{}{
				{
					NameConfig: map[string]interface{}{
						"name": "find",
						"options": map[string]interface{}{
							"httpMethod":       "GET",
							"validationMethod": "NO_VALIDATION",
						},
						"run_as_authed_user":           false,
						"run_as_user_id":               "",
						"run_as_user_id_script_source": "",
						"can_evaluate":                 map[string]interface{}{},
						"respond_result":               true,
						"create_user_on_auth":          false,
						"fetch_custom_user_data":       false,
					},
					NameSource: `
exports = function({ query }) {
    const {a, b, c} = query

    const filter = {}
    if (!!a) {
      filter.a = a
    }
    if (!!b) {
      filter.b = b
    }
    if (!!c) {
      filter.c = c
    }

    return context.services
      .get('mongodb-atlas')
      .db('test')
      .collection('coll2')
      .find(filter)
};
`,
				},
			},
			Rules: []map[string]interface{}{
				{
					"name":    "access",
					"actions": []interface{}{"get", "post", "put", "delete", "patch", "head"},
					"when": map[string]interface{}{
						"%%args.url.host": map[string]interface{}{"%in": []interface{}{"*"}},
					},
				},
			},
		},
		{
			Config: map[string]interface{}{
				"type":    "mongodb",
				"name":    "mdbSvc",
				"config":  map[string]interface{}{},
				"version": float64(1),
			},
			DefaultRule: map[string]interface{}{
				"roles": []interface{}{
					map[string]interface{}{
						"name": "owner",
						"apply_when": map[string]interface{}{
							"userId": "%%user.id",
						},
						"read": true,
					},
				},
			},
			Rules: []map[string]interface{}{
				{
					"database":   "foo",
					"collection": "bar",
				},
			},
		},
	},
	Triggers: []map[string]interface{}{
		{
			"name":          "yell",
			"type":          "SCHEDULED",
			"config":        map[string]interface{}{"schedule": "0 0 * * 1"},
			"function_name": "test",
			"disabled":      false,
		},
	},
	GraphQL: GraphQLStructure{
		Config: map[string]interface{}{
			"use_natural_pluralization": true,
		},
		CustomResolvers: []map[string]interface{}{
			{
				"function_name":       "addOne",
				"on_type":             "Query",
				"field_name":          "result",
				"input_type_format":   "scalar",
				"input_type":          "number",
				"payload_type_format": "scalar",
				"payload_type":        "number",
			},
		},
	},
	Values: []map[string]interface{}{
		{
			"name":        "SECRET",
			"value":       "secret",
			"from_secret": true,
		},
		{
			"name":        "VALUE",
			"value":       "eggcorn",
			"from_secret": false,
		},
	},
	Hosting: map[string]interface{}{
		"enabled":            true,
		"app_default_domain": "full-tkdcx.stitch-statichosting-dev.baas-dev.10gen.cc",
	},
	Security: map[string]interface{}{
		"allowed_request_origins": []interface{}{
			"http://localhost:8080",
		},
	},
	CustomUserDataConfig: map[string]interface{}{
		"enabled":            true,
		"mongo_service_name": "mongodb-atlas",
		"database_name":      "test",
		"collection_name":    "coll3",
		"user_id_field":      "xref",
	},
	LogForwarders: []map[string]interface{}{
		{
			"name":         "forwarder",
			"log_types":    []interface{}{"auth"},
			"log_statuses": []interface{}{"error", "success"},
			"policy": map[string]interface{}{
				"type": "single",
			},
			"action": map[string]interface{}{
				"type": "function",
				"name": "test",
			},
			"disabled": false,
		},
	},
}}}

var allowedRequestOrigins = []string{"http://localhost:8080"}

var appSecurity = map[string]interface{}{"allowed_request_origins": []interface{}{"http://localhost:8080"}}

var appHosting = map[string]interface{}{
	"enabled":            true,
	"app_default_domain": "full-tkdcx.stitch-statichosting-dev.baas-dev.10gen.cc",
}

var appSync = map[string]interface{}{"development_mode_enabled": false}

var appCustomUserDataConfig = map[string]interface{}{
	"enabled":            true,
	"mongo_service_name": "mongodb-atlas",
	"database_name":      "test",
	"collection_name":    "coll3",
	"user_id_field":      "xref",
}

var appGraphQLStructure = GraphQLStructure{
	Config: map[string]interface{}{
		"use_natural_pluralization": true,
	},
}

var appData20180301Local = AppDataV1{AppStructureV1{
	ConfigVersion:        realm.AppConfigVersion20180301,
	Name:                 "20180301-local",
	Location:             realm.LocationVirginia,
	DeploymentModel:      realm.DeploymentModelGlobal,
	Security:             appSecurity,
	Hosting:              appHosting,
	CustomUserDataConfig: appCustomUserDataConfig,
	Sync:                 appSync,
}}

var appData20180301Remote = AppDataV1{AppStructureV1{
	ConfigVersion:        realm.AppConfigVersion20180301,
	ID:                   "20180301-remote-abcde",
	Name:                 "20180301-remote",
	Location:             realm.LocationVirginia,
	DeploymentModel:      realm.DeploymentModelGlobal,
	Security:             appSecurity,
	Hosting:              appHosting,
	CustomUserDataConfig: appCustomUserDataConfig,
	Sync:                 appSync,
}}

var appData20180301Nested = AppDataV1{AppStructureV1{
	ConfigVersion:        realm.AppConfigVersion20180301,
	Name:                 "20180301-nested",
	Location:             realm.LocationVirginia,
	DeploymentModel:      realm.DeploymentModelGlobal,
	Security:             appSecurity,
	Hosting:              appHosting,
	CustomUserDataConfig: appCustomUserDataConfig,
	Sync:                 appSync,
	GraphQL:              appGraphQLStructure,
}}

var appMeta20180301 = AppMeta{
	GroupID:       "groupID",
	AppID:         "appID",
	ConfigVersion: realm.AppConfigVersion20180301,
}

var appData20200603Local = AppDataV1{AppStructureV1{
	ConfigVersion:        realm.AppConfigVersion20200603,
	Name:                 "20200603-local",
	Location:             realm.LocationVirginia,
	DeploymentModel:      realm.DeploymentModelGlobal,
	Security:             appSecurity,
	Hosting:              appHosting,
	CustomUserDataConfig: appCustomUserDataConfig,
	Sync:                 appSync,
}}

var appData20200603Remote = AppDataV1{AppStructureV1{
	ConfigVersion:        realm.AppConfigVersion20200603,
	ID:                   "20200603-remote-abcde",
	Name:                 "20200603-remote",
	Location:             realm.LocationVirginia,
	DeploymentModel:      realm.DeploymentModelGlobal,
	Security:             appSecurity,
	Hosting:              appHosting,
	CustomUserDataConfig: appCustomUserDataConfig,
	Sync:                 appSync,
}}

var appData20200603Nested = AppDataV1{AppStructureV1{
	ConfigVersion:        realm.AppConfigVersion20200603,
	Name:                 "20200603-nested",
	Location:             realm.LocationVirginia,
	DeploymentModel:      realm.DeploymentModelGlobal,
	Security:             appSecurity,
	Hosting:              appHosting,
	CustomUserDataConfig: appCustomUserDataConfig,
	Sync:                 appSync,
	GraphQL:              appGraphQLStructure,
}}

var appMeta20200603 = AppMeta{
	GroupID:       "groupID",
	AppID:         "appID",
	ConfigVersion: realm.AppConfigVersion20200603,
}

var appData20210101Local = AppDataV2{AppStructureV2{
	ConfigVersion:         realm.AppConfigVersion20210101,
	Name:                  "20210101-local",
	Location:              realm.LocationVirginia,
	DeploymentModel:       realm.DeploymentModelGlobal,
	AllowedRequestOrigins: allowedRequestOrigins,
}}

var appData20210101Remote = AppDataV2{AppStructureV2{
	ConfigVersion:         realm.AppConfigVersion20210101,
	ID:                    "20210101-remote-abcde",
	Name:                  "20210101-remote",
	Location:              realm.LocationVirginia,
	DeploymentModel:       realm.DeploymentModelGlobal,
	AllowedRequestOrigins: allowedRequestOrigins,
}}

var appData20210101Nested = AppDataV2{AppStructureV2{
	ConfigVersion:         realm.AppConfigVersion20210101,
	Name:                  "20210101-nested",
	Location:              realm.LocationVirginia,
	DeploymentModel:       realm.DeploymentModelGlobal,
	AllowedRequestOrigins: allowedRequestOrigins,
	GraphQL:               appGraphQLStructure,
}}

var appMeta20210101 = AppMeta{
	GroupID:       "groupID",
	AppID:         "appID",
	ConfigVersion: realm.AppConfigVersion20210101,
}
