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

	// Root returns the root view returns by visiting the base url of the admin API
	Root() (*Root, error)

	// Self returns the user associated with the credentials in use
	Self() (*User, error)

	// User returns the user with the given ID unless it cannot be found in which case
	// an error will be returned
	User(userID string) (*User, error)

	// Group returns the group with the given ID unless it cannot be found in which case
	// an error will be returned
	Group(groupID string) (*Group, error)

	// AtlasGroup returns the group with the given ID unless it either cannot be found
	// or is not an atlas group
	AtlasGroup(groupID string) (*Group, error)

	// AtlasCluster returns the cluster with the given name unless it cannot be found
	AtlasCluster(groupID string, clusterName string) (*AtlasCluster, error)

	// CreateAtlasCluster creates a cluster with the given configuration
	CreateAtlasCluster(groupID string, cluster CreateAtlasCluster) error

	// DeleteAtlasCluster deletes the cluster with the given name under the group specified
	DeleteAtlasCluster(groupID, clusterName string) error

	// AtlasIPWhitelistEntries returns information about the group's IP Whitelist
	AtlasIPWhitelistEntries(groupID string) (*AtlasIPWhitelistGetResponse, error)

	// AddAtlasIPWhitelistEntries adds the given entries to the group's IP Whitelist
	AddAtlasIPWhitelistEntries(groupID string, entries ...AtlasIPWhitelistEntry) error

	// AddAtlasDBUser adds a user to the group unless it already exists
	AddAtlasDBUser(groupID string, user *AtlasDBUser) error

	// UpdateAtlasDBUser updates the given user in the group unless it does not exist
	UpdateAtlasDBUser(groupID string, user *AtlasDBUser) error

	// AddOrUpdateAtlasDBUser adds or updates the given user in the group
	AddOrUpdateAtlasDBUser(groupID string, user *AtlasDBUser) error
}

type simpleClient struct {
	transport        *digest.Transport
	publicAPIBaseURL string
	atlasAPIBaseURL  string
}

// NewClient constructs and returns a new Client given a username, API key,
// the public Cloud API base URL, and the atlas API base url
func NewClient(publicAPIBaseURL, atlasAPIBaseURL string) Client {
	return &simpleClient{
		publicAPIBaseURL: publicAPIBaseURL,
		atlasAPIBaseURL:  atlasAPIBaseURL,
	}
}

func (client simpleClient) WithAuth(username, apiKey string) Client {
	// digest.NewTransport will use http.DefaultTransport
	client.transport = digest.NewTransport(username, apiKey)
	return &client
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

func (client *simpleClient) Root() (*Root, error) {
	resp, err := client.do(http.MethodGet, client.publicAPIBaseURL, nil, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	var root Root
	if err := dec.Decode(&root); err != nil {
		return nil, err
	}

	return &root, nil
}

func (client *simpleClient) Self() (*User, error) {
	root, err := client.Root()
	if err != nil {
		return nil, err
	}

	var userLink *Link
	for _, link := range root.Links {
		if link.Rel == RelationUser {
			userLink = &link
			break
		}
	}

	if userLink == nil {
		return nil, errors.New("failed to find self user link in root")
	}

	resp, err := client.do(http.MethodGet, userLink.HRef, nil, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error finding self: %v", resp.Status)
	}

	dec := json.NewDecoder(resp.Body)
	var user User
	if err := dec.Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (client *simpleClient) User(userID string) (*User, error) {
	resp, err := client.do(http.MethodGet, fmt.Sprintf("%s/users/%s", client.publicAPIBaseURL, userID), nil, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("user '%s' not found", userID)
	}

	dec := json.NewDecoder(resp.Body)
	var user User
	if err := dec.Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (client *simpleClient) Group(groupID string) (*Group, error) {
	resp, err := client.do(http.MethodGet, fmt.Sprintf("%s/groups/%s", client.publicAPIBaseURL, groupID), nil, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("group '%s' not found", groupID)
	}

	dec := json.NewDecoder(resp.Body)
	var group Group
	if err := dec.Decode(&group); err != nil {
		return nil, err
	}

	return &group, nil
}

func (client *simpleClient) AtlasGroup(groupID string) (*Group, error) {
	// Atlas group should have a clusters endpoint
	resp, err := client.do(http.MethodGet, fmt.Sprintf("%s/groups/%s/clusters", client.atlasAPIBaseURL, groupID), nil, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("group '%s' not found", groupID)
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error finding group '%s'", groupID)
	}

	return &Group{ID: groupID}, nil
}

func (client *simpleClient) AtlasCluster(groupID string, clusterName string) (*AtlasCluster, error) {
	resp, err := client.do(
		http.MethodGet,
		fmt.Sprintf("%s/groups/%s/clusters/%s",
			client.atlasAPIBaseURL,
			groupID,
			clusterName,
		),
		nil,
		true,
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("cluster '%s' not found", clusterName)
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error finding cluster '%s'", clusterName)
	}

	dec := json.NewDecoder(resp.Body)
	var cluster AtlasCluster
	if err := dec.Decode(&cluster); err != nil {
		return nil, err
	}

	return &cluster, nil
}

func (client *simpleClient) CreateAtlasCluster(groupID string, cluster CreateAtlasCluster) error {
	resp, err := client.do(
		http.MethodPost,
		fmt.Sprintf("%s/groups/%s/clusters",
			client.atlasAPIBaseURL,
			groupID,
		),
		cluster,
		true,
	)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("error creating cluster '%s'", cluster.Name)
	}

	return nil
}

func (client *simpleClient) DeleteAtlasCluster(groupID, clusterName string) error {
	resp, err := client.do(
		http.MethodDelete,
		fmt.Sprintf("%s/groups/%s/clusters/%s",
			client.atlasAPIBaseURL,
			groupID,
			clusterName,
		),
		nil,
		true,
	)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("error deleting cluster '%s'", clusterName)
	}
	return nil
}

func (client *simpleClient) AtlasIPWhitelistEntries(groupID string) (*AtlasIPWhitelistGetResponse, error) {
	resp, err := client.do(
		http.MethodGet,
		fmt.Sprintf("%s/groups/%s/whitelist",
			client.atlasAPIBaseURL,
			groupID,
		),
		nil,
		true,
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error finding IP whitelist entries: %v", resp.Status)
	}

	dec := json.NewDecoder(resp.Body)
	var whitelistResponse AtlasIPWhitelistGetResponse
	if err := dec.Decode(&whitelistResponse); err != nil {
		return nil, err
	}

	return &whitelistResponse, nil
}

func (client *simpleClient) AddAtlasIPWhitelistEntries(
	groupID string,
	entries ...AtlasIPWhitelistEntry,
) error {
	resp, err := client.do(
		http.MethodPost,
		fmt.Sprintf("%s/groups/%s/whitelist",
			client.atlasAPIBaseURL,
			groupID,
		),
		entries,
		true,
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("error adding IP whitelist entries: %v", resp.Status)
	}

	return nil
}

var errAtlasDBUserAlreadyExists = errors.New("atlas DB user already exists")

func (client *simpleClient) AddAtlasDBUser(groupID string, user *AtlasDBUser) error {
	resp, err := client.do(
		http.MethodPost,
		fmt.Sprintf("%s/groups/%s/databaseUsers",
			client.atlasAPIBaseURL,
			groupID,
		),
		user,
		true,
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		if resp.StatusCode == http.StatusConflict {
			return errAtlasDBUserAlreadyExists
		}
		return fmt.Errorf("error adding DB user: %v", resp.Status)
	}

	return nil
}

func (client *simpleClient) UpdateAtlasDBUser(groupID string, user *AtlasDBUser) error {
	resp, err := client.do(
		http.MethodPatch,
		fmt.Sprintf("%s/groups/%s/databaseUsers/admin/%s",
			client.atlasAPIBaseURL,
			groupID,
			user.Username,
		),
		user,
		true,
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("user '%s' not found", user.Username)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error updating DB user: %v", resp.Status)
	}

	return nil
}

func (client *simpleClient) AddOrUpdateAtlasDBUser(groupID string, user *AtlasDBUser) error {
	err := client.AddAtlasDBUser(groupID, user)
	if err == nil {
		return nil
	}

	if err != errAtlasDBUserAlreadyExists {
		return err
	}

	return client.UpdateAtlasDBUser(groupID, user)
}
