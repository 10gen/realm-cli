package mdbcloud

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/10gen/stitch-cli/utils"

	"github.com/edaniels/digest"
)

var errCommonServerError = "an unexpected server error has occurred"

type groupResponse struct {
	Results []Group `json:"results"`
}

type errResponse struct {
	Detail    string `json:"detail"`
	Error     int    `json:"error"`
	ErrorCode string `json:"errorCode"`
}

// Group represents a mongodb atlas group
type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Client provides access to the MongoDB Cloud Manager APIs
type Client interface {
	WithAuth(username, apiKey string) Client
	Groups() ([]Group, error)
	GroupByName(string) (*Group, error)
	DeleteDatabaseUser(groupID, username string) error
}

type simpleClient struct {
	transport       *digest.Transport
	atlasAPIBaseURL string
}

// NewClient constructs and returns a new Client given a username, API key,
// the public Cloud API base URL, and the atlas API base url
func NewClient(atlasAPIBaseURL string) Client {
	return &simpleClient{
		atlasAPIBaseURL: atlasAPIBaseURL,
	}
}

func (client simpleClient) WithAuth(username, apiKey string) Client {
	// digest.NewTransport will use http.DefaultTransport
	client.transport = digest.NewTransport(username, apiKey)
	return &client
}

// Groups returns all available Groups for the user
func (client *simpleClient) Groups() ([]Group, error) {
	resp, err := client.do(
		http.MethodGet,
		fmt.Sprintf("%s/api/public/v1.0/groups", client.atlasAPIBaseURL),
		nil,
		true,
	)
	errPrefix := "failed to fetch available Projects: %s"
	if err != nil {
		return nil, fmt.Errorf(errPrefix, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(errPrefix, resp.Status)
	}

	dec := json.NewDecoder(resp.Body)
	var groupResp groupResponse
	if decodeErr := dec.Decode(&groupResp); decodeErr != nil {
		return nil, decodeErr
	}

	return groupResp.Results, nil
}

func (client *simpleClient) GroupByName(groupName string) (*Group, error) {
	resp, err := client.do(
		http.MethodGet,
		fmt.Sprintf("%s/api/public/v1.0/groups/byName/%s", client.atlasAPIBaseURL, groupName),
		nil,
		true,
	)
	errPrefix := "failed to fetch Project '%s': %s"
	if err != nil {
		return nil, fmt.Errorf(errPrefix, groupName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(errPrefix, groupName, resp.Status)
	}

	dec := json.NewDecoder(resp.Body)
	var groupResp Group
	if err := dec.Decode(&groupResp); err != nil {
		return nil, err
	}

	return &groupResp, nil
}

// DeleteDatabaseUser deletes the database user with the provided username
func (client *simpleClient) DeleteDatabaseUser(groupID, username string) error {
	resp, err := client.do(
		http.MethodDelete,
		fmt.Sprintf("%s/api/atlas/v1.0/groups/%s/databaseUsers/admin/%s",
			client.atlasAPIBaseURL,
			groupID,
			username,
		),
		nil,
		true,
	)
	errPrefix := "failed to delete User '%s': %s"
	if err != nil {
		return fmt.Errorf(errPrefix, username, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf(errPrefix, username, resp.Status)
	}

	return nil
}

func (client *simpleClient) do(
	method, url string,
	body interface{},
	needAuth bool, // nolint: unparam
) (*http.Response, error) {

	var bodyReader io.Reader
	if body != nil {
		md, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(md)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, errors.New(errCommonServerError)
	}

	if body != nil {
		req.Header.Add("Content-Type", string(utils.MediaTypeJSON))
	}

	req.Header.Add("User-Agent", "MongoDB-Stitch-CLI")

	cl := http.Client{}
	cl.Timeout = time.Second * 20
	if client.transport == nil {
		if needAuth {
			return nil, errors.New("expected to have auth context")
		}
		return cl.Do(req)
	}
	cl.Transport = client.transport

	resp, err := cl.Do(req)
	if err != nil {
		if netErr, isNetErr := err.(net.Error); isNetErr && netErr.Timeout() {
			return nil, fmt.Errorf(errCommonServerError+": request timed out after %s", cl.Timeout.String())
		}
		return nil, errors.New(errCommonServerError)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		defer resp.Body.Close()
		var cloudErr errResponse
		dec := json.NewDecoder(resp.Body)
		errPrefix := "failed to authenticate with MongoDB Cloud API: %s"
		if err := dec.Decode(&cloudErr); err != nil {
			return nil, fmt.Errorf(errPrefix, err)
		}
		return nil, fmt.Errorf(errPrefix, cloudErr.Detail)
	}

	if resp.StatusCode == http.StatusForbidden {
		defer resp.Body.Close()
		return nil, fmt.Errorf(
			"(%s) Please check your Atlas API Whitelist entries at https://cloud.mongodb.com/v2#/account/publicApi to ensure that requests from this IP address are allowed",
			resp.Status,
		)
	}

	return resp, nil
}
