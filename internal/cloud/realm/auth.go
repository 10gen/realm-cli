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

// Session is the Realm session
type Session struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// AuthProfile contains all of the roles associated with a user
type AuthProfile struct {
	Roles []Role `json:"roles"`
}

// Role contains a GrouID field that maps to a project ID
type Role struct {
	GroupID string `json:"group_id"`
}

// AllGroupIDs returns all available group ids for a given user
func (pd AuthProfile) AllGroupIDs() []string {
	var arr []string
	set := map[string]struct{}{}
	for _, role := range pd.Roles {
		if role.GroupID == "" {
			continue
		}
		if _, ok := set[role.GroupID]; ok {
			continue
		}
		arr = append(arr, role.GroupID)
		set[role.GroupID] = struct{}{}
	}
	return arr
}

var (
	authenticatePath = adminAPI + "/auth/providers/mongodb-cloud/login"
	authProfilePath  = adminAPI + "/auth/profile"
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

func (c *client) GetAuthProfile() (AuthProfile, error) {
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
