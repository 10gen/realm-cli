package local

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/10gen/realm-cli/internal/cloud/realm"
)

// AppConfigJSON is the app config.json data
type AppConfigJSON struct {
	AppDataV1
}

// NewAppConfigJSON returns app config.json data
func NewAppConfigJSON(name string, meta realm.AppMeta) AppConfigJSON {
	return AppConfigJSON{AppDataV1{AppStructureV1{
		Name:            name,
		Location:        meta.Location,
		DeploymentModel: meta.DeploymentModel,
	}}}
}

// AppStitchJSON is the app stitch.json data
type AppStitchJSON struct {
	AppDataV1
}

// NewAppStitchJSON returns app stitch.json data
func NewAppStitchJSON(name string, meta realm.AppMeta) AppStitchJSON {
	return AppStitchJSON{AppDataV1{AppStructureV1{
		Name:            name,
		Location:        meta.Location,
		DeploymentModel: meta.DeploymentModel,
	}}}
}

// AppRealmConfigJSON is the app realm_config.json data
type AppRealmConfigJSON struct {
	AppDataV2
}

// NewAppRealmConfigJSON returns app realm_config.json data
func NewAppRealmConfigJSON(name string, meta realm.AppMeta) AppRealmConfigJSON {
	return AppRealmConfigJSON{AppDataV2{AppStructureV2{
		Name:            name,
		Location:        meta.Location,
		DeploymentModel: meta.DeploymentModel,
	}}}
}

// set of write options
const (
	exportedJSONPrefix = ""
	exportedJSONIndent = "    "
)

// MarshalJSON returns the json representation of the passed in interface
func MarshalJSON(o interface{}) ([]byte, error) {
	data, err := json.MarshalIndent(o, exportedJSONPrefix, exportedJSONIndent)
	if err != nil {
		return nil, err
	}
	data = append(data, "\n"...)
	return data, nil
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

		if err := WriteFile(path, zipFile.Mode(), data); err != nil {
			return err
		}
	}
	return nil
}

func mkdir(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory at %s: %w", path, err)
	}
	return nil
}

// WriteFile writes the file to the specified path with the specified permissions
func WriteFile(path string, perm os.FileMode, r io.Reader) error {
	if err := mkdir(filepath.Dir(path)); err != nil {
		return err
	}

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
