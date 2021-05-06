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
}

// RedactedPrivateAPIKey returns the user's private API key with sensitive information redacted
func (creds Credentials) RedactedPrivateAPIKey() string {
	redact := func(s string) string {
		return strings.Repeat("*", len(s))
	}

	parts := strings.Split(creds.PrivateAPIKey, "-")
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
