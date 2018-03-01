package commands

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/10gen/stitch-cli/models"
	"github.com/10gen/stitch-cli/user"
	u "github.com/10gen/stitch-cli/utils/test"
	gc "github.com/smartystreets/goconvey/convey"

	"github.com/mitchellh/cli"
)

func TestExportCommand(t *testing.T) {
	setup := func() (*ExportCommand, *cli.MockUi) {
		mockUI := cli.NewMockUi()
		cmd, err := NewExportCommandFactory(mockUI)()
		if err != nil {
			panic(err)
		}

		exportCommand := cmd.(*ExportCommand)
		exportCommand.storage = u.NewEmptyStorage()
		return exportCommand, mockUI
	}

	t.Run("should require an app-id", func(t *testing.T) {
		exportCommand, mockUI := setup()
		exitCode := exportCommand.Run([]string{})
		u.So(t, exitCode, gc.ShouldEqual, 1)

		u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, errAppIDRequired.Error())
	})

	t.Run("should require the user to be logged in", func(t *testing.T) {
		exportCommand, mockUI := setup()
		exitCode := exportCommand.Run([]string{`--app-id=my-cool-app`})
		u.So(t, exitCode, gc.ShouldEqual, 1)

		u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, user.ErrNotLoggedIn.Error())
	})

	t.Run("when the user is logged in", func(t *testing.T) {
		setup := func() (*ExportCommand, *cli.MockUi) {
			mockUI := cli.NewMockUi()
			cmd, err := NewExportCommandFactory(mockUI)()
			if err != nil {
				panic(err)
			}

			exportCommand := cmd.(*ExportCommand)

			return exportCommand, mockUI
		}

		t.Run("on success", func(t *testing.T) {
			type testCase struct {
				Description         string
				ExpectedDestination string
				Args                []string
			}

			dest := "my_app_123456"
			zipData := "myZipData"
			appID := "my-cool-app"

			for _, tc := range []testCase{
				{
					Description:         "it writes response data to the default directory",
					ExpectedDestination: dest,
					Args:                []string{`--app-id=` + appID},
				},
				{
					Description:         "it writes response data to the provided destination directory",
					ExpectedDestination: "some/other/directory/my_app",
					Args:                []string{`--app-id=` + appID, `--output=some/other/directory/my_app`},
				},
			} {
				t.Run(tc.Description, func(t *testing.T) {
					exportCommand, mockUI := setup()

					responseGroupID := "group-id"
					responseAppID := "app-id"

					mockStitchClient := u.MockStitchClient{
						FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
							u.So(t, clientAppID, gc.ShouldEqual, appID)

							return &models.App{
								ClientAppID: clientAppID,
								GroupID:     responseGroupID,
								ID:          responseAppID,
							}, nil
						},
						ExportFn: func(groupID, appID string) (string, io.ReadCloser, error) {
							u.So(t, groupID, gc.ShouldEqual, responseGroupID)
							u.So(t, appID, gc.ShouldEqual, responseAppID)

							return dest, u.NewResponseBody(strings.NewReader(zipData)), nil
						},
					}

					exportCommand.stitchClient = &mockStitchClient
					exportCommand.user = &user.User{
						APIKey:      "my-api-key",
						AccessToken: u.GenerateValidAccessToken(),
					}

					destination := ""
					var zipData string

					exportCommand.exportToDirectory = func(dest string, r io.Reader) error {
						destination = dest
						b, err := ioutil.ReadAll(r)
						if err != nil {
							panic(err)
						}
						zipData = string(b)
						return nil
					}

					exitCode := exportCommand.Run(tc.Args)
					u.So(t, exitCode, gc.ShouldEqual, 0)
					u.So(t, mockUI.ErrorWriter.String(), gc.ShouldBeEmpty)
					u.So(t, destination, gc.ShouldEqual, tc.ExpectedDestination)
					u.So(t, zipData, gc.ShouldEqual, zipData)
				})
			}
		})

		t.Run("returns an error when the response from the API is unexpected", func(t *testing.T) {
			exportCommand, mockUI := setup()

			mockStitchClient := u.MockStitchClient{
				FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
					return &models.App{
						ClientAppID: clientAppID,
						GroupID:     "group-id",
						ID:          "app-id",
					}, nil
				},
				ExportFn: func(groupID, appID string) (string, io.ReadCloser, error) {
					return "", nil, fmt.Errorf("oh noes")
				},
			}

			exportCommand.stitchClient = &mockStitchClient
			exportCommand.user = &user.User{
				APIKey:      "my-api-key",
				AccessToken: u.GenerateValidAccessToken(),
			}

			exitCode := exportCommand.Run([]string{`--app-id=my-cool-app`})
			u.So(t, exitCode, gc.ShouldEqual, 1)

			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, "oh noes")
		})
	})
}
