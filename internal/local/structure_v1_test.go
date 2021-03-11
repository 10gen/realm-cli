package local

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestWriteFunctionsV1(t *testing.T) {
	tmpDir, cleanupTmpDir, err := u.NewTempDir("")
	assert.Nil(t, err)
	defer cleanupTmpDir()

	t.Run("should write functions to disk", func(t *testing.T) {
		data := []map[string]interface{}{
			{
				"config": map[string]interface{}{
					"name":    "test",
					"private": true,
				},
				"source": "exports = function(){\n  console.log('Hello World!');\n};",
			},
		}

		err := writeFunctionsV1(tmpDir, data)
		assert.Nil(t, err)

		key, err := ioutil.ReadFile(filepath.Join(tmpDir, NameFunctions, "test", FileConfig.String()))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "name": "test",
    "private": true
}
`, string(key))

		superSecret, err := ioutil.ReadFile(filepath.Join(tmpDir, NameFunctions, "test", FileSource.String()))
		assert.Nil(t, err)
		assert.Equal(t, `exports = function(){
  console.log('Hello World!');
};`, string(superSecret))
	})
}

func TestWriteAuthProviders(t *testing.T) {
	tmpDir, cleanupTmpDir, err := u.NewTempDir("")
	assert.Nil(t, err)
	defer cleanupTmpDir()

	t.Run("should write auth to disk", func(t *testing.T) {
		data := []map[string]interface{}{
			{
				"name":     "api-key",
				"type":     "api-key",
				"disabled": true,
			},
		}

		err := writeAuthProviders(tmpDir, data)
		assert.Nil(t, err)

		providers, err := ioutil.ReadFile(filepath.Join(tmpDir, NameAuthProviders, "api-key"+extJSON))
		assert.Nil(t, err)
		assert.Equal(t, `{
    "disabled": true,
    "name": "api-key",
    "type": "api-key"
}
`, string(providers))
	})
}
