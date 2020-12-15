package realm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/10gen/realm-cli/internal/utils/api"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	apiKeysPathPattern      = appPathPattern + "/api_keys"
	pendingUsersPathPattern = appPathPattern + "/user_registrations/pending_users"
	usersPathPattern        = appPathPattern + "/users"
	userPathPattern         = usersPathPattern + "/%s"
	userDisablePathPattern  = userPathPattern + "/disable"
	userLogoutPathPattern   = userPathPattern + "/logout"

	usersQueryStatus        = "status"
	usersQueryProviderTypes = "provider_types"
)

func (c *client) CreateAPIKey(groupID, appID, apiKeyName string) (APIKey, error) {
	res, resErr := c.doJSON(
		http.MethodPost,
		fmt.Sprintf(apiKeysPathPattern, groupID, appID),
		createAPIKeyRequest{apiKeyName},
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
		createUserRequest{email, password},
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

func (c *client) DeleteUser(groupID, appID, userID string) error {
	res, resErr := c.do(
		http.MethodDelete,
		fmt.Sprintf(userPathPattern, groupID, appID, userID),
		api.RequestOptions{UseAuth: true},
	)
	if resErr != nil {
		return resErr
	}
	if res.StatusCode != http.StatusNoContent {
		return UnmarshalServerError(res)
	}
	return nil
}

func (c *client) DisableUser(groupID, appID, userID string) error {
	res, resErr := c.do(
		http.MethodPut,
		fmt.Sprintf(userDisablePathPattern, groupID, appID, userID),
		api.RequestOptions{UseAuth: true},
	)
	if resErr != nil {
		return resErr
	}
	if res.StatusCode != http.StatusNoContent {
		return UnmarshalServerError(res)
	}
	return nil
}

func (c *client) FindUsers(groupID, appID string, filter UserFilter) ([]User, error) {
	if filter.Pending {
		return c.getPendingUsers(groupID, appID, filter.IDs)
	}

	if len(filter.IDs) == 0 {
		return c.getUsers(groupID, appID, filter.State, filter.Providers)
	}

	users := make([]User, 0, len(filter.IDs))
	for _, userID := range filter.IDs {
		user, err := c.getUser(groupID, appID, userID)
		if err != nil {
			return nil, err
		}

		var include bool
		switch filter.State {
		case UserStateEnabled:
			include = !user.Disabled
		case UserStateDisabled:
			include = user.Disabled
		default:
			include = true
		}

		if include {
			users = append(users, user)
		}
	}

	if len(filter.Providers) == 0 {
		return users, nil
	}

	providers := make(map[string]struct{}, len(filter.Providers))
	for _, provider := range filter.Providers {
		providers[provider] = struct{}{}
	}

	filtered := make([]User, 0, len(users))
	for _, user := range users {
		if _, ok := providers[user.Type]; !ok {
			continue
		}
		filtered = append(filtered, user)
	}
	return filtered, nil
}

func (c *client) RevokeUserSessions(groupID, appID, userID string) error {
	res, resErr := c.do(
		http.MethodPut,
		fmt.Sprintf(userLogoutPathPattern, groupID, appID, userID),
		api.RequestOptions{UseAuth: true},
	)
	if resErr != nil {
		return resErr
	}
	if res.StatusCode != http.StatusNoContent {
		return UnmarshalServerError(res)
	}
	return nil
}

func (c *client) getPendingUsers(groupID, appID string, userIDs []string) ([]User, error) {
	res, resErr := c.do(
		http.MethodGet,
		fmt.Sprintf(pendingUsersPathPattern, groupID, appID),
		api.RequestOptions{UseAuth: true},
	)
	if resErr != nil {
		return nil, resErr
	}
	if res.StatusCode != http.StatusOK {
		return nil, UnmarshalServerError(res)
	}

	dec := json.NewDecoder(res.Body)
	defer res.Body.Close()

	var users []User
	if err := dec.Decode(&users); err != nil {
		return nil, err
	}

	if len(userIDs) == 0 {
		return users, nil
	}

	userIDSet := make(map[string]struct{}, len(userIDs))
	for _, userID := range userIDs {
		userIDSet[userID] = struct{}{}
	}

	filtered := make([]User, 0, len(users))
	for _, user := range users {
		if _, ok := userIDSet[user.ID]; !ok {
			continue
		}
		filtered = append(users, user)
	}
	return filtered, nil
}

func (c *client) getUser(groupID, appID, userID string) (User, error) {
	res, resErr := c.do(
		http.MethodGet,
		fmt.Sprintf(userPathPattern, groupID, appID, userID),
		api.RequestOptions{UseAuth: true},
	)
	if resErr != nil {
		return User{}, resErr
	}
	if res.StatusCode != http.StatusOK {
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

func (c *client) getUsers(groupID, appID string, userState UserState, providerTypes []string) ([]User, error) {
	options := api.RequestOptions{
		UseAuth: true,
		Query:   make(map[string]string),
	}
	if userState != UserStateNil {
		options.Query[usersQueryStatus] = string(userState)
	}
	if len(providerTypes) > 0 {
		options.Query[usersQueryProviderTypes] = strings.Join(providerTypes, ",")
	}

	res, resErr := c.do(http.MethodGet, fmt.Sprintf(usersPathPattern, groupID, appID), options)
	if resErr != nil {
		return nil, resErr
	}
	if res.StatusCode != http.StatusOK {
		return nil, UnmarshalServerError(res)
	}

	dec := json.NewDecoder(res.Body)
	defer res.Body.Close()

	var users []User
	if err := dec.Decode(&users); err != nil {
		return nil, err
	}
	return users, nil
}

// APIKey is a Realm application api key
type APIKey struct {
	ID       string `json:"_id"`
	Name     string `json:"name"`
	Disabled bool   `json:"disabled"`
	Key      string `json:"key"`
}

// UserFilter represents the optional filter parameters available for lists of users
type UserFilter struct {
	IDs       []string
	Pending   bool
	Providers []string
	State     UserState
}

// UserState is a Realm application user state
type UserState string

// set of supported user state values
const (
	UserStateNil      UserState = ""
	UserStateEnabled  UserState = "enabled"
	UserStateDisabled UserState = "disabled"
)

// User is a Realm application user
type User struct {
	ID                     string                 `json:"_id"`
	Identities             []UserIdentity         `json:"identities,omitempty"`
	Type                   string                 `json:"type"`
	Disabled               bool                   `json:"disabled"`
	Data                   map[string]interface{} `json:"data,omitempty"`
	CreationDate           int64                  `json:"creation_date"`
	LastAuthenticationDate int64                  `json:"last_authentication_date"`
}

// UserIdentity is a Realm application user identity
type UserIdentity struct {
	UID          string                 `json:"id"`
	ProviderType string                 `json:"provider_type"`
	ProviderID   primitive.ObjectID     `json:"provider_id"`
	ProviderData map[string]interface{} `json:"provider_data,omitempty"`
}

type createAPIKeyRequest struct {
	Name string `json:"name"`
}

type createUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
