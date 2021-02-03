package local

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/10gen/realm-cli/internal/cloud/realm"
)

// AppData is the Realm app data
type AppData interface {
	ConfigVersion() realm.AppConfigVersion
	ID() string
	Name() string
	Location() realm.Location
	DeploymentModel() realm.DeploymentModel
	LoadData(rootDir string) error
}

// set of supported local names
const (
	extJS   = ".js"
	extJSON = ".json"

	// app configs
	NameRealmConfig = "realm_config"
	NameConfig      = "config"
	NameStitch      = "stitch"

	// misc files
	NameSecrets = "secrets"
	NameSource  = "source"

	// auth
	NameAuth          = "auth"
	NameAuthProviders = "auth_providers"

	// functions
	NameFunctions   = "functions"
	nameNodeModules = "node_modules"

	// graphql
	NameGraphQL         = "graphql"
	NameCustomResolvers = "custom_resolvers"

	// hosting
	NameHosting = "hosting"

	// services
	NameIncomingWebhooks = "incoming_webhooks"
	NameRules            = "rules"
	NameServices         = "services"

	// triggers
	NameTriggers = "triggers"

	// sync
	NameSync = "sync"

	// values
	NameValues = "values"
)

// set of supported local files
var (
	FileRealmConfig = File{NameRealmConfig, extJSON}
	FileConfig      = File{NameConfig, extJSON}
	FileStitch      = File{NameStitch, extJSON}

	FileSecrets = File{NameSecrets, extJSON}

	FileSource = File{NameSource, extJS}
)

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// File is a local Realm app file
type File struct {
	Name string
	Ext  string
}

func (f File) String() string { return f.Name + f.Ext }

// TODO(REALMC-7653): recursively walk the functions directory and collect all .js files
// func walk(rootDir string, fn func(file os.FileInfo, path string) error) error {
// 	dw := directoryWalker{path: rootDir}
// 	if err := dw.walk(func(f os.FileInfo, p string) error {
// 		if f.IsDir() {
// 			return walk(p, fn)
// 		}
// 		return fn(f, p)
// 	}); err != nil {
// 		return err
// 	}
// 	return nil
// }

type directoryWalker struct {
	path            string
	continueOnError bool
	failOnNotExist  bool
	onlyDirs        bool
	onlyFiles       bool
}

func (dw directoryWalker) walk(fn func(file os.FileInfo, path string) error) error {
	if _, err := os.Stat(dw.path); err != nil {
		if os.IsNotExist(err) && !dw.failOnNotExist {
			return nil
		}
		return err
	}
	files, filesErr := ioutil.ReadDir(dw.path)
	if filesErr != nil {
		return filesErr
	}
	for _, file := range files {
		if dw.onlyDirs && !file.IsDir() || dw.onlyFiles && file.IsDir() {
			continue
		}
		err := fn(file, filepath.Join(dw.path, file.Name()))
		if err != nil {
			if dw.continueOnError {
				continue
			}
			return err
		}
	}
	return nil
}

func unmarshalDirectoryFlat(path string) ([]map[string]interface{}, error) {
	var out []map[string]interface{}

	dw := directoryWalker{path: path, onlyFiles: true}
	if walkErr := dw.walk(func(file os.FileInfo, path string) error {
		switch filepath.Ext(path) {
		case extJSON:
			data, dataErr := readFile(path)
			if dataErr != nil {
				return dataErr
			}

			var o map[string]interface{}
			if err := json.Unmarshal(data, &o); err != nil {
				return err
			}
			out = append(out, o)
		}
		return nil
	}); walkErr != nil {
		return nil, walkErr
	}

	return out, nil
}

type optionsUnmarshalJSON struct {
	failOnEmpty   bool
	failOnMissing bool
}

func readFile(path string) ([]byte, error) {
	return readFileWithOptions(path, optionsUnmarshalJSON{})
}

func readFileWithOptions(path string, opts optionsUnmarshalJSON) ([]byte, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) && !opts.failOnMissing {
			return nil, nil
		}
		return nil, err
	}

	data, readErr := ioutil.ReadFile(path)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read file at %s: %w", path, readErr)
	}

	if len(data) == 0 {
		if opts.failOnEmpty {
			return nil, fmt.Errorf("no file contents at %s", path)
		}
		return nil, nil
	}

	return data, nil
}
