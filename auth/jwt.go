package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// JWT related errors
var (
	ErrInvalidToken = errors.New("the provided authentication token is invalid")
)

// NewJWT returns a new JWT or an error if the provided token is invalid
func NewJWT(s string) (*JWT, error) {
	if s == "" {
		return nil, ErrInvalidToken
	}

	return parse(s)
}

// JWT represents a basic auth token
type JWT struct {
	Exp int64 `json:"exp,omitempty"`
}

// Expired returns a boolean representing whether or not the token is expired
func (jwt *JWT) Expired() bool {
	return time.Now().Unix() > jwt.Exp
}

func parse(s string) (*JWT, error) {
	b, err := base64.StdEncoding.DecodeString(strings.Split(s, ".")[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var jwt JWT
	decoder := json.NewDecoder(bytes.NewReader(b))
	if err := decoder.Decode(&jwt); err != nil {
		return nil, ErrInvalidToken
	}

	return &jwt, nil
}
