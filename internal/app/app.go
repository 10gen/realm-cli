package app

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/10gen/realm-cli/internal/cloud/realm"
)

// ReadPackage reads the app package
func ReadPackage(wd string) (map[string]interface{}, error) {
	appDir, appDirOK, appDirErr := ResolveDirectory(wd)
	if appDirErr != nil {
		return nil, appDirErr
	}
	if !appDirOK {
		return map[string]interface{}{}, nil
	}

	switch appDir.ConfigVersion {
	case realm.AppConfigVersion20180301, realm.AppConfigVersion20200603:
		return readAppStructureV1(appDir)
	}
	return readAppStructureV2(appDir)
}

const (
	exportedJSONPrefix = ""
	exportedJSONIndent = "    "
)

// WriteDefaultConfig writes a default Realm app config to disk
func WriteDefaultConfig(path string, config Config) error {
	// TODO(REALMC-7653): marshal config according to version
	data, dataErr := json.MarshalIndent(config, exportedJSONPrefix, exportedJSONIndent)
	if dataErr != nil {
		return fmt.Errorf("failed to write app config: %w", dataErr)
	}

	file, fileErr := configVersionFile(realm.DefaultAppConfigVersion)
	if fileErr != nil {
		return fileErr
	}

	return writeFile(filepath.Join(path, file.String()), 0666, bytes.NewReader(data))
}

// WriteZip writes the zip contents to the specified filepath
func WriteZip(wd string, zipPkg *zip.Reader) error {
	if err := mkdir(wd); err != nil {
		return err
	}
	for _, zipFile := range zipPkg.File {
		path := filepath.Join(wd, zipFile.Name)

		if zipFile.FileInfo().IsDir() {
			if err := mkdir(path); err != nil {
				return err
			}
			continue
		}

		data, openErr := zipFile.Open()
		if openErr != nil {
			return openErr
		}
		defer data.Close()

		if err := writeFile(path, zipFile.Mode(), data); err != nil {
			return err
		}
	}
	return nil
}

// ResolveConfig resolves the MongoDB Realm application configuration
// based on the current working directory
// Empty data will be returned if called outside a project directory
func ResolveConfig(wd string) (string, Config, error) {
	appDir, appDirOK, appDirErr := ResolveDirectory(wd)
	if appDirErr != nil {
		return "", Config{}, appDirErr
	}
	if !appDirOK {
		return "", Config{}, nil
	}

	var config Config
	if err := unmarshalAppConfigInto(appDir, &config); err != nil {
		return "", Config{}, err
	}
	return appDir.Path, config, nil
}

// Directory represents the metadata associated with a Realm app project directory
type Directory struct {
	Path          string
	ConfigVersion realm.AppConfigVersion
}

const (
	maxDirectoryContainSearchDepth = 8
)

var (
	validConfigs = []File{FileRealmConfig, FileConfig, FileStitch}
)

// ResolveDirectory searches upwards from the current working directory
// for the root directory of a MongoDB Realm application project
// and returns the root directory path along with the project's config version
func ResolveDirectory(wd string) (Directory, bool, error) {
	wd, wdErr := filepath.Abs(wd)
	if wdErr != nil {
		return Directory{}, false, wdErr
	}

	for i := 0; i < maxDirectoryContainSearchDepth; i++ {
		for _, config := range validConfigs {
			path := filepath.Join(wd, config.String())
			if _, err := os.Stat(path); err == nil {
				return Directory{filepath.Dir(path), fileConfigVersion(config)}, true, nil
			}
		}

		if wd == "/" {
			break
		}
		wd = filepath.Clean(filepath.Join(wd, ".."))
	}

	return Directory{}, false, nil
}
