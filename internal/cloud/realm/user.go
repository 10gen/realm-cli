package realm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	apiKeysPathPattern = adminAPI + "/groups/%s/apps/%s/api_keys"
	usersPathPattern   = adminAPI + "/groups/%s/apps/%s/users"
)

func (c *client) CreateAPIKey(groupID, appID, apiKeyName string) (APIKey, error) {
	res, resErr := c.doJSON(
		http.MethodPost,
		fmt.Sprintf(apiKeysPathPattern, groupID, appID),
		createAPIKeyPayload{apiKeyName},
		api.RequestOptions{UseAuth: true},
	)
	if resErr != nil {
		return APIKey{}, resErr
	}

	if res.StatusCode != http.StatusCreated {
		return APIKey{}, UnmarshalServerError(res)
	}

	dec := json.NewDecoder(res.Body)
	defer res.Body.Close()

	var apiKey APIKey
	if err := dec.Decode(&apiKey); err != nil {
		return APIKey{}, err
	}
	return apiKey, nil
}

func (c *client) CreateUser(groupID, appID, email, password string) (User, error) {
	res, resErr := c.doJSON(
		http.MethodPost,
		fmt.Sprintf(usersPathPattern, groupID, appID),
		createUserPayload{email, password},
		api.RequestOptions{UseAuth: true},
	)
	if resErr != nil {
		return User{}, resErr
	}

	if res.StatusCode != http.StatusCreated {
		return User{}, UnmarshalServerError(res)
	}

	dec := json.NewDecoder(res.Body)
	defer res.Body.Close()

	var user User
	if err := dec.Decode(&user); err != nil {
		return User{}, err
	}
	return user, nil
}

// PartialAPIKey is a partial Realm application api key
type PartialAPIKey struct {
	ID       string `json:"_id"`
	Name     string `json:"name"`
	Disabled bool   `json:"disabled"`
}

// APIKey is a Realm application api key
type APIKey struct {
	PartialAPIKey
	Key string `json:"key"`
}

// User is a Realm application user
type User struct {
	ID                     string                 `json:"_id"`
	Identities             []interface{}          `json:"identities,omitempty"`
	Type                   string                 `json:"type"`
	Disabled               bool                   `json:"disabled"`
	Data                   map[string]interface{} `json:"data,omitempty"`
	CreationDate           int64                  `json:"creation_date"`
	LastAuthenticationDate int64                  `json:"last_authentication_date"`
}

type createAPIKeyPayload struct {
	Name string `json:"name"`
}

type createUserPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
