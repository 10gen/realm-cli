package mdbcloud

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/10gen/stitch-cli/utils"

	"github.com/edaniels/digest"
)

var errCommonServerError = fmt.Errorf("an unexpected server error has occurred")

type groupResponse struct {
	Results []Group `json:"results"`
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
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch available Project IDs: %s", resp.Status)
	}

	dec := json.NewDecoder(resp.Body)
	var groupResponse groupResponse
	if err := dec.Decode(&groupResponse); err != nil {
		return nil, err
	}

	return groupResponse.Results, nil
}

func (client *simpleClient) GroupByName(groupName string) (*Group, error) {
	resp, err := client.do(
		http.MethodGet,
		fmt.Sprintf("%s/api/public/v1.0/groups/byName/%s", client.atlasAPIBaseURL, groupName),
		nil,
		true,
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("no project found with name %s", groupName)
		}
		return nil, fmt.Errorf("failed to fetch available Project IDs: %s", resp.Status)
	}

	dec := json.NewDecoder(resp.Body)
	var groupResponse Group
	if err := dec.Decode(&groupResponse); err != nil {
		return nil, err
	}

	return &groupResponse, nil
}

func (client *simpleClient) do(
	method, url string, // nolint: unparam
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
		return nil, errCommonServerError
	}

	if body != nil {
		req.Header.Add("Content-Type", string(utils.MediaTypeJSON))
	}

	req.Header.Add("User-Agent", "MongoDB-Stitch-CLI")

	cl := http.Client{}
	cl.Timeout = time.Second * 5
	if client.transport == nil {
		if needAuth {
			return nil, errors.New("expected to have auth context")
		}
		return cl.Do(req)
	}
	cl.Transport = client.transport

	resp, err := cl.Do(req)
	if err != nil {
		return nil, errCommonServerError
	}

	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to authenticate with MongoDB Cloud API")
	}

	return resp, nil
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
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("error deleting database user '%s'", username)
	}
	return nil
}
