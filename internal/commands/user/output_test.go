package user

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestTableHeaders(t *testing.T) {
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
			assert.Equal(t, tc.expectedHeaders, tableHeaders(tc.authProviderType))
		})
	}
}

func TestTableRow(t *testing.T) {
	for _, tc := range []struct {
		description      string
		authProviderType realm.AuthProviderType
		user             realm.User
		expectedRow      map[string]interface{}
	}{
		{
			description:      "should show name for apikey type user",
			authProviderType: realm.AuthProviderTypeAPIKey,
			user: realm.User{
				ID:   "user-1",
				Type: "type-1",
				Data: map[string]interface{}{"name": "name-1"},
			},
			expectedRow: map[string]interface{}{
				"ID":   "user-1",
				"Name": "name-1",
				"Type": "type-1",
			},
		},
		{
			description:      "should show email for local-userpass type user",
			authProviderType: realm.AuthProviderTypeUserPassword,
			user: realm.User{
				ID:   "user-1",
				Type: "type-1",
				Data: map[string]interface{}{"email": "user-1@test.com"},
			},
			expectedRow: map[string]interface{}{
				"ID":    "user-1",
				"Email": "user-1@test.com",
				"Type":  "type-1",
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			output := userOutput{user: tc.user}
			row := tableRow(tc.authProviderType, output, func(uo userOutput, m map[string]interface{}) {})
			assert.Equal(t, tc.expectedRow, row)
		})
	}
}
