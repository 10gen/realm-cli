package realm

import (
	"encoding/json"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	authenticatePath = adminAPI + "/auth/providers/mongodb-cloud/login"
	authProfilePath  = adminAPI + "/auth/profile"
	authSessionPath  = adminAPI + "/auth/session"
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
		api.RequestOptions{NoAuth: true, PreventRefresh: true},
	)
	if resErr != nil {
		return Session{}, resErr
	}
	if res.StatusCode != http.StatusOK {
		return Session{}, api.ErrUnexpectedStatusCode{"authenticate", res.StatusCode}
	}
	defer res.Body.Close()

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
	res, resErr := c.do(http.MethodGet, authProfilePath, api.RequestOptions{})
	if resErr != nil {
		return AuthProfile{}, resErr
	}
	if res.StatusCode != http.StatusOK {
		return AuthProfile{}, api.ErrUnexpectedStatusCode{"get auth profile", res.StatusCode}
	}
	defer res.Body.Close()

	var profile AuthProfile
	if err := json.NewDecoder(res.Body).Decode(&profile); err != nil {
		return AuthProfile{}, err
	}
	return profile, nil
}

func (c *client) getAuthToken(options api.RequestOptions) (string, error) {
	requiresAccessToken := !options.NoAuth
	requiresRefreshToken := options.RefreshAuth

	if requiresAccessToken || requiresRefreshToken {
		if c.profile == nil {
			return "", ErrInvalidSession{}
		}

		session := c.profile.Session()
		if requiresRefreshToken {
			if session.RefreshToken == "" {
				return "", ErrInvalidSession{}
			}
			return session.RefreshToken, nil
		}

		if requiresAccessToken {
			if session.AccessToken == "" {
				return "", ErrInvalidSession{}
			}
			return session.AccessToken, nil
		}
	}

	return "", nil
}

func (c *client) refreshAuth() error {
	res, resErr := c.do(
		http.MethodPost,
		authSessionPath,
		api.RequestOptions{RefreshAuth: true},
	)
	if resErr != nil {
		return resErr
	}
	if res.StatusCode != http.StatusCreated {
		return ErrInvalidSession{}
	}
	defer res.Body.Close()

	var s Session
	if err := json.NewDecoder(res.Body).Decode(&s); err != nil {
		return err
	}

	session := c.profile.Session()
	session.AccessToken = s.AccessToken
	c.profile.SetSession(session)

	return c.profile.Save()
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
