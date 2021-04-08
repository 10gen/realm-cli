package realm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	dependenciesPathPattern        = appPathPattern + "/dependencies"
	dependenciesArchivePathPattern = dependenciesPathPattern + "/archive"
	dependenciesDiffPathPattern    = dependenciesPathPattern + "/diff"

	paramFile = "file"
)

func (c *client) ImportDependencies(groupID, appID, uploadPath string) error {
	file, fileErr := os.Open(uploadPath)
	if fileErr != nil {
		return fileErr
	}
	defer file.Close()

	fileInfo, fileInfoErr := file.Stat()
	if fileInfoErr != nil {
		return fileInfoErr
	}

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	form, formErr := w.CreateFormFile(paramFile, fileInfo.Name())
	if formErr != nil {
		return formErr
	}

	if _, err := io.Copy(form, file); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	res, resErr := c.do(
		http.MethodPost,
		fmt.Sprintf(dependenciesPathPattern, groupID, appID),
		api.RequestOptions{
			Body:        body,
			ContentType: w.FormDataContentType(),
		},
	)
	if resErr != nil {
		return resErr
	}
	if res.StatusCode != http.StatusNoContent {
		return api.ErrUnexpectedStatusCode{"import dependencies", res.StatusCode}
	}
	return nil
}

func (c *client) ExportDependencies(groupID, appID string) (string, io.ReadCloser, error) {
	res, resErr := c.do(http.MethodGet, fmt.Sprintf(dependenciesArchivePathPattern, groupID, appID), api.RequestOptions{})
	if resErr != nil {
		return "", nil, resErr
	}
	if res.StatusCode != http.StatusOK {
		return "", nil, api.ErrUnexpectedStatusCode{"export dependencies", res.StatusCode}
	}

	_, mediaParams, mediaErr := mime.ParseMediaType(res.Header.Get(api.HeaderContentDisposition))
	if mediaErr != nil {
		return "", nil, mediaErr
	}

	filename := mediaParams[mediaParamFilename]
	if filename == "" {
		return "", nil, errors.New("export response is missing filename")
	}

	return filename, res.Body, nil
}

func (c *client) DiffDependencies(groupID, appID, uploadPath string) (DependenciesDiff, error) {
	file, err := os.Open(uploadPath)
	if err != nil {
		return DependenciesDiff{}, err
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return DependenciesDiff{}, err
	}

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	form, err := w.CreateFormFile(paramFile, fi.Name())
	if err != nil {
		return DependenciesDiff{}, err
	}

	if _, err := io.Copy(form, file); err != nil {
		return DependenciesDiff{}, err
	}
	if err := w.Close(); err != nil {
		return DependenciesDiff{}, err
	}

	res, err := c.do(
		http.MethodPost,
		fmt.Sprintf(dependenciesDiffPathPattern, groupID, appID),
		api.RequestOptions{
			Body:        body,
			ContentType: w.FormDataContentType(),
		},
	)
	if err != nil {
		return DependenciesDiff{}, err
	}
	if res.StatusCode != http.StatusOK {
		return DependenciesDiff{}, api.ErrUnexpectedStatusCode{"diff dependencies", res.StatusCode}
	}
	defer res.Body.Close()

	var diff DependenciesDiff
	if err := json.NewDecoder(res.Body).Decode(&diff); err != nil {
		return DependenciesDiff{}, err
	}

	return diff, nil
}
