package auth

import (
	"errors"
	"strings"
)

// Errors related to auth
var (
	ErrInvalidAPIKey       = errors.New("realm: invalid API key")
	ErrInvalidPublicAPIKey = errors.New("realm: invalid username or public API key")
	ErrInvalidPassword     = errors.New("realm: invalid password")
)

// ProviderType represents available types of auth providers
type ProviderType string

// Available ProviderTypes
const (
	ProviderTypeAPIKey           ProviderType = "mongodb-cloud"
	ProviderTypeUsernamePassword ProviderType = "local-userpass"
)

const (
	usernameField = "username"
	passwordField = "password"
	apiKeyField   = "apiKey"
)

// AuthenticationProvider represents an authentication method
type AuthenticationProvider interface {
	Type() ProviderType
	Payload() map[string]string
	Validate() error
}

// Response represents the response payload from the API containing an access token and refresh token
type Response struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// ValidAPIKey locally checks if the given API key is valid
func ValidAPIKey(apiKey string) bool {
	return len(apiKey) > 0 && strings.Contains(apiKey, "-")
}

// ValidAccessToken locally checks if the given access token is valid
func ValidAccessToken(accessToken string) bool {
	return len(accessToken) > 0
}

// APIKeyProvider is an AuthenticationProvider using a cloud API Key
type APIKeyProvider struct {
	providerType ProviderType
	payload      map[string]string
}

// NewAPIKeyProvider returns a new APIKeyProvider given an API Key and Username
func NewAPIKeyProvider(username, apiKey string) *APIKeyProvider {
	return &APIKeyProvider{
		providerType: ProviderTypeAPIKey,
		payload: map[string]string{
			usernameField: username,
			apiKeyField:   apiKey,
		},
	}
}

// Type returns the auth provider type
func (p *APIKeyProvider) Type() ProviderType {
	return p.providerType
}

// Payload returns the provider's auth payload
func (p *APIKeyProvider) Payload() map[string]string {
	return p.payload
}

// Validate will determine if a given provider is valid
func (p *APIKeyProvider) Validate() error {
	if apiKey, ok := p.payload[apiKeyField]; !ok || !ValidAPIKey(apiKey) {
		return ErrInvalidAPIKey
	}

	if username, ok := p.payload[usernameField]; !ok || username == "" {
		return ErrInvalidPublicAPIKey
	}

	return nil
}

// UsernamePasswordProvider is an AuthenticationProvider using an email and password
type UsernamePasswordProvider struct {
	providerType ProviderType
	payload      map[string]string
}

// NewUsernamePasswordProvider returns a new UsernamePasswordProvider given a username and password
func NewUsernamePasswordProvider(username, password string) *UsernamePasswordProvider {
	return &UsernamePasswordProvider{
		providerType: ProviderTypeUsernamePassword,
		payload: map[string]string{
			usernameField: username,
			passwordField: password,
		},
	}
}

// Validate will determine if a given provider is valid
func (p *UsernamePasswordProvider) Validate() error {
	if username, ok := p.payload[usernameField]; !ok || username == "" {
		return ErrInvalidPublicAPIKey
	}

	if password, ok := p.payload[passwordField]; !ok || password == "" {
		return ErrInvalidPassword
	}

	return nil
}

// Type returns the auth provider type
func (p *UsernamePasswordProvider) Type() ProviderType {
	return p.providerType
}

// Payload returns the provider's auth payload
func (p *UsernamePasswordProvider) Payload() map[string]string {
	return p.payload
}
