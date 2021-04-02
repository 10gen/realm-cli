package commands

import (
	"io"
	"testing"

	"github.com/10gen/realm-cli/utils/telemetry"

	"github.com/10gen/realm-cli/models"
	"github.com/10gen/realm-cli/user"
	u "github.com/10gen/realm-cli/utils/test"
	"github.com/mitchellh/cli"
	gc "github.com/smartystreets/goconvey/convey"
)

func setUpBasicDiffCommand() (*DiffCommand, *cli.MockUi) {
	mockUI := cli.NewMockUi()
	mockService := &telemetry.Service{}
	cmd, err := NewDiffCommandFactory(mockUI, mockService)()
	if err != nil {
		panic(err)
	}

	diffCommand := cmd.(*DiffCommand)
	diffCommand.storage = u.NewEmptyStorage()
	diffCommand.writeToDirectory = func(dest string, r io.Reader, overwrite bool) error {
		return nil
	}
	diffCommand.writeAppConfigToFile = func(dest string, app models.AppInstanceData) error {
		return nil
	}

	mockRealmClient := &u.MockRealmClient{
		DiffFn: func(groupID, appID string, appData []byte, strategy string) ([]string, error) {
			return []string{"sample-diff-contents"}, nil
		},
		FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
			return &models.App{
				GroupID: "group-id",
				ID:      "app-id",
			}, nil
		},
	}
	diffCommand.realmClient = mockRealmClient
	return diffCommand, mockUI
}

func TestDiffCommand(t *testing.T) {
	validArgs := []string{"--app-id=my-app-abcdef"}

	t.Run("should require the user to be logged in", func(t *testing.T) {
		diffCommand, mockUI := setUpBasicDiffCommand()
		exitCode := diffCommand.Run(validArgs)
		u.So(t, exitCode, gc.ShouldEqual, 1)

		u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, user.ErrNotLoggedIn.Error())
	})

	t.Run("when the user is logged in", func(t *testing.T) {
		setup := func() (*DiffCommand, *cli.MockUi) {
			diffCommand, mockUI := setUpBasicDiffCommand()

			diffCommand.user = &user.User{
				APIKey:      "my-api-key",
				AccessToken: u.GenerateValidAccessToken(),
			}

			return diffCommand, mockUI
		}

		type testCase struct {
			Description      string
			Args             []string
			ExpectedExitCode int
			ExpectedError    string
			WorkingDirectory string
			RealmClient      u.MockRealmClient
		}

		for _, tc := range []testCase{
			{
				Description:      "it fails if given an invalid flagAppPath",
				Args:             append([]string{"--path=/somewhere/bogus"}, validArgs...),
				ExpectedExitCode: 1,
				ExpectedError:    "directory does not exist",
			},
			{
				Description:      "it succeeds if given a valid flagAppPath",
				Args:             append([]string{"--path=../testdata/full_app"}, validArgs...),
				ExpectedExitCode: 0,
				RealmClient: u.MockRealmClient{
					DiffFn: func(groupID, appID string, appData []byte, strategy string) ([]string, error) {
						return []string{"sample-diff-contents"}, nil
					},
					FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
						return &models.App{
							GroupID: "group-id",
							ID:      "app-id",
						}, nil
					},
				},
			},
		} {
			t.Run(tc.Description, func(t *testing.T) {
				diffCommand, mockUI := setup()

				diffCommand.realmClient = &tc.RealmClient
				diffCommand.workingDirectory = tc.WorkingDirectory

				exitCode := diffCommand.Run(tc.Args)
				u.So(t, exitCode, gc.ShouldEqual, tc.ExpectedExitCode)
				u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, tc.ExpectedError)
			})
		}

	})

}
