package user

import (
	"strings"
)

// Session is the CLI profile session
type Session struct {
	AccessToken  string
	RefreshToken string
}

// Credentials are the user credentials
type Credentials struct {
	PublicAPIKey  string
	PrivateAPIKey string
	Username      string
	Password      string
}

// RedactKey returns the user's private API key with sensitive information redacted
func RedactKey(privateAPIKey string) string {
	parts := strings.Split(privateAPIKey, "-")
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return Redact(parts[0])
	default:
		lastIdx := len(parts) - 1

		out := make([]string, len(parts))
		for i := 0; i < lastIdx; i++ {
			out[i] = Redact(parts[i])
		}
		out[lastIdx] = parts[lastIdx]

		return strings.Join(out, "-")
	}
}

//Redact returns sensitive information redacted
func Redact(s string) string {
	return strings.Repeat("*", len(s))
}
