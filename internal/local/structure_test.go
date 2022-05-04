package local

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestWriteSecrets(t *testing.T) {
	tmpDir, cleanupTmpDir, err := u.NewTempDir("")
	assert.Nil(t, err)
	defer cleanupTmpDir()

	t.Run("should write secrets to disk", func(t *testing.T) {
		data := SecretsStructure{
			AuthProviders: map[string]map[string]string{
				"provider": {"name": "super-secret", "value": "super-secret-value"},
			},
			Services: map[string]map[string]string{
				"svc": {"name": "super-secret", "value": "super-secret-value"},
			},
		}

		err := writeSecrets(tmpDir, data)
		assert.Nil(t, err)

		secrets, err := ioutil.ReadFile(filepath.Join(tmpDir, FileSecrets.String()))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "auth_providers": {
        "provider": {
            "name": "super-secret",
            "value": "super-secret-value"
        }
    },
    "services": {
        "svc": {
            "name": "super-secret",
            "value": "super-secret-value"
        }
    }
}
`, string(secrets))
	})
}

func TestWriteEnvironments(t *testing.T) {
	tmpDir, cleanupTmpDir, err := u.NewTempDir("")
	assert.Nil(t, err)
	defer cleanupTmpDir()

	t.Run("should write environments to disk", func(t *testing.T) {
		data := map[string]map[string]interface{}{
			"development.json": {
				"values": map[string]interface{}{
					"greeting": "hello",
				},
			},
			"no-environment.json": {
				"values": map[string]interface{}{
					"greeting": "hello",
				},
			},
			"production.json": {
				"values": map[string]interface{}{
					"greeting": "hello",
				},
			},
			"qa.json": {
				"values": map[string]interface{}{
					"greeting": "hello",
				},
			},
			"testing.json": {
				"values": map[string]interface{}{
					"greeting": "hello",
				},
			},
		}

		err := writeEnvironments(tmpDir, data)
		assert.Nil(t, err)

		for _, name := range []string{"development", "no-environment", "production", "qa", "testing"} {
			environment, err := ioutil.ReadFile(filepath.Join(tmpDir, NameEnvironments, name+extJSON))
			assert.Nil(t, err)
			assert.Equal(t, `{
    "values": {
        "greeting": "hello"
    }
}
`, string(environment))
		}
	})
}

func TestWriteValues(t *testing.T) {
	tmpDir, cleanupTmpDir, err := u.NewTempDir("")
	assert.Nil(t, err)
	defer cleanupTmpDir()

	t.Run("should write values to disk", func(t *testing.T) {
		data := []map[string]interface{}{
			{
				"name":        "key",
				"value":       "value",
				"from_secret": false,
			},
			{
				"name":        "super-secret",
				"value":       "super-secret-value",
				"from_secret": true,
			},
		}

		err := writeValues(tmpDir, data)
		assert.Nil(t, err)

		key, err := ioutil.ReadFile(filepath.Join(tmpDir, NameValues, "key"+extJSON))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "from_secret": false,
    "name": "key",
    "value": "value"
}
`, string(key))

		superSecret, err := ioutil.ReadFile(filepath.Join(tmpDir, NameValues, "super-secret"+extJSON))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "from_secret": true,
    "name": "super-secret",
    "value": "super-secret-value"
}
`, string(superSecret))
	})
}

func TestWriteGraphQL(t *testing.T) {
	tmpDir, cleanupTmpDir, err := u.NewTempDir("")
	assert.Nil(t, err)
	defer cleanupTmpDir()

	t.Run("should write graphql to disk", func(t *testing.T) {
		data := GraphQLStructure{
			Config: map[string]interface{}{
				"use_natural_pluralization": true,
			},
			CustomResolvers: []map[string]interface{}{
				{
					"field_name":          "data",
					"function_name":       "addOne",
					"input_type":          "number",
					"input_type_format":   "scalar",
					"on_type":             "Query",
					"payload_type":        "number",
					"payload_type_format": "scalar",
				},
			},
		}

		err := writeGraphQL(tmpDir, data)
		assert.Nil(t, err)

		config, err := ioutil.ReadFile(filepath.Join(tmpDir, NameGraphQL, FileConfig.String()))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "use_natural_pluralization": true
}
`, string(config))

		resolver, err := ioutil.ReadFile(filepath.Join(tmpDir, NameGraphQL, NameCustomResolvers, "query_data"+extJSON))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "field_name": "data",
    "function_name": "addOne",
    "input_type": "number",
    "input_type_format": "scalar",
    "on_type": "Query",
    "payload_type": "number",
    "payload_type_format": "scalar"
}
`, string(resolver))
	})
}

func TestWriteServices(t *testing.T) {
	tmpDir, cleanupTmpDir, err := u.NewTempDir("")
	assert.Nil(t, err)
	defer cleanupTmpDir()

	t.Run("should write services to disk", func(t *testing.T) {
		data := []ServiceStructure{{
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
			Rules: []map[string]interface{}{
				{
					"name": "access",
					"actions": []string{
						"get",
						"post",
						"put",
						"delete",
						"patch",
						"head",
					},
					"when": map[string]interface{}{
						"%%args.url.host": map[string]interface{}{
							"%in": []string{
								"*",
							},
						},
					},
				},
			},
		}}

		err := writeServices(tmpDir, data)
		assert.Nil(t, err)

		config, err := ioutil.ReadFile(filepath.Join(tmpDir, NameServices, "http", FileConfig.String()))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "config": {},
    "name": "http",
    "type": "http",
    "version": 1
}
`, string(config))

		webhook, err := ioutil.ReadFile(filepath.Join(tmpDir, NameServices, "http", NameIncomingWebhooks, "find", FileConfig.String()))
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

		src, err := ioutil.ReadFile(filepath.Join(tmpDir, NameServices, "http", NameIncomingWebhooks, "find", FileSource.String()))
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

		rule, err := ioutil.ReadFile(filepath.Join(tmpDir, NameServices, "http", NameRules, "access"+extJSON))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "actions": [
        "get",
        "post",
        "put",
        "delete",
        "patch",
        "head"
    ],
    "name": "access",
    "when": {
        "%%args.url.host": {
            "%in": [
                "*"
            ]
        }
    }
}
`, string(rule))

		_, err = ioutil.ReadFile(filepath.Join(tmpDir, NameServices, "http", FileDefaultRule.String()))
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "no such file or directory"),
			fmt.Sprintf("expected 'no such file or directory' in error message but got '%s'", err.Error()))
	})

	t.Run("should write services with v2 Rules to disk", func(t *testing.T) {
		mdbSvcName := "mdbSvc"
		data := []ServiceStructure{{
			Config: map[string]interface{}{
				"name":    mdbSvcName,
				"type":    "mongodb",
				"config":  map[string]interface{}{},
				"version": 1,
			},
			Rules: []map[string]interface{}{
				{
					"database":   "foo",
					"collection": "bar",
				},
			},
		}}

		err := writeServices(tmpDir, data)
		assert.Nil(t, err)

		config, err := ioutil.ReadFile(filepath.Join(tmpDir, NameServices, mdbSvcName, FileConfig.String()))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "config": {},
    "name": "mdbSvc",
    "type": "mongodb",
    "version": 1
}
`, string(config))

		_, err = ioutil.ReadFile(filepath.Join(tmpDir, NameServices, mdbSvcName, FileDefaultRule.String()))
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "no such file or directory"),
			fmt.Sprintf("expected 'no such file or directory' in error message but got '%s'", err.Error()))

		rule, err := ioutil.ReadFile(filepath.Join(tmpDir, NameServices, mdbSvcName, NameRules, "foo.bar"+extJSON))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "collection": "bar",
    "database": "foo"
}
`, string(rule))
	})

	t.Run("should write services with v2 Rules and a default rule to disk", func(t *testing.T) {
		mdbSvcName := "mdbSvc"
		data := []ServiceStructure{{
			Config: map[string]interface{}{
				"name":    mdbSvcName,
				"type":    "mongodb",
				"config":  map[string]interface{}{},
				"version": 1,
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

		err := writeServices(tmpDir, data)
		assert.Nil(t, err)

		config, err := ioutil.ReadFile(filepath.Join(tmpDir, NameServices, mdbSvcName, FileConfig.String()))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "config": {},
    "name": "mdbSvc",
    "type": "mongodb",
    "version": 1
}
`, string(config))

		defaultRule, err := ioutil.ReadFile(filepath.Join(tmpDir, NameServices, mdbSvcName, FileDefaultRule.String()))
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

		rule, err := ioutil.ReadFile(filepath.Join(tmpDir, NameServices, mdbSvcName, NameRules, "foo.bar"+extJSON))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "collection": "bar",
    "database": "foo"
}
`, string(rule))
	})
}

func TestWriteTriggers(t *testing.T) {
	tmpDir, cleanupTmpDir, err := u.NewTempDir("")
	assert.Nil(t, err)
	defer cleanupTmpDir()

	t.Run("should write values to disk", func(t *testing.T) {
		data := []map[string]interface{}{
			{
				"name":          "yell",
				"type":          "SCHEDULED",
				"config":        map[string]interface{}{"schedule": "0 0 * * 1"},
				"function_name": "test",
				"disabled":      false,
			},
		}

		err := writeTriggers(tmpDir, data)
		assert.Nil(t, err)

		key, err := ioutil.ReadFile(filepath.Join(tmpDir, NameTriggers, "yell"+extJSON))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "config": {
        "schedule": "0 0 * * 1"
    },
    "disabled": false,
    "function_name": "test",
    "name": "yell",
    "type": "SCHEDULED"
}
`, string(key))
	})
}

func TestWriteLogForwarders(t *testing.T) {
	tmpDir, cleanupTmpDir, err := u.NewTempDir("")
	assert.Nil(t, err)
	defer cleanupTmpDir()

	t.Run("should write log forwarders to disk", func(t *testing.T) {
		data := []map[string]interface{}{
			{
				"name": "lf1",
				"log_types": []interface{}{
					"auth",
					"function",
				},
				"log_statuses": []interface{}{
					"error",
				},
				"policy": map[string]interface{}{
					"type": "single",
				},
				"action": map[string]interface{}{
					"type": "function",
					"name": "function0",
				},
				"disabled": false,
			},
			{
				"name": "lf2",
				"log_types": []interface{}{
					"auth",
					"graphql",
				},
				"log_statuses": []interface{}{
					"error",
					"success",
				},
				"policy": map[string]interface{}{
					"type": "batch",
				},
				"action": map[string]interface{}{
					"type": "function",
					"name": "function0",
				},
				"disabled": false,
			},
		}

		err := writeLogForwarders(tmpDir, data)
		assert.Nil(t, err)

		lf1, err := ioutil.ReadFile(filepath.Join(tmpDir, NameLogForwarders, "lf1"+extJSON))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "action": {
        "name": "function0",
        "type": "function"
    },
    "disabled": false,
    "log_statuses": [
        "error"
    ],
    "log_types": [
        "auth",
        "function"
    ],
    "name": "lf1",
    "policy": {
        "type": "single"
    }
}
`, string(lf1))

		lf2, err := ioutil.ReadFile(filepath.Join(tmpDir, NameLogForwarders, "lf2"+extJSON))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "action": {
        "name": "function0",
        "type": "function"
    },
    "disabled": false,
    "log_statuses": [
        "error",
        "success"
    ],
    "log_types": [
        "auth",
        "graphql"
    ],
    "name": "lf2",
    "policy": {
        "type": "batch"
    }
}
`, string(lf2))
	})
}

func TestWriteEndpoints(t *testing.T) {
	tmpDir, cleanupTmpDir, err := u.NewTempDir("")
	assert.Nil(t, err)
	defer cleanupTmpDir()

	t.Run("should write endpoints to disk", func(t *testing.T) {
		data := EndpointStructure{
			Configs: []map[string]interface{}{
				{
					"create_user_on_auth":    true,
					"disabled":               true,
					"fetch_custom_user_data": true,
					"function_name":          "test",
					"http_method":            "GET",
					"respond_result":         true,
					"route":                  "/hello/world",
					"secret_name":            "super_secret",
					"validation_method":      "VERIFY_PAYLOAD",
				},
				{
					"function_name":     "test",
					"http_method":       "POST",
					"route":             "/hello/world",
					"validation_method": "NO_VALIDATION",
				},
			},
		}

		err := writeEndpoints(tmpDir, data)
		assert.Nil(t, err)

		config, err := ioutil.ReadFile(filepath.Join(tmpDir, NameHTTPEndpoints, FileConfig.String()))
		assert.Nil(t, err)
		assert.Equal(t, `[
    {
        "create_user_on_auth": true,
        "disabled": true,
        "fetch_custom_user_data": true,
        "function_name": "test",
        "http_method": "GET",
        "respond_result": true,
        "route": "/hello/world",
        "secret_name": "super_secret",
        "validation_method": "VERIFY_PAYLOAD"
    },
    {
        "function_name": "test",
        "http_method": "POST",
        "route": "/hello/world",
        "validation_method": "NO_VALIDATION"
    }
]
`, string(config))
	})
}
