package app

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestReadPackage(t *testing.T) {
	wd, wdErr := os.Getwd()
	assert.Nil(t, wdErr)

	testRoot := wd
	projectRoot := filepath.Join(testRoot, "testdata", "full_project")

	pkg, err := ReadPackage(projectRoot)
	assert.Nil(t, err)
	for field := range fullPkg {
		assert.Equalf(t,
			fullPkg[field],
			pkg[field],
			"%q must be equal\n  got:    %v (%T)\n  wanted: %v (%T)",
			field, pkg[field], pkg[field], fullPkg[field], fullPkg[field],
		)
	}
}

func TestWriteDefaultConfig(t *testing.T) {
	t.Run("Should write the app config contents successfully", func(t *testing.T) {
		for _, tc := range []struct {
			description string
			config      Config
			contents    string
		}{
			{
				description: "With zero contents",
				contents: `{
    "config_version": 0,
    "name": "",
    "location": "",
    "deployment_model": "",
    "security": {},
    "custom_user_data_config": {
        "enabled": false
    },
    "sync": {
        "development_mode_enabled": false
    }
}`,
			},
			{
				description: "With a full 20180301 config",
				config: Config{
					ConfigVersion:        realm.AppConfigVersion20180301,
					ID:                   "test-abcde",
					Name:                 "test",
					Location:             realm.LocationVirginia,
					DeploymentModel:      realm.DeploymentModelGlobal,
					Security:             SecurityConfig{[]string{"http://localhost:8080"}},
					CustomUserDataConfig: CustomUserDataConfig{true},
					Sync:                 SyncConfig{true},
				},
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
}`,
			},
			{
				description: "With a full 20200603 config",
				config: Config{
					ConfigVersion:        realm.AppConfigVersion20200603,
					ID:                   "test-abcde",
					Name:                 "test",
					Location:             realm.LocationVirginia,
					DeploymentModel:      realm.DeploymentModelGlobal,
					Security:             SecurityConfig{[]string{"http://localhost:8080"}},
					CustomUserDataConfig: CustomUserDataConfig{true},
					Sync:                 SyncConfig{true},
				},
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
}`,
			},
			{
				description: "With a full 20180301 config",
				config: Config{
					ConfigVersion:        realm.AppConfigVersion20180301,
					ID:                   "test-abcde",
					Name:                 "test",
					Location:             realm.LocationVirginia,
					DeploymentModel:      realm.DeploymentModelGlobal,
					Security:             SecurityConfig{[]string{"http://localhost:8080"}},
					CustomUserDataConfig: CustomUserDataConfig{true},
					Sync:                 SyncConfig{true},
				},
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
}`,
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				tmpDir, cleanupTmpDir, tmpDirErr := u.NewTempDir("")
				assert.Nil(t, tmpDirErr)
				defer cleanupTmpDir()

				assert.Nil(t, WriteDefaultConfig(tmpDir, tc.config))

				data, dataErr := ioutil.ReadFile(filepath.Join(tmpDir, FileConfig.String()))
				assert.Nil(t, dataErr)
				assert.Equal(t, tc.contents, string(data))
			})
		}
	})
}

func TestResolveData(t *testing.T) {
	wd, wdErr := os.Getwd()
	assert.Nil(t, wdErr)

	testRoot := wd
	projectRoot := filepath.Join(testRoot, "testdata", "project")

	projectConfig := Config{
		ConfigVersion:   realm.AppConfigVersion20210101,
		ID:              "eggcorn-abcde",
		Name:            "eggcorn",
		Location:        realm.LocationVirginia,
		DeploymentModel: realm.DeploymentModelGlobal,
		Security:        SecurityConfig{},
	}

	t.Run("With a working directory outside of the root of a project directory", func(t *testing.T) {
		t.Run("Resolving the app directory should return an empty string", func(t *testing.T) {
			path, insideProject, err := ResolveDirectory(testRoot)
			assert.Nil(t, err)
			assert.False(t, insideProject, "expected to be outside project")
			assert.Equal(t, Directory{}, path)
		})

		t.Run("Resolving the app data should successfully return empty data", func(t *testing.T) {
			appDir, config, err := ResolveConfig(testRoot)
			assert.Nil(t, err)
			assert.Equal(t, "", appDir)
			assert.Equal(t, Config{}, config)
		})
	})

	t.Run("With a working directory at the root of a project directory", func(t *testing.T) {
		t.Run("Resolving the app directory should return the working directory", func(t *testing.T) {
			path, insideProject, err := ResolveDirectory(projectRoot)
			assert.Nil(t, err)
			assert.True(t, insideProject, "expected to be inside project")
			assert.Equal(t, Directory{Path: projectRoot, ConfigVersion: realm.AppConfigVersion20200603}, path)
		})

		t.Run("Resolving the app data should successfully return project data", func(t *testing.T) {
			appDir, config, err := ResolveConfig(projectRoot)
			assert.Nil(t, err)
			assert.Equal(t, projectRoot, appDir)
			assert.Equal(t, projectConfig, config)
		})
	})

	t.Run("With a working directory nested deeply inside a project directory", func(t *testing.T) {
		nestedRoot := filepath.Join(projectRoot, "l1", "l2", "l3")

		t.Run("Resolving the app directory should return the working directory", func(t *testing.T) {
			path, insideProject, err := ResolveDirectory(nestedRoot)
			assert.Nil(t, err)
			assert.True(t, insideProject, "expected to be inside project")
			assert.Equal(t, Directory{Path: projectRoot, ConfigVersion: realm.AppConfigVersion20200603}, path)
		})

		t.Run("Resolving the app data should successfully return project data", func(t *testing.T) {
			appDir, config, err := ResolveConfig(nestedRoot)
			assert.Nil(t, err)
			assert.Equal(t, projectConfig, config)
			assert.Equal(t, projectRoot, appDir)
		})

		t.Run("Resolving the app data should return empty data if it exceeds the max search depth", func(t *testing.T) {
			superNestedRoot := filepath.Join(nestedRoot, "l4", "l5", "l6", "l7", "l8", "l9")

			appDir, config, err := ResolveConfig(superNestedRoot)
			assert.Nil(t, err)
			assert.Equal(t, "", appDir)
			assert.Equal(t, Config{}, config)
		})
	})

	t.Run("Resolving the app data should fail when a project has an empty configuration", func(t *testing.T) {
		emptyProjectRoot := filepath.Join(testRoot, "testdata", "empty_project")

		expectedErr := fmt.Errorf(
			"no file contents at %s",
			filepath.Join(emptyProjectRoot, FileConfig.String()),
		)

		_, _, err := ResolveConfig(filepath.Join(emptyProjectRoot, "l1", "l2", "l3"))
		assert.Equal(t, expectedErr, err)
	})
}

var fullPkg = map[string]interface{}{
	FieldConfigVersion:   float64(20200603),
	"app_id":             "full-abcde",
	FieldName:            "full",
	FieldLocation:        "US-VA",
	FieldDeploymentModel: "GLOBAL",
	NameAuthProviders: []map[string]interface{}{
		{
			"name":     "api-key",
			"type":     "api-key",
			"disabled": bool(false),
		},
	},
	NameFunctions: []map[string]interface{}{
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
	NameServices: []map[string]interface{}{
		{
			NameConfig: map[string]interface{}{
				"name":    "http",
				"type":    "http",
				"config":  map[string]interface{}{},
				"version": float64(1),
			},
			NameIncomingWebhooks: []map[string]interface{}{
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
			NameRules: []map[string]interface{}{
				{
					"name":    "access",
					"actions": []interface{}{"get", "post", "put", "delete", "patch", "head"},
					"when": map[string]interface{}{
						"%%args.url.host": map[string]interface{}{"%in": []interface{}{"*"}},
					},
				},
			},
		},
	},
	NameTriggers: []map[string]interface{}{
		{
			"name":          "yell",
			"type":          "SCHEDULED",
			"config":        map[string]interface{}{"schedule": "0 0 * * 1"},
			"function_name": "test",
			"disabled":      false,
		},
	},
	NameGraphQL: map[string]interface{}{
		NameConfig: map[string]interface{}{
			"use_natural_pluralization": true,
		},
		NameCustomResolvers: []map[string]interface{}{},
	},
	NameValues: []map[string]interface{}{
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
	"hosting": map[string]interface{}{
		"enabled":            bool(true),
		"app_default_domain": "full-tkdcx.stitch-statichosting-dev.baas-dev.10gen.cc",
	},
	"security": map[string]interface{}{
		"allowed_request_origins": []interface{}{
			"http://localhost:8080",
		},
	},
	"custom_user_data_config": map[string]interface{}{
		"enabled":            bool(true),
		"mongo_service_name": "mongodb-atlas",
		"database_name":      "test",
		"collection_name":    "coll3",
		"user_id_field":      "xref",
	},
}
