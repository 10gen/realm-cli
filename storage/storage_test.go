package storage_test

import (
	"testing"

	"github.com/10gen/realm-cli/user"
	u "github.com/10gen/realm-cli/utils/test"

	gc "github.com/smartystreets/goconvey/convey"
)

func TestStorageMigration(t *testing.T) {
	t.Run("supports migrating user data when writing", func(t *testing.T) {
		s := u.NewEmptyStorage()
		user := &user.User{
			Username: "my-username",
			APIKey:   "my-api-key",
		}

		err := s.WriteUserConfig(user)
		u.So(t, err, gc.ShouldBeNil)

		migratedUser, err := s.ReadUserConfig()
		u.So(t, err, gc.ShouldBeNil)

		u.So(t, migratedUser.PublicAPIKey, gc.ShouldEqual, user.Username)
		u.So(t, migratedUser.PrivateAPIKey, gc.ShouldEqual, user.APIKey)
	})

	t.Run("supports migrating user data when reading", func(t *testing.T) {
		s := u.NewPopulatedDeprecatedStorage("my-username", "my-api-key")

		migratedUser, err := s.ReadUserConfig()
		u.So(t, err, gc.ShouldBeNil)

		u.So(t, migratedUser.PublicAPIKey, gc.ShouldEqual, "my-username")
		u.So(t, migratedUser.PrivateAPIKey, gc.ShouldEqual, "my-api-key")
	})
}
