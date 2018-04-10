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

// Client provides access to the MongoDB Cloud Manager APIs
type Client interface {
	WithAuth(username, apiKey string) Client

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

func (client *simpleClient) DeleteDatabaseUser(groupID, username string) error {
	resp, err := client.do(
		http.MethodDelete,
		fmt.Sprintf("%s/groups/%s/databaseUsers/admin/%s",
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
