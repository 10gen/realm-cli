package local

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestParseFunctionsV2(t *testing.T) {
	wd, wdErr := os.Getwd()
	assert.Nil(t, wdErr)

	testRoot := filepath.Join(wd, "testdata/functions")

	t.Run("should return the parsed functions directory with nested javascript files", func(t *testing.T) {
		functions, err := parseFunctionsV2(testRoot)
		assert.Nil(t, err)
		assert.Equal(t, FunctionsStructure{
			Configs: []map[string]interface{}{{
				"name":    "bar",
				"private": true,
			}},
			Sources: map[string]string{
				"eggcorn.js": `exports = function () {
  console.log('eggcorn');
};
`,
				"foo/bar.js": `exports = function () {
  console.log('foobar');
};
`,
			},
		}, functions)
	})
}

func TestParseDataSources(t *testing.T) {
	wd, wdErr := os.Getwd()
	assert.Nil(t, wdErr)

	testRoot := filepath.Join(wd, "testdata/data_sources")

	t.Run("should return the parsed data sources directory with nested rules and schema", func(t *testing.T) {
		dataSources, err := parseDataSources(testRoot)
		assert.Nil(t, err)
		assert.Equal(t, []DataSourceStructure{{
			Config: map[string]interface{}{
				"type":   "mongodb-atlas",
				"name":   "mongodb-atlas",
				"config": map[string]interface{}{},
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
					"database":      "foo",
					"collection":    "bar",
					"schema":        map[string]interface{}{"title": "foo.bar schema"},
					"relationships": map[string]interface{}{},
				},
				{
					"database":   "foo",
					"collection": "onlyRulesColl",
				},
				{
					"database":   "foo",
					"collection": "onlySchemasColl",
					"schema": map[string]interface{}{
						"title": "soloSchema",
						"properties": map[string]interface{}{
							"name": map[string]interface{}{
								"bsonType": "string",
							},
							"country": map[string]interface{}{
								"bsonType": "string",
							},
						},
					},
					"relationships": map[string]interface{}{
						"name": map[string]interface{}{
							"ref":         "#/relationship/mongodb-atlas/foo/onlySchemasColl",
							"foreign_key": "country",
							"is_list":     false,
						},
						"country": map[string]interface{}{
							"ref":         "#/relationship/mongodb-atlas/foo/onlySchemasColl",
							"foreign_key": "name",
							"is_list":     false,
						},
					},
				},
				{
					"database":   "test",
					"collection": "test",
					"schema":     map[string]interface{}{"title": "test.test schema"},
					"relationships": map[string]interface{}{
						"user_id": map[string]interface{}{
							"ref":         "#/relationship/mongodb-atlas/foo/bar",
							"source_key":  "user_id",
							"foreign_key": "user_id",
							"is_list":     false,
						},
					},
				},
			},
		}}, dataSources)
	})
}

func TestWriteFunctionsV2(t *testing.T) {
	tmpDir, cleanupTmpDir, err := u.NewTempDir("")
	assert.Nil(t, err)
	defer cleanupTmpDir()

	t.Run("should write functions to disk", func(t *testing.T) {
		data := FunctionsStructure{
			Configs: []map[string]interface{}{
				{
					"name":    "test",
					"private": true,
				},
			},
			Sources: map[string]string{
				"nested/test.js": "exports = function(){\n  console.log('Hello World!');\n};",
			},
		}

		err := writeFunctionsV2(tmpDir, data)
		assert.Nil(t, err)

		key, err := ioutil.ReadFile(filepath.Join(tmpDir, NameFunctions, FileConfig.String()))
		assert.Nil(t, err)
		assert.Equal(t, `[
    {
        "name": "test",
        "private": true
    }
]
`, string(key))

		superSecret, err := ioutil.ReadFile(filepath.Join(tmpDir, NameFunctions, "nested/test.js"))
		assert.Nil(t, err)
		assert.Equal(t, "exports = function(){\n  console.log('Hello World!');\n};", string(superSecret))
	})
}

func TestWriteAuth(t *testing.T) {
	tmpDir, cleanupTmpDir, err := u.NewTempDir("")
	assert.Nil(t, err)
	defer cleanupTmpDir()

	t.Run("should write auth to disk", func(t *testing.T) {
		data := AuthStructure{
			CustomUserData: map[string]interface{}{"enabled": false},
			Providers: map[string]interface{}{
				"api-key": map[string]interface{}{
					"name":     "api-key",
					"type":     "api-key",
					"disabled": true,
				},
			},
		}

		err := writeAuth(tmpDir, data)
		assert.Nil(t, err)

		providers, err := ioutil.ReadFile(filepath.Join(tmpDir, NameAuth, FileProviders.String()))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "api-key": {
        "disabled": true,
        "name": "api-key",
        "type": "api-key"
    }
}
`, string(providers))

		userData, err := ioutil.ReadFile(filepath.Join(tmpDir, NameAuth, FileCustomUserData.String()))
		assert.Nil(t, err)
		assert.Equal(t, "{\n    \"enabled\": false\n}\n", string(userData))
	})
}

func TestWriteSync(t *testing.T) {
	tmpDir, cleanupTmpDir, err := u.NewTempDir("")
	assert.Nil(t, err)
	defer cleanupTmpDir()

	t.Run("should write sync to disk", func(t *testing.T) {
		data := SyncStructure{Config: map[string]interface{}{"development_mode_enabled": false}}

		writeSync(tmpDir, data)

		sync, err := ioutil.ReadFile(filepath.Join(tmpDir, NameSync, FileConfig.String()))
		assert.Nil(t, err)
		assert.Equal(t, "{\n    \"development_mode_enabled\": false\n}\n", string(sync))
	})
}

func TestWriteDataSources(t *testing.T) {
	tmpDir, cleanupTmpDir, err := u.NewTempDir("")
	assert.Nil(t, err)
	defer cleanupTmpDir()

	mdbAtlasStr := "mongodb-atlas"

	t.Run("should write services to disk", func(t *testing.T) {
		data := []DataSourceStructure{{
			Config: map[string]interface{}{
				"name": mdbAtlasStr,
				"type": mdbAtlasStr,
				"config": map[string]interface{}{
					"clusterName":         "Cluster0",
					"wireProtocolEnabled": true,
				},
			},
			Rules: []map[string]interface{}{
				{
					"database":   "foo",
					"collection": "bar",
				},
			},
		}}

		err := writeDataSources(tmpDir, data)
		assert.Nil(t, err)

		config, err := ioutil.ReadFile(filepath.Join(tmpDir, NameDataSources, mdbAtlasStr, FileConfig.String()))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "config": {
        "clusterName": "Cluster0",
        "wireProtocolEnabled": true
    },
    "name": "mongodb-atlas",
    "type": "mongodb-atlas"
}
`, string(config))

		_, err = ioutil.ReadFile(filepath.Join(tmpDir, NameDataSources, mdbAtlasStr, FileDefaultRule.String()))
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "no such file or directory"),
			fmt.Sprintf("expected 'no such file or directory' in error message but got '%s'", err.Error()))

		rule, err := ioutil.ReadFile(filepath.Join(tmpDir, NameDataSources, mdbAtlasStr, "foo", "bar", FileRules.String()))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "collection": "bar",
    "database": "foo"
}
`, string(rule))

		// the rule in this data source doesn't contain a schema or relationships, so assert that those files are not present
		_, err = ioutil.ReadFile(filepath.Join(tmpDir, NameDataSources, mdbAtlasStr, "foo", "bar", FileSchema.String()))
		assert.NotNil(t, err)
		assert.True(t, os.IsNotExist(err), "schema.json must not exist")

		_, err = ioutil.ReadFile(filepath.Join(tmpDir, NameDataSources, mdbAtlasStr, "foo", "bar", FileRelationships.String()))
		assert.NotNil(t, err)
		assert.True(t, os.IsNotExist(err), "relationships.json must not exist")
	})

	t.Run("should write services with a default rule to disk", func(t *testing.T) {
		data := []DataSourceStructure{{
			Config: map[string]interface{}{
				"name": mdbAtlasStr,
				"type": mdbAtlasStr,
				"config": map[string]interface{}{
					"clusterName":         "Cluster0",
					"wireProtocolEnabled": true,
				},
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
		}}

		err := writeDataSources(tmpDir, data)
		assert.Nil(t, err)

		config, err := ioutil.ReadFile(filepath.Join(tmpDir, NameDataSources, mdbAtlasStr, FileConfig.String()))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "config": {
        "clusterName": "Cluster0",
        "wireProtocolEnabled": true
    },
    "name": "mongodb-atlas",
    "type": "mongodb-atlas"
}
`, string(config))

		defaultRule, err := ioutil.ReadFile(filepath.Join(tmpDir, NameDataSources, mdbAtlasStr, FileDefaultRule.String()))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "roles": [
        {
            "apply_when": {
                "userId": "%%user.id"
            },
            "name": "owner",
            "read": true
        }
    ]
}
`, string(defaultRule))

		rule, err := ioutil.ReadFile(filepath.Join(tmpDir, NameDataSources, mdbAtlasStr, "foo", "bar", FileRules.String()))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "collection": "bar",
    "database": "foo"
}
`, string(rule))

		// the rule in this data source doesn't contain a schema or relationships, so assert that those files are not present
		_, err = ioutil.ReadFile(filepath.Join(tmpDir, NameDataSources, mdbAtlasStr, "foo", "bar", FileSchema.String()))
		assert.NotNil(t, err)
		assert.True(t, os.IsNotExist(err), "schema.json must not exist")

		_, err = ioutil.ReadFile(filepath.Join(tmpDir, NameDataSources, mdbAtlasStr, "foo", "bar", FileRelationships.String()))
		assert.NotNil(t, err)
		assert.True(t, os.IsNotExist(err), "relationships.json must not exist")
	})

	t.Run("should write schemas and relationships to disk if they are included in the rule object", func(t *testing.T) {
		data := []DataSourceStructure{{
			Config: map[string]interface{}{
				"name": mdbAtlasStr,
				"type": mdbAtlasStr,
				"config": map[string]interface{}{
					"clusterName":         "Cluster0",
					"wireProtocolEnabled": true,
				},
			},
			Rules: []map[string]interface{}{
				{
					"database":   "foo",
					"collection": "bar",
					"schema": map[string]interface{}{
						"title": "foo.bar schema",
					},
					"relationships": map[string]interface{}{
						"user_id": map[string]interface{}{
							"ref":         "#/relationship/another/db/coll",
							"source_key":  "user_id",
							"foreign_key": "user_id",
							"is_list":     false,
						},
					},
				},
			},
		}}

		err := writeDataSources(tmpDir, data)
		assert.Nil(t, err)

		config, err := ioutil.ReadFile(filepath.Join(tmpDir, NameDataSources, mdbAtlasStr, FileConfig.String()))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "config": {
        "clusterName": "Cluster0",
        "wireProtocolEnabled": true
    },
    "name": "mongodb-atlas",
    "type": "mongodb-atlas"
}
`, string(config))

		rule, err := ioutil.ReadFile(filepath.Join(tmpDir, NameDataSources, mdbAtlasStr, "foo", "bar", FileRules.String()))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "collection": "bar",
    "database": "foo"
}
`, string(rule))

		schema, err := ioutil.ReadFile(filepath.Join(tmpDir, NameDataSources, mdbAtlasStr, "foo", "bar", FileSchema.String()))
		assert.Nil(t, err)
		assert.Equal(t, "{\n    \"title\": \"foo.bar schema\"\n}\n", string(schema))

		relationships, err := ioutil.ReadFile(filepath.Join(tmpDir, NameDataSources, mdbAtlasStr, "foo", "bar", FileRelationships.String()))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "user_id": {
        "foreign_key": "user_id",
        "is_list": false,
        "ref": "#/relationship/another/db/coll",
        "source_key": "user_id"
    }
}
`, string(relationships))
	})
}

func TestWriteHTTPEndpoints(t *testing.T) {
	tmpDir, cleanupTmpDir, err := u.NewTempDir("")
	assert.Nil(t, err)
	defer cleanupTmpDir()

	t.Run("should write http endpoints to disk", func(t *testing.T) {
		data := []HTTPServiceStructure{{
			Config: map[string]interface{}{
				"name":    "http",
				"type":    "http",
				"config":  map[string]interface{}{},
				"version": 1,
			},
			IncomingWebhooks: []map[string]interface{}{
				{
					"name":                         "find",
					"run_as_authed_user":           false,
					"run_as_user_id":               "",
					"run_as_user_id_script_source": "",
					"can_evaluate":                 map[string]interface{}{},
					"options": map[string]interface{}{
						"httpMethod":       "GET",
						"validationMethod": "NO_VALIDATION",
					},
					"respond_result":         true,
					"fetch_custom_user_data": false,
					"create_user_on_auth":    false,
					"source": `
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
			Rules: []map[string]interface{}{{
				"name":    "rule",
				"actions": []interface{}{"get"},
				"when": map[string]interface{}{`%%args.url.host`: map[string]interface{}{
					`%in`: []interface{}{"google.com"},
				}},
			}},
		}}

		err := writeHTTPServices(tmpDir, data)
		assert.Nil(t, err)

		config, err := ioutil.ReadFile(filepath.Join(tmpDir, NameHTTPEndpoints, "http", FileConfig.String()))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "config": {},
    "name": "http",
    "type": "http",
    "version": 1
}
`, string(config))

		webhook, err := ioutil.ReadFile(filepath.Join(tmpDir, NameHTTPEndpoints, "http", NameIncomingWebhooks, "find", FileConfig.String()))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "can_evaluate": {},
    "create_user_on_auth": false,
    "fetch_custom_user_data": false,
    "name": "find",
    "options": {
        "httpMethod": "GET",
        "validationMethod": "NO_VALIDATION"
    },
    "respond_result": true,
    "run_as_authed_user": false,
    "run_as_user_id": "",
    "run_as_user_id_script_source": ""
}
`, string(webhook))

		src, err := ioutil.ReadFile(filepath.Join(tmpDir, NameHTTPEndpoints, "http", NameIncomingWebhooks, "find", FileSource.String()))
		assert.Nil(t, err)
		assert.Equal(t, `
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
`, string(src))

		rule, err := ioutil.ReadFile(filepath.Join(tmpDir, NameHTTPEndpoints, "http", NameRules, "rule.json"))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "actions": [
        "get"
    ],
    "name": "rule",
    "when": {
        "%%args.url.host": {
            "%in": [
                "google.com"
            ]
        }
    }
}
`, string(rule))
	})
}
