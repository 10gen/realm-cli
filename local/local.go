package local

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/10gen/stitch-cli/app"
)

var Config string

const stitchConfig = "stitch.toml"

var ErrConfigNotFound = errors.New("stitch config not found.")

func GetApp() (a app.App, ok bool) {
	path, err := findStitchConfig()
	if err != nil {
		return
	}
	payload, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	a, err = app.Import(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "stitch: failed to parse config: %s", err)
		return
	}
	ok = true
	return
}

func findStitchConfig() (path string, err error) {
	if Config != "" {
		if _, err = os.Stat(Config); err != nil {
			err = errors.New("stitch: specified local config does not exist.")
			fmt.Fprintf(os.Stderr, "%s", err)
			return
		}
		path = Config
		return
	}

	var wd string
	wd, err = os.Getwd()
	if err != nil {
		return
	}

	for {
		path = filepath.Join(wd, stitchConfig)
		if _, err = os.Stat(path); err == nil {
			return
		}

		if wd == "/" {
			break
		}
		wd = filepath.Clean(filepath.Join(wd, ".."))
	}

	return "", ErrConfigNotFound
}
