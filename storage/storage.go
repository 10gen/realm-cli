package storage

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/10gen/stitch-cli/user"

	"gopkg.in/yaml.v2"
)

// New returns a new Storage given a Strategy
func New(strategy Strategy) *Storage {
	return &Storage{
		strategy: strategy,
	}
}

// Storage represents something that can write user data to some form of Storage
type Storage struct {
	strategy Strategy
}

// WriteUserConfig writes the user data to Storage
func (s *Storage) WriteUserConfig(u *user.User) error {
	// TODO remove after personal API key support has been fully removed
	if u.PublicAPIKey != "" {
		u.Username = ""
	}

	if u.PrivateAPIKey != "" {
		u.APIKey = ""
	}

	raw, err := yaml.Marshal(u)
	if err != nil {
		return err
	}

	return s.strategy.Write(raw)
}

// ReadUserConfig reads the user data from Storage
func (s *Storage) ReadUserConfig() (*user.User, error) {
	b, err := s.strategy.Read()
	if err != nil {
		return nil, err
	}

	var user user.User
	if err := yaml.Unmarshal(b, &user); err != nil {
		return nil, err
	}

	// TODO remove after personal API key support has been fully removed
	if user.Username != "" && user.PublicAPIKey == "" {
		user.PublicAPIKey = user.Username
	}

	if user.APIKey != "" && user.PrivateAPIKey == "" {
		user.PrivateAPIKey = user.APIKey
	}

	return &user, nil
}

// Clear clears out a user's data from Storage
func (s *Storage) Clear() error {
	return s.WriteUserConfig(&user.User{})
}

// FileStrategy is a Storage that reads/persists data to/from a file at the provided path
type FileStrategy struct {
	path string
}

// Read reads data from the file at the provided path
func (fs *FileStrategy) Read() ([]byte, error) {
	if _, err := os.Stat(fs.path); os.IsNotExist(err) {
		return []byte{}, nil
	}

	return ioutil.ReadFile(fs.path)
}

// Write writes data to the file at the provided path
func (fs *FileStrategy) Write(data []byte) error {
	if err := os.MkdirAll(filepath.Dir(fs.path), 0700); err != nil {
		return err
	}

	return ioutil.WriteFile(fs.path, data, 0600)
}

// NewFileStrategy returns a new FileStrategy given a location on disk to store data
func NewFileStrategy(path string) (Strategy, error) {
	return &FileStrategy{
		path: path,
	}, nil
}

// Strategy represents a means of reading and writing data
type Strategy interface {
	Read() ([]byte, error)
	Write(data []byte) error
}
