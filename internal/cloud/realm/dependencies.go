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
	dependenciesStatusPathPattern  = dependenciesPathPattern + "/status"
	dependenciesExportPathPattern  = dependenciesPathPattern + "/export"

	paramFile = "file"
)

// DependenciesStatus is used to get information from a dependencies status request
type DependenciesStatus struct {
	State   string `json:"status"`
	Message string `json:"status_message"`
}

// set of known dependencies status states
const (
	DependenciesStateCreated    = "created"
	DependenciesStateSuccessful = "successful"
	DependenciesStateFailed     = "failed"
)

func (c *client) DependenciesStatus(groupID, appID string) (DependenciesStatus, error) {
	res, err := c.do(
		http.MethodGet,
		fmt.Sprintf(dependenciesStatusPathPattern, groupID, appID),
		api.RequestOptions{},
	)
	if err != nil {
		return DependenciesStatus{}, err
	}
	if res.StatusCode != http.StatusOK {
		return DependenciesStatus{}, api.ErrUnexpectedStatusCode{"get dependencies status", res.StatusCode}
	}
	defer res.Body.Close()
	var status DependenciesStatus
	if err := json.NewDecoder(res.Body).Decode(&status); err != nil {
		return DependenciesStatus{}, err
	}
	return status, nil
}

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

	form, err := w.CreateFormFile(paramFile, fileInfo.Name())
	if err != nil {
		return err
	}

	if _, err := io.Copy(form, file); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	res, err := c.do(
		http.MethodPut,
		fmt.Sprintf(dependenciesPathPattern, groupID, appID),
		api.RequestOptions{
			Body:        body,
			ContentType: w.FormDataContentType(),
		},
	)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusNoContent {
		return api.ErrUnexpectedStatusCode{"import dependencies", res.StatusCode}
	}

	return nil
}

func (c *client) ExportDependencies(groupID, appID string) (string, io.ReadCloser, error) {
	res, resErr := c.do(http.MethodGet, fmt.Sprintf(dependenciesExportPathPattern, groupID, appID), api.RequestOptions{})
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

func (c *client) ExportDependenciesArchive(groupID, appID string) (string, io.ReadCloser, error) {
	res, resErr := c.do(http.MethodGet, fmt.Sprintf(dependenciesArchivePathPattern, groupID, appID), api.RequestOptions{})
	if resErr != nil {
		return "", nil, resErr
	}
	if res.StatusCode != http.StatusOK {
		return "", nil, api.ErrUnexpectedStatusCode{"export dependencies archive", res.StatusCode}
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
