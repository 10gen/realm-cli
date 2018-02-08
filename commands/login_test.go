package commands

import (
	"net/http"
	"strings"
	"testing"

	"github.com/10gen/stitch-cli/auth"
	"github.com/10gen/stitch-cli/user"
	u "github.com/10gen/stitch-cli/utils/test"
	gc "github.com/smartystreets/goconvey/convey"

	"github.com/mitchellh/cli"
)

func TestLoginCommand(t *testing.T) {
	t.Run("required arguments", func(t *testing.T) {
		setup := func() (*LoginCommand, *cli.MockUi) {
			mockUI := cli.NewMockUi()
			cmd, err := NewLoginCommandFactory(mockUI)()
			if err != nil {
				panic(err)
			}

			loginCommand := cmd.(*LoginCommand)
			loginCommand.storage = u.NewEmptyStorage()

			return loginCommand, mockUI
		}

		t.Run("should require an api key", func(t *testing.T) {
			loginCommand, mockUI := setup()
			exitCode := loginCommand.Run([]string{})
			u.So(t, exitCode, gc.ShouldEqual, 1)

			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, errAPIKeyRequired.Error())
		})

		t.Run("should require a username", func(t *testing.T) {
			loginCommand, mockUI := setup()
			exitCode := loginCommand.Run([]string{`--api-key=my-api-key`})
			u.So(t, exitCode, gc.ShouldEqual, 1)

			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, errUsernameRequired.Error())
		})
	})

	t.Run("when the user is not logged in", func(t *testing.T) {
		setup := func() *LoginCommand {
			mockUI := cli.NewMockUi()
			cmd, err := NewLoginCommandFactory(mockUI)()
			if err != nil {
				panic(err)
			}

			loginCommand := cmd.(*LoginCommand)

			mockClient := u.NewMockClient(
				[]*http.Response{
					{
						StatusCode: http.StatusOK,
						Body: u.NewAuthResponseBody(auth.Response{
							AccessToken:  "new.access.token",
							RefreshToken: "new.refresh.token",
						}),
					},
				},
			)

			loginCommand.client = mockClient
			loginCommand.storage = u.NewEmptyStorage()

			return loginCommand
		}

		t.Run("logs the user in and updates auth data", func(t *testing.T) {
			loginCommand := setup()
			exitCode := loginCommand.Run([]string{`--api-key=my-api-key`, `--username=my.username`})
			u.So(t, exitCode, gc.ShouldEqual, 0)

			validUser := &user.User{
				APIKey:       "my-api-key",
				Username:     "my.username",
				AccessToken:  "new.access.token",
				RefreshToken: "new.refresh.token",
			}

			u.So(t, loginCommand.user, gc.ShouldResemble, validUser)

			storedUser, err := loginCommand.storage.ReadUserConfig()
			u.So(t, err, gc.ShouldBeNil)

			u.So(t, storedUser, gc.ShouldResemble, validUser)
		})
	})

	t.Run("when the user is logged in", func(t *testing.T) {
		setup := func() (*LoginCommand, *cli.MockUi, *u.MockClient) {
			mockUI := cli.NewMockUi()
			cmd, err := NewLoginCommandFactory(mockUI)()
			if err != nil {
				panic(err)
			}

			loginCommand := cmd.(*LoginCommand)

			mockClient := u.NewMockClient(
				[]*http.Response{
					{
						StatusCode: http.StatusOK,
						Body: u.NewAuthResponseBody(auth.Response{
							AccessToken:  "new.access.token",
							RefreshToken: "new.refresh.token",
						}),
					},
				},
			)

			loginCommand.client = mockClient
			loginCommand.user = &user.User{
				APIKey: "my-existing-api-key",
			}
			loginCommand.storage = u.NewEmptyStorage()

			return loginCommand, mockUI, mockClient
		}

		t.Run("does not prompt the user when the `y` flag has been provided", func(t *testing.T) {
			loginCommand, _, _ := setup()
			exitCode := loginCommand.Run([]string{`--api-key=my-api-key`, `--username=my.username`, `-y`})
			u.So(t, exitCode, gc.ShouldEqual, 0)

			validUser := &user.User{
				APIKey:       "my-api-key",
				Username:     "my.username",
				AccessToken:  "new.access.token",
				RefreshToken: "new.refresh.token",
			}

			u.So(t, loginCommand.user, gc.ShouldResemble, validUser)

			storedUser, err := loginCommand.storage.ReadUserConfig()
			u.So(t, err, gc.ShouldBeNil)

			u.So(t, storedUser, gc.ShouldResemble, validUser)
		})

		t.Run("prompts the user for confirmation and cancels the request if 'n' is entered", func(t *testing.T) {
			loginCommand, mockUI, _ := setup()

			mockUI.InputReader = strings.NewReader("n\n")
			exitCode := loginCommand.Run([]string{`--api-key=my-api-key`, `--username=my.username`})
			u.So(t, exitCode, gc.ShouldEqual, 0)

			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "you are already logged in, this action will deauthenticate the existing user")
			u.So(t, loginCommand.user.APIKey, gc.ShouldEqual, "my-existing-api-key")
		})

		t.Run("prompts the user for confirmation and continues if 'y' is entered", func(t *testing.T) {
			loginCommand, mockUI, _ := setup()

			mockUI.InputReader = strings.NewReader("y\n")
			exitCode := loginCommand.Run([]string{`--api-key=my-api-key`, `--username=my.username`})
			u.So(t, exitCode, gc.ShouldEqual, 0)

			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "you are already logged in, this action will deauthenticate the existing user")

			validUser := &user.User{
				APIKey:       "my-api-key",
				Username:     "my.username",
				AccessToken:  "new.access.token",
				RefreshToken: "new.refresh.token",
			}
			u.So(t, loginCommand.user, gc.ShouldResemble, validUser)
		})
	})
}
