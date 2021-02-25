package utils

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/10gen/realm-cli/models"
	"github.com/mitchellh/go-homedir"
)

const (
	// FunctionsRoot is the root directory for functions and dependencies
	FunctionsRoot = "functions"

	jsonExt              = ".json"
	jsExt                = ".js"
	authProvidersName    = "auth_providers"
	appConfigName        = "config"
	configName           = "config"
	triggersName         = "triggers"
	incomingWebhooksName = "incoming_webhooks"
	rulesName            = "rules"
	secretsName          = "secrets"
	servicesName         = "services"
	environmentsName     = "environments"
	sourceName           = "source"
	valuesName           = "values"
	graphQLName          = "graphql"
	customResolversName  = "custom_resolvers"
)

//
var (
	// HostingRoot is the root directory for hosting assets and attributes
	HostingRoot = "hosting"
	// HostingFilesDirectory is the directory to place the static hosting assets
	HostingFilesDirectory = fmt.Sprintf("%s/files", HostingRoot)
	// HostingAttributes is the file that stores the static hosting asset descriptions struct
	HostingAttributes = fmt.Sprintf("%s/metadata.json", HostingRoot)
	// HostingCacheFileName is the file that stores the cached hosting asset data
	HostingCacheFileName = ".asset-cache.json"

	errAppNotFound = errors.New("could not find realm app")
)

const maxDirectoryContainSearchDepth = 8

// GetDirectoryContainingFile searches upwards for a valid Realm app directory
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
		if processErr := processFile(filepath.Join(dest, zipFile.Name), zipFile); processErr != nil {
			return processErr
		}
	}

	return err
}

// WriteFileToDir writes the data to dest and creates the necessary directories along the path
func WriteFileToDir(dest string, data io.Reader) error {
	// make all subdirectories if necessary
	err := os.MkdirAll(path.Dir(dest), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create sub-directory %q: %s", dest, err)
	}

	// now we create the file
	f, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create file %q: %s", dest, err)
	}
	defer f.Close()

	_, err = io.Copy(f, data)
	if err != nil {
		return fmt.Errorf("failed to copy file %q: %s", dest, err)
	}

	return nil
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

// UnmarshalFromDir unmarshals a Realm app from the given directory into a map[string]interface{}
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

	values, err := unmarshalJSONFiles(filepath.Join(path, valuesName), true)
	if err != nil {
		return app, err
	}

	if len(values) != 0 {
		app[valuesName] = values
	}

	authProviders, err := unmarshalJSONFiles(filepath.Join(path, authProvidersName), true)
	if err != nil {
		return app, err
	}

	if len(authProviders) != 0 {
		app[authProvidersName] = authProviders
	}

	functions, err := unmarshalFunctionDirectories(filepath.Join(path, FunctionsRoot), true)
	if err != nil {
		return app, err
	}

	if len(functions) != 0 {
		app[FunctionsRoot] = functions
	}

	triggers, err := unmarshalJSONFiles(filepath.Join(path, triggersName), true)
	if err != nil {
		return app, err
	}

	if len(triggers) != 0 {
		app[triggersName] = triggers
	}

	graphQL, err := unmarshalGraphQLDirectories(filepath.Join(path, graphQLName), true)
	if err != nil {
		return app, err
	}

	app[graphQLName] = graphQL

	services, err := unmarshalServiceDirectories(filepath.Join(path, servicesName), true)
	if err != nil {
		return app, err
	}
	app[servicesName] = services

	environmentsPath := filepath.Join(path, environmentsName)
	_, err = os.Stat(environmentsPath)
	if err == nil {
		// ignore environments folder if it's missing
		environments, err := unmarshalJSONFilesWithFilenames(environmentsPath)
		if err != nil {
			return app, err
		}
		app[environmentsName] = environments
	}

	return app, nil
}

func unmarshalJSONFiles(path string, ignoreDirErr bool) ([]interface{}, error) {
	fileInfos, err := ioutil.ReadDir(path)
	if err != nil && !ignoreDirErr {
		return []interface{}{}, err
	}
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

func unmarshalJSONFilesWithFilenames(path string) (map[string]interface{}, error) {
	fileInfos, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	files := make(map[string]interface{}, len(fileInfos))

	for _, fileInfo := range fileInfos {
		jsonFilePath := filepath.Join(path, fileInfo.Name())
		if filepath.Ext(jsonFilePath) != jsonExt {
			continue
		}

		var f interface{}
		if err := readAndUnmarshalJSONInto(jsonFilePath, &f); err != nil {
			return nil, err
		}

		files[fileInfo.Name()] = f
	}

	return files, nil
}

func unmarshalFunctionDirectories(path string, ignoreDirErr bool) ([]interface{}, error) {
	fileInfos, err := ioutil.ReadDir(path)
	if err != nil && !ignoreDirErr {
		return []interface{}{}, err
	}
	directories := []interface{}{}

	err = iterDirectories(func(info os.FileInfo, path string) error {
		// we skip over node_modules since we upload that as a single entity
		if strings.Contains(path, "node_modules") {
			return nil
		}
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

func unmarshalGraphQLDirectories(path string, ignoreDirErr bool) (map[string]interface{}, error) {
	fileInfos, err := ioutil.ReadDir(path)
	if err != nil && !ignoreDirErr {
		return map[string]interface{}{}, err
	}

	gqlServices := map[string]interface{}{}
	gqlServices[customResolversName] = []interface{}{}
	gqlConfigFilename := configName + jsonExt

	// find the graphql config file
	for _, fi := range fileInfos {
		if fi.Name() == gqlConfigFilename {
			var config map[string]interface{}
			if err := readAndUnmarshalJSONInto(filepath.Join(path, fi.Name()), &config); err != nil {
				return map[string]interface{}{}, err
			}
			gqlServices[configName] = config
		}
	}
	err = iterDirectories(func(info os.FileInfo, path string) error {
		gqlSvcFileInfos, err := ioutil.ReadDir(path)
		if err != nil {
			return err
		}

		for _, fileInfo := range gqlSvcFileInfos {
			var config map[string]interface{}
			if err := readAndUnmarshalJSONInto(filepath.Join(path, fileInfo.Name()), &config); err != nil {
				return err
			}

			// As we add graphql-related services we can expand this to unmarshal each one accordingly
			if strings.Contains(path, customResolversName) {
				if customResolvers, ok := gqlServices[customResolversName].([]interface{}); ok {
					gqlServices[customResolversName] = append(customResolvers, config)
				}
			}
		}

		return nil
	}, path, fileInfos)

	if err != nil {
		return nil, err
	}

	return gqlServices, nil
}

func unmarshalServiceDirectories(path string, ignoreDirErr bool) ([]interface{}, error) {
	fileInfos, err := ioutil.ReadDir(path)
	if err != nil && !ignoreDirErr {
		return []interface{}{}, err
	}
	services := []interface{}{}

	err = iterDirectories(func(info os.FileInfo, path string) error {
		svc := map[string]interface{}{}

		var config map[string]interface{}
		if err := readAndUnmarshalJSONInto(filepath.Join(path, configName+jsonExt), &config); err != nil {
			return err
		}

		svc[configName] = config

		incomingWebhooks, err := unmarshalFunctionDirectories(filepath.Join(path, incomingWebhooksName), true)
		if err != nil {
			return err
		}

		svc[incomingWebhooksName] = incomingWebhooks

		rules, err := unmarshalJSONFiles(filepath.Join(path, rulesName), true)
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
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	if err := json.Unmarshal(data, out); err != nil {
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

// GenerateFileHashStr takes a file name and opens and generates a string of the hash.Hash for that file
func GenerateFileHashStr(fName string) (string, error) {
	f, err := os.Open(fName)
	if err != nil {
		return "", err
	}

	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// ResolveAppDirectory returns the directory path for an app
func ResolveAppDirectory(appPath, workingDirectory string) (string, error) {
	if appPath != "" {
		path, err := homedir.Expand(appPath)
		if err != nil {
			return "", err
		}

		if _, err := os.Stat(path); err != nil {
			return "", errors.New("directory does not exist")
		}
		return path, nil
	}

	return GetDirectoryContainingFile(workingDirectory, models.AppConfigFileName)
}

// ResolveAppInstanceData loads data for an app from a config.json file located in the provided directory path,
// merging in any overridden parameters from command line flags
func ResolveAppInstanceData(appID, path string) (models.AppInstanceData, error) {
	appInstanceDataFromFile := models.AppInstanceData{}
	err := appInstanceDataFromFile.UnmarshalFile(path)

	if os.IsNotExist(err) {
		return models.AppInstanceData{
			models.AppIDField: appID,
		}, nil
	}

	if err != nil {
		return nil, err
	}

	if appID != "" {
		appInstanceDataFromFile[models.AppIDField] = appID
	}

	return appInstanceDataFromFile, nil
}
