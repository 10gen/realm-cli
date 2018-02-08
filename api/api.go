package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/10gen/stitch-cli/auth"
	"github.com/10gen/stitch-cli/user"
)

// DefaultBaseURL is the default cloud prod base url for Stitch apps
const DefaultBaseURL = "https://stitch.mongodb.com/api/admin/v3.0"

const (
	authSessionRoute = "/auth/session"
)

// Client represents something that is capable of making HTTP requests
type Client interface {
	ExecuteRequest(method, path string, options RequestOptions) (*http.Response, error)
}

// RequestOptions represents a simple set of options to use with HTTP requests
type RequestOptions struct {
	Body   io.Reader
	Header http.Header
}

type basicAPIClient struct{}

// ExecuteRequest makes an HTTP request to the provided path
func (apiClient *basicAPIClient) ExecuteRequest(method, path string, options RequestOptions) (*http.Response, error) {
	req, err := http.NewRequest(method, path, options.Body)
	if err != nil {
		return nil, err
	}

	req.Header = options.Header

	client := &http.Client{}

	return client.Do(req)
}

// NewClient returns a new Client
func NewClient() Client {
	return &basicAPIClient{}
}

// NewAuthClient returns a new *AuthClient
func NewAuthClient(baseURL string, client Client, user *user.User) *AuthClient {
	return &AuthClient{
		baseURL: baseURL,
		Client:  client,
		user:    user,
	}
}

// AuthClient is a Client that is aware of a User's auth credentials
type AuthClient struct {
	Client
	baseURL string
	user    *user.User
}

// RefreshAuth makes a call to the session endpoint using the user's refresh token in order to obtain a new access token
func (ac *AuthClient) RefreshAuth() (auth.Response, error) {
	res, err := ac.Client.ExecuteRequest(http.MethodPost, ac.baseURL+authSessionRoute, RequestOptions{
		Header: http.Header{
			"Authorization": []string{"Bearer " + ac.user.RefreshToken},
		},
	})
	if err != nil {
		return auth.Response{}, err
	}

	if res.StatusCode != http.StatusCreated {
		return auth.Response{}, fmt.Errorf("%s: failed to refresh auth", res.Status)
	}

	decoder := json.NewDecoder(res.Body)
	defer res.Body.Close()

	var authResponse auth.Response
	if err := decoder.Decode(&authResponse); err != nil {
		return auth.Response{}, err
	}

	return authResponse, nil
}

// ExecuteRequest makes a call to the provided path, supplying the user's access token
func (ac *AuthClient) ExecuteRequest(method, path string, options RequestOptions) (*http.Response, error) {
	res, err := ac.Client.ExecuteRequest(method, path, RequestOptions{
		Header: http.Header{
			"Authorization": []string{"Bearer " + ac.user.AccessToken},
		},
	})
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusUnauthorized {
		res.Body.Close()
		authResponse, err := ac.RefreshAuth()
		if err != nil {
			return nil, err
		}

		return ac.Client.ExecuteRequest(method, path, RequestOptions{
			Header: http.Header{
				"Authorization": []string{"Bearer " + authResponse.AccessToken},
			},
		})
	}

	return res, err
}
