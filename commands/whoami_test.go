package commands

import (
	"net/http"
	"testing"

	"github.com/10gen/stitch-cli/storage"
	"github.com/10gen/stitch-cli/user"
	u "github.com/10gen/stitch-cli/utils/test"

	"github.com/mitchellh/cli"
	gc "github.com/smartystreets/goconvey/convey"
)

func TestWhoamiCommand(t *testing.T) {
	setup := func(inMemoryUser *user.User, storage *storage.Storage) (*WhoamiCommand, *cli.MockUi) {
		mockUI := cli.NewMockUi()
		cmd, err := NewWhoamiCommandFactory(mockUI)()
		if err != nil {
			panic(err)
		}

		whoamiCommand := cmd.(*WhoamiCommand)

		whoamiCommand.client = u.NewMockClient([]*http.Response{})
		whoamiCommand.user = inMemoryUser
		whoamiCommand.storage = storage

		return whoamiCommand, mockUI
	}

	for _, testCase := range []struct {
		description      string
		user             *user.User
		storage          (func() *storage.Storage)
		expectedExitCode int
		expectedMessage  string
	}{
		{
			description: "with no user present",
			storage: func() *storage.Storage {
				return u.NewEmptyStorage()
			},
			expectedExitCode: 0,
			expectedMessage:  "no user info available",
		},
		{
			description: "after reading in memory user details",
			user: &user.User{
				Username: "in-memory.username",
				APIKey:   "in-memory-api-key",
			},
			storage: func() *storage.Storage {
				return u.NewEmptyStorage()
			},
			expectedExitCode: 0,
			expectedMessage:  "in-memory.username [API Key: **-******-***-key]",
		},
		{
			description: "after reading in storage user details",
			storage: func() *storage.Storage {
				usr := &user.User{
					Username: "storage.username",
					APIKey:   "storage-api-key",
				}

				strg := u.NewEmptyStorage()
				strg.WriteUserConfig(usr)
				return strg
			},
			expectedExitCode: 0,
			expectedMessage:  "storage.username [API Key: *******-***-key]",
		},
	} {
		t.Run("it displays the correct information "+testCase.description, func(t *testing.T) {
			whoamiCommand, mockUI := setup(testCase.user, testCase.storage())

			exitCode := whoamiCommand.Run([]string{})
			u.So(t, exitCode, gc.ShouldEqual, testCase.expectedExitCode)

			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, testCase.expectedMessage)
		})
	}
}
