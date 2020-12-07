package realm

import (
	"encoding/json"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

type loginPayload struct {
	PublicAPIKey  string `json:"username"`
	PrivateAPIKey string `json:"apiKey"`
}

// Session is the Realm login response
type Session struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// UserProfile contains all of the roles associated with a user
type UserProfile struct {
	Roles []Role `json:"roles"`
}

// Role contains a GrouID field that maps to a project ID
type Role struct {
	GroupID string `json:"group_id"`
}

// AllGroupIDs returns all available group ids for a given user
func (pd UserProfile) AllGroupIDs() []string {
	groupIDs := []string{}
	set := make(map[string]bool)
	for _, role := range pd.Roles {
		if !set[role.GroupID] {
			if role.GroupID == "" {
				continue
			}
			groupIDs = append(groupIDs, role.GroupID)
			set[role.GroupID] = true
		}
	}

	return groupIDs
}

var (
	authenticatePath = adminAPI + "/auth/providers/mongodb-cloud/login"
)

func (c *client) Authenticate(publicAPIKey, privateAPIKey string) (Session, error) {
	res, resErr := c.doJSON(http.MethodPost, authenticatePath, loginPayload{publicAPIKey, privateAPIKey}, api.RequestOptions{})
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

func (c *client) GetUserProfile() (UserProfile, error) {
	res, resErr := c.do(http.MethodGet, userProfileEndpoint, api.RequestOptions{UseAuth: true})
	if resErr != nil {
		return UserProfile{}, resErr
	}

	if res.StatusCode != http.StatusOK {
		return UserProfile{}, UnmarshalServerError(res)
	}

	dec := json.NewDecoder(res.Body)
	defer res.Body.Close()

	var profile UserProfile
	if err := dec.Decode(&profile); err != nil {
		return UserProfile{}, err
	}

	return profile, nil
}
