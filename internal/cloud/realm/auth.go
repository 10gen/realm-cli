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

// AuthResponse is the Realm login response
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (c *client) Authenticate(publicAPIKey, privateAPIKey string) (AuthResponse, error) {
	path := adminAPI + "/auth/providers/mongodb-cloud/login"

	opts, optsErr := api.JSONRequestOptions(loginPayload{publicAPIKey, privateAPIKey})
	if optsErr != nil {
		return AuthResponse{}, optsErr
	}

	res, resErr := c.do(http.MethodPost, path, opts)
	if resErr != nil {
		return AuthResponse{}, resErr
	}

	if res.StatusCode != http.StatusOK {
		return AuthResponse{}, UnmarshalServerError(res)
	}

	dec := json.NewDecoder(res.Body)
	defer res.Body.Close()

	var auth AuthResponse
	if err := dec.Decode(&auth); err != nil {
		return AuthResponse{}, err
	}
	return auth, nil
}
