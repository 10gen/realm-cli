package auth

import (
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestUser(t *testing.T) {
	t.Run("User should redact its private API key by displaying only the last portion", func(t *testing.T) {
		for _, tc := range []struct {
			description   string
			privateAPIKey string
			display       string
		}{
			{
				description:   "With an empty key",
				privateAPIKey: "",
				display:       "",
			},
			{
				description:   "With a key that has no dashes",
				privateAPIKey: "password",
				display:       "********",
			},
			{
				description:   "With a key that has one dash",
				privateAPIKey: "api-key",
				display:       "***-key",
			},
			{
				description:   "With a key that has many dashes",
				privateAPIKey: "some-super-secret-key",
				display:       "****-*****-******-key",
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				user := User{PrivateAPIKey: tc.privateAPIKey}
				assert.Equal(t, tc.display, user.RedactedPrivateAPIKey())
			})
		}
	})
}
