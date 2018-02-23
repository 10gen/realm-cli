package commands

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/10gen/stitch-cli/models"
	"github.com/10gen/stitch-cli/user"
	u "github.com/10gen/stitch-cli/utils/test"
	gc "github.com/smartystreets/goconvey/convey"

	"github.com/mitchellh/cli"
)

func TestImportCommand(t *testing.T) {
	setUpBasicCommand := func() (*ImportCommand, *cli.MockUi) {
		mockUI := cli.NewMockUi()
		cmd, err := NewImportCommandFactory(mockUI)()
		if err != nil {
			panic(err)
		}

		importCommand := cmd.(*ImportCommand)
		importCommand.storage = u.NewEmptyStorage()
		importCommand.writeToDirectory = func(dest string, r io.Reader) error {
			return nil
		}

		mockStitchClient := &u.MockStitchClient{
			ExportFn: func(groupID, appID string) (string, io.ReadCloser, error) {
				return "", u.NewResponseBody(bytes.NewReader([]byte{})), nil
			},
			ImportFn: func(groupID, appID string) error {
				return nil
			},
			FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
				return &models.App{
					GroupID: "group-id",
					ID:      "app-id",
				}, nil
			},
		}
		importCommand.stitchClient = mockStitchClient
		return importCommand, mockUI
	}

	validArgs := []string{"--app-id=my-app-abcdef"}

	t.Run("should require the user to be logged in", func(t *testing.T) {
		importCommand, mockUI := setUpBasicCommand()
		exitCode := importCommand.Run(validArgs)
		u.So(t, exitCode, gc.ShouldEqual, 1)

		u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, user.ErrNotLoggedIn.Error())
	})

	t.Run("when the user is logged in", func(t *testing.T) {
		setup := func() (*ImportCommand, *cli.MockUi) {
			importCommand, mockUI := setUpBasicCommand()

			importCommand.user = &user.User{
				APIKey:      "my-api-key",
				AccessToken: u.GenerateValidAccessToken(),
			}

			return importCommand, mockUI
		}

		type testCase struct {
			Description      string
			Args             []string
			ExpectedExitCode int
			ExpectedError    string
			WorkingDirectory string
			StitchClient     u.MockStitchClient
		}

		for _, tc := range []testCase{
			{
				Description:      "it does not import if the user does not confirm the diff",
				Args:             append([]string{"--path=../testdata/full_app"}, validArgs...),
				ExpectedExitCode: 0,
				StitchClient: u.MockStitchClient{
					ImportFn: func(groupID, appID string) error {
						return nil
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
				importCommand, mockUI := setup()

				// Mock a "no" response when we prompt the user to confirm the diff
				mockUI.InputReader = strings.NewReader("n\n")
				importCommand.stitchClient = &tc.StitchClient

				exitCode := importCommand.Run(tc.Args)

				mockClient := importCommand.stitchClient.(*u.MockStitchClient)

				u.So(t, exitCode, gc.ShouldEqual, tc.ExpectedExitCode)
				u.So(t, mockUI.ErrorWriter.String(), gc.ShouldBeEmpty)
				u.So(t, len(mockClient.ImportFnCalls), gc.ShouldEqual, 0)
			})
		}

		for _, tc := range []testCase{
			{
				Description:      "it fails if an app-id is not provided",
				Args:             []string{"--path=../testdata/simple_app"},
				ExpectedExitCode: 1,
				ExpectedError:    errAppIDRequired.Error(),
			},
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
				StitchClient: u.MockStitchClient{
					ExportFn: func(groupID, appID string) (string, io.ReadCloser, error) {
						return "", u.NewResponseBody(bytes.NewReader([]byte{})), nil
					},
					ImportFn: func(groupID, appID string) error {
						return nil
					},
					FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
						return &models.App{
							GroupID: "group-id",
							ID:      "app-id",
						}, nil
					},
				},
			},
			{
				Description:      "reports an error if it fails to fetch the app by clientID",
				Args:             append([]string{"--path=../testdata/full_app"}, validArgs...),
				ExpectedExitCode: 1,
				ExpectedError:    "oh no failed to fetch app",
				StitchClient: u.MockStitchClient{
					ExportFn: func(groupID, appID string) (string, io.ReadCloser, error) {
						return "", nil, fmt.Errorf("oh no")
					},
					ImportFn: func(groupID, appID string) error {
						return nil
					},
					FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
						return nil, fmt.Errorf("oh no failed to fetch app")
					},
				},
			},
			{
				Description:      "it fails with an error if the response to the import request is invalid",
				Args:             append([]string{"--path=../testdata/simple_app"}, validArgs...),
				ExpectedExitCode: 1,
				ExpectedError:    "oh noes",
				StitchClient: u.MockStitchClient{
					ImportFn: func(groupID, appID string) error {
						return fmt.Errorf("oh noes")
					},
					FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
						return &models.App{
							GroupID: "group-id",
							ID:      "app-id",
						}, nil
					},
				},
			},
			{
				Description:      "it succeeds if it can grab instance data from the app config file at the provided path",
				Args:             []string{"--path=../testdata/simple_app_with_instance_data"},
				ExpectedExitCode: 0,
				StitchClient: u.MockStitchClient{
					ExportFn: func(groupID, appID string) (string, io.ReadCloser, error) {
						return "", u.NewResponseBody(strings.NewReader("export response")), nil
					},
					ImportFn: func(groupID, appID string) error {
						return nil
					},
					FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
						return &models.App{
							GroupID: "group-id",
							ID:      "app-id",
						}, nil
					},
				},
			},
			{
				Description:      "it succeeds if it can grab instance data from the app config file in the current directory",
				Args:             []string{},
				ExpectedExitCode: 0,
				WorkingDirectory: "../testdata/simple_app_with_instance_data",
				StitchClient: u.MockStitchClient{
					ExportFn: func(groupID, appID string) (string, io.ReadCloser, error) {
						return "", u.NewResponseBody(strings.NewReader("export response")), nil
					},
					ImportFn: func(groupID, appID string) error {
						return nil
					},
					FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
						return &models.App{
							GroupID: "group-id",
							ID:      "app-id",
						}, nil
					},
				},
			},
			{
				Description:      "reports an error if it fails to export the app",
				Args:             append([]string{"--path=../testdata/full_app"}, validArgs...),
				ExpectedExitCode: 1,
				ExpectedError:    "failed to sync app",
				StitchClient: u.MockStitchClient{
					ExportFn: func(groupID, appID string) (string, io.ReadCloser, error) {
						return "", nil, fmt.Errorf("oh no")
					},
					ImportFn: func(groupID, appID string) error {
						return nil
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
				importCommand, mockUI := setup()

				// Mock a "yes" response when we prompt the user to confirm the diff
				mockUI.InputReader = strings.NewReader("y\n")
				importCommand.stitchClient = &tc.StitchClient
				importCommand.workingDirectory = tc.WorkingDirectory

				exitCode := importCommand.Run(tc.Args)

				u.So(t, exitCode, gc.ShouldEqual, tc.ExpectedExitCode)
				u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, tc.ExpectedError)
			})
		}

		t.Run("syncing data after a successful import", func(t *testing.T) {
			t.Run("on success", func(t *testing.T) {
				type testCase struct {
					Description       string
					Args              []string
					WorkingDirectory  string
					ExpectedDirectory string
				}

				for _, tc := range []testCase{
					{
						Description:       "it writes data to the provided directory",
						Args:              append([]string{"--path=../testdata/simple_app"}, validArgs...),
						WorkingDirectory:  "",
						ExpectedDirectory: abs("../testdata/simple_app"),
					},
					{
						Description:       "it writes data to the working directory when using app config file data",
						Args:              []string{},
						WorkingDirectory:  "../testdata/simple_app_with_instance_data",
						ExpectedDirectory: abs("../testdata/simple_app_with_instance_data"),
					},
				} {
					t.Run(tc.Description, func(t *testing.T) {
						importCommand, mockUI := setup()
						mockUI.InputReader = strings.NewReader("y\n")
						importCommand.workingDirectory = tc.WorkingDirectory
						mockStitchClient := &u.MockStitchClient{
							ExportFn: func(groupID, appID string) (string, io.ReadCloser, error) {
								return "", u.NewResponseBody(strings.NewReader("export response")), nil
							},
							ImportFn: func(groupID, appID string) error {
								return nil
							},
							FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
								return &models.App{
									GroupID: "group-id",
									ID:      "app-id",
								}, nil
							},
						}
						importCommand.stitchClient = mockStitchClient

						destinationDirectory := ""
						writeContent := ""

						importCommand.writeToDirectory = func(dest string, zipData io.Reader) error {
							b, err := ioutil.ReadAll(zipData)
							if err != nil {
								return err
							}
							destinationDirectory = dest
							writeContent = string(b)
							return nil
						}

						exitCode := importCommand.Run(tc.Args)
						u.So(t, exitCode, gc.ShouldEqual, 0)
						u.So(t, mockUI.ErrorWriter.String(), gc.ShouldBeEmpty)
						u.So(t, abs(destinationDirectory), gc.ShouldEqual, tc.ExpectedDirectory)
						u.So(t, writeContent, gc.ShouldEqual, "export response")
					})
				}
			})
		})
	})
}

func abs(path string) string {
	p, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return p
}
