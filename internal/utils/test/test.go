package testutils

import (
	"io/ioutil"
	"os"
)

func NewTempDir(name string) (string, func(), error) {
	dir, err := ioutil.TempDir("", name)
	if err != nil {
		return "", nil, err
	}
	return dir, func() { os.RemoveAll(dir) }, nil
}
