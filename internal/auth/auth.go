package auth

import (
	"strings"
)

// Service is an auth service
type Service interface {
	ClearSession()
	Save() error
	Session() Session
	SetSession(session Session)
	User() User
	SetUser(user User)
}

// Session is the CLI profile session
type Session struct {
	AccessToken  string
	RefreshToken string
}

// User is the CLI profile user
type User struct {
	PublicAPIKey  string
	PrivateAPIKey string
}

// RedactedPrivateAPIKey returns the user's private API key with sensitive information redacted
func (u User) RedactedPrivateAPIKey() string {
	redact := func(s string) string {
		return strings.Repeat("*", len(s))
	}

	parts := strings.Split(u.PrivateAPIKey, "-")
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return redact(parts[0])
	default:
		lastIdx := len(parts) - 1

		out := make([]string, len(parts))
		for i := 0; i < lastIdx; i++ {
			out[i] = redact(parts[i])
		}
		out[lastIdx] = parts[lastIdx]

		return strings.Join(out, "-")
	}
}
