package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/10gen/stitch-cli/auth"
	"github.com/10gen/stitch-cli/models"
)

const (
	authProviderLoginRoute = adminBaseURL + "/auth/providers/%s/login"
	appExportRoute         = adminBaseURL + "/groups/%s/apps/%s/export?template=%t"
	appImportRoute         = adminBaseURL + "/groups/%s/apps/%s/import"
	appsByGroupIDRoute     = adminBaseURL + "/groups/%s/apps"
	userProfileRoute       = adminBaseURL + "/auth/profile"
)

var (
	errExportMissingFilename = errors.New("the app export response did not specify a filename")
	errGroupNotFound         = errors.New("group could not be found")
)

// ErrAppNotFound is used when an app cannot be found by client app ID
type ErrAppNotFound struct {
	ClientAppID string
}

func (eanf ErrAppNotFound) Error() string {
	return fmt.Sprintf("Unable to find app with ID: %q", eanf.ClientAppID)
}

// ErrStitchResponse represents a response from a Stitch API call
type ErrStitchResponse struct {
	data errStitchResponseData
}

// Error returns a stringified error message
func (esr ErrStitchResponse) Error() string {
	return fmt.Sprintf("error: %s", esr.data.Error)
}

// UnmarshalJSON unmarshals JSON data into an ErrStitchResponse
func (esr *ErrStitchResponse) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &esr.data)
}

type errStitchResponseData struct {
	Error string `json:"error"`
}

// UnmarshalStitchError unmarshals an *http.Response into an ErrStitchResponse. If the Body does not
// contain content it uses the provided Status
func UnmarshalStitchError(res *http.Response) error {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(res.Body); err != nil {
		return err
	}

	str := buf.String()
	if str == "" {
		return ErrStitchResponse{
			data: errStitchResponseData{
				Error: res.Status,
			},
		}
	}

	var stitchResponse ErrStitchResponse
	if err := json.NewDecoder(&buf).Decode(&stitchResponse); err != nil {
		stitchResponse.data.Error = str
	}

	return stitchResponse
}

// StitchClient represents a Client that can be used to call the Stitch Admin API
type StitchClient interface {
	Authenticate(authProvider auth.AuthenticationProvider) (*auth.Response, error)
	Export(groupID, appID string, isTemplated bool) (string, io.ReadCloser, error)
	Import(groupID, appID string, appData []byte, strategy string) error
	Diff(groupID, appID string, appData []byte, strategy string) ([]string, error)
	FetchAppByGroupIDAndClientAppID(groupID, clientAppID string) (*models.App, error)
	FetchAppByClientAppID(clientAppID string) (*models.App, error)
	FetchAppsByGroupID(groupID string) ([]*models.App, error)
	CreateEmptyApp(groupID, appName string) (*models.App, error)
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
func (sc *basicStitchClient) Export(groupID, appID string, isTemplated bool) (string, io.ReadCloser, error) {
	res, err := sc.ExecuteRequest(http.MethodGet, fmt.Sprintf(appExportRoute, groupID, appID, isTemplated), RequestOptions{})
	if err != nil {
		return "", nil, err
	}

	if res.StatusCode != http.StatusOK {
		defer res.Body.Close()
		return "", nil, UnmarshalStitchError(res)
	}

	_, params, err := mime.ParseMediaType(res.Header.Get("Content-Disposition"))
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

func (sc *basicStitchClient) CreateEmptyApp(groupID, appName string) (*models.App, error) {
	res, err := sc.ExecuteRequest(
		http.MethodPost,
		fmt.Sprintf(appsByGroupIDRoute, groupID),
		RequestOptions{Body: strings.NewReader(fmt.Sprintf(`{"name":"%s"}`, appName))},
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

func findAppByClientAppID(apps []*models.App, clientAppID string) *models.App {
	for _, app := range apps {
		if app.ClientAppID == clientAppID {
			return app
		}
	}

	return nil
}
