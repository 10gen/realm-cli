package testutils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/10gen/stitch-cli/api"
	"github.com/10gen/stitch-cli/auth"
	"github.com/10gen/stitch-cli/models"
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
func NewPopulatedStorage(apiKey, refreshToken, accessToken string) *storage.Storage {
	b, err := yaml.Marshal(user.User{
		APIKey:       apiKey,
		Username:     "user.name",
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
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

	tokenString := base64.StdEncoding.EncodeToString(tokenBytes)

	return fmt.Sprintf("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.%s.RuF0KMEBAalfnsdMeozpQLQ_2hK27l9omxtTp8eF1yI", tokenString)
}

// MockStitchClient satisfies an api.StitchClient
type MockStitchClient struct {
	FetchAppByClientAppIDFn func(clientAppID string) (*models.App, error)
	ExportFn                func(groupID, appID string) (string, io.ReadCloser, error)
	ImportFn                func(groupID, appID string) error
	ImportFnCalls           [][]string
}

// Authenticate will authenticate a user given an auth.AuthenticationProvider
func (msc *MockStitchClient) Authenticate(authProvider auth.AuthenticationProvider) (*auth.Response, error) {
	return nil, nil
}

// Export will download a Stitch app as a .zip
func (msc *MockStitchClient) Export(groupID, appID string) (string, io.ReadCloser, error) {
	if msc.ExportFn != nil {
		return msc.ExportFn(groupID, appID)
	}

	return "", nil, nil
}

// Diff will execute a dry-run of an import, returning a diff of proposed changes
func (msc *MockStitchClient) Diff(groupID, appID string, appData []byte) ([]string, error) {
	return []string{}, nil
}

// Import will push a local Stitch app to the server
func (msc *MockStitchClient) Import(groupID, appID string, appData []byte) error {
	if msc.ImportFn != nil {
		msc.ImportFnCalls = append(msc.ImportFnCalls, []string{groupID, appID})
		return msc.ImportFn(groupID, appID)
	}
	return nil
}

// FetchAppByClientAppID fetches a Stitch app given a clientAppID
func (msc *MockStitchClient) FetchAppByClientAppID(clientAppID string) (*models.App, error) {
	if msc.FetchAppByClientAppIDFn != nil {
		return msc.FetchAppByClientAppIDFn(clientAppID)
	}

	return nil, nil
}
