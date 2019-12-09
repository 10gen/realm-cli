package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/10gen/stitch-cli/auth"
	"github.com/10gen/stitch-cli/hosting"
	"github.com/10gen/stitch-cli/models"
	"github.com/10gen/stitch-cli/secrets"
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

	userProfileRoute       = adminBaseURL + "/auth/profile"
	authProviderLoginRoute = adminBaseURL + "/auth/providers/%s/login"

	appsByGroupIDRoute      = adminBaseURL + "/groups/%s/apps"
	atlasAppsByGroupIDRoute = appsByGroupIDRoute + "?product=atlas"
	appImportRoute          = adminBaseURL + "/groups/%s/apps/%s/import"
	appExportRoute          = adminBaseURL + "/groups/%s/apps/%s/export?%s"

	draftsRoute      = adminBaseURL + "/groups/%s/apps/%s/drafts"
	draftByIDRoute   = adminBaseURL + "/groups/%s/apps/%s/drafts/%s"
	deployDraftRoute = adminBaseURL + "/groups/%s/apps/%s/drafts/%s/deployment"
	diffDraftRoute   = adminBaseURL + "/groups/%s/apps/%s/drafts/%s/diff"

	deploymentByIDRoute = adminBaseURL + "/groups/%s/apps/%s/deployments/%s"

	hostingInvalidateCacheRoute = adminBaseURL + "/groups/%s/apps/%s/hosting/cache"
	hostingAssetsRoute          = adminBaseURL + "/groups/%s/apps/%s/hosting/assets"
	hostingAssetRoute           = adminBaseURL + "/groups/%s/apps/%s/hosting/assets/asset"

	secretsRoute = adminBaseURL + "/groups/%s/apps/%s/secrets"
	secretRoute  = adminBaseURL + "/groups/%s/apps/%s/secrets/%s"

	dependenciesRoute = adminBaseURL + "/groups/%s/apps/%s/dependencies"
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
	AddSecret(groupID, appID string, secret secrets.Secret) error
	Authenticate(authProvider auth.AuthenticationProvider) (*auth.Response, error)
	CopyAsset(groupID, appID, fromPath, toPath string) error
	CreateDraft(groupID, appID string) (*models.AppDraft, error)
	CreateEmptyApp(groupID, appName, location, deploymentModel string) (*models.App, error)
	DeleteAsset(groupID, appID, path string) error
	DeployDraft(groupID, appID, draftID string) (*models.Deployment, error)
	Diff(groupID, appID string, appData []byte, strategy string) ([]string, error)
	DiscardDraft(groupID, appID, draftID string) error
	DraftDiff(groupID, appID, draftID string) (*models.DraftDiff, error)
	Export(groupID, appID string, strategy ExportStrategy) (string, io.ReadCloser, error)
	FetchAppByClientAppID(clientAppID string) (*models.App, error)
	FetchAppByGroupIDAndClientAppID(groupID, clientAppID string) (*models.App, error)
	FetchAppsByGroupID(groupID string) ([]*models.App, error)
	GetDeployment(groupID, appID, deploymentID string) (*models.Deployment, error)
	GetDrafts(groupID, appID string) ([]models.AppDraft, error)
	Import(groupID, appID string, appData []byte, strategy string) error
	InvalidateCache(groupID, appID, path string) error
	ListAssetsForAppID(groupID, appID string) ([]hosting.AssetMetadata, error)
	ListSecrets(groupID, appID string) ([]secrets.Secret, error)
	MoveAsset(groupID, appID, fromPath, toPath string) error
	RemoveSecretByID(groupID, appID, secretID string) error
	RemoveSecretByName(groupID, appID, secretName string) error
	SetAssetAttributes(groupID, appID, path string, attributes ...hosting.AssetAttribute) error
	UpdateSecretByID(groupID, appID, secretID, secretValue string) error
	UpdateSecretByName(groupID, appID, secretName, secretValue string) error
	UploadAsset(groupID, appID, path, hash string, size int64, body io.Reader, attributes ...hosting.AssetAttribute) error
	UploadDependencies(groupID, appID, fullPath string) error
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

func (sc *basicStitchClient) CreateDraft(groupID, appID string) (*models.AppDraft, error) {
	res, err := sc.ExecuteRequest(http.MethodPost, fmt.Sprintf(draftsRoute, groupID, appID), RequestOptions{})
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		return nil, UnmarshalStitchError(res)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var draft models.AppDraft
	err = json.Unmarshal(bytes, &draft)
	if err != nil {
		return nil, err
	}

	return &draft, nil
}

func (sc *basicStitchClient) DeployDraft(groupID, appID, draftID string) (*models.Deployment, error) {
	res, err := sc.ExecuteRequest(http.MethodPost, fmt.Sprintf(deployDraftRoute, groupID, appID, draftID), RequestOptions{})
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		return nil, UnmarshalStitchError(res)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var deployment models.Deployment
	err = json.Unmarshal(bytes, &deployment)
	if err != nil {
		return nil, err
	}

	return &deployment, nil
}

func (sc *basicStitchClient) DiscardDraft(groupID, appID, draftID string) error {
	res, err := sc.ExecuteRequest(http.MethodDelete, fmt.Sprintf(draftByIDRoute, groupID, appID, draftID), RequestOptions{})
	if err != nil {
		return nil
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		return UnmarshalStitchError(res)
	}

	return nil
}

func (sc *basicStitchClient) GetDeployment(groupID, appID, deploymentID string) (*models.Deployment, error) {
	res, err := sc.ExecuteRequest(http.MethodGet, fmt.Sprintf(deploymentByIDRoute, groupID, appID, deploymentID), RequestOptions{})
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, UnmarshalStitchError(res)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var deployment models.Deployment
	err = json.Unmarshal(bytes, &deployment)
	if err != nil {
		return nil, err
	}

	return &deployment, nil
}

func (sc *basicStitchClient) GetDrafts(groupID, appID string) ([]models.AppDraft, error) {
	res, err := sc.ExecuteRequest(http.MethodGet, fmt.Sprintf(draftsRoute, groupID, appID), RequestOptions{})
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, UnmarshalStitchError(res)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var drafts []models.AppDraft
	err = json.Unmarshal(bytes, &drafts)
	if err != nil {
		return nil, err
	}

	return drafts, nil
}

func (sc *basicStitchClient) DraftDiff(groupID, appID, draftID string) (*models.DraftDiff, error) {
	res, err := sc.ExecuteRequest(http.MethodGet, fmt.Sprintf(diffDraftRoute, groupID, appID, draftID), RequestOptions{})
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, UnmarshalStitchError(res)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var diff models.DraftDiff
	err = json.Unmarshal(bytes, &diff)
	if err != nil {
		return nil, err
	}

	return &diff, nil
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
	return sc.fetchAppsByGropuIDFromEndpoint(groupID, appsByGroupIDRoute)
}

func (sc *basicStitchClient) FetchAtlasAppsByGroupID(groupID string) ([]*models.App, error) {
	return sc.fetchAppsByGropuIDFromEndpoint(groupID, atlasAppsByGroupIDRoute)
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
	errChan := make(chan error, 2)
	go func() {
		defer func() {
			bodyWriter.Close()
			// If building the request failed, force the reader side to fail
			// so that ExecuteRequest returns the error. This behaves equivalent to
			// .Close() if err is nil.
			errChan <- pipeWriter.CloseWithError(err)
		}()
		// Create the first part and write the metadata into it
		metaWriter, formErr := bodyWriter.CreateFormField(metadataParam)
		if err != nil {
			errChan <- fmt.Errorf("failed to create metadata multipart field: %s", formErr)
		}

		if _, metaErr := metaWriter.Write(metaPart); metaErr != nil {
			errChan <- fmt.Errorf("failed to write metadata to body: %s", metaErr)
		}

		// Create the second part, stream the file body into it, then close it.
		fileWriter, fileErr := bodyWriter.CreateFormField(fileParam)
		if fileErr != nil {
			errChan <- fmt.Errorf("failed to create metadata multipart field: %s", fileErr)
		}

		if _, copyErr := io.Copy(fileWriter, body); copyErr != nil {
			errChan <- fmt.Errorf("failed to write file to body: %s", copyErr)
		}
		errChan <- nil
	}()

	res, err := sc.ExecuteRequest(
		http.MethodPut,
		fmt.Sprintf(hostingAssetRoute, groupID, appID),
		RequestOptions{
			Body:   pipeReader,
			Header: http.Header{"Content-Type": {"multipart/mixed; boundary=" + bodyWriter.Boundary()}},
		},
	)
	if err := <-errChan; err != nil {
		return err
	}
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

		// Check if the clientAppID is referring to an Atlas trigger
		apps, err = sc.FetchAtlasAppsByGroupID(groupID)
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

// ListSecrets list secrets for the app
func (sc *basicStitchClient) ListSecrets(groupID, appID string) ([]secrets.Secret, error) {
	res, err := sc.ExecuteRequest(
		http.MethodGet,
		fmt.Sprintf(secretsRoute, groupID, appID),
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
	var secrets []secrets.Secret
	if err := dec.Decode(&secrets); err != nil {
		return nil, err
	}

	return secrets, nil
}

// AddSecret creates a secret for the app
func (sc *basicStitchClient) AddSecret(groupID, appID string, secret secrets.Secret) error {
	payload, err := json.Marshal(secret)
	if err != nil {
		return err
	}

	res, err := sc.ExecuteRequest(
		http.MethodPost,
		fmt.Sprintf(secretsRoute, groupID, appID),
		RequestOptions{
			Body: bytes.NewReader(payload),
		},
	)
	if err != nil {
		return nil
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		return UnmarshalStitchError(res)
	}

	return nil
}

// UpdateSecretByID updates a secret's value from the app
func (sc *basicStitchClient) UpdateSecretByID(groupID, appID, secretID, secretValue string) error {
	appSecrets, err := sc.ListSecrets(groupID, appID)
	if err != nil {
		return err
	}

	var secretToUpdate *secrets.Secret
	for _, s := range appSecrets {
		if s.ID == secretID {
			secretToUpdate = &s
			break
		}
	}

	if secretToUpdate == nil {
		return fmt.Errorf("secret not found: %s", secretID)
	}

	secretToUpdate.Value = secretValue
	payload, err := json.Marshal(secretToUpdate)
	if err != nil {
		return err
	}

	res, err := sc.ExecuteRequest(
		http.MethodPut,
		fmt.Sprintf(secretRoute, groupID, appID, secretID),
		RequestOptions{
			Body: bytes.NewReader(payload),
		},
	)
	if err != nil {
		return err
	}

	return checkStatusNoContent(res, err, "failed to update secret")
}

// UpdateSecretByName updates a secret's value from the app
func (sc *basicStitchClient) UpdateSecretByName(groupID, appID, secretName, secretValue string) error {
	appSecrets, err := sc.ListSecrets(groupID, appID)
	if err != nil {
		return err
	}

	var secretToUpdate *secrets.Secret
	for _, s := range appSecrets {
		if s.Name == secretName {
			secretToUpdate = &s
			break
		}
	}

	if secretToUpdate == nil {
		return fmt.Errorf("secret not found: %s", secretName)
	}

	secretToUpdate.Value = secretValue
	payload, err := json.Marshal(secretToUpdate)
	if err != nil {
		return err
	}

	res, err := sc.ExecuteRequest(
		http.MethodPut,
		fmt.Sprintf(secretRoute, groupID, appID, secretToUpdate.ID),
		RequestOptions{
			Body: bytes.NewReader(payload),
		},
	)
	if err != nil {
		return err
	}

	return checkStatusNoContent(res, err, "failed to update secret")
}

// RemoveSecretByID deletes a secret from the app
func (sc *basicStitchClient) RemoveSecretByID(groupID, appID, secretID string) error {
	res, err := sc.ExecuteRequest(
		http.MethodDelete,
		fmt.Sprintf(secretRoute, groupID, appID, secretID),
		RequestOptions{},
	)
	if err != nil {
		return err
	}

	return checkStatusNoContent(res, err, "failed to remove secret")
}

// RemoveSecretByName deletes a secret from the app
func (sc *basicStitchClient) RemoveSecretByName(groupID, appID, secretName string) error {
	secrets, err := sc.ListSecrets(groupID, appID)
	if err != nil {
		return err
	}

	var secretID string
	for _, s := range secrets {
		if s.Name == secretName {
			secretID = s.ID
			break
		}
	}

	if secretID == "" {
		return fmt.Errorf("secret not found: %s", secretName)
	}

	return sc.RemoveSecretByID(groupID, appID, secretID)
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

func (sc *basicStitchClient) UploadDependencies(groupID, appID, fullPath string) error {
	body, formatDataContentType, err := newMultipartMessage(fullPath)
	if err != nil {
		return err
	}

	res, err := sc.ExecuteRequest(
		http.MethodPost,
		fmt.Sprintf(dependenciesRoute, groupID, appID),
		RequestOptions{
			Body:   body,
			Header: http.Header{"Content-Type": {formatDataContentType}},
		},
	)

	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return UnmarshalStitchError(res)
	}
	return nil
}

func newMultipartMessage(fullPath string) (io.Reader, string, error) {
	file, openErr := os.Open(fullPath)
	if openErr != nil {
		return nil, "", fmt.Errorf("failed to open the dependencies file '%s': %s", fullPath, openErr)
	}
	defer file.Close()
	fileInfo, statErr := file.Stat()
	if statErr != nil {
		return nil, "", errors.New("failed to grab the dependencies file info")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	defer writer.Close()
	part, formErr := writer.CreateFormFile(fileParam, fileInfo.Name())
	if formErr != nil {
		return nil, "", fmt.Errorf("failed to create multipart form file: %s", formErr)
	}

	_, copyErr := io.Copy(part, file)
	if copyErr != nil {
		return nil, "", fmt.Errorf("failed to write file to body: %s", copyErr)
	}

	return body, writer.FormDataContentType(), nil
}

func (sc *basicStitchClient) fetchAppsByGropuIDFromEndpoint(groupID string, endpoint string) ([]*models.App, error) {
	res, err := sc.ExecuteRequest(http.MethodGet, fmt.Sprintf(endpoint, groupID), RequestOptions{})
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
