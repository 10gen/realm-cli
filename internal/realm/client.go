package realm

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Client is a Realm client
type Client interface {
	Login(publicAPIKey, privateAPIKey string) (AuthResponse, error)
}

type client struct {
	baseURL string
}

// NewClient creates a new Realm client
func NewClient(baseURL string) Client {
	return &client{baseURL}
}

// AuthResponse is the Realm login response
type AuthResponse struct {
	AccessToken  string
	RefreshToken string
}

func (c *client) Login(publicAPIKey, privateAPIKey string) (AuthResponse, error) {
	// TODO(REALMC-7341): Implement login requests
	return AuthResponse{
		AccessToken:  primitive.NewObjectID().Hex(),
		RefreshToken: primitive.NewObjectID().Hex(),
	}, nil
}
