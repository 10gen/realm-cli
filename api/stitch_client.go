package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/10gen/stitch-cli/auth"
	"github.com/10gen/stitch-cli/hosting"
	"github.com/10gen/stitch-cli/models"
)

// ExportStrategy is the enumeration of possible strategies which can be used
// when exporting a stitch application
type ExportStrategy string

const (
	// ExportStrategyNone will result in no extra configuration into the call to Export
	ExportStrategyNone ExportStrategy = "none"
	// ExportStrategyTemplate will result in the `template` querystring parameter getting added to the call to Export
	ExportStrategyTemplate ExportStrategy = "template"
	// ExportStrategySourceControl will result in the `source_control` querystring parameter getting added to the call to Export
	ExportStrategySourceControl ExportStrategy = "source_control"

	authProviderLoginRoute      = adminBaseURL + "/auth/providers/%s/login"
	appExportRoute              = adminBaseURL + "/groups/%s/apps/%s/export?%s"
	appImportRoute              = adminBaseURL + "/groups/%s/apps/%s/import"
	appsByGroupIDRoute          = adminBaseURL + "/groups/%s/apps"
	userProfileRoute            = adminBaseURL + "/auth/profile"
	hostingAssetRoute           = adminBaseURL + "/groups/%s/apps/%s/hosting/assets/asset"
	hostingAssetsRoute          = adminBaseURL + "/groups/%s/apps/%s/hosting/assets"
	hostingInvalidateCacheRoute = adminBaseURL + "/groups/%s/apps/%s/hosting/cache"
)

var (
	errExportMissingFilename = errors.New("the app export response did not specify a filename")
	errGroupNotFound         = errors.New("group could not be found")
)

const (
	metadataParam = "meta"
	fileParam     = "file"
	pathParam     = "path"
)

type copyPayload struct {
	CopyFrom string `json:"copy_from"`
	CopyTo   string `json:"copy_to"`
}

type movePayload struct {
	MoveFrom string `json:"move_from"`
	MoveTo   string `json:"move_to"`
}

type setAttributesPayload struct {
	Attributes []hosting.AssetAttribute `json:"attributes"`
}

type invalidateCachePayload struct {
	Invalidate bool   `json:"invalidate"`
	Path       string `json:"path"`
}

// StitchClient represents a Client that can be used to call the Stitch Admin API
type StitchClient interface {
	Authenticate(authProvider auth.AuthenticationProvider) (*auth.Response, error)
	Export(groupID, appID string, strategy ExportStrategy) (string, io.ReadCloser, error)
	Import(groupID, appID string, appData []byte, strategy string) error
	Diff(groupID, appID string, appData []byte, strategy string) ([]string, error)
	FetchAppByGroupIDAndClientAppID(groupID, clientAppID string) (*models.App, error)
	FetchAppByClientAppID(clientAppID string) (*models.App, error)
	FetchAppsByGroupID(groupID string) ([]*models.App, error)
	CreateEmptyApp(groupID, appName, location, deploymentModel string) (*models.App, error)
	UploadAsset(groupID, appID, path, hash string, size int64, body io.Reader, attributes ...hosting.AssetAttribute) error
	CopyAsset(groupID, appID, fromPath, toPath string) error
	MoveAsset(groupID, appID, fromPath, toPath string) error
	DeleteAsset(groupID, appID, path string) error
	SetAssetAttributes(groupID, appID, path string, attributes ...hosting.AssetAttribute) error
	ListAssetsForAppID(groupID, appID string) ([]hosting.AssetMetadata, error)
	InvalidateCache(groupID, appID, path string) error
}

// NewStitchClient returns a new StitchClient to be used for making calls to the Stitch Admin API
func NewStitchClient(client Client) StitchClient {
	return &basicStitchClient{
		Client: client,
	}
}

type basicStitchClient struct {
	Client
}

// Authenticate will authenticate a user given an api key and username
func (sc *basicStitchClient) Authenticate(authProvider auth.AuthenticationProvider) (*auth.Response, error) {
	body, err := json.Marshal(authProvider.Payload())
	if err != nil {
		return nil, err
	}

	res, err := sc.Client.ExecuteRequest(http.MethodPost, fmt.Sprintf(authProviderLoginRoute, authProvider.Type()), RequestOptions{
		Body: bytes.NewReader(body),
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
	})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: failed to authenticate: %s", res.Status, UnmarshalStitchError(res))
	}

	decoder := json.NewDecoder(res.Body)

	var authResponse auth.Response
	if err := decoder.Decode(&authResponse); err != nil {
		return nil, err
	}

	return &authResponse, nil
}

// Export will download a Stitch app as a .zip
func (sc *basicStitchClient) Export(groupID, appID string, strategy ExportStrategy) (string, io.ReadCloser, error) {
	url := fmt.Sprintf(appExportRoute, groupID, appID, "")
	if strategy == ExportStrategyTemplate {
		url = fmt.Sprintf(appExportRoute, groupID, appID, "template=true")
	} else if strategy == ExportStrategySourceControl {
		url = fmt.Sprintf(appExportRoute, groupID, appID, "source_control=true")
	}

	res, err := sc.ExecuteRequest(http.MethodGet, url, RequestOptions{})
	if err != nil {
		return "", nil, err
	}

	if res.StatusCode != http.StatusOK {
		defer res.Body.Close()
		return "", nil, UnmarshalStitchError(res)
	}

	_, params, err := mime.ParseMediaType(res.Header.Get(hosting.AttributeContentDisposition))
	if err != nil {
		res.Body.Close()
		return "", nil, err
	}

	filename := params["filename"]
	if len(filename) == 0 {
		res.Body.Close()
		return "", nil, errExportMissingFilename
	}

	return filename, res.Body, nil
}

// Diff will execute a dry-run of an import, returning a diff of proposed changes
func (sc *basicStitchClient) Diff(groupID, appID string, appData []byte, strategy string) ([]string, error) {
	res, err := sc.invokeImportRoute(groupID, appID, appData, strategy, true)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, UnmarshalStitchError(res)
	}

	var diffs []string
	if err := json.NewDecoder(res.Body).Decode(&diffs); err != nil {
		return nil, err
	}

	return diffs, nil
}

// Import will push a local Stitch app to the server
func (sc *basicStitchClient) Import(groupID, appID string, appData []byte, strategy string) error {
	res, err := sc.invokeImportRoute(groupID, appID, appData, strategy, false)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		return UnmarshalStitchError(res)
	}

	return nil
}

func (sc *basicStitchClient) invokeImportRoute(groupID, appID string, appData []byte, strategy string, diff bool) (*http.Response, error) {
	url := fmt.Sprintf(appImportRoute, groupID, appID)

	url += fmt.Sprintf("?strategy=%s", strategy)
	if diff {
		url += "&diff=true"
	}

	return sc.ExecuteRequest(http.MethodPost, url, RequestOptions{Body: bytes.NewReader(appData)})
}

func (sc *basicStitchClient) FetchAppsByGroupID(groupID string) ([]*models.App, error) {
	res, err := sc.ExecuteRequest(http.MethodGet, fmt.Sprintf(appsByGroupIDRoute, groupID), RequestOptions{})
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusNotFound {
			return nil, errGroupNotFound
		}
		return nil, UnmarshalStitchError(res)
	}

	dec := json.NewDecoder(res.Body)
	var apps []*models.App
	if err := dec.Decode(&apps); err != nil {
		return nil, err
	}

	return apps, nil
}

// FetchAppByGroupIDAndClientAppID fetches a Stitch app given a groupID and clientAppID
func (sc *basicStitchClient) FetchAppByGroupIDAndClientAppID(groupID, clientAppID string) (*models.App, error) {
	return sc.findProjectAppByClientAppID([]string{groupID}, clientAppID)
}

// FetchAppByClientAppID fetches a Stitch app given a clientAppID
func (sc *basicStitchClient) FetchAppByClientAppID(clientAppID string) (*models.App, error) {
	res, err := sc.ExecuteRequest(http.MethodGet, userProfileRoute, RequestOptions{})
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, UnmarshalStitchError(res)
	}

	dec := json.NewDecoder(res.Body)
	var profileData models.UserProfile
	if err := dec.Decode(&profileData); err != nil {
		return nil, err
	}

	return sc.findProjectAppByClientAppID(profileData.AllGroupIDs(), clientAppID)
}

// UploadAsset creates a pipe and writes the asset to an http.POST along with its metadata
func (sc *basicStitchClient) UploadAsset(groupID, appID, path, hash string, size int64, body io.Reader, attributes ...hosting.AssetAttribute) error {
	// The upload request consists of a multipart body with two parts:
	// 1) the metadata, as json, and 2) the file data itself.

	// First build the metadata part
	metaPart, err := json.Marshal(hosting.AssetMetadata{
		AppID:    appID,
		FilePath: path,
		FileHash: hash,
		FileSize: size,
		Attrs:    attributes,
	})
	if err != nil {
		return err
	}

	// Construct a pipe stream: the reader side will be consumed and sent as the
	// body of the outgoing request, and the writer side we can use to
	// asynchronously populate it.
	pipeReader, pipeWriter := io.Pipe()

	bodyWriter := multipart.NewWriter(pipeWriter)
	go func() (err error) {
		defer func() {
			bodyWriter.Close()
			// If building the request failed, force the reader side to fail
			// so that ExecuteRequest returns the error. This behaves equivalent to
			// .Close() if err is nil.
			pipeWriter.CloseWithError(err)
		}()
		// Create the first part and write the metadata into it
		metaWriter, formErr := bodyWriter.CreateFormField(metadataParam)
		if err != nil {
			return fmt.Errorf("failed to create metadata multipart field: %s", formErr)
		}

		if _, metaErr := metaWriter.Write(metaPart); metaErr != nil {
			return fmt.Errorf("failed to write metadata to body: %s", metaErr)
		}

		// Create the second part, stream the file body into it, then close it.
		fileWriter, fileErr := bodyWriter.CreateFormField(fileParam)
		if fileErr != nil {
			return fmt.Errorf("failed to create metadata multipart field: %s", fileErr)
		}

		if _, copyErr := io.Copy(fileWriter, body); copyErr != nil {
			return fmt.Errorf("failed to write file to body: %s", copyErr)
		}
		return nil
	}()

	res, err := sc.ExecuteRequest(
		http.MethodPut,
		fmt.Sprintf(hostingAssetRoute, groupID, appID),
		RequestOptions{
			Body:   pipeReader,
			Header: http.Header{"Content-Type": {"multipart/mixed; boundary=" + bodyWriter.Boundary()}},
		},
	)
	return checkStatusNoContent(res, err, "failed to upload asset")
}

// SetAssetAttributes sets the asset at the given path to have the provided AssetAttributes
func (sc *basicStitchClient) SetAssetAttributes(groupID, appID, path string, attributes ...hosting.AssetAttribute) error {
	attrs, err := json.Marshal(setAttributesPayload{attributes})
	if err != nil {
		return err
	}

	res, err := sc.ExecuteRequest(
		http.MethodPatch,
		fmt.Sprintf(hostingAssetRoute+"?%s=%s", groupID, appID, pathParam, path),
		RequestOptions{
			Body: bytes.NewReader(attrs),
		},
	)
	return checkStatusNoContent(res, err, "failed to update asset")
}

// CopyAsset moves an asset from location fromPath to location toPath
func (sc *basicStitchClient) CopyAsset(groupID, appID, fromPath, toPath string) error {
	payload, err := json.Marshal(copyPayload{fromPath, toPath})
	if err != nil {
		return err
	}

	res, err := sc.invokePostRoute(groupID, appID, bytes.NewReader(payload))
	return checkStatusNoContent(res, err, "failed to copy asset")
}

// MoveAsset moves an asset from location fromPath to location toPath
func (sc *basicStitchClient) MoveAsset(groupID, appID, fromPath, toPath string) error {
	payload, err := json.Marshal(movePayload{fromPath, toPath})
	if err != nil {
		return err
	}

	res, err := sc.invokePostRoute(groupID, appID, bytes.NewReader(payload))
	return checkStatusNoContent(res, err, "failed to move asset")
}

func (sc *basicStitchClient) invokePostRoute(groupID, appID string, payload io.Reader) (*http.Response, error) {
	return sc.ExecuteRequest(
		http.MethodPost,
		fmt.Sprintf(hostingAssetsRoute, groupID, appID),
		RequestOptions{
			Body: payload,
		},
	)
}

// DeleteAsset deletes the asset at the given path
func (sc *basicStitchClient) DeleteAsset(groupID, appID, path string) error {
	res, err := sc.ExecuteRequest(
		http.MethodDelete,
		fmt.Sprintf(hostingAssetRoute+"?%s=%s", groupID, appID, pathParam, path),
		RequestOptions{},
	)
	return checkStatusNoContent(res, err, "failed to delete asset")
}

func (sc *basicStitchClient) findProjectAppByClientAppID(groupIDs []string, clientAppID string) (*models.App, error) {
	for _, groupID := range groupIDs {
		apps, err := sc.FetchAppsByGroupID(groupID)
		if err != nil && err != errGroupNotFound {
			return nil, err
		}

		if app := findAppByClientAppID(apps, clientAppID); app != nil {
			return app, nil
		}
	}

	return nil, ErrAppNotFound{clientAppID}
}

func (sc *basicStitchClient) CreateEmptyApp(groupID, appName, location, deploymentModel string) (*models.App, error) {
	res, err := sc.ExecuteRequest(
		http.MethodPost,
		fmt.Sprintf(appsByGroupIDRoute, groupID),
		RequestOptions{Body: strings.NewReader(fmt.Sprintf(`{"name":"%s","location":"%s","deployment_model":"%s"}`, appName, location, deploymentModel))},
	)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		return nil, UnmarshalStitchError(res)
	}

	dec := json.NewDecoder(res.Body)
	var app models.App
	if err := dec.Decode(&app); err != nil {
		return nil, err
	}

	return &app, nil
}

func (sc *basicStitchClient) ListAssetsForAppID(groupID, appID string) ([]hosting.AssetMetadata, error) {
	res, err := sc.ExecuteRequest(
		http.MethodGet,
		fmt.Sprintf(hostingAssetsRoute+"?recursive=true", groupID, appID),
		RequestOptions{},
	)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, UnmarshalStitchError(res)
	}

	dec := json.NewDecoder(res.Body)
	var assetMetadata []hosting.AssetMetadata
	if err := dec.Decode(&assetMetadata); err != nil {
		return nil, err
	}

	return assetMetadata, nil
}

// InvalidateCache requests cache invalidation for the resource at the given
// path in the app's CloudFront distribution
func (sc *basicStitchClient) InvalidateCache(groupID, appID, path string) error {
	payload, err := json.Marshal(invalidateCachePayload{Invalidate: true, Path: path})
	if err != nil {
		return err
	}

	res, err := sc.ExecuteRequest(
		http.MethodPut,
		fmt.Sprintf(hostingInvalidateCacheRoute, groupID, appID),
		RequestOptions{
			Body: bytes.NewReader(payload),
		},
	)
	return checkStatusNoContent(res, err, "failed to invalidate cache")
}

func checkStatusNoContent(res *http.Response, requestErr error, errMessage string) error {
	if requestErr != nil {
		return requestErr
	}
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("%s: %s: %s", res.Status, errMessage, UnmarshalStitchError(res))
	}
	return nil
}

func findAppByClientAppID(apps []*models.App, clientAppID string) *models.App {
	for _, app := range apps {
		if app.ClientAppID == clientAppID {
			return app
		}
	}

	return nil
}
