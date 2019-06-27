package testutils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/10gen/stitch-cli/api"
	"github.com/10gen/stitch-cli/api/mdbcloud"
	"github.com/10gen/stitch-cli/auth"
	"github.com/10gen/stitch-cli/hosting"
	"github.com/10gen/stitch-cli/models"
	"github.com/10gen/stitch-cli/secrets"
	"github.com/10gen/stitch-cli/storage"
	"github.com/10gen/stitch-cli/user"

	"github.com/smartystreets/goconvey/convey/gotest"
	"gopkg.in/yaml.v2"
)

// Assertion is a func that checks some condition for use in a test
type Assertion func(actual interface{}, expected ...interface{}) string

type failureView struct {
	Message  string `json:"Message"`
	Expected string `json:"Expected"`
	Actual   string `json:"Actual"`
}

// So runs an assertion and fails the test if necessary
func So(t *testing.T, actual interface{}, assert Assertion, expected ...interface{}) {
	t.Helper()
	file, line, _ := gotest.ResolveExternalCaller()
	if result := assert(actual, expected...); result != "" {
		fv := failureView{}
		err := json.Unmarshal([]byte(result), &fv)
		errMessage := result
		if err == nil {
			errMessage = fv.Message
		}
		formatted := fmt.Sprintf(
			"\n* %s\nLine %d:\n%s\n",
			file,
			line,
			errMessage,
		)
		t.Fatal(formatted)
	}
}

// NewMockClient returns a new MockClient
func NewMockClient(responses []*http.Response) *MockClient {
	return &MockClient{
		RequestData: []RequestData{},
		Responses:   responses,
	}
}

// A MockClient is a new api.Client that can be used to mock out HTTP requests and return responses
type MockClient struct {
	RequestData   []RequestData
	Responses     []*http.Response
	responseIndex int
}

// ExecuteRequest satisfies the api.Client interface, records request data, and returns the provided responses in order
func (mc *MockClient) ExecuteRequest(method, path string, options api.RequestOptions) (*http.Response, error) {
	mc.RequestData = append(mc.RequestData, RequestData{
		Method:  method,
		Path:    path,
		Options: options,
	})

	response := mc.Responses[mc.responseIndex]
	mc.responseIndex++

	return response, nil
}

// RequestData represents a given request made to the MockClient
type RequestData struct {
	Method  string
	Path    string
	Options api.RequestOptions
}

// ResponseBody is a io.ReadCloser that can be used as a net/http.Body
type ResponseBody struct {
	*bytes.Buffer
}

// Close satisfies the io.ReadCloser interface
func (ar *ResponseBody) Close() error {
	return nil
}

// NewAuthResponseBody returns a new ResponseBody populated with auth.Response data
func NewAuthResponseBody(data auth.Response) *ResponseBody {
	authResponseBytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	rb := ResponseBody{}
	rb.Buffer = bytes.NewBuffer(authResponseBytes)

	return &rb
}

// NewResponseBody returns a new ResponseBody
func NewResponseBody(data io.Reader) *ResponseBody {
	rb := ResponseBody{}
	b, err := ioutil.ReadAll(data)
	if err != nil {
		panic(err)
	}
	rb.Buffer = bytes.NewBuffer(b)

	return &rb
}

// NewEmptyStorage creates a new empty MemoryStrategy
func NewEmptyStorage() *storage.Storage {
	return storage.New(NewMemoryStrategy([]byte{}))
}

// NewPopulatedStorage creates a new MemoryStrategy populated with data
func NewPopulatedStorage(privateAPIKey, refreshToken, accessToken string) *storage.Storage {
	b, err := yaml.Marshal(user.User{
		PublicAPIKey:  "user.name",
		PrivateAPIKey: privateAPIKey,
		RefreshToken:  refreshToken,
		AccessToken:   accessToken,
	})
	if err != nil {
		panic(err)
	}

	return storage.New(NewMemoryStrategy(b))
}

// NewPopulatedDeprecatedStorage creates a new MemoryStrategy populated with data in the old deprecated format
func NewPopulatedDeprecatedStorage(username, apiKey string) *storage.Storage {
	b, err := yaml.Marshal(user.User{
		Username: username,
		APIKey:   apiKey,
	})
	if err != nil {
		panic(err)
	}

	return storage.New(NewMemoryStrategy(b))
}

// MemoryStrategy is a storage.Strategy that stores data in memory
type MemoryStrategy struct {
	data []byte
}

// Write records the provided data to memory storage
func (ms *MemoryStrategy) Write(data []byte) error {
	ms.data = data
	return nil
}

// Read reads the data currently stored in memory storage
func (ms *MemoryStrategy) Read() ([]byte, error) {
	return ms.data, nil
}

// NewMemoryStrategy returns a new MemoryStrategy
func NewMemoryStrategy(data []byte) *MemoryStrategy {
	return &MemoryStrategy{
		data: data,
	}
}

// GenerateValidAccessToken generates and returns a valid access token *from the future*
func GenerateValidAccessToken() string {
	token := auth.JWT{
		Exp: time.Now().Add(time.Hour).Unix(),
	}

	tokenBytes, err := json.Marshal(token)
	if err != nil {
		panic(err)
	}

	tokenString := base64.RawStdEncoding.EncodeToString(tokenBytes)

	return fmt.Sprintf("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.%s.RuF0KMEBAalfnsdMeozpQLQ_2hK27l9omxtTp8eF1yI", tokenString)
}

// MockStitchClient satisfies an api.StitchClient
type MockStitchClient struct {
	CreateEmptyAppFn                  func(groupID, appName, locationName, deploymentModelName string) (*models.App, error)
	FetchAppByGroupIDAndClientAppIDFn func(groupID, clientAppID string) (*models.App, error)
	FetchAppByClientAppIDFn           func(clientAppID string) (*models.App, error)
	FetchAppsByGroupIDFn              func(groupID string) ([]*models.App, error)
	ListAssetsForAppIDFn              func(groupID, appID string) ([]string, []hosting.AssetDescription, error)
	UploadAssetFn                     func(groupID, appID, path, hash string, size int64, body io.Reader, attributes ...hosting.AssetAttribute) error
	CopyAssetFn                       func(groupID, appID, fromPath, toPath string) error
	MoveAssetFn                       func(groupID, appID, fromPath, toPath string) error
	DeleteAssetFn                     func(groupID, appID, path string) error
	SetAssetAttributesFn              func(groupID, appID, path string, attributes ...hosting.AssetAttribute) error
	ExportFn                          func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error)
	ExportFnCalls                     [][]string
	ImportFn                          func(groupID, appID string, appData []byte, strategy string) error
	ImportFnCalls                     [][]string
	DiffFn                            func(groupID, appID string, appData []byte, strategy string) ([]string, error)
	InvalidateCacheFn                 func(groupID, appID, path string) error
	ListSecretsFn                     func(groupID, appID string) ([]secrets.Secret, error)
	AddSecretFn                       func(groupID, appID string, secret secrets.Secret) error
	UpdateSecretByIDFn                func(groupID, appID, secretID, secretValue string) error
	UpdateSecretByNameFn              func(groupID, appID, secretName, secretValue string) error
	RemoveSecretByIDFn                func(groupID, appID, secretID string) error
	RemoveSecretByNameFn              func(groupID, appID, secretName string) error
}

var _ api.StitchClient = (*MockStitchClient)(nil)

// Authenticate will authenticate a user given an auth.AuthenticationProvider
func (msc *MockStitchClient) Authenticate(authProvider auth.AuthenticationProvider) (*auth.Response, error) {
	return nil, nil
}

// Export will download a Stitch app as a .zip
func (msc *MockStitchClient) Export(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
	if msc.ExportFn != nil {
		msc.ExportFnCalls = append(msc.ExportFnCalls, []string{groupID, appID, string(strategy)})
		return msc.ExportFn(groupID, appID, strategy)
	}

	return "", nil, nil
}

// CreateDraft returns a mock AppDraft
func (msc *MockStitchClient) CreateDraft(groupID, appID string) (*models.AppDraft, error) {
	return &models.AppDraft{ID: "draft-id"}, nil
}

// DeployDraft returns a mock Deployment
func (msc *MockStitchClient) DeployDraft(groupID, appID, draftID string) (*models.Deployment, error) {
	return &models.Deployment{ID: "deployment-id"}, nil
}

// DiscardDraft does nothing
func (msc *MockStitchClient) DiscardDraft(groupID, appID, draftID string) error {
	return nil
}

// DraftDiff returns an empty DraftDiff
func (msc *MockStitchClient) DraftDiff(groupID, appID, draftID string) (*models.DraftDiff, error) {
	return &models.DraftDiff{}, nil
}

// GetDeployment returns a mock Deployment
func (msc *MockStitchClient) GetDeployment(groupID, appID, deploymentID string) (*models.Deployment, error) {
	return &models.Deployment{ID: "deployment-id"}, nil
}

// GetDrafts returns an empty list of AppDrafts
func (msc *MockStitchClient) GetDrafts(groupID, appID string) ([]models.AppDraft, error) {
	return []models.AppDraft{}, nil
}

// Diff will execute a dry-run of an import, returning a diff of proposed changes
func (msc *MockStitchClient) Diff(groupID, appID string, appData []byte, strategy string) ([]string, error) {
	if msc.DiffFn != nil {
		return msc.DiffFn(groupID, appID, appData, strategy)
	}

	return []string{}, nil
}

// FetchAppsByGroupID does nothing
func (msc *MockStitchClient) FetchAppsByGroupID(groupID string) ([]*models.App, error) {
	if msc.FetchAppsByGroupIDFn != nil {
		return msc.FetchAppsByGroupIDFn(groupID)
	}

	return nil, errors.New("someone should test me")
}

// CreateEmptyApp does nothing
func (msc *MockStitchClient) CreateEmptyApp(groupID, appName, locationName, deploymentModelName string) (*models.App, error) {
	if msc.CreateEmptyAppFn != nil {
		return msc.CreateEmptyAppFn(groupID, appName, locationName, deploymentModelName)
	}

	return nil, errors.New("someone should test me")
}

// Import will push a local Stitch app to the server
func (msc *MockStitchClient) Import(groupID, appID string, appData []byte, strategy string) error {
	if msc.ImportFn != nil {
		msc.ImportFnCalls = append(msc.ImportFnCalls, []string{groupID, appID})
		return msc.ImportFn(groupID, appID, appData, strategy)
	}
	return nil
}

// FetchAppByGroupIDAndClientAppID fetches a Stitch app given a groupID and clientAppID
func (msc *MockStitchClient) FetchAppByGroupIDAndClientAppID(groupID, clientAppID string) (*models.App, error) {
	if msc.FetchAppByGroupIDAndClientAppIDFn != nil {
		return msc.FetchAppByGroupIDAndClientAppIDFn(groupID, clientAppID)
	}

	return nil, api.ErrAppNotFound{clientAppID}
}

// FetchAppByClientAppID fetches a Stitch app given a clientAppID
func (msc *MockStitchClient) FetchAppByClientAppID(clientAppID string) (*models.App, error) {
	if msc.FetchAppByClientAppIDFn != nil {
		return msc.FetchAppByClientAppIDFn(clientAppID)
	}

	return nil, api.ErrAppNotFound{clientAppID}
}

// UploadAsset uploads an asset
func (msc *MockStitchClient) UploadAsset(groupID, appID, path, hash string, size int64, body io.Reader, attributes ...hosting.AssetAttribute) error {
	if msc.UploadAssetFn != nil {
		return msc.UploadAssetFn(groupID, appID, path, hash, size, body, attributes...)
	}

	return nil
}

// CopyAsset copies an asset
func (msc *MockStitchClient) CopyAsset(groupID, appID, fromPath, toPath string) error {
	if msc.CopyAssetFn != nil {
		return msc.CopyAssetFn(groupID, appID, fromPath, toPath)
	}

	return nil
}

// MoveAsset moves an asset
func (msc *MockStitchClient) MoveAsset(groupID, appID, fromPath, toPath string) error {
	if msc.MoveAssetFn != nil {
		return msc.MoveAssetFn(groupID, appID, fromPath, toPath)
	}

	return nil
}

// DeleteAsset deletes an asset
func (msc *MockStitchClient) DeleteAsset(groupID, appID, path string) error {
	if msc.DeleteAssetFn != nil {
		return msc.DeleteAssetFn(groupID, appID, path)
	}

	return nil
}

// SetAssetAttributes sets an asset's attributes
func (msc *MockStitchClient) SetAssetAttributes(groupID, appID, path string, attributes ...hosting.AssetAttribute) error {
	if msc.SetAssetAttributesFn != nil {
		return msc.SetAssetAttributesFn(groupID, appID, path, attributes...)
	}

	return nil
}

// ListAssetsForAppID fetches a Stitch app given a clientAppID
func (msc *MockStitchClient) ListAssetsForAppID(groupID, appID string) ([]hosting.AssetMetadata, error) {
	assetMetadata := []hosting.AssetMetadata{
		{
			FilePath: "/bar/shouldRemainSame.txt",
			URL:      "URL/bar/shouldRemainSame.txt",
			Attrs: []hosting.AssetAttribute{
				{Name: "Content-Type", Value: "html"},
			},
		},
		{
			FilePath: "/bar/shouldBeRemoved.txt",
			URL:      "URL/bar/shouldBeRemoved.txt",
			Attrs: []hosting.AssetAttribute{
				{Name: "Content-Type", Value: "text/plain"},
			},
		},
		{
			FilePath: "/bar/attrsShouldAllRemain.html",
			URL:      "URL/bar/attrsShouldAllRemain.html",
			Attrs: []hosting.AssetAttribute{
				{Name: "Content-Disposition", Value: "inline"},
				{Name: "Content-Type", Value: "htmp"},
				{Name: "Content-Language", Value: "fr"},
				{Name: "Content-Encoding", Value: "utf-8"},
				{Name: "Cache-Control", Value: "true"},
			},
		},
		{
			FilePath: "/bar/attrsShouldRemoveAllButOne.html",
			URL:      "URL/bar/attrsShouldAllRemain.html",
			Attrs: []hosting.AssetAttribute{
				{Name: "Content-Disposition", Value: "inline"},
				{Name: "content-type", Value: "htmp"},
				{Name: "content-language", Value: "fr"},
			},
		},
		{
			FilePath: "/bar/shouldBeRemoved.html",
			URL:      "URL/bar/shouldBeRemoved.html",
			Attrs:    []hosting.AssetAttribute{},
		},
		{
			FilePath: "/bar/shouldBeRemoved",
			URL:      "URL/bar/shouldBeRemoved",
			Attrs: []hosting.AssetAttribute{
				{Name: "Content-Type", Value: "htmp"},
				{Name: "Content-Language", Value: "fr"},
			},
		},
	}
	return assetMetadata, nil
}

// InvalidateCache requests cache invalidation for the asset at the argued path
func (msc *MockStitchClient) InvalidateCache(groupID, appID, path string) error {
	if msc.InvalidateCacheFn != nil {
		return msc.InvalidateCacheFn(groupID, appID, path)
	}

	return nil
}

// ListSecrets lists the secrets of an app
func (msc *MockStitchClient) ListSecrets(groupID, appID string) ([]secrets.Secret, error) {
	if msc.ListSecretsFn != nil {
		return msc.ListSecretsFn(groupID, appID)
	}

	return nil, nil
}

// AddSecret adds a secret to the app
func (msc *MockStitchClient) AddSecret(groupID, appID string, secret secrets.Secret) error {
	if msc.AddSecretFn != nil {
		return msc.AddSecretFn(groupID, appID, secret)
	}

	return nil
}

// UpdateSecretByID updates a secret from the app
func (msc *MockStitchClient) UpdateSecretByID(groupID, appID, secretID, secretValue string) error {
	if msc.UpdateSecretByIDFn != nil {
		return msc.UpdateSecretByIDFn(groupID, appID, secretID, secretValue)
	}

	return nil
}

// UpdateSecretByName updates a secret from the app
func (msc *MockStitchClient) UpdateSecretByName(groupID, appID, secretName, secretValue string) error {
	if msc.UpdateSecretByNameFn != nil {
		return msc.UpdateSecretByNameFn(groupID, appID, secretName, secretValue)
	}

	return nil
}

// RemoveSecretByID removes a secret from the app
func (msc *MockStitchClient) RemoveSecretByID(groupID, appID, secretID string) error {
	if msc.RemoveSecretByIDFn != nil {
		return msc.RemoveSecretByIDFn(groupID, appID, secretID)
	}

	return nil
}

// RemoveSecretByName removes a secret from the app
func (msc *MockStitchClient) RemoveSecretByName(groupID, appID, secretName string) error {
	if msc.RemoveSecretByNameFn != nil {
		return msc.RemoveSecretByNameFn(groupID, appID, secretName)
	}

	return nil
}

// MockMDBClient satisfies a mdbcloud.Client
type MockMDBClient struct {
	WithAuthFn           func(username, apiKey string) mdbcloud.Client
	GroupsFn             func() ([]mdbcloud.Group, error)
	GroupByNameFn        func(string) (*mdbcloud.Group, error)
	DeleteDatabaseUserFn func(groupId, username string) error
}

// WithAuth will authenticate a user given username and apiKey
func (mmc MockMDBClient) WithAuth(username, apiKey string) mdbcloud.Client {
	return nil
}

// Groups will return a list of groups available
func (mmc *MockMDBClient) Groups() ([]mdbcloud.Group, error) {
	if mmc.GroupsFn != nil {
		return mmc.GroupsFn()
	}
	return nil, errors.New("someone should test me")
}

// GroupByName will look up the Group given a name
func (mmc *MockMDBClient) GroupByName(groupName string) (*mdbcloud.Group, error) {
	if mmc.GroupByNameFn != nil {
		return mmc.GroupByNameFn(groupName)
	}
	return nil, errors.New("someone should test me")
}

// DeleteDatabaseUser does nothing
func (mmc *MockMDBClient) DeleteDatabaseUser(groupID, username string) error {
	if mmc.DeleteDatabaseUserFn != nil {
		return mmc.DeleteDatabaseUserFn(groupID, username)
	}
	return nil
}

// MongoDBCloudEnv represents ENV variables required for running tests against cloud
type MongoDBCloudEnv struct {
	CloudAPIBaseURL     string
	StitchServerBaseURL string
	APIKey              string
	Username            string
	AdminUsername       string
	AdminAPIKey         string
	GroupID             string
}

// ENV returns the current MongoDBCloudEnv configuration
func ENV() MongoDBCloudEnv {
	defaultServerURL := "http://localhost:9090"

	cloudAPIBaseURL := os.Getenv("STITCH_MONGODB_CLOUD_API_BASE_URL")
	if cloudAPIBaseURL == "" {
		cloudAPIBaseURL = defaultServerURL
	}

	stitchServerBaseURL := os.Getenv("STITCH_SERVER_BASE_URL")
	if stitchServerBaseURL == "" {
		stitchServerBaseURL = defaultServerURL
	}

	return MongoDBCloudEnv{
		CloudAPIBaseURL:     cloudAPIBaseURL,
		StitchServerBaseURL: stitchServerBaseURL,
		APIKey:              os.Getenv("STITCH_MONGODB_CLOUD_API_KEY"),
		Username:            os.Getenv("STITCH_MONGODB_CLOUD_USERNAME"),
		GroupID:             os.Getenv("STITCH_MONGODB_CLOUD_GROUP_ID"),
		AdminUsername:       os.Getenv("STITCH_MONGODB_CLOUD_ADMIN_USERNAME"),
		AdminAPIKey:         os.Getenv("STITCH_MONGODB_CLOUD_ADMIN_API_KEY"),
	}
}

var mongoDBCloudNotRunning = false

// MustSkipf skips a test suite, but panics if STITCH_NO_SKIP_TEST is set, indicating
// that skipping is not permitted.
func MustSkipf(t *testing.T, format string, args ...interface{}) {
	if len(os.Getenv("STITCH_NO_SKIP_TEST")) > 0 {
		panic("test was skipped, but STITCH_NO_SKIP_TEST is set.")
	}
	t.Skipf(format, args...)
}

// SkipUnlessMongoDBCloudRunning skips tests if there is no cloud instance running at
// the chosen base URL
var SkipUnlessMongoDBCloudRunning = func() func(t *testing.T) {
	return func(t *testing.T) {
		cloudEnv := ENV()

		if mongoDBCloudNotRunning {
			MustSkipf(t, "MongoDB Cloud not running at %s", cloudEnv.CloudAPIBaseURL)
			return
		}
		req, err := http.NewRequest(http.MethodGet, cloudEnv.CloudAPIBaseURL, nil)
		if err != nil {
			panic(err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			MustSkipf(t, "MongoDB Cloud not running at %s", cloudEnv.CloudAPIBaseURL)
			return
		}
	}
}()
