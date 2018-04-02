package harness

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/10gen/stitch-cli/utils"
	"github.com/10gen/stitch-cli/utils/test"
)

// CloudPrivateAPIClient is a client for interacting with MongoDB Cloud's
// private API used by the UI
type CloudPrivateAPIClient struct {
	t          *testing.T
	httpClient *http.Client

	username  *string
	password  *string
	groupName *string
	groupID   *string
	loggedIn  bool
	csrfToken *string
	csrfTime  *string
}

// NewCloudPrivateAPIClient returns a new CloudPrivateAPIClient that
// has not yet logged in
func NewCloudPrivateAPIClient(t *testing.T) *CloudPrivateAPIClient {
	httpClient := http.Client{}
	httpClient.Jar, _ = cookiejar.New(nil)

	return &CloudPrivateAPIClient{
		t:          t,
		httpClient: &httpClient,
	}
}

func (client *CloudPrivateAPIClient) do(
	method string,
	url string,
	body io.Reader,
	headers http.Header,
) (*http.Response, error) {

	baseURL := testutils.MongoDBCloudPrivateAPIBaseURL()

	req, err := http.NewRequest(method, baseURL+url, body)
	if err != nil {
		return nil, err
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	if client.csrfToken != nil && client.csrfTime != nil {
		req.Header.Add("X-CSRF-Token", *client.csrfToken)
		req.Header.Add("X-CSRF-Time", *client.csrfTime)
	}

	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	client.t.Logf("Executing request method='%v' url='%v'", method, baseURL+url)
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Do executes a given method and url against the API
func (client *CloudPrivateAPIClient) Do(method string, url string) (*http.Response, error) {
	return client.do(method, url, nil, nil)
}

// DoJSON executes a given method and url against the API with JSON attached
func (client *CloudPrivateAPIClient) DoJSON(method string, url string, body interface{}) (*http.Response, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return client.do(method, url, bytes.NewReader(bodyBytes), nil)
}

// PostForm posts the given form to url against the API
func (client *CloudPrivateAPIClient) PostForm(url string, form url.Values) (*http.Response, error) {

	baseURL := testutils.MongoDBCloudPrivateAPIBaseURL()

	req, err := http.NewRequest(http.MethodPost, baseURL+url, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	if client.csrfToken != nil && client.csrfTime != nil {
		req.Header.Add("X-CSRF-Token", *client.csrfToken)
		req.Header.Add("X-CSRF-Time", *client.csrfTime)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client.t.Logf("Executing request method='%v' url='%v' form='%v'", http.MethodPost, baseURL+url, form)
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// RegisterUser registers a new user and associates it with this client
func (client *CloudPrivateAPIClient) RegisterUser() error {
	username := fmt.Sprintf("test-%s@domain.com", utils.RandomAlphaNumericString(12))
	password := "PassWord101#"
	registerPayload := map[string]interface{}{
		"username":          username,
		"password":          password,
		"firstName":         "FirstName",
		"lastName":          "LastName",
		"company":           "MyCompany",
		"newGroup":          true,
		"jobResponsibility": "IT Executive (CIO, CTO, VP Engineering, etc.)",
		"phoneNumber":       "867-5309",
	}

	resp, err := client.DoJSON(http.MethodPost, "/user/registerCall", registerPayload)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to register user (username=%q password=%q): %v", username, password, resp)
	}

	client.loggedIn = true
	client.username = &username
	client.password = &password

	return nil
}

// PlanType represents a type of plan within MongoDB Cloud
type PlanType uint8

// The set of known plan types
const (
	PlanTypeClassic PlanType = iota
	PlanTypeBasic
	PlanTypePremium
	PlanTypeStandard
	PlanTypeFreeTier
	PlanTypeOnPrem
	PlanTypeNDS
)

// String returns the API name of the plan
func (p PlanType) String() string {
	return planTypeToString[p]
}

var planTypeToString = map[PlanType]string{
	PlanTypeClassic:  "CLASSIC",
	PlanTypeBasic:    "BASIC",
	PlanTypePremium:  "PREMIUM",
	PlanTypeStandard: "STANDARD",
	PlanTypeFreeTier: "FREETIER",
	PlanTypeOnPrem:   "ONPREM",
	PlanTypeNDS:      "NDS",
}

// CreateGroup creates a group owned by the current user of the given plan type
func (client *CloudPrivateAPIClient) CreateGroup(
	planType PlanType, // nolint: interfacer
) error {
	groupName := fmt.Sprintf("test-%s", utils.RandomAlphaNumericString(12))

	form := url.Values{}
	form.Add("company", groupName)
	form.Add("groupType", planType.String())

	resp, err := client.PostForm("/user/addGroupSubmit", form)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create group: %v", resp)
	}

	var resM map[string]interface{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&resM); err != nil {
		return err
	}

	client.groupName = &groupName
	groupID := resM["newObjId"].(string)
	client.groupID = &groupID

	return client.captureCSRFToken()
}

func (client *CloudPrivateAPIClient) captureCSRFToken() error {
	resp, err := client.Do(http.MethodGet, "/")
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to capture CSRF token: %v", resp)
	}

	rd, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	pat := regexp.MustCompile("meta name=\"csrf-token\" content=\"([a-z0-9]+)\"")
	matches := pat.FindStringSubmatch(string(rd))
	if len(matches) != 2 {
		return errors.New("CSRF token not found")
	}
	client.csrfToken = &matches[1]

	pat = regexp.MustCompile("meta name=\"csrf-time\" content=\"([0-9]+)\"")
	matches = pat.FindStringSubmatch(string(rd))
	if len(matches) != 2 {
		return errors.New("CSRF time not found")
	}
	client.csrfTime = &matches[1]

	return nil
}

func (client *CloudPrivateAPIClient) refreshPasswordAuth() error {

	form := url.Values{}
	form.Add("password", client.Password())

	resp, err := client.PostForm("/user/checkPassword", form)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to refresh password auth: %v", resp)
	}

	return nil
}

// CreateAPIKey creates an API key for the current user
func (client *CloudPrivateAPIClient) CreateAPIKey() (string, string, error) {

	if err := client.refreshPasswordAuth(); err != nil {
		return "", "", err
	}

	form := url.Values{}
	form.Add("desc", utils.RandomAlphaNumericString(12))

	resp, err := client.PostForm("/settings/addPublicApiKey", form)
	if err != nil {
		return "", "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("failed to create API key: %v", resp)
	}

	var apiKey map[string]interface{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&apiKey); err != nil {
		return "", "", err
	}

	return apiKey["id"].(string), apiKey["key"].(string), nil
}

// DisableAPIKey disables the given API key
func (client *CloudPrivateAPIClient) DisableAPIKey(id string) error {

	if err := client.refreshPasswordAuth(); err != nil {
		return err
	}

	resp, err := client.Do(http.MethodPut, fmt.Sprintf("/settings/disablePublicApiKey/%s", id))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to disable API key: %v", resp)
	}

	return nil
}

// AddUserToGroup adds the given user to the group belonging to the current user
// with the given role
func (client *CloudPrivateAPIClient) AddUserToGroup(username string, groupID string, role string) error {

	if err := client.refreshPasswordAuth(); err != nil {
		return err
	}

	form := url.Values{}
	form.Add("username", username)
	form.Add("role", role)

	resp, err := client.PostForm(fmt.Sprintf("/user/addUser/%s", groupID), form)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to add user to group: %v", resp)
	}

	return nil
}

// Username returns the currently logged in user's username
func (client *CloudPrivateAPIClient) Username() string {
	if client.username == nil {
		client.t.Log("must log into cloud before getting username")
		client.t.FailNow()
		return ""
	}

	return *client.username
}

// Password returns the currently logged in user's password
func (client *CloudPrivateAPIClient) Password() string {
	if client.password == nil {
		client.t.Log("must log into cloud before getting password")
		client.t.FailNow()
		return ""
	}

	return *client.password
}

// GroupID returns the currently logged in user's current Group ID
func (client *CloudPrivateAPIClient) GroupID() string {
	if client.groupID == nil {
		client.t.Log("must log into cloud before getting group ID")
		client.t.FailNow()
		return ""
	}

	return *client.groupID
}
