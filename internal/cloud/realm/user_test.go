package realm

import (
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestAuthProviderTypeDisplay(t *testing.T) {
	for _, tc := range []struct {
		apt            AuthProviderType
		expectedOutput string
	}{
		{AuthProviderTypeAnonymous, "Anonymous"},
		{AuthProviderTypeUserPassword, "User/Password"},
		{AuthProviderTypeAPIKey, "ApiKey"},
		{AuthProviderTypeApple, "Apple"},
		{AuthProviderTypeGoogle, "Google"},
		{AuthProviderTypeFacebook, "Facebook"},
		{AuthProviderTypeCustomToken, "Custom JWT"},
		{AuthProviderTypeCustomFunction, "Custom Function"},
		{AuthProviderType("invalid_provider_type"), "Unknown"},
	} {
		t.Run(fmt.Sprintf("should return %s", tc.expectedOutput), func(t *testing.T) {
			assert.Equal(t, tc.apt.Display(), tc.expectedOutput)
		})
	}
}

func TestNewAuthProviderTypes(t *testing.T) {
	for _, tc := range []struct {
		inSlice  []string
		outSlice AuthProviderTypes
	}{
		{
			inSlice: []string{"anon-user", "local-userpass", "api-key"},
			outSlice: AuthProviderTypes{
				AuthProviderTypeAnonymous,
				AuthProviderTypeUserPassword,
				AuthProviderTypeAPIKey,
			},
		},
		{
			inSlice: []string{"anon-user", "local-userpass", "api-key", "oauth2-facebook", "oauth2-google", "oauth2-apple"},
			outSlice: AuthProviderTypes{
				AuthProviderTypeAnonymous,
				AuthProviderTypeUserPassword,
				AuthProviderTypeAPIKey,
				AuthProviderTypeFacebook,
				AuthProviderTypeGoogle,
				AuthProviderTypeApple,
			},
		},
	} {
		t.Run("should return provider type slice", func(t *testing.T) {
			assert.Equal(t, NewAuthProviderTypes(tc.inSlice...), tc.outSlice)
		})
	}
}

func TestAuthProviderTypesJoin(t *testing.T) {
	for _, tc := range []struct {
		apts           AuthProviderTypes
		sep            string
		expectedOutput string
	}{
		{
			apts: []AuthProviderType{
				AuthProviderTypeAnonymous,
				AuthProviderTypeUserPassword,
				AuthProviderTypeAPIKey,
			},
			sep:            ",",
			expectedOutput: "anon-user,local-userpass,api-key",
		},
		{
			apts: []AuthProviderType{
				AuthProviderTypeAnonymous,
				AuthProviderTypeUserPassword,
				AuthProviderTypeAPIKey,
				AuthProviderTypeFacebook,
				AuthProviderTypeGoogle,
				AuthProviderTypeApple,
			},
			sep:            ", ",
			expectedOutput: "anon-user, local-userpass, api-key, oauth2-facebook, oauth2-google, oauth2-apple",
		},
	} {
		t.Run(fmt.Sprintf("should return %s", tc.expectedOutput), func(t *testing.T) {
			assert.Equal(t, tc.apts.join(tc.sep), tc.expectedOutput)
		})
	}
}
