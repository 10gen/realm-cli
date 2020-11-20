package cli

import (
	"fmt"

	"github.com/mitchellh/go-homedir"
)

const (
	profileDir = ".config/realm-cli"
)

func homeDir() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", home, profileDir), nil
}
