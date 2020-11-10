package mock

import (
	"github.com/10gen/realm-cli/internal/cli"
)

// ConfigStore represents a mocked ConfigStore
type ConfigStore struct {
	cli.ConfigStore
	ClearConfigFn func() error
	ReadConfigFn  func() (cli.Config, error)
	WriteConfigFn func(config cli.Config) error
}

// ClearConfig calls the mocked ClearConfig implementation if provided,
// otherwise calls the mock's underlying store's ClearConfig implementation.
// NOTE: may panic if underlying store is not set
func (cs ConfigStore) ClearConfig() error {
	if cs.ClearConfigFn == nil {
		return cs.ConfigStore.ClearConfig()
	}
	return cs.ClearConfigFn()
}

// ReadConfig calls the mocked ReadConfig implementation if provided,
// otherwise calls the mock's underlying store's ReadConfig implementation.
// NOTE: may panic if underlying store is not set
func (cs ConfigStore) ReadConfig() (cli.Config, error) {
	if cs.ReadConfigFn == nil {
		return cs.ConfigStore.ReadConfig()
	}
	return cs.ReadConfigFn()
}

// WriteConfig calls the mocked WriteConfig implementation if provided,
// otherwise calls the mock's underlying store's WriteConfig implementation.
// NOTE: may panic if underlying store is not set
func (cs ConfigStore) WriteConfig(config cli.Config) error {
	if cs.WriteConfigFn == nil {
		return cs.ConfigStore.WriteConfig(config)
	}
	return cs.WriteConfigFn(config)
}
