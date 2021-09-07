package realm_test

import (
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TODO(REALMC-9207): Unskip tests once backend is fully implemented
func TestRealmIPAccess(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	t.Run("should fail without an auth client", func(t *testing.T) {
		client := realm.NewClient(u.RealmServerURL())

		_, err := client.AllowedIPCreate(primitive.NewObjectID().Hex(), primitive.NewObjectID().Hex(), "0.0.0.0/0", "comment", false)
		assert.Equal(t, realm.ErrInvalidSession{}, err)
	})

	t.Run("with an active session", func(t *testing.T) {
		client := newAuthClient(t)
		groupID := u.CloudGroupID()

		testApp, teardown := setupTestApp(t, client, groupID, "accesslist-test")
		defer teardown()

		t.Run("should have no allowed ips upon app initialization", func(t *testing.T) {
			allowedIPs, err := client.AllowedIPs(groupID, testApp.ID)
			assert.Nil(t, err)
			assert.Equal(t, 0, len(allowedIPs))
		})

		t.Run("should create an allowed ip", func(t *testing.T) {
			address := "0.0.0.0"
			comment := "comment"
			useCurrent := false
			allowedIP, err := client.AllowedIPCreate(groupID, testApp.ID, address, comment, useCurrent)

			assert.Nil(t, err)
			assert.Equal(t, address, allowedIP.Address)
			assert.Equal(t, comment, allowedIP.Comment)

			t.Run("and list all app allowed ips", func(t *testing.T) {
				allowedIPs, err := client.AllowedIPs(groupID, testApp.ID)
				assert.Nil(t, err)
				assert.Equal(t, []realm.AllowedIP{allowedIP}, allowedIPs)
			})

			t.Run("and should update the allowed ip address", func(t *testing.T) {
				assert.Nil(t, client.AllowedIPUpdate(groupID, testApp.ID, allowedIP.ID, "1.1.1.1", "new comment"))

				t.Run("and list the new allowed ip address and comment", func(t *testing.T) {
					allowedIPs, err := client.AllowedIPs(groupID, testApp.ID)
					assert.Nil(t, err)
					assert.Equal(t, "1.1.1.1", allowedIPs[0].Address)
					assert.Equal(t, "new comment", allowedIPs[0].Comment)
				})

				t.Run("and return an error if we can't find the allowed ip", func(t *testing.T) {
					err := client.AllowedIPUpdate(groupID, testApp.ID, "notFound", "notUsed", "notUsed")
					assert.Equal(t, realm.ServerError{Message: "allowed IP not found: 'notFound'"}, err)
				})
			})

			t.Run("and should delete the allowed ip", func(t *testing.T) {
				assert.Nil(t, client.AllowedIPDelete(groupID, testApp.ID, allowedIP.Address))

				t.Run("and list no more allowed ips", func(t *testing.T) {
					allowedIPs, err := client.AllowedIPs(groupID, testApp.ID)
					assert.Nil(t, err)
					assert.Equal(t, []realm.AllowedIP{}, allowedIPs)
				})

				t.Run("and return an error if we can't find the allowed ip", func(t *testing.T) {
					err := client.AllowedIPDelete(groupID, testApp.ID, allowedIP.Address)
					assert.Equal(t, realm.ServerError{Message: fmt.Sprintf("allowed ip not found: '%s'", allowedIP.Address)}, err)
				})
			})
		})
	})
}
