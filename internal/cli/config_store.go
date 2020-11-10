package cli

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// ConfigStore represents the CLI config store
type ConfigStore interface {
	ClearConfig() error
	ReadConfig() (Config, error)
	WriteConfig(config Config) error
}

// NewFileSystemConfigStore creates a new file system config store
func NewFileSystemConfigStore(path string) ConfigStore {
	return &fileSystemConfigStore{path}
}

type fileSystemConfigStore struct {
	path string
}

func (fss *fileSystemConfigStore) ClearConfig() error {
	if err := fss.WriteConfig(Config{}); err != nil {
		return NewPrivilegedErr("failed to clear CLI config", err)
	}
	return nil
}

func (fss *fileSystemConfigStore) ReadConfig() (Config, error) {
	if _, err := os.Stat(fss.path); os.IsNotExist(err) {
		return Config{}, nil
	}
	raw, readErr := ioutil.ReadFile(fss.path)
	if readErr != nil {
		return Config{}, NewPrivilegedErr("failed to read CLI config", readErr)
	}

	var config Config
	if err := yaml.Unmarshal(raw, &config); err != nil {
		return Config{}, NewErrw("config is invalid yaml", err)
	}
	return config, nil
}

func (fss *fileSystemConfigStore) WriteConfig(config Config) error {
	raw, yamlErr := yaml.Marshal(config)
	if yamlErr != nil {
		return NewErrw("config is invalid yaml", yamlErr)
	}
	if err := os.MkdirAll(filepath.Dir(fss.path), 0700); err != nil {
		return NewPrivilegedErr("failed to write CLI config", err)
	}
	if err := ioutil.WriteFile(fss.path, raw, 0600); err != nil {
		return NewPrivilegedErr("failed to write CLI config", err)
	}
	return nil
}
