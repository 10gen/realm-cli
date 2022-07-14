package local

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/10gen/realm-cli/internal/cli/user"

	"gopkg.in/yaml.v2"
)

// Environment contains the name, full filepath, and credentials of a profile
type Environment struct {
	Name        string
	Filepath    string
	Credentials Credentials
}

// Credentials are the apikeys associated with the profiles
type Credentials struct {
	PublicAPIKey  string `yaml:"public_api_key"`
	PrivateAPIKey string `yaml:"private_api_key"`
}

// LoadEnvironments returns a list of each profile's environment containing name, filepath, and api keys
func LoadEnvironments() ([]Environment, error) {
	dir, err := user.HomeDir()
	if err != nil {
		return nil, err
	}

	dirEntryList, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	environments := make([]Environment, 0, len(dirEntryList))
	for _, v := range dirEntryList {
		if v.IsDir() {
			continue
		}
		filename := v.Name()
		ext := filepath.Ext(filename)
		if ext != "."+user.ProfileType {
			continue
		}
		name := filename[0 : len(filename)-len(ext)]

		data, err := ioutil.ReadFile(filepath.Join(dir, filename))
		if err != nil {
			return nil, fmt.Errorf("failed to read profile: %s", name)
		}

		var profile map[string]Credentials
		if err := yaml.Unmarshal(data, &profile); err != nil {
			return nil, fmt.Errorf("failed to read profile: %s", name)
		}

		environments = append(environments, Environment{
			Name:        name,
			Filepath:    filepath.Join(dir, v.Name()),
			Credentials: profile[name],
		})
	}
	return environments, nil
}
