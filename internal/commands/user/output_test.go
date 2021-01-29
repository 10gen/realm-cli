package user

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestUserTableHeaders(t *testing.T) {
	for _, tc := range []struct {
		description      string
		authProviderType realm.AuthProviderType
		expectedHeaders  []string
	}{
		{
			description:      "should show name for apikey",
			authProviderType: realm.AuthProviderTypeAPIKey,
			expectedHeaders:  []string{"Name", "ID", "Type"},
		},
		{
			description:      "should show email for local-userpass",
			authProviderType: realm.AuthProviderTypeUserPassword,
			expectedHeaders:  []string{"Email", "ID", "Type"},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expectedHeaders, userTableHeaders(tc.authProviderType))
		})
	}
}

func TestUserTableRow(t *testing.T) {
	for _, tc := range []struct {
		description      string
		userEnable       bool
		authProviderType realm.AuthProviderType
		output           userOutput
		expectedRow      map[string]interface{}
	}{
		{
			description:      "should show name for apikey type user",
			userEnable:       false,
			authProviderType: realm.AuthProviderTypeAPIKey,
			output: userOutput{
				user: realm.User{
					ID:         "user-1",
					Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeAPIKey}},
					Type:       "type-1",
					Data:       map[string]interface{}{"name": "name-1"},
				},
				err: nil,
			},
			expectedRow: map[string]interface{}{
				"ID":   "user-1",
				"Name": "name-1",
				"Type": "type-1",
			},
		},
		{
			description:      "should show email for local-userpass type user",
			userEnable:       false,
			authProviderType: realm.AuthProviderTypeUserPassword,
			output: userOutput{
				user: realm.User{
					ID:         "user-1",
					Identities: []realm.UserIdentity{{ProviderType: realm.AuthProviderTypeUserPassword}},
					Type:       "type-1",
					Data:       map[string]interface{}{"email": "user-1@test.com"},
				},
				err: nil,
			},
			expectedRow: map[string]interface{}{
				"ID":    "user-1",
				"Email": "user-1@test.com",
				"Type":  "type-1",
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expectedRow, userTableRow(tc.authProviderType, tc.output, func(uo userOutput, m map[string]interface{}) {}))
		})
	}
}
