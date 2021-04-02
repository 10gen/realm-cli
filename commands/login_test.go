package commands

import (
	"net/http"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/utils/telemetry"

	"github.com/10gen/realm-cli/auth"
	"github.com/10gen/realm-cli/user"
	u "github.com/10gen/realm-cli/utils/test"
	gc "github.com/smartystreets/goconvey/convey"

	"github.com/mitchellh/cli"
)

func TestLoginCommand(t *testing.T) {
	t.Run("required arguments", func(t *testing.T) {
		setup := func() (*LoginCommand, *cli.MockUi) {
			mockUI := cli.NewMockUi()
			mockService := &telemetry.Service{}
			cmd, err := NewLoginCommandFactory(mockUI, mockService)()
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

			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, auth.ErrInvalidAPIKey.Error())
		})

		t.Run("should require a username when using personal API keys", func(t *testing.T) {
			loginCommand, mockUI := setup()
			exitCode := loginCommand.Run([]string{`--api-key=my-api-key`})
			u.So(t, exitCode, gc.ShouldEqual, 1)

			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, auth.ErrInvalidPublicAPIKey.Error())
		})

		t.Run("should require an api key", func(t *testing.T) {
			loginCommand, mockUI := setup()
			exitCode := loginCommand.Run([]string{`--private-api-key=my-api-key`})
			u.So(t, exitCode, gc.ShouldEqual, 1)

			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, auth.ErrInvalidPublicAPIKey.Error())
		})
	})

	t.Run("when the user is not logged in", func(t *testing.T) {
		setup := func() (*LoginCommand, *cli.MockUi) {
			mockUI := cli.NewMockUi()
			mockService := &telemetry.Service{}
			cmd, err := NewLoginCommandFactory(mockUI, mockService)()
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

			return loginCommand, mockUI
		}

		t.Run("logs the user in and updates auth data", func(t *testing.T) {
			loginCommand, mockUI := setup()
			exitCode := loginCommand.Run([]string{`--api-key=my-api-key`, `--private-api-key=my-private-api-key`})
			u.So(t, exitCode, gc.ShouldEqual, 0)

			validUser := &user.User{
				PublicAPIKey:  "my-api-key",
				PrivateAPIKey: "my-private-api-key",
				AccessToken:   "new.access.token",
				RefreshToken:  "new.refresh.token",
			}

			u.So(t, loginCommand.user, gc.ShouldResemble, validUser)

			storedUser, err := loginCommand.storage.ReadUserConfig()
			u.So(t, err, gc.ShouldBeNil)

			u.So(t, storedUser, gc.ShouldResemble, validUser)

			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "you have successfully logged in as my-api-key")
		})
	})

	t.Run("when the user is logged in", func(t *testing.T) {
		setup := func() (*LoginCommand, *cli.MockUi, *u.MockClient) {
			mockUI := cli.NewMockUi()
			mockService := &telemetry.Service{}
			cmd, err := NewLoginCommandFactory(mockUI, mockService)()
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
				PublicAPIKey:  "user.name",
				PrivateAPIKey: "my-existing-api-key",
				AccessToken:   "my-existing-token",
				RefreshToken:  "my-refresh-token",
			}
			loginCommand.storage = u.NewEmptyStorage()

			return loginCommand, mockUI, mockClient
		}

		t.Run("does not prompt the user when the `y` flag has been provided", func(t *testing.T) {
			loginCommand, _, _ := setup()
			exitCode := loginCommand.Run([]string{`--api-key=my-api-key`, `--username=my.username`, `-y`})
			u.So(t, exitCode, gc.ShouldEqual, 0)

			validUser := &user.User{
				PublicAPIKey:  "my.username",
				PrivateAPIKey: "my-api-key",
				AccessToken:   "new.access.token",
				RefreshToken:  "new.refresh.token",
			}

			u.So(t, loginCommand.user, gc.ShouldResemble, validUser)

			storedUser, err := loginCommand.storage.ReadUserConfig()
			u.So(t, err, gc.ShouldBeNil)

			u.So(t, storedUser, gc.ShouldResemble, validUser)
		})

		t.Run("prompts the user for confirmation and cancels the request if 'n' is entered", func(t *testing.T) {
			loginCommand, mockUI, _ := setup()

			mockUI.InputReader = strings.NewReader("n\n")
			exitCode := loginCommand.Run([]string{`--api-key=my-api-key-updated`, `--username=my.username`})
			u.So(t, exitCode, gc.ShouldEqual, 0)

			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "you are already logged in as user.name, this action will deauthenticate the existing user [apiKey: **-********-***-key]")
			u.So(t, loginCommand.user.PrivateAPIKey, gc.ShouldEqual, "my-existing-api-key")
		})

		t.Run("prompts the user for confirmation and continues if 'y' is entered", func(t *testing.T) {
			loginCommand, mockUI, _ := setup()

			mockUI.InputReader = strings.NewReader("y\n")
			exitCode := loginCommand.Run([]string{`--api-key=my-api-key`, `--username=my.username`})
			u.So(t, exitCode, gc.ShouldEqual, 0)

			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "you are already logged in as user.name, this action will deauthenticate the existing user [apiKey: **-********-***-key]")

			validUser := &user.User{
				PublicAPIKey:  "my.username",
				PrivateAPIKey: "my-api-key",
				AccessToken:   "new.access.token",
				RefreshToken:  "new.refresh.token",
			}

			u.So(t, loginCommand.user, gc.ShouldResemble, validUser)
		})
	})
}
