package testutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/10gen/stitch-cli/api"
	"github.com/10gen/stitch-cli/auth"
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
