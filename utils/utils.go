package utils

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	jsonExt                string = ".json"
	jsExt                         = ".js"
	authProvidersName             = "auth_providers"
	appConfigName                 = "stitch"
	configName                    = "config"
	functionsName                 = "functions"
	eventSubscriptionsName        = "event_subscriptions"
	incomingWebhooksName          = "incoming_webhooks"
	rulesName                     = "rules"
	secretsName                   = "secrets"
	servicesName                  = "services"
	sourceName                    = "source"
	valuesName                    = "values"
)

var (
	errAppNotFound = errors.New("could not find stitch app")
)

// ReadAndUnmarshalInto unmarshals data from the given path into an interface{} using the provided marshalFn
func ReadAndUnmarshalInto(marshalFn func(in []byte, out interface{}) error, path string, out interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	if err := marshalFn(data, out); err != nil {
		return fmt.Errorf("failed to parse %s: %s", path, err)
	}

	return nil
}

const maxDirectoryContainSearchDepth = 8

// GetDirectoryContainingFile searches upwards for a valid Stitch app directory
func GetDirectoryContainingFile(wd, filename string) (string, error) {
	wd, err := filepath.Abs(wd)
	if err != nil {
		return "", err
	}

	for i := 0; i < maxDirectoryContainSearchDepth; i++ {
		path := filepath.Join(wd, filename)
		if _, err := os.Stat(path); err == nil {
			return filepath.Dir(path), nil
		}

		if wd == "/" {
			break
		}

		wd = filepath.Clean(filepath.Join(wd, ".."))
	}

	return "", errAppNotFound
}

// WriteZipToDir takes a destination and an io.Reader containing zip data and unpacks it
func WriteZipToDir(dest string, zipData io.Reader, overwrite bool) error {
	if _, err := os.Open(dest); !overwrite && err == nil {
		return fmt.Errorf("failed to create directory %q: directory already exists", dest)
	}

	b, err := ioutil.ReadAll(zipData)
	if err != nil {
		return err
	}

	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return err
	}

	err = os.MkdirAll(dest, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory %q: %s", dest, err)
	}

	for _, zipFile := range r.File {
		if err := processFile(filepath.Join(dest, zipFile.Name), zipFile); err != nil {
			return err
		}
	}

	return err
}

func processFile(path string, zipFile *zip.File) error {
	fileData, err := zipFile.Open()
	if err != nil {
		return fmt.Errorf("failed to extract file %q: %s", path, err)
	}
	defer fileData.Close()

	if zipFile.FileInfo().IsDir() {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create sub-directory %q: %s", path, err)
		}
	} else {
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zipFile.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file %q: %s", path, err)
		}
		defer f.Close()

		_, err = io.Copy(f, fileData)
		if err != nil {
			return fmt.Errorf("failed to extract file %q: %s", path, err)
		}
	}

	return nil
}

// UnmarshalFromDir unmarshals a Stitch app from the given directory into a map[string]interface{}
func UnmarshalFromDir(path string) (map[string]interface{}, error) {
	app := map[string]interface{}{}

	if err := readAndUnmarshalJSONInto(filepath.Join(path, appConfigName+jsonExt), &app); err != nil {
		return app, err
	}

	if _, err := os.Stat(filepath.Join(path, secretsName+jsonExt)); err == nil {
		var secrets interface{}
		if err := readAndUnmarshalJSONInto(filepath.Join(path, secretsName+jsonExt), &secrets); err != nil {
			return app, err
		}

		app[secretsName] = secrets
	}

	values, err := unmarshalJSONFiles(filepath.Join(path, valuesName))
	if err != nil {
		return app, err
	}

	if len(values) != 0 {
		app[valuesName] = values
	}

	authProviders, err := unmarshalJSONFiles(filepath.Join(path, authProvidersName))
	if err != nil {
		return app, err
	}

	if len(authProviders) != 0 {
		app[authProvidersName] = authProviders
	}

	functions, err := unmarshalFunctionDirectories(filepath.Join(path, functionsName))
	if err != nil {
		return app, err
	}

	if len(functions) != 0 {
		app[functionsName] = functions
	}

	eventSubscriptions, err := unmarshalJSONFiles(filepath.Join(path, eventSubscriptionsName))
	if err != nil {
		return app, err
	}

	if len(eventSubscriptions) != 0 {
		app[eventSubscriptionsName] = eventSubscriptions
	}

	services, err := unmarshalServiceDirectories(filepath.Join(path, servicesName))
	if err != nil {
		return app, err
	}

	app[servicesName] = services

	return app, nil
}

func unmarshalJSONFiles(path string) ([]interface{}, error) {
	fileInfos, _ := ioutil.ReadDir(path)
	files := make([]interface{}, 0, len(fileInfos))

	for _, fileInfo := range fileInfos {
		jsonFilePath := filepath.Join(path, fileInfo.Name())
		if filepath.Ext(jsonFilePath) != jsonExt {
			continue
		}

		var f interface{}
		if err := readAndUnmarshalJSONInto(jsonFilePath, &f); err != nil {
			return []interface{}{}, err
		}

		files = append(files, f)
	}

	return files, nil
}

func unmarshalFunctionDirectories(path string) ([]interface{}, error) {
	fileInfos, _ := ioutil.ReadDir(path)
	directories := []interface{}{}

	err := iterDirectories(func(info os.FileInfo, path string) error {
		var config interface{}
		if err := readAndUnmarshalJSONInto(filepath.Join(path, configName+jsonExt), &config); err != nil {
			return err
		}

		sourceBytes, err := ioutil.ReadFile(filepath.Join(path, sourceName+jsExt))
		if err != nil {
			return err
		}

		directory := map[string]interface{}{}
		directory[configName] = config
		directory[sourceName] = string(sourceBytes)

		directories = append(directories, directory)

		return nil
	}, path, fileInfos)

	if err != nil {
		return nil, err
	}

	return directories, nil
}

func unmarshalServiceDirectories(path string) ([]interface{}, error) {
	fileInfos, _ := ioutil.ReadDir(path)
	services := []interface{}{}

	err := iterDirectories(func(info os.FileInfo, path string) error {
		svc := map[string]interface{}{}

		var config map[string]interface{}
		if err := readAndUnmarshalJSONInto(filepath.Join(path, configName+jsonExt), &config); err != nil {
			return err
		}

		svc[configName] = config

		incomingWebhooks, err := unmarshalFunctionDirectories(filepath.Join(path, incomingWebhooksName))
		if err != nil {
			return err
		}

		svc[incomingWebhooksName] = incomingWebhooks

		rules, err := unmarshalJSONFiles(filepath.Join(path, rulesName))
		if err != nil {
			return err
		}

		svc[rulesName] = rules

		services = append(services, svc)

		return nil
	}, path, fileInfos)

	if err != nil {
		return nil, err
	}

	return services, nil
}

func iterDirectories(iterFn func(info os.FileInfo, path string) error, path string, fileInfos []os.FileInfo) error {
	for _, fileInfo := range fileInfos {
		fileNamePath := filepath.Join(path, fileInfo.Name())
		if info, err := os.Stat(fileNamePath); !info.IsDir() || err != nil {
			continue
		}

		if err := iterFn(fileInfo, fileNamePath); err != nil {
			return err
		}
	}

	return nil
}

func readAndUnmarshalJSONInto(path string, out interface{}) error {
	if err := ReadAndUnmarshalInto(json.Unmarshal, path, out); err != nil {
		return fmt.Errorf("failed to parse %s: %s", path, err)
	}

	return nil
}

// MediaType defines the type of HTTP media in a request/response
type MediaType string

// The set of known MediaTypes
const (
	MediaTypeTextPlain         = MediaType("text/plain; charset=utf-8")
	MediaTypeOctetStream       = MediaType("application/octet-stream")
	MediaTypeHTML              = MediaType("text/html")
	MediaTypeJSON              = MediaType("application/json")
	MediaTypeMultipartFormData = MediaType("multipart/form-data")
	MediaTypeFormURLEncoded    = MediaType("application/x-www-form-urlencoded")
	MediaTypeZip               = MediaType("application/zip")
)
