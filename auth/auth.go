package auth

import (
	"errors"
)

// Errors related to auth
var (
	ErrInvalidAPIKey = errors.New("stitch: invalid API key")
)

// Response represents the response payload from the API containing an access token and refresh token
type Response struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// ValidAPIKey locally checks if the given API key is valid
func ValidAPIKey(apiKey string) bool {
	return len(apiKey) > 0
}
