package local

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestParseFunctionsV2(t *testing.T) {
	wd, wdErr := os.Getwd()
	assert.Nil(t, wdErr)

	testRoot := filepath.Join(wd, "testdata/functions")

	t.Run("should return the parsed functions directory with nested javascript files", func(t *testing.T) {
		functions, err := parseFunctionsV2(testRoot)
		assert.Nil(t, err)
		assert.Equal(t, &FunctionsStructure{
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
			Rules: []map[string]interface{}{
				{
					"database":   "foo",
					"collection": "bar",
					"schema":     map[string]interface{}{"title": "foo.bar schema"},
				},
				{
					"database":   "test",
					"collection": "test",
					"schema":     map[string]interface{}{"title": "test.test schema"},
				},
			},
		}}, dataSources)
	})
}
