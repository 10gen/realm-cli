package cli

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/so"
)

func TestConfigStoreEmpty(t *testing.T) {
}

// NOTE: this test relies on a tmp dir to write and read files from
// as such, the sub-tests here introduce side-effects as they execute
func TestConfigStore(t *testing.T) {
	dir, rmDir, dirErr := u.NewTempDir("home")
	so.So(t, dirErr, so.ShouldBeNil)
	defer rmDir()

	path := filepath.Join(dir, "realm-cli", "realm-cli")

	configStore := NewFileSystemConfigStore(path)

	validConfig := Config{
		PublicAPIKey:  "public-api-key",
		PrivateAPIKey: "private-api-key",
		AccessToken:   "access-token",
		RefreshToken:  "refresh-token",
	}

	t.Run("Should read a non-existent file without error and return a zero-value config", func(t *testing.T) {
		_, pathErr := os.Stat(path)
		so.So(t, os.IsNotExist(pathErr), so.ShouldBeTrue)

		config, err := configStore.ReadConfig()
		so.So(t, err, so.ShouldBeNil)

		so.So(t, config.PublicAPIKey, so.ShouldBeZeroValue)
		so.So(t, config.PrivateAPIKey, so.ShouldBeZeroValue)
		so.So(t, config.RefreshToken, so.ShouldBeZeroValue)
		so.So(t, config.AccessToken, so.ShouldBeZeroValue)
	})

	t.Run("Should write to a non-existent file the config contents", func(t *testing.T) {
		_, pathErr := os.Stat(path)
		so.So(t, os.IsNotExist(pathErr), so.ShouldBeTrue)

		so.So(t, configStore.WriteConfig(validConfig), so.ShouldBeNil)

		_, pathErr = os.Stat(path)
		so.So(t, os.IsNotExist(pathErr), so.ShouldBeFalse)
	})

	t.Run("Should read a valid config successfully", func(t *testing.T) {
		config, err := configStore.ReadConfig()
		so.So(t, err, so.ShouldBeNil)
		so.So(t, config, so.ShouldResemble, validConfig)
	})

	t.Run("Should clear a config successfully and leave the file remaining", func(t *testing.T) {
		so.So(t, configStore.ClearConfig(), so.ShouldBeNil)

		config, err := configStore.ReadConfig()
		so.So(t, err, so.ShouldBeNil)

		so.So(t, config.PublicAPIKey, so.ShouldBeZeroValue)
		so.So(t, config.PrivateAPIKey, so.ShouldBeZeroValue)
		so.So(t, config.RefreshToken, so.ShouldBeZeroValue)
		so.So(t, config.AccessToken, so.ShouldBeZeroValue)

		_, pathErr := os.Stat(path)
		so.So(t, os.IsNotExist(pathErr), so.ShouldBeFalse)
	})
}

func TestConfigStoreFailures(t *testing.T) {
	dir, rmDir, dirErr := u.NewTempDir("realm-cli")
	so.So(t, dirErr, so.ShouldBeNil)
	defer rmDir()

	path := filepath.Join(dir, "realm-cli")

	dirConfigStore := NewFileSystemConfigStore(dir)
	fileConfigStore := NewFileSystemConfigStore(path)

	t.Run("Should fail to read invalid file", func(t *testing.T) {

		_, err := dirConfigStore.ReadConfig()
		so.So(t, err, so.ShouldNotBeNil)
		so.So(t, err.Error(), so.ShouldEqual, "failed to read CLI config: is a directory")
	})

	t.Run("Should fail to read invalid yaml", func(t *testing.T) {
		badYaml := []byte(`public-api-key - my-key
`)
		so.So(t, ioutil.WriteFile(path, badYaml, 0600), so.ShouldBeNil)
		defer func() { os.Remove(path) }()

		_, err := fileConfigStore.ReadConfig()
		so.So(t, err, so.ShouldNotBeNil)
		so.So(t, err.Error(), so.ShouldEqual, "config is invalid yaml")
	})

	t.Run("Should fail to write invalid file", func(t *testing.T) {
		err := dirConfigStore.WriteConfig(Config{})
		so.So(t, err, so.ShouldNotBeNil)
		so.So(t, err.Error(), so.ShouldEqual, "failed to write CLI config: is a directory")
	})
}
