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
)

// set of supported auth errors
var (
	ErrInvalidSession = errors.New("invalid session")
)

func (c *client) Authenticate(publicAPIKey, privateAPIKey string) (Session, error) {
	res, resErr := c.doJSON(http.MethodPost, authenticatePath, authPayload{publicAPIKey, privateAPIKey}, api.RequestOptions{})
	if resErr != nil {
		return Session{}, resErr
	}
	if res.StatusCode != http.StatusOK {
		return Session{}, UnmarshalServerError(res)
	}

	dec := json.NewDecoder(res.Body)
	defer res.Body.Close()

	var session Session
	if err := dec.Decode(&session); err != nil {
		return Session{}, err
	}
	return session, nil
}

func (c *client) AuthProfile() (AuthProfile, error) {
	res, resErr := c.do(http.MethodGet, authProfilePath, api.RequestOptions{UseAuth: true})
	if resErr != nil {
		return AuthProfile{}, resErr
	}
	if res.StatusCode != http.StatusOK {
		return AuthProfile{}, UnmarshalServerError(res)
	}

	dec := json.NewDecoder(res.Body)
	defer res.Body.Close()

	var profile AuthProfile
	if err := dec.Decode(&profile); err != nil {
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

// Session is the Realm session
type Session struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type authPayload struct {
	PublicAPIKey  string `json:"username"`
	PrivateAPIKey string `json:"apiKey"`
}

// AuthProfile is the user's profile
type AuthProfile struct {
	Roles []Role `json:"roles"`
}

// Role is a user role
type Role struct {
	GroupID string `json:"group_id"`
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
