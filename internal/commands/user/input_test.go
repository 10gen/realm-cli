package user

import (
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestProviderTypeDisplayUser(t *testing.T) {
	testUsers := []realm.User{
		{
			ID:         "user-1",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeAnonymous}},
		},
		{
			ID:         "user-2",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeUserPassword}},
			Data:       map[string]interface{}{"email": "user-2@test.com"},
		},
		{
			ID:         "user-3",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeAPIKey}},
			Data:       map[string]interface{}{"name": "name-3"},
		},
		{
			ID:         "user-4",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeApple}},
		},
		{
			ID:         "user-5",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeGoogle}},
		},
		{
			ID:         "user-6",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeFacebook}},
		},
		{
			ID:         "user-7",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeCustomToken}},
		},
		{
			ID:         "user-8",
			Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeCustomFunction}},
		},
	}
	for _, tc := range []struct {
		apt            realm.AuthProviderType
		user           realm.User
		expectedOutput string
	}{
		{
			apt:            realm.AuthProviderTypeAnonymous,
			user:           testUsers[0],
			expectedOutput: "Anonymous - user-1",
		},
		{
			apt:            realm.AuthProviderTypeUserPassword,
			user:           testUsers[1],
			expectedOutput: "User/Password - user-2@test.com - user-2",
		},
		{
			apt:            realm.AuthProviderTypeAPIKey,
			user:           testUsers[2],
			expectedOutput: "ApiKey - name-3 - user-3",
		},
		{
			apt:            realm.AuthProviderTypeApple,
			user:           testUsers[3],
			expectedOutput: "Apple - user-4",
		},
		{
			apt:            realm.AuthProviderTypeGoogle,
			user:           testUsers[4],
			expectedOutput: "Google - user-5",
		},
		{
			apt:            realm.AuthProviderTypeFacebook,
			user:           testUsers[5],
			expectedOutput: "Facebook - user-6",
		},
		{
			apt:            realm.AuthProviderTypeCustomToken,
			user:           testUsers[6],
			expectedOutput: "Custom JWT - user-7",
		},
		{
			apt:            realm.AuthProviderTypeCustomFunction,
			user:           testUsers[7],
			expectedOutput: "Custom Function - user-8",
		},
	} {
		t.Run(fmt.Sprintf("should return %s", tc.expectedOutput), func(t *testing.T) {
			assert.Equal(t, displayUser(tc.apt, tc.user), tc.expectedOutput)
		})
	}
}
