package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/10gen/stitch-cli/api"
	"github.com/10gen/stitch-cli/hosting"
	"github.com/10gen/stitch-cli/models"
	"github.com/10gen/stitch-cli/user"
	"github.com/10gen/stitch-cli/utils"
	u "github.com/10gen/stitch-cli/utils/test"
	gc "github.com/smartystreets/goconvey/convey"

	"github.com/mitchellh/cli"
	"github.com/mitchellh/go-homedir"
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

	t.Run("help output", func(t *testing.T) {
		t.Run("describes --for-source-control", func(t *testing.T) {
			exportCommand, _ := setup()
			helpOutput := exportCommand.Help()
			u.So(t, helpOutput, gc.ShouldContainSubstring, "--for-source-control")
		})
	})

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
			exportCommand.storage = u.NewEmptyStorage()

			return exportCommand, mockUI
		}

		t.Run("on success", func(t *testing.T) {
			type testCase struct {
				Description          string
				ExpectedDestination  string
				ExpectedMetadataFile string
				Args                 []string

				ExpectedGroupID                         string
				FetchAppByClientIDInvocations           int
				FetchAppByGroupIDAndClientIDInvocations int
			}

			zipFileName := "my_app_123456.zip"
			zipData := "myZipData"
			appID := "my-cool-app-123456"

			assetDescriptions := []hosting.AssetDescription{
				{
					FilePath: "/bar/shouldRemainSame.txt",
					Attrs: []hosting.AssetAttribute{
						{Name: "Content-Type", Value: "html"},
					},
				},
				{
					FilePath: "/bar/attrsShouldAllRemain.html",
					Attrs: []hosting.AssetAttribute{
						{Name: "Content-Disposition", Value: "inline"},
						{Name: "Content-Type", Value: "htmp"},
						{Name: "Content-Language", Value: "fr"},
						{Name: "Content-Encoding", Value: "utf-8"},
						{Name: "Cache-Control", Value: "true"},
					},
				},
				{
					FilePath: "/bar/attrsShouldRemoveAllButOne.html",
					Attrs: []hosting.AssetAttribute{
						{Name: "Content-Disposition", Value: "inline"},
					},
				},
				{
					FilePath: "/bar/shouldBeRemoved",
					Attrs: []hosting.AssetAttribute{
						{Name: "Content-Type", Value: "htmp"},
						{Name: "Content-Language", Value: "fr"},
					},
				},
			}
			assetDescriptionData, err := json.Marshal(assetDescriptions)
			if err != nil {
				panic(err)
			}
			expectedMetadataFile := string(assetDescriptionData)
			expectedAssetFile := "here is my fake file it means nothing"

			homeDir, err := homedir.Dir()
			u.So(t, err, gc.ShouldBeNil)

			for _, tc := range []testCase{
				{
					Description:         "it writes response data to the default directory",
					ExpectedDestination: "my_app",
					Args:                []string{`--app-id=` + appID},

					ExpectedGroupID:               "group-id",
					FetchAppByClientIDInvocations: 1,
				},
				{
					Description:         "it overrides the project ID and writes response data to the default directory",
					ExpectedDestination: "my_app",
					Args:                []string{`--app-id=` + appID, `--project-id=project-id`},

					ExpectedGroupID:                         "project-id",
					FetchAppByGroupIDAndClientIDInvocations: 1,
				},
				{
					Description:         "it writes response data to the provided destination directory using the '--output' flag",
					ExpectedDestination: "some/other/directory/my_app",
					Args:                []string{`--app-id=` + appID, `--output=some/other/directory/my_app`},

					ExpectedGroupID:               "group-id",
					FetchAppByClientIDInvocations: 1,
				},
				{
					Description:         "it writes response data to an expanded home directory output path using the '--output' flag",
					ExpectedDestination: fmt.Sprintf("%s%smy_app", homeDir, string(os.PathSeparator)),
					Args:                []string{`--app-id=` + appID, `--output=~/my_app`},

					ExpectedGroupID:               "group-id",
					FetchAppByClientIDInvocations: 1,
				},
				{
					Description:         "it writes response data to the provided destination directory using the '-o' flag",
					ExpectedDestination: "some/other/directory/my_app",
					Args:                []string{`--app-id=` + appID, `-o`, `some/other/directory/my_app`},

					ExpectedGroupID:               "group-id",
					FetchAppByClientIDInvocations: 1,
				},
				{
					Description:         "it writes response data to an expanded home directory output path using the '-o' flag",
					ExpectedDestination: fmt.Sprintf("%s%smy_app", homeDir, string(os.PathSeparator)),
					Args:                []string{`--app-id=` + appID, `-o`, `~/my_app`},

					ExpectedGroupID:               "group-id",
					FetchAppByClientIDInvocations: 1,
				},
				{
					Description:          "it writes response data to the default directory and includes hosting assets",
					ExpectedDestination:  fmt.Sprintf("%s%smy_app", homeDir, string(os.PathSeparator)),
					Args:                 []string{`--app-id=` + appID, `--include-hosting=true`, `--output=~/my_app`},
					ExpectedMetadataFile: expectedMetadataFile,

					ExpectedGroupID:               "group-id",
					FetchAppByClientIDInvocations: 1,
				},
			} {
				t.Run(tc.Description, func(t *testing.T) {
					exportCommand, mockUI := setup()

					responseAppID := "app-id"

					var fetchAppByClientAppID, fetchAppByGroupIDAndClientAppID int

					mockStitchClient := u.MockStitchClient{
						FetchAppByGroupIDAndClientAppIDFn: func(groupID, clientAppID string) (*models.App, error) {
							fetchAppByGroupIDAndClientAppID++

							u.So(t, clientAppID, gc.ShouldEqual, appID)

							return &models.App{
								ClientAppID: clientAppID,
								GroupID:     groupID,
								ID:          responseAppID,
							}, nil
						},
						FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
							fetchAppByClientAppID++

							u.So(t, clientAppID, gc.ShouldEqual, appID)

							return &models.App{
								ClientAppID: clientAppID,
								GroupID:     "group-id",
								ID:          responseAppID,
							}, nil
						},
						ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
							u.So(t, groupID, gc.ShouldEqual, tc.ExpectedGroupID)
							u.So(t, appID, gc.ShouldEqual, responseAppID)

							return zipFileName, u.NewResponseBody(strings.NewReader(zipData)), nil
						},
					}

					exportCommand.stitchClient = &mockStitchClient
					exportCommand.user = &user.User{
						APIKey:      "my-api-key",
						AccessToken: u.GenerateValidAccessToken(),
					}

					destination := ""
					var zipData string

					exportCommand.exportToDirectory = func(dest string, r io.Reader, overwrite bool) error {
						destination = dest
						b, err := ioutil.ReadAll(r)
						if err != nil {
							return err
						}
						zipData = string(b)
						return nil
					}

					metadataStr := ""
					var fileStrs []string

					exportCommand.writeFileToDirectory = func(dest string, data io.Reader) error {
						b, err := ioutil.ReadAll(data)
						if err != nil {
							return err
						}

						if strings.HasSuffix(dest, utils.HostingAttributes) {
							metadataStr = string(b)
						} else {
							fileStrs = append(fileStrs, string(b))
						}
						return nil
					}

					exportCommand.getAssetAtURL = func(url string) (io.ReadCloser, error) {
						reader := strings.NewReader("here is my fake file it means nothing")
						return ioutil.NopCloser(reader), nil
					}

					exitCode := exportCommand.Run(tc.Args)
					u.So(t, exitCode, gc.ShouldEqual, 0)
					u.So(t, mockUI.ErrorWriter.String(), gc.ShouldBeEmpty)
					u.So(t, destination, gc.ShouldEqual, tc.ExpectedDestination)
					u.So(t, zipData, gc.ShouldEqual, zipData)
					u.So(t, metadataStr, gc.ShouldEqual, tc.ExpectedMetadataFile)
					for _, fileStr := range fileStrs {
						u.So(t, fileStr, gc.ShouldEqual, expectedAssetFile)
					}

					u.So(t, fetchAppByClientAppID, gc.ShouldEqual, tc.FetchAppByClientIDInvocations)
					u.So(t, fetchAppByGroupIDAndClientAppID, gc.ShouldEqual, tc.FetchAppByGroupIDAndClientIDInvocations)
				})
			}
		})

		t.Run("--for-source-control", func(t *testing.T) {
			t.Run("calls StitchClient.Export properly", func(t *testing.T) {
				exportCommand, _ := setup()
				exportCommand.stitchClient = &u.MockStitchClient{
					FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
						return &models.App{
							ClientAppID: clientAppID,
							GroupID:     "group-id",
							ID:          "app-id",
						}, nil
					},
					ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
						u.So(t, strategy, gc.ShouldEqual, api.ExportStrategySourceControl)
						return "", u.NewResponseBody(strings.NewReader("")), nil
					},
				}

				exportCommand.user = &user.User{APIKey: "my-api-key", AccessToken: u.GenerateValidAccessToken()}
				exportCommand.Run([]string{"--app-id=my-cool-app", "--for-source-control=true"})
			})
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
				ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
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
