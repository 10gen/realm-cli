package commands

import (
	"testing"

	"github.com/10gen/stitch-cli/storage"
	"github.com/10gen/stitch-cli/user"
	u "github.com/10gen/stitch-cli/utils/test"
	gc "github.com/smartystreets/goconvey/convey"

	"github.com/mitchellh/cli"
)

func TestLogoutCommand(t *testing.T) {
	setup := func(storage *storage.Storage) (*LogoutCommand, *cli.MockUi) {
		mockUI := cli.NewMockUi()
		cmd, err := NewLogoutCommandFactory(mockUI)()
		if err != nil {
			panic(err)
		}

		logoutCommand := cmd.(*LogoutCommand)
		logoutCommand.storage = storage

		return logoutCommand, mockUI
	}

	t.Run("clears out the storage", func(t *testing.T) {
		logoutCommand, _ := setup(u.NewPopulatedStorage("apikey", "refresh", "access"))

		res := logoutCommand.Run([]string{})
		u.So(t, res, gc.ShouldEqual, 0)

		storedUser, err := logoutCommand.storage.ReadUserConfig()
		u.So(t, err, gc.ShouldBeNil)
		u.So(t, storedUser, gc.ShouldResemble, &user.User{})
	})

	t.Run("plays nicely when the user is not logged in", func(t *testing.T) {
		logoutCommand, _ := setup(u.NewEmptyStorage())

		res := logoutCommand.Run([]string{})
		u.So(t, res, gc.ShouldEqual, 0)
	})
}
