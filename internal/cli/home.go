package cli

import (
	"fmt"

	"github.com/mitchellh/go-homedir"
)

const (
	// DirProfile is the CLI profile directory
	DirProfile = ".config/realm-cli"
)

// HomeDir returns the CLI home directory
func HomeDir() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", home, DirProfile), nil
}
