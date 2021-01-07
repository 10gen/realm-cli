package realm

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	authenticatePath = adminAPI + "/auth/providers/mongodb-cloud/login"
	authProfilePath  = adminAPI + "/auth/profile"
	authSessionPath  = adminAPI + "/auth/session"
)

// set of supported auth errors
var (
	ErrInvalidSession = errors.New("invalid session")
)

// Session is the Realm session
type Session struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type authenticateRequest struct {
	PublicAPIKey  string `json:"username"`
	PrivateAPIKey string `json:"apiKey"`
}

func (c *client) Authenticate(publicAPIKey, privateAPIKey string) (Session, error) {
	res, resErr := c.doJSON(
		http.MethodPost,
		authenticatePath,
		authenticateRequest{publicAPIKey, privateAPIKey},
		api.RequestOptions{PreventRefresh: true},
	)
	if resErr != nil {
		return Session{}, resErr
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return Session{}, parseResponseError(res)
	}

	var session Session
	if err := json.NewDecoder(res.Body).Decode(&session); err != nil {
		return Session{}, err
	}
	return session, nil
}

// AuthProfile is the user's profile
type AuthProfile struct {
	Roles []Role `json:"roles"`
}

// Role is a user role
type Role struct {
	GroupID string `json:"group_id"`
}

func (c *client) AuthProfile() (AuthProfile, error) {
	res, resErr := c.do(http.MethodGet, authProfilePath, api.RequestOptions{UseAuth: true})
	if resErr != nil {
		return AuthProfile{}, resErr
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return AuthProfile{}, parseResponseError(res)
	}

	var profile AuthProfile
	if err := json.NewDecoder(res.Body).Decode(&profile); err != nil {
		return AuthProfile{}, err
	}
	return profile, nil
}

func (c *client) getAuth(options api.RequestOptions) (string, error) {
	if options.UseAuth {
		if c.session.AccessToken == "" {
			return "", ErrInvalidSession
		}
		return c.session.AccessToken, nil
	}

	if options.RefreshAuth {
		if c.session.RefreshToken == "" {
			return "", ErrInvalidSession
		}
		return c.session.RefreshToken, nil
	}

	return "", nil
}

func (c *client) refreshAuth() (string, error) {
	res, resErr := c.do(http.MethodPost, authSessionPath, api.RequestOptions{RefreshAuth: true, PreventRefresh: true})
	if resErr != nil {
		return "", resErr
	}
	if res.StatusCode != http.StatusCreated {
		return "", ErrInvalidSession
	}
	defer res.Body.Close()

	var session Session
	if err := json.NewDecoder(res.Body).Decode(&session); err != nil {
		return "", err
	}
	return session.AccessToken, nil
}

// AllGroupIDs returns all group ids associated with the user's profile
func (profile AuthProfile) AllGroupIDs() []string {
	groupIDSet := map[string]struct{}{"": struct{}{}}

	var groupIDs []string
	for _, role := range profile.Roles {
		if _, ok := groupIDSet[role.GroupID]; ok {
			continue
		}
		groupIDs = append(groupIDs, role.GroupID)
		groupIDSet[role.GroupID] = struct{}{}
	}
	return groupIDs
}
