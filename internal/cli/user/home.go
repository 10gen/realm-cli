package user

import (
	"fmt"

	"github.com/mitchellh/go-homedir"
)

const (
	servicePath = ".config/realm-cli"
)

// HomeDir returns the CLI home directory
func HomeDir() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", home, servicePath), nil
}
