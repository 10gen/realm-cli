package realm

import (
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

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
