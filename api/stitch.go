package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/10gen/stitch-cli/auth"
)

const (
	authProviderCloudLoginRoute = "/auth/providers/mongodb-cloud/login"
	appExportRoute              = "/groups/%s/apps/%s/export"
)

// NewStitchClient returns a new StitchClient to be used for making calls to the Stitch Admin API
func NewStitchClient(baseURL string, client Client) StitchClient {
	return StitchClient{
		baseURL: baseURL,
		Client:  client,
	}
}

// StitchClient represents a Client that can be used to call the Stitch Admin API
type StitchClient struct {
	Client
	baseURL string
}

// Authenticate will authenticate a user given an api key and username
func (sc StitchClient) Authenticate(apiKey, username string) (*auth.Response, error) {
	body, err := json.Marshal(map[string]string{
		"apiKey":   apiKey,
		"username": username,
	})
	if err != nil {
		return nil, err
	}

	res, err := sc.Client.ExecuteRequest(http.MethodPost, sc.baseURL+authProviderCloudLoginRoute, RequestOptions{
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
		return nil, fmt.Errorf("%s: failed to authenticate", res.Status)
	}

	decoder := json.NewDecoder(res.Body)

	var authResponse auth.Response
	if err := decoder.Decode(&authResponse); err != nil {
		return nil, err
	}

	return &authResponse, nil
}

// Export will download a Stitch app as a .zip
func (sc StitchClient) Export(groupID, appID string) (*http.Response, error) {
	return sc.ExecuteRequest(http.MethodGet, sc.baseURL+fmt.Sprintf(appExportRoute, groupID, appID), RequestOptions{})
}
