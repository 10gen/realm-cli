package app

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func mkdir(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory at %s: %w", path, err)
	}
	return nil
}

type fileIterator struct {
	path            string
	continueOnError bool
	failOnMissing   bool
	onlyDirs        bool
	onlyFiles       bool
}

func (fi fileIterator) forEach(fn func(file os.FileInfo, path string) error) error {
	if _, err := os.Stat(fi.path); err != nil {
		if os.IsNotExist(err) && !fi.failOnMissing {
			return nil
		}
		return err
	}
	files, filesErr := ioutil.ReadDir(fi.path)
	if filesErr != nil {
		return filesErr
	}
	for _, file := range files {
		if fi.onlyDirs && !file.IsDir() || fi.onlyFiles && file.IsDir() {
			continue
		}
		err := fn(file, filepath.Join(fi.path, file.Name()))
		if err != nil {
			if fi.continueOnError {
				continue
			}
			return err
		}
	}
	return nil
}

func writeFile(path string, perm os.FileMode, r io.Reader) error {
	f, openErr := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if openErr != nil {
		return fmt.Errorf("failed to open file at %s: %s", path, openErr)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("failed to write file at %s: %w", path, err)
	}
	return nil
}

type unmarshalOptions struct {
	path          string
	failOnEmpty   bool
	failOnMissing bool
}

func unmarshalJSONInto(out interface{}, opts unmarshalOptions) error {
	if _, err := os.Stat(opts.path); err != nil {
		if os.IsNotExist(err) && !opts.failOnMissing {
			return nil
		}
		return err
	}

	data, readErr := ioutil.ReadFile(opts.path)
	if readErr != nil {
		return fmt.Errorf("failed to read file at %s: %s", opts.path, readErr)
	}

	if len(data) == 0 {
		if opts.failOnEmpty {
			return fmt.Errorf("no file contents at %s", opts.path)
		}
		return nil
	}
	return json.Unmarshal(data, out)
}

func unmarshalDirectoryInto(path string, name string, pkg map[string]interface{}) error {
	var out []interface{}

	fi := fileIterator{path: filepath.Join(path, name), onlyFiles: true}
	if err := fi.forEach(func(file os.FileInfo, path string) error {
		switch filepath.Ext(path) {
		case extJSON:
			var o interface{}
			if err := unmarshalJSONInto(&o, unmarshalOptions{path: path}); err != nil {
				return err
			}
			out = append(out, o)
		}
		return nil
	}); err != nil {
		return err
	}

	if len(out) != 0 {
		pkg[name] = out
	}
	return nil
}

func unmarshalAppConfigInto(appDir Directory, out interface{}) error {
	file, fileErr := configVersionFile(appDir.ConfigVersion)
	if fileErr != nil {
		return fileErr
	}

	return unmarshalJSONInto(out, unmarshalOptions{
		path:        filepath.Join(appDir.Path, file.String()),
		failOnEmpty: true,
	})
}

const (
	nodeModules = "node_modules"
)

func unmarshalFunctionsInto(path, name string, pkg map[string]interface{}) error {
	var out []interface{}

	fi := fileIterator{path: filepath.Join(path, name), onlyDirs: true}
	if err := fi.forEach(func(file os.FileInfo, path string) error {
		if strings.Contains(path, nodeModules) {
			return nil // skip node_modules since we upload that as a single entity
		}

		var cfg interface{}
		if err := unmarshalJSONInto(&cfg, unmarshalOptions{path: filepath.Join(path, FileConfig.String())}); err != nil {
			return err
		}

		src, srcErr := ioutil.ReadFile(filepath.Join(path, FileSource.String()))
		if srcErr != nil {
			return srcErr
		}

		o := map[string]interface{}{}
		o[NameConfig] = cfg
		o[NameSource] = src

		out = append(out, o)
		return nil
	}); err != nil {
		return err
	}

	if len(out) != 0 {
		pkg[name] = out
	}
	return nil
}

func unmarshalGraphQLInto(appDir Directory, pkg map[string]interface{}) error {
	out := map[string]interface{}{}

	var cfg interface{}
	if err := unmarshalJSONInto(&cfg, unmarshalOptions{path: filepath.Join(appDir.Path, FileGraphQLConfig.String())}); err != nil {
		return err
	}
	out[NameConfig] = cfg

	out[NameCustomResolvers] = []interface{}{}

	fi := fileIterator{path: filepath.Join(appDir.Path, DirGraphQLCustomResolvers.String()), onlyFiles: true}
	if err := fi.forEach(func(file os.FileInfo, path string) error {
		var customResolver interface{}
		if err := unmarshalJSONInto(&customResolver, unmarshalOptions{path: path}); err != nil {
			return err
		}

		if customResolvers, ok := out[NameCustomResolvers].([]interface{}); ok {
			out[NameCustomResolvers] = append(customResolvers, customResolver)
		}
		return nil
	}); err != nil {
		return err
	}

	pkg[NameGraphQL] = out
	return nil
}

func unmarshalSecretsInto(appDir Directory, pkg map[string]interface{}) error {
	filePath := filepath.Join(appDir.Path, FileSecrets.String())
	if _, statErr := os.Stat(filePath); statErr == nil {
		var out interface{}

		if err := unmarshalJSONInto(&out, unmarshalOptions{path: filePath}); err != nil {
			return err
		}
		pkg[NameSecrets] = out
	}
	return nil
}

func unmarshalServicesInto(appDir Directory, pkg map[string]interface{}) error {
	var out []interface{}

	fi := fileIterator{path: filepath.Join(appDir.Path, NameServices), onlyDirs: true}
	if err := fi.forEach(func(file os.FileInfo, path string) error {
		o := map[string]interface{}{}

		var cfg interface{}
		if err := unmarshalJSONInto(&cfg, unmarshalOptions{path: filepath.Join(path, FileConfig.String())}); err != nil {
			return err
		}
		o[NameConfig] = cfg

		if _, statErr := os.Stat(filepath.Join(path, NameIncomingWebhooks)); statErr == nil {
			if err := unmarshalFunctionsInto(path, NameIncomingWebhooks, o); err != nil {
				return err
			}
		}

		if err := unmarshalDirectoryInto(path, NameRules, o); err != nil {
			return err
		}

		out = append(out, o)
		return nil
	}); err != nil {
		return err
	}

	if len(out) != 0 {
		pkg[NameServices] = out
	}
	return nil
}
