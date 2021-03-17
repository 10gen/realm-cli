package realm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/10gen/realm-cli/internal/utils/api"
	"github.com/10gen/realm-cli/internal/utils/flags"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	apiKeysPathPattern      = appPathPattern + "/api_keys"
	pendingUsersPathPattern = appPathPattern + "/user_registrations/pending_users"
	usersPathPattern        = appPathPattern + "/users"
	userPathPattern         = usersPathPattern + "/%s"
	userDisablePathPattern  = userPathPattern + "/disable"
	userEnablePathPattern   = userPathPattern + "/enable"
	userLogoutPathPattern   = userPathPattern + "/logout"

	usersQueryStatus        = "status"
	usersQueryProviderTypes = "provider_types"
)

// UserState is a Realm app user state
type UserState string

// String returns the user state string
func (us UserState) String() string { return string(us) }

// Type returns the user state type
func (us UserState) Type() string { return flags.TypeString }

// Set validates and sets the user state value
func (us *UserState) Set(val string) error {
	newUserState := UserState(val)

	if !isValidUserState(newUserState) {
		return errInvalidUserState
	}

	*us = newUserState
	return nil
}

// set of supported user state values
const (
	UserStateNil      UserState = ""
	UserStateEnabled  UserState = "enabled"
	UserStateDisabled UserState = "disabled"
)

var (
	errInvalidUserState = func() error {
		allUserStateTypes := []string{UserStateEnabled.String(), UserStateDisabled.String()}
		return fmt.Errorf("unsupported value, use one of [%s] instead", strings.Join(allUserStateTypes, ", "))
	}()
)

func isValidUserState(us UserState) bool {
	switch us {
	case
		UserStateNil, // allow state to be optional
		UserStateEnabled,
		UserStateDisabled:
		return true
	}
	return false
}

// APIKey is a Realm app api key
type APIKey struct {
	ID       string `json:"_id"`
	Name     string `json:"name"`
	Disabled bool   `json:"disabled"`
	Key      string `json:"key"`
}

// User is a Realm app user
type User struct {
	ID                     string                 `json:"_id"`
	Identities             []UserIdentity         `json:"identities,omitempty"`
	Type                   string                 `json:"type"`
	Disabled               bool                   `json:"disabled"`
	Data                   map[string]interface{} `json:"data,omitempty"`
	CreationDate           int64                  `json:"creation_date"`
	LastAuthenticationDate int64                  `json:"last_authentication_date"`
}

// UserIdentity is a Realm app user identity
type UserIdentity struct {
	UID          string                 `json:"id"`
	ProviderType AuthProviderType       `json:"provider_type"`
	ProviderID   primitive.ObjectID     `json:"provider_id"`
	ProviderData map[string]interface{} `json:"provider_data,omitempty"`
}

// AuthProviderType is a Realm app auth provider type
type AuthProviderType string

// set of supported auth provider type values
const (
	AuthProviderTypeEmpty          AuthProviderType = ""
	AuthProviderTypeUserPassword   AuthProviderType = "local-userpass"
	AuthProviderTypeAPIKey         AuthProviderType = "api-key"
	AuthProviderTypeFacebook       AuthProviderType = "oauth2-facebook"
	AuthProviderTypeGoogle         AuthProviderType = "oauth2-google"
	AuthProviderTypeAnonymous      AuthProviderType = "anon-user"
	AuthProviderTypeCustomToken    AuthProviderType = "custom-token"
	AuthProviderTypeApple          AuthProviderType = "oauth2-apple"
	AuthProviderTypeCustomFunction AuthProviderType = "custom-function"
)

// set of supported auth constants
var (
	ValidAuthProviderTypes = []AuthProviderType{
		AuthProviderTypeUserPassword,
		AuthProviderTypeAPIKey,
		AuthProviderTypeFacebook,
		AuthProviderTypeGoogle,
		AuthProviderTypeAnonymous,
		AuthProviderTypeCustomToken,
		AuthProviderTypeApple,
		AuthProviderTypeCustomFunction,
	}
)

// String returns the auth provider type string
func (pt AuthProviderType) String() string { return string(pt) }

// Display returns the auth provider type display string
func (pt AuthProviderType) Display() string {
	switch pt {
	case AuthProviderTypeAnonymous:
		return "Anonymous"
	case AuthProviderTypeUserPassword:
		return "User/Password"
	case AuthProviderTypeAPIKey:
		return "ApiKey"
	case AuthProviderTypeApple:
		return "Apple"
	case AuthProviderTypeGoogle:
		return "Google"
	case AuthProviderTypeFacebook:
		return "Facebook"
	case AuthProviderTypeCustomToken:
		return "Custom JWT"
	case AuthProviderTypeCustomFunction:
		return "Custom Function"
	default:
		return "Unknown"
	}
}

// AuthProviderTypes is a Realm app auth provider type slice
type AuthProviderTypes []AuthProviderType

// NewAuthProviderTypes returns an AuthProviderTypes from the provided strings
func NewAuthProviderTypes(apts ...string) AuthProviderTypes {
	authProviderTypes := make([]AuthProviderType, len(apts))
	for i, apt := range apts {
		authProviderTypes[i] = AuthProviderType(apt)
	}
	return authProviderTypes
}

func (apts AuthProviderTypes) join(sep string) string {
	var sb strings.Builder
	for i, apt := range apts {
		if i != 0 {
			sb.WriteString(sep)
		}
		sb.WriteString(apt.String())
	}
	return sb.String()

}

type createAPIKeyRequest struct {
	Name string `json:"name"`
}

func (c *client) CreateAPIKey(groupID, appID, apiKeyName string) (APIKey, error) {
	res, resErr := c.doJSON(
		http.MethodPost,
		fmt.Sprintf(apiKeysPathPattern, groupID, appID),
		createAPIKeyRequest{apiKeyName},
		api.RequestOptions{},
	)
	if resErr != nil {
		return APIKey{}, resErr
	}
	if res.StatusCode != http.StatusCreated {
		return APIKey{}, api.ErrUnexpectedStatusCode{"create api key", res.StatusCode}
	}
	defer res.Body.Close()

	var apiKey APIKey
	if err := json.NewDecoder(res.Body).Decode(&apiKey); err != nil {
		return APIKey{}, err
	}
	return apiKey, nil
}

type createUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c *client) CreateUser(groupID, appID, email, password string) (User, error) {
	res, resErr := c.doJSON(
		http.MethodPost,
		fmt.Sprintf(usersPathPattern, groupID, appID),
		createUserRequest{email, password},
		api.RequestOptions{},
	)
	if resErr != nil {
		return User{}, resErr
	}
	if res.StatusCode != http.StatusCreated {
		return User{}, api.ErrUnexpectedStatusCode{"create user", res.StatusCode}
	}
	defer res.Body.Close()

	var user User
	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		return User{}, err
	}
	return user, nil
}

func (c *client) DeleteUser(groupID, appID, userID string) error {
	res, resErr := c.do(
		http.MethodDelete,
		fmt.Sprintf(userPathPattern, groupID, appID, userID),
		api.RequestOptions{},
	)
	if resErr != nil {
		return resErr
	}
	if res.StatusCode != http.StatusNoContent {
		return api.ErrUnexpectedStatusCode{Action: "delete user", Actual: res.StatusCode}
	}
	return nil
}

func (c *client) DisableUser(groupID, appID, userID string) error {
	res, resErr := c.do(
		http.MethodPut,
		fmt.Sprintf(userDisablePathPattern, groupID, appID, userID),
		api.RequestOptions{},
	)
	if resErr != nil {
		return resErr
	}
	if res.StatusCode != http.StatusNoContent {
		return api.ErrUnexpectedStatusCode{Action: "disable user", Actual: res.StatusCode}
	}
	return nil
}

func (c *client) EnableUser(groupID, appID, userID string) error {
	res, resErr := c.do(
		http.MethodPut,
		fmt.Sprintf(userEnablePathPattern, groupID, appID, userID),
		api.RequestOptions{},
	)
	if resErr != nil {
		return resErr
	}
	if res.StatusCode != http.StatusNoContent {
		return api.ErrUnexpectedStatusCode{Action: "enable user", Actual: res.StatusCode}
	}
	return nil
}

// UserFilter represents the optional filter parameters available for lists of users
type UserFilter struct {
	IDs       []string
	Pending   bool
	Providers []AuthProviderType
	State     UserState
}

func (c *client) FindUsers(groupID, appID string, filter UserFilter) ([]User, error) {
	if filter.Pending {
		return c.getPendingUsers(groupID, appID, filter.IDs)
	}
	if len(filter.IDs) == 0 {
		return c.getUsers(groupID, appID, filter.State, filter.Providers)
	}
	return c.getUsersByIDs(groupID, appID, filter.IDs, filter.State, filter.Providers)
}

func (c *client) RevokeUserSessions(groupID, appID, userID string) error {
	res, resErr := c.do(
		http.MethodPut,
		fmt.Sprintf(userLogoutPathPattern, groupID, appID, userID),
		api.RequestOptions{},
	)
	if resErr != nil {
		return resErr
	}
	if res.StatusCode != http.StatusNoContent {
		return api.ErrUnexpectedStatusCode{Action: "revoke user", Actual: res.StatusCode}
	}
	return nil
}

func (c *client) getPendingUsers(groupID, appID string, userIDs []string) ([]User, error) {
	res, resErr := c.do(
		http.MethodGet,
		fmt.Sprintf(pendingUsersPathPattern, groupID, appID),
		api.RequestOptions{},
	)
	if resErr != nil {
		return nil, resErr
	}
	if res.StatusCode != http.StatusOK {
		return nil, api.ErrUnexpectedStatusCode{"get pending users", res.StatusCode}
	}
	defer res.Body.Close()

	var users []User
	if err := json.NewDecoder(res.Body).Decode(&users); err != nil {
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
		api.RequestOptions{},
	)
	if resErr != nil {
		return User{}, resErr
	}
	if res.StatusCode != http.StatusOK {
		return User{}, api.ErrUnexpectedStatusCode{"get user", res.StatusCode}
	}
	defer res.Body.Close()

	var user User
	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		return User{}, err
	}
	return user, nil
}

func (c *client) getUsers(groupID, appID string, userState UserState, authProviderTypes AuthProviderTypes) ([]User, error) {
	options := api.RequestOptions{Query: make(map[string]string)}
	if userState != UserStateNil {
		options.Query[usersQueryStatus] = string(userState)
	}
	if len(authProviderTypes) > 0 {
		options.Query[usersQueryProviderTypes] = authProviderTypes.join(",")
	}

	res, resErr := c.do(http.MethodGet, fmt.Sprintf(usersPathPattern, groupID, appID), options)
	if resErr != nil {
		return nil, resErr
	}
	if res.StatusCode != http.StatusOK {
		return nil, api.ErrUnexpectedStatusCode{"get users", res.StatusCode}
	}
	defer res.Body.Close()

	var users []User
	if err := json.NewDecoder(res.Body).Decode(&users); err != nil {
		return nil, err
	}
	return users, nil
}

func (c *client) getUsersByIDs(groupID, appID string, userIDs []string, userState UserState, authProviderTypes []AuthProviderType) ([]User, error) {
	users := make([]User, 0, len(userIDs))
	for _, userID := range userIDs {
		user, err := c.getUser(groupID, appID, userID)
		if err != nil {
			return nil, err
		}

		if userMatchesState(user, userState) {
			users = append(users, user)
		}
	}

	if len(authProviderTypes) == 0 {
		return users, nil
	}

	providers := make(map[AuthProviderType]struct{}, len(authProviderTypes))
	for _, provider := range authProviderTypes {
		providers[provider] = struct{}{}
	}

	filtered := make([]User, 0, len(users))
	for _, user := range users {
		var matchedProvider bool
		for _, identity := range user.Identities {
			if _, ok := providers[identity.ProviderType]; !ok {
				continue
			}
			matchedProvider = true
			break
		}
		if matchedProvider {
			filtered = append(filtered, user)
		}
	}
	return filtered, nil
}

func userMatchesState(user User, userState UserState) bool {
	if userState == UserStateEnabled {
		return !user.Disabled
	}
	if userState == UserStateDisabled {
		return user.Disabled
	}
	return true
}
