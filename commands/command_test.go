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

const expiredAccessToken string = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MTgwMzE0ODAsImlhdCI6MTUxODAyOTY4MCwiaXNzIjoiNWE3YjRiNzA3YTRhMTExYzQ5ZjFmZjQwIiwic3RpdGNoX2RhdGEiOnsibW9uZ29kYi9jbG91ZC1hcGlLZXkiOiJjd0FBQUFWMllXeDFaUUJRQUFBQUFOZzQ1NFFIUHVsQ0U3UTNzaVI2b0NlTkZZaFlXbzZtZjIwakRtYjRpV0tWczlTU0VhYUF6bUtVN25mRUxmRHNBZVBiSW1XQUlHa0FJMG1KTVhieng5Q3hSOVRRSGdIOS9OT3ZLQkp3VEZNc0NHVnVZM0o1Y0hSbFpGOTJZV3gxWlFBQkFBPT0iLCJtb25nb2RiL2Nsb3VkLXVzZXJuYW1lIjoiWXdBQUFBVjJZV3gxWlFCQUFBQUFBTTBwanYreHdhbk5EQXBFL2VrTG5RbmJuTCt0d3JWa2twV3ZmYVRsOUdDNmYybWd2NTRidklxMGU1dFZmZlNaTjZDUFc0UkJqSjlIQ3AzdGc3SUgySGtJWlc1amNubHdkR1ZrWDNaaGJIVmxBQUVBIn0sInN0aXRjaF9kZXZJZCI6IjAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMCIsInN0aXRjaF9kb21haW5JZCI6IjAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMCIsInN1YiI6IjU5ZTBmMWMzM2JmOTk0MjU1M2Q4ZWM1ZCIsInR5cCI6ImFjY2VzcyJ9.CdH7tgQ6gWbZigsM2ksgD3TNzHhrajsCfRE6uTNDlEE"
const updatedAccessToken string = "updated.access.token"

func TestBaseCommandClient(t *testing.T) {
	setup := func() *BaseCommand {
		return &BaseCommand{}
	}

	t.Run("should set the client", func(t *testing.T) {
		base := setup()
		u.So(t, base.client, gc.ShouldBeNil)

		_, err := base.Client()
		u.So(t, err, gc.ShouldBeNil)

		u.So(t, base.client, gc.ShouldNotBeNil)
	})
}

func TestBaseCommandUser(t *testing.T) {
	setup := func() *BaseCommand {
		return &BaseCommand{
			storage: u.NewPopulatedStorage("key", "refresh", "access"),
		}
	}

	t.Run("should load the user from storage", func(t *testing.T) {
		base := setup()
		u.So(t, base.user, gc.ShouldBeNil)

		usr, err := base.User()
		u.So(t, err, gc.ShouldBeNil)

		u.So(t, usr.APIKey, gc.ShouldEqual, "key")
		u.So(t, usr.AccessToken, gc.ShouldEqual, "access")
		u.So(t, usr.RefreshToken, gc.ShouldEqual, "refresh")
	})

	t.Run("should set the user", func(t *testing.T) {
		base := setup()
		u.So(t, base.user, gc.ShouldBeNil)

		_, err := base.User()
		u.So(t, err, gc.ShouldBeNil)

		u.So(t, base.user, gc.ShouldNotBeNil)
	})
}

func TestBaseCommandAuthClient(t *testing.T) {
	t.Run("with an empty token", func(t *testing.T) {
		setup := func() *BaseCommand {
			return &BaseCommand{
				user:    &user.User{},
				storage: u.NewEmptyStorage(),
			}
		}

		t.Run("should return an error", func(t *testing.T) {
			base := setup()
			_, err := base.AuthClient()
			u.So(t, err, gc.ShouldEqual, auth.ErrInvalidToken)
		})
	})

	t.Run("with an invalid token", func(t *testing.T) {
		setup := func() *BaseCommand {
			return &BaseCommand{
				user: &user.User{
					AccessToken: "invalid.access.token",
				},
				storage: u.NewEmptyStorage(),
			}
		}

		t.Run("should return an error", func(t *testing.T) {
			base := setup()
			_, err := base.AuthClient()
			u.So(t, err, gc.ShouldEqual, auth.ErrInvalidToken)
		})
	})

	t.Run("with an expired token", func(t *testing.T) {
		setup := func() *BaseCommand {
			mockClient := u.NewMockClient(
				[]*http.Response{
					{
						StatusCode: http.StatusCreated,
						Body: u.NewAuthResponseBody(auth.Response{
							AccessToken: updatedAccessToken,
						}),
					},
				},
			)

			return &BaseCommand{
				user: &user.User{
					AccessToken: expiredAccessToken,
				},
				client:  mockClient,
				storage: u.NewEmptyStorage(),
			}
		}

		t.Run("should update the user's access token", func(t *testing.T) {
			base := setup()
			_, err := base.AuthClient()
			u.So(t, err, gc.ShouldBeNil)

			u.So(t, base.user.AccessToken, gc.ShouldEqual, updatedAccessToken)
		})

		t.Run("should update the stored access token", func(t *testing.T) {
			base := setup()
			_, err := base.AuthClient()
			u.So(t, err, gc.ShouldBeNil)

			userFromStorage, err := base.storage.ReadUserConfig()
			u.So(t, err, gc.ShouldBeNil)
			u.So(t, userFromStorage.AccessToken, gc.ShouldEqual, updatedAccessToken)
		})
	})
}

func TestBaseCommandAsk(t *testing.T) {
	t.Run("should handle valid input", func(t *testing.T) {
		type testCase struct {
			input          string
			expectedResult bool
			expectedErr    error
		}

		testCases := []testCase{
			{
				input:          "y",
				expectedResult: true,
				expectedErr:    nil,
			},
			{
				input:          "   y   ",
				expectedResult: true,
				expectedErr:    nil,
			},
			{
				input:          "yes",
				expectedResult: true,
				expectedErr:    nil,
			},
			{
				input:          "Y",
				expectedResult: true,
				expectedErr:    nil,
			},
			{
				input:          "YES",
				expectedResult: true,
				expectedErr:    nil,
			},
			{
				input:          "n",
				expectedResult: false,
				expectedErr:    nil,
			},
			{
				input:          "   n   ",
				expectedResult: false,
				expectedErr:    nil,
			},
			{
				input:          "N",
				expectedResult: false,
				expectedErr:    nil,
			},
			{
				input:          "no",
				expectedResult: false,
				expectedErr:    nil,
			},
			{
				input:          "NO",
				expectedResult: false,
				expectedErr:    nil,
			},
		}

		for _, tc := range testCases {
			mockUI := cli.NewMockUi()
			baseCommand := &BaseCommand{
				UI: mockUI,
			}

			mockUI.InputReader = strings.NewReader(tc.input + "\n")

			result, err := baseCommand.AskYesNo("lemme ask you a question [y/n]: ")
			u.So(t, err, gc.ShouldEqual, tc.expectedErr)
			u.So(t, result, gc.ShouldEqual, tc.expectedResult)
		}
	})

	t.Run("should prompt the user again for invalid input", func(t *testing.T) {
		type testCase struct {
			input string
		}

		testCases := []testCase{
			{
				input: "nah",
			},
		}

		for _, tc := range testCases {
			mockUI := cli.NewMockUi()
			baseCommand := &BaseCommand{
				UI: mockUI,
			}

			mockUI.InputReader = strings.NewReader(tc.input + "\n no\n")
			_, err := baseCommand.AskYesNo("lemme ask you a question [y/n]: ")
			u.So(t, err, gc.ShouldBeNil)

			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "Could not understand response")
		}
	})
}
