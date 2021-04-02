package commands

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/utils/telemetry"

	"github.com/golang/mock/gomock"

	"github.com/10gen/realm-cli/api"
	"github.com/10gen/realm-cli/api/mdbcloud"
	mock_api "github.com/10gen/realm-cli/api/mocks"
	"github.com/10gen/realm-cli/hosting"
	"github.com/10gen/realm-cli/models"
	"github.com/10gen/realm-cli/user"
	"github.com/10gen/realm-cli/utils"
	u "github.com/10gen/realm-cli/utils/test"
	gc "github.com/smartystreets/goconvey/convey"

	"github.com/mitchellh/cli"
)

func setUpBasicCommand() (*ImportCommand, *cli.MockUi) {
	mockUI := cli.NewMockUi()
	mockService := &telemetry.Service{}
	cmd, err := NewImportCommandFactory(mockUI, mockService)()
	if err != nil {
		panic(err)
	}

	importCommand := cmd.(*ImportCommand)
	importCommand.storage = u.NewEmptyStorage()
	importCommand.writeToDirectory = func(dest string, r io.Reader, overwrite bool) error {
		return nil
	}
	importCommand.writeAppConfigToFile = func(dest string, app models.AppInstanceData) error {
		return nil
	}

	mockRealmClient := &u.MockRealmClient{
		ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
			return "", u.NewResponseBody(bytes.NewReader([]byte{})), nil
		},
		DiffFn: func(groupID, appID string, appData []byte, strategy string) ([]string, error) {
			return []string{"sample-diff-contents"}, nil
		},
		ImportFn: func(groupID, appID string, appData []byte, strategy string) error {
			return nil
		},
		FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
			return &models.App{
				GroupID: "group-id",
				ID:      "app-id",
			}, nil
		},
	}
	importCommand.realmClient = mockRealmClient
	return importCommand, mockUI
}

func TestImportNewApp(t *testing.T) {
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
			Description             string
			Args                    []string
			ExpectedExitCode        int
			RealmClient             u.MockRealmClient
			AtlasClient             u.MockMDBClient
			ProjectInput            string
			LocationInput           string
			DeploymentModelInput    string
			ExpectedOutput          string
			ExpectedLocation        string
			ExpectedDeploymentModel string
		}

		t.Run("supports creating and importing a new app", func(t *testing.T) {
			realmClient := u.MockRealmClient{
				ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
					return "", u.NewResponseBody(bytes.NewReader([]byte{})), nil
				},
				CreateEmptyAppFn: func(groupID, appName, locationName, deploymentModelName string) (*models.App, error) {
					return &models.App{Name: appName, ClientAppID: appName + "-abcdef"}, nil
				},
				FetchAppsByGroupIDFn: func(groupID string) ([]*models.App, error) {
					return []*models.App{}, nil
				},
				FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
					return nil, api.ErrAppNotFound{ClientAppID: clientAppID}
				},
			}

			importCommand, mockUI := setup()

			// Mock responses for prompts
			confirmCreateApp := "y\n"
			enterAppName := "My-Test-app\n"
			enterLocation := "US-VA\n"
			enterDeploymentModel := "GLOBAL\n"
			mockUI.InputReader = strings.NewReader(confirmCreateApp + enterAppName + enterLocation + enterDeploymentModel)
			importCommand.realmClient = &realmClient

			var writeToDirectoryCallCount int
			importCommand.writeToDirectory = func(dest string, zipData io.Reader, overwrite bool) error {
				writeToDirectoryCallCount++
				return nil
			}

			var writeAppConfigCallCount int
			importCommand.writeAppConfigToFile = func(dest string, app models.AppInstanceData) error {
				writeAppConfigCallCount++
				return nil
			}

			exitCode := importCommand.Run([]string{"--project-id=59dbcb07127ab4131c54e810", "--path=../testdata/new_app"})
			u.So(t, exitCode, gc.ShouldEqual, 0)

			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldBeEmpty)
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "New app created: My-Test-app-abcdef")
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "Successfully imported 'My-Test-app-abcdef'")

			mockRealmClient := importCommand.realmClient.(*u.MockRealmClient)
			u.So(t, mockRealmClient.ExportFnCalls, gc.ShouldHaveLength, 1)

			u.So(t, writeToDirectoryCallCount, gc.ShouldEqual, 1)
			u.So(t, writeAppConfigCallCount, gc.ShouldEqual, 1)
		})

		for _, tc := range []testCase{
			{
				Description:      "supports creating new app when providing a project name that is returned in list",
				Args:             []string{"--path=../testdata/new_app"},
				ExpectedExitCode: 0,
				RealmClient: u.MockRealmClient{
					ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
						return "", u.NewResponseBody(bytes.NewReader([]byte{})), nil
					},
					CreateEmptyAppFn: func(groupID, appName, locationName, deploymentModelName string) (*models.App, error) {
						return &models.App{Name: appName, ClientAppID: appName + "-abcdef"}, nil
					},
					FetchAppsByGroupIDFn: func(groupID string) ([]*models.App, error) {
						return []*models.App{}, nil
					},
					FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
						return nil, api.ErrAppNotFound{ClientAppID: clientAppID}
					},
				},
				AtlasClient: u.MockMDBClient{
					GroupsFn: func() ([]mdbcloud.Group, error) {
						return []mdbcloud.Group{{ID: "59dbcb07127ab4131c54e810", Name: "My-Group"}}, nil
					},
				},
				ProjectInput: "My-Group\n",
			},
			{
				Description:      "supports creating new app when providing a project id that is returned in list",
				Args:             []string{"--path=../testdata/new_app"},
				ExpectedExitCode: 0,
				RealmClient: u.MockRealmClient{
					ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
						return "", u.NewResponseBody(bytes.NewReader([]byte{})), nil
					},
					CreateEmptyAppFn: func(groupID, appName, locationName, deploymentModelName string) (*models.App, error) {
						return &models.App{Name: appName, ClientAppID: appName + "-abcdef"}, nil
					},
					FetchAppsByGroupIDFn: func(groupID string) ([]*models.App, error) {
						return []*models.App{}, nil
					},
					FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
						return nil, api.ErrAppNotFound{ClientAppID: clientAppID}
					},
				},
				AtlasClient: u.MockMDBClient{
					GroupsFn: func() ([]mdbcloud.Group, error) {
						return []mdbcloud.Group{{ID: "59dbcb07127ab4131c54e810", Name: "My-Group"}}, nil
					},
				},
				ProjectInput: "59dbcb07127ab4131c54e810\n",
			},
			{
				Description:      "supports creating new app when provided project name is not returned in list",
				Args:             []string{"--path=../testdata/new_app"},
				ExpectedExitCode: 0,
				RealmClient: u.MockRealmClient{
					ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
						return "", u.NewResponseBody(bytes.NewReader([]byte{})), nil
					},
					CreateEmptyAppFn: func(groupID, appName, locationName, deploymentModelName string) (*models.App, error) {
						return &models.App{Name: appName, ClientAppID: appName + "-abcdef"}, nil
					},
					FetchAppsByGroupIDFn: func(groupID string) ([]*models.App, error) {
						return []*models.App{}, nil
					},
					FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
						return nil, api.ErrAppNotFound{ClientAppID: clientAppID}
					},
				},
				AtlasClient: u.MockMDBClient{
					GroupsFn: func() ([]mdbcloud.Group, error) {
						return []mdbcloud.Group{{ID: "59dbcb07127ab4131c54e810", Name: "My-Group"}}, nil
					},
					GroupByNameFn: func(groupName string) (*mdbcloud.Group, error) {
						return &mdbcloud.Group{ID: "87aabc17127ab4229c54e742", Name: "Other-Group"}, nil
					},
				},
				ProjectInput: "Other-Group\n",
			},
			{
				Description:      "supports creating new app when provided project id is not returned in list",
				Args:             []string{"--path=../testdata/new_app"},
				ExpectedExitCode: 0,
				RealmClient: u.MockRealmClient{
					ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
						return "", u.NewResponseBody(bytes.NewReader([]byte{})), nil
					},
					CreateEmptyAppFn: func(groupID, appName, locationName, deploymentModelName string) (*models.App, error) {
						return &models.App{Name: appName, ClientAppID: appName + "-abcdef"}, nil
					},
					FetchAppsByGroupIDFn: func(groupID string) ([]*models.App, error) {
						return []*models.App{}, nil
					},
					FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
						return nil, api.ErrAppNotFound{ClientAppID: clientAppID}
					},
				},
				AtlasClient: u.MockMDBClient{
					GroupsFn: func() ([]mdbcloud.Group, error) {
						return []mdbcloud.Group{{ID: "59dbcb07127ab4131c54e810", Name: "My-Group"}}, nil
					},
				},
				ProjectInput: "87aabc17127ab4229c54e742\n",
			},
		} {
			t.Run(tc.Description, func(t *testing.T) {
				importCommand, mockUI := setup()

				// Mock responses for prompts
				confirmCreateApp := "y\n"
				enterAppName := "My-Test-app\n"
				enterProjectName := tc.ProjectInput
				enterLocation := "US-VA\n"
				enterDeploymentModel := "GLOBAL\n"
				mockUI.InputReader = strings.NewReader(confirmCreateApp + enterAppName + enterProjectName + enterLocation + enterDeploymentModel)
				importCommand.realmClient = &tc.RealmClient
				importCommand.atlasClient = &tc.AtlasClient

				var writeToDirectoryCallCount int
				importCommand.writeToDirectory = func(dest string, zipData io.Reader, overwrite bool) error {
					writeToDirectoryCallCount++
					return nil
				}

				var writeAppConfigCallCount int
				importCommand.writeAppConfigToFile = func(dest string, app models.AppInstanceData) error {
					writeAppConfigCallCount++
					return nil
				}

				exitCode := importCommand.Run(tc.Args)
				u.So(t, exitCode, gc.ShouldEqual, tc.ExpectedExitCode)

				u.So(t, mockUI.ErrorWriter.String(), gc.ShouldBeEmpty)
				u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "New app created: My-Test-app-abcdef")
				u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "Successfully imported 'My-Test-app-abcdef'")

				mockClient := importCommand.realmClient.(*u.MockRealmClient)
				u.So(t, mockClient.ExportFnCalls, gc.ShouldHaveLength, 1)

				u.So(t, writeToDirectoryCallCount, gc.ShouldEqual, 1)
				u.So(t, writeAppConfigCallCount, gc.ShouldEqual, 1)
			})
		}

		t.Run("returns an error when an invalid project name is entered", func(t *testing.T) {
			realmClient := u.MockRealmClient{
				ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
					return "", u.NewResponseBody(bytes.NewReader([]byte{})), nil
				},
				CreateEmptyAppFn: func(groupID, appName, locationName, deploymentModelName string) (*models.App, error) {
					return &models.App{Name: appName, ClientAppID: appName + "-abcdef"}, nil
				},
				FetchAppsByGroupIDFn: func(groupID string) ([]*models.App, error) {
					return []*models.App{}, nil
				},
				FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
					return nil, api.ErrAppNotFound{ClientAppID: clientAppID}
				},
			}
			atlasClient := u.MockMDBClient{
				GroupsFn: func() ([]mdbcloud.Group, error) {
					return []mdbcloud.Group{{ID: "59dbcb07127ab4131c54e810", Name: "My-Group"}}, nil
				},
				GroupByNameFn: func(groupName string) (*mdbcloud.Group, error) {
					return nil, fmt.Errorf("no project found with name %s", groupName)
				},
			}

			importCommand, mockUI := setup()

			// Mock responses for prompts
			confirmCreateApp := "y\n"
			enterAppName := "My-Test-app\n"
			enterProjectName := "group\n"
			mockUI.InputReader = strings.NewReader(confirmCreateApp + enterAppName + enterProjectName)
			importCommand.realmClient = &realmClient
			importCommand.atlasClient = &atlasClient

			var writeToDirectoryCallCount int
			importCommand.writeToDirectory = func(dest string, zipData io.Reader, overwrite bool) error {
				writeToDirectoryCallCount++
				return nil
			}

			var writeAppConfigCallCount int
			importCommand.writeAppConfigToFile = func(dest string, app models.AppInstanceData) error {
				writeAppConfigCallCount++
				return nil
			}
			exitCode := importCommand.Run([]string{"--path=../testdata/new_app"})
			u.So(t, exitCode, gc.ShouldEqual, 1)

			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldEqual, "no project found with name "+enterProjectName)
		})

		//include multi-region
		t.Run("supports creating app with non-default location and deployment model", func(t *testing.T) {
			realmClient := u.MockRealmClient{
				ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
					return "", u.NewResponseBody(bytes.NewReader([]byte{})), nil
				},
				CreateEmptyAppFn: func(groupID, appName, locationName, deploymentModelName string) (*models.App, error) {
					return &models.App{Name: appName, ClientAppID: appName + "-abcdef"}, nil
				},
				FetchAppsByGroupIDFn: func(groupID string) ([]*models.App, error) {
					return []*models.App{}, nil
				},
				FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
					return nil, api.ErrAppNotFound{ClientAppID: clientAppID}
				},
			}

			importCommand, mockUI := setup()

			// Mock responses for prompts
			confirmCreateApp := "y\n"
			enterAppName := "My-Test-app\n"
			enterLocation := "IE\n"
			enterDeploymentModel := "LOCAL\n"
			mockUI.InputReader = strings.NewReader(confirmCreateApp + enterAppName + enterLocation + enterDeploymentModel)

			origCreateEmptyAppFn := realmClient.CreateEmptyAppFn
			defer func() {
				realmClient.CreateEmptyAppFn = origCreateEmptyAppFn
			}()

			var createAppLocation, createAppDeploymentModelName string
			realmClient.CreateEmptyAppFn = func(groupID, appName, locationName, deploymentModelName string) (*models.App, error) {
				createAppLocation = locationName
				createAppDeploymentModelName = deploymentModelName
				return &models.App{Name: appName, ClientAppID: appName + "-abcdef"}, nil
			}

			importCommand.realmClient = &realmClient
			exitCode := importCommand.Run([]string{"--project-id=59dbcb07127ab4131c54e810", "--path=../testdata/new_app"})
			u.So(t, exitCode, gc.ShouldEqual, 0)

			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldBeEmpty)
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "New app created: My-Test-app-abcdef")
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "Successfully imported 'My-Test-app-abcdef'")

			u.So(t, createAppLocation, gc.ShouldEqual, "IE")
			u.So(t, createAppDeploymentModelName, gc.ShouldEqual, "LOCAL")

		})

		for _, tc := range []testCase{
			{
				Description:      "uses location from config file as suggested default",
				Args:             []string{"--path=../testdata/simple_app_with_deployment_config"},
				ExpectedExitCode: 0,
				RealmClient: u.MockRealmClient{
					ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
						return "", u.NewResponseBody(bytes.NewReader([]byte{})), nil
					},
					CreateEmptyAppFn: func(groupID, appName, locationName, deploymentModelName string) (*models.App, error) {
						return &models.App{Name: appName, ClientAppID: appName + "-abcdef"}, nil
					},
					FetchAppsByGroupIDFn: func(groupID string) ([]*models.App, error) {
						return []*models.App{}, nil
					},
					FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
						return nil, api.ErrAppNotFound{ClientAppID: clientAppID}
					},
				},
				LocationInput:        "US-VA\n",
				DeploymentModelInput: "GLOBAL\n",
				ExpectedOutput:       "Location [IE]",
			},
			{
				Description:      "uses deployment model from config file as suggested default",
				Args:             []string{"--path=../testdata/simple_app_with_deployment_config"},
				ExpectedExitCode: 0,
				RealmClient: u.MockRealmClient{
					ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
						return "", u.NewResponseBody(bytes.NewReader([]byte{})), nil
					},
					CreateEmptyAppFn: func(groupID, appName, locationName, deploymentModelName string) (*models.App, error) {
						return &models.App{Name: appName, ClientAppID: appName + "-abcdef"}, nil
					},
					FetchAppsByGroupIDFn: func(groupID string) ([]*models.App, error) {
						return []*models.App{}, nil
					},
					FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
						return nil, api.ErrAppNotFound{ClientAppID: clientAppID}
					},
				},
				LocationInput:        "US-VA\n",
				DeploymentModelInput: "GLOBAL\n",
				ExpectedOutput:       "Deployment Model [LOCAL]",
			},
			{
				Description:      "returns an error when an invalid location is entered",
				Args:             []string{"--path=../testdata/new_app"},
				ExpectedExitCode: 1,
				RealmClient: u.MockRealmClient{
					ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
						return "", u.NewResponseBody(bytes.NewReader([]byte{})), nil
					},
					CreateEmptyAppFn: func(groupID, appName, locationName, deploymentModelName string) (*models.App, error) {
						return &models.App{Name: appName, ClientAppID: appName + "-abcdef"}, nil
					},
					FetchAppsByGroupIDFn: func(groupID string) ([]*models.App, error) {
						return []*models.App{}, nil
					},
					FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
						return nil, api.ErrAppNotFound{ClientAppID: clientAppID}
					},
				},
				LocationInput:        "test\n",
				DeploymentModelInput: "GLOBAL\n",
				ExpectedOutput:       "Could not understand response, valid values are US-VA, US-OR, IE, AU:",
			},
			{
				Description:      "returns an error when an invalid deployment model is entered",
				Args:             []string{"--path=../testdata/new_app"},
				ExpectedExitCode: 0,
				RealmClient: u.MockRealmClient{
					ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
						return "", u.NewResponseBody(bytes.NewReader([]byte{})), nil
					},
					CreateEmptyAppFn: func(groupID, appName, locationName, deploymentModelName string) (*models.App, error) {
						return &models.App{Name: appName, ClientAppID: appName + "-abcdef"}, nil
					},
					FetchAppsByGroupIDFn: func(groupID string) ([]*models.App, error) {
						return []*models.App{}, nil
					},
					FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
						return nil, api.ErrAppNotFound{ClientAppID: clientAppID}
					},
				},
				LocationInput:        "US-VA\n",
				DeploymentModelInput: "test\n",
				ExpectedOutput:       "Could not understand response, valid values are GLOBAL, LOCAL",
			},
		} {
			t.Run(tc.Description, func(t *testing.T) {
				importCommand, mockUI := setup()

				// Mock responses for prompts
				confirmCreateApp := "y\n"
				enterAppName := "My-Test-app\n"
				enterLocation := tc.LocationInput
				enterDeploymentModel := tc.DeploymentModelInput
				mockUI.InputReader = strings.NewReader(confirmCreateApp + enterAppName + enterLocation + enterDeploymentModel)
				importCommand.realmClient = &tc.RealmClient

				exitCode := importCommand.Run(append([]string{"--project-id=59dbcb07127ab4131c54e810"}, tc.Args...))
				u.So(t, exitCode, gc.ShouldEqual, tc.ExpectedExitCode)

				u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, tc.ExpectedOutput)
			})
		}
	})
}

func TestImportCommand(t *testing.T) {
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
			RealmClient      u.MockRealmClient
		}

		t.Run("it does not import if the user does not confirm the diff", func(t *testing.T) {
			realmClient := u.MockRealmClient{
				ImportFn: func(groupID, appID string, appData []byte, strategy string) error {
					return nil
				},
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

			importCommand, mockUI := setup()

			// Mock a "no" response when we prompt the user to confirm the diff
			mockUI.InputReader = strings.NewReader("n\n")
			importCommand.realmClient = &realmClient

			exitCode := importCommand.Run(append([]string{"--path=../testdata/full_app"}, validArgs...))

			mockClient := importCommand.realmClient.(*u.MockRealmClient)

			u.So(t, exitCode, gc.ShouldEqual, 0)
			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldBeEmpty)
			u.So(t, len(mockClient.ImportFnCalls), gc.ShouldEqual, 0)
		})

		t.Run("it asks the user to discard existing drafts", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			realmClient := mock_api.NewMockRealmClient(ctrl)
			defer ctrl.Finish()

			realmClient.EXPECT().FetchAppByClientAppID("my-app-abcdef").Return(&models.App{GroupID: "group-id", ID: "app-id"}, nil)
			realmClient.EXPECT().Diff("group-id", "app-id", gomock.Any(), gomock.Any()).Return([]string{"changes"}, nil)
			realmClient.EXPECT().CreateDraft("group-id", "app-id").Return(nil, api.UnmarshalRealmError(&http.Response{
				Body: u.NewResponseBody(strings.NewReader(`{ "error_code": "DraftAlreadyExists" }`)),
			}))
			realmClient.EXPECT().GetDrafts("group-id", "app-id").Return([]models.AppDraft{
				{ID: "draft-id"},
			}, nil)
			realmClient.EXPECT().DraftDiff("group-id", "app-id", "draft-id").Return(&models.DraftDiff{
				Diffs: []string{"just", "some", "diffs"},
			}, nil)

			importCommand, mockUI := setup()
			mockUI.InputReader = strings.NewReader("y\n")
			importCommand.realmClient = realmClient
			importCommand.Run(append([]string{"--path=../testdata/full_app"}, validArgs...))

			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "Would you like to discard these changes?")
		})

		t.Run("it cancels the import if the user doesn't discard draft", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			realmClient := mock_api.NewMockRealmClient(ctrl)
			defer ctrl.Finish()

			realmClient.EXPECT().FetchAppByClientAppID("my-app-abcdef").Return(&models.App{GroupID: "group-id", ID: "app-id"}, nil)
			realmClient.EXPECT().Diff("group-id", "app-id", gomock.Any(), gomock.Any()).Return([]string{"changes"}, nil)
			realmClient.EXPECT().CreateDraft("group-id", "app-id").Return(nil, api.UnmarshalRealmError(&http.Response{
				Body: u.NewResponseBody(strings.NewReader(`{ "error_code": "DraftAlreadyExists" }`)),
			}))
			realmClient.EXPECT().GetDrafts("group-id", "app-id").Return([]models.AppDraft{
				{ID: "draft-id"},
			}, nil)
			realmClient.EXPECT().DraftDiff("group-id", "app-id", "draft-id").Return(&models.DraftDiff{
				Diffs: []string{"just", "some", "diffs"},
			}, nil)

			importCommand, mockUI := setup()
			mockUI.InputReader = strings.NewReader("y\nn\n")
			importCommand.realmClient = realmClient
			importCommand.Run(append([]string{"--path=../testdata/full_app"}, validArgs...))

			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "Would you like to discard these changes?")
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "Cancelling import.")
		})

		t.Run("it discards the draft and deploys the new draft", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			realmClient := mock_api.NewMockRealmClient(ctrl)
			defer ctrl.Finish()

			realmClient.EXPECT().FetchAppByClientAppID("my-app-abcdef").Return(&models.App{GroupID: "group-id", ID: "app-id"}, nil)
			realmClient.EXPECT().Diff("group-id", "app-id", gomock.Any(), gomock.Any()).Return([]string{"changes"}, nil)
			realmClient.EXPECT().CreateDraft("group-id", "app-id").Return(nil, api.UnmarshalRealmError(&http.Response{
				Body: u.NewResponseBody(strings.NewReader(`{ "error_code": "DraftAlreadyExists" }`)),
			}))
			realmClient.EXPECT().GetDrafts("group-id", "app-id").Return([]models.AppDraft{
				{ID: "draft-id"},
			}, nil)
			realmClient.EXPECT().DraftDiff("group-id", "app-id", "draft-id").Return(&models.DraftDiff{
				Diffs: []string{"just", "some", "diffs"},
			}, nil)
			realmClient.EXPECT().DiscardDraft("group-id", "app-id", "draft-id").Return(nil)
			realmClient.EXPECT().CreateDraft("group-id", "app-id").Return(&models.AppDraft{ID: "draft-id-2"}, nil)
			realmClient.EXPECT().Import("group-id", "app-id", gomock.Any(), gomock.Any()).Return(nil)
			realmClient.EXPECT().DeployDraft("group-id", "app-id", "draft-id-2").Return(&models.Deployment{
				Status: models.DeploymentStatusSuccessful,
			}, nil)
			realmClient.EXPECT().Export("group-id", "app-id", api.ExportStrategyNone).Return("", u.NewResponseBody(bytes.NewReader([]byte{})), nil)

			importCommand, mockUI := setup()
			mockUI.InputReader = strings.NewReader("y\ny\n")
			importCommand.realmClient = realmClient
			importCommand.Run(append([]string{"--path=../testdata/full_app"}, validArgs...))

			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "Would you like to discard these changes?")
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "Discarding existing draft...")
		})

		t.Run("it asks the user to discard empty drafts", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			realmClient := mock_api.NewMockRealmClient(ctrl)
			defer ctrl.Finish()

			realmClient.EXPECT().FetchAppByClientAppID("my-app-abcdef").Return(&models.App{GroupID: "group-id", ID: "app-id"}, nil)
			realmClient.EXPECT().Diff("group-id", "app-id", gomock.Any(), gomock.Any()).Return([]string{"changes"}, nil)
			realmClient.EXPECT().CreateDraft("group-id", "app-id").Return(nil, api.UnmarshalRealmError(&http.Response{
				Body: u.NewResponseBody(strings.NewReader(`{ "error_code": "DraftAlreadyExists" }`)),
			}))
			realmClient.EXPECT().GetDrafts("group-id", "app-id").Return([]models.AppDraft{
				{ID: "draft-id"},
			}, nil)
			realmClient.EXPECT().DraftDiff("group-id", "app-id", "draft-id").Return(&models.DraftDiff{}, nil) // empty diff
			realmClient.EXPECT().DiscardDraft("group-id", "app-id", "draft-id").Return(nil)
			realmClient.EXPECT().CreateDraft("group-id", "app-id").Return(&models.AppDraft{ID: "draft-id-2"}, nil)
			realmClient.EXPECT().Import("group-id", "app-id", gomock.Any(), gomock.Any()).Return(nil)
			realmClient.EXPECT().DeployDraft("group-id", "app-id", "draft-id-2").Return(&models.Deployment{
				Status: models.DeploymentStatusSuccessful,
			}, nil)
			realmClient.EXPECT().Export("group-id", "app-id", api.ExportStrategyNone).Return("", u.NewResponseBody(bytes.NewReader([]byte{})), nil)

			importCommand, mockUI := setup()
			mockUI.InputReader = strings.NewReader("y\ny\n")
			importCommand.realmClient = realmClient
			importCommand.Run(append([]string{"--path=../testdata/full_app"}, validArgs...))

			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "An empty draft already exists for your app, would you like to discard it first?")
		})

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
					ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
						return "", u.NewResponseBody(bytes.NewReader([]byte{})), nil
					},
					ImportFn: func(groupID, appID string, appData []byte, strategy string) error {
						return nil
					},
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
			{
				Description:      "reports an error if it fails to fetch the app by clientID",
				Args:             append([]string{"--path=../testdata/full_app"}, validArgs...),
				ExpectedExitCode: 1,
				ExpectedError:    "oh no failed to fetch app",
				RealmClient: u.MockRealmClient{
					ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
						return "", nil, fmt.Errorf("oh no")
					},
					ImportFn: func(groupID, appID string, appData []byte, strategy string) error {
						return nil
					},
					DiffFn: func(groupID, appID string, appData []byte, strategy string) ([]string, error) {
						return []string{"sample-diff-contents"}, nil
					},
					FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
						return nil, fmt.Errorf("oh no failed to fetch app")
					},
				},
			},
			{
				Description:      "successfully imports even if the config.json file is empty",
				Args:             append([]string{"--path=../testdata/simple_app_empty_config_json"}, validArgs...),
				ExpectedExitCode: 0,
				RealmClient: u.MockRealmClient{
					ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
						return "", u.NewResponseBody(strings.NewReader("export response")), nil
					},
					ImportFn: func(groupID, appID string, appData []byte, strategy string) error {
						return nil
					},
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
			{
				Description:      "it fails with an error if the response to the import request is invalid",
				Args:             append([]string{"--path=../testdata/simple_app"}, validArgs...),
				ExpectedExitCode: 1,
				ExpectedError:    "oh noes",
				RealmClient: u.MockRealmClient{
					ImportFn: func(groupID, appID string, appData []byte, strategy string) error {
						return fmt.Errorf("oh noes")
					},
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
			{
				Description:      "it succeeds if it can grab instance data from the app config file at the provided path",
				Args:             []string{"--path=../testdata/simple_app_with_instance_data"},
				ExpectedExitCode: 0,
				RealmClient: u.MockRealmClient{
					ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
						return "", u.NewResponseBody(strings.NewReader("export response")), nil
					},
					ImportFn: func(groupID, appID string, appData []byte, strategy string) error {
						return nil
					},
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
			{
				Description:      "it succeeds if using a specific project-id",
				Args:             []string{"--path=../testdata/simple_app_with_instance_data", "--project-id=myprojectid"},
				ExpectedExitCode: 0,
				RealmClient: u.MockRealmClient{
					ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
						return "", u.NewResponseBody(strings.NewReader("export response")), nil
					},
					ImportFn: func(groupID, appID string, appData []byte, strategy string) error {
						return nil
					},
					DiffFn: func(groupID, appID string, appData []byte, strategy string) ([]string, error) {
						return []string{"sample-diff-contents"}, nil
					},
					FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
						return nil, errors.New("this should not be called")
					},
					FetchAppByGroupIDAndClientAppIDFn: func(groupID, clientAppID string) (*models.App, error) {
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
				RealmClient: u.MockRealmClient{
					ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
						return "", u.NewResponseBody(strings.NewReader("export response")), nil
					},
					ImportFn: func(groupID, appID string, appData []byte, strategy string) error {
						return nil
					},
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
			{
				Description:      "reports an error if it fails to export the app",
				Args:             append([]string{"--path=../testdata/full_app"}, validArgs...),
				ExpectedExitCode: 1,
				ExpectedError:    "failed to sync app",
				RealmClient: u.MockRealmClient{
					ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
						return "", nil, fmt.Errorf("oh no")
					},
					ImportFn: func(groupID, appID string, appData []byte, strategy string) error {
						return nil
					},
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
				importCommand, mockUI := setup()

				// Mock a "yes" response when we prompt the user to confirm the diff
				mockUI.InputReader = strings.NewReader("y\n")
				importCommand.realmClient = &tc.RealmClient
				importCommand.workingDirectory = tc.WorkingDirectory

				exitCode := importCommand.Run(tc.Args)
				u.So(t, exitCode, gc.ShouldEqual, tc.ExpectedExitCode)
				u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, tc.ExpectedError)
			})
		}

		//include hosting
		for _, tc := range []testCase{
			{
				Description:      "it succeeds if given a valid flagAppPath and flagIncludeHosting",
				Args:             append([]string{"--path=../testdata/full_app", "--include-hosting"}, validArgs...),
				ExpectedExitCode: 0,
				RealmClient: u.MockRealmClient{
					ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
						return "", u.NewResponseBody(bytes.NewReader([]byte{})), nil
					},
					ImportFn: func(groupID, appID string, appData []byte, strategy string) error {
						return nil
					},
					DiffFn: func(groupID, appID string, appData []byte, strategy string) ([]string, error) {
						return []string{"sample-diff-contents"}, nil
					},
					UploadAssetFn: func(groupID, appID, path, hash string, size int64, body io.Reader, attributes ...hosting.AssetAttribute) error {
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
				Description:      "it succeeds if given a valid flagAppPath, flagIncludeHosting, and flagResetCDNCache",
				Args:             append([]string{"--path=../testdata/full_app", "--include-hosting", "--reset-cdn-cache"}, validArgs...),
				ExpectedExitCode: 0,
				RealmClient: u.MockRealmClient{
					ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
						return "", u.NewResponseBody(bytes.NewReader([]byte{})), nil
					},
					ImportFn: func(groupID, appID string, appData []byte, strategy string) error {
						return nil
					},
					DiffFn: func(groupID, appID string, appData []byte, strategy string) ([]string, error) {
						return []string{"sample-diff-contents"}, nil
					},
					UploadAssetFn: func(groupID, appID, path, hash string, size int64, body io.Reader, attributes ...hosting.AssetAttribute) error {
						return nil
					},
					FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
						return &models.App{
							GroupID: "group-id",
							ID:      "app-id",
						}, nil
					},
					InvalidateCacheFn: func(groupID, appID, path string) error {
						return nil
					},
				},
			},
		} {
			t.Run(tc.Description, func(t *testing.T) {
				importCommand, mockUI := setup()

				// Mock a "yes" response when we prompt the user to confirm the diff
				mockUI.InputReader = strings.NewReader("y\n")
				importCommand.realmClient = &tc.RealmClient
				importCommand.workingDirectory = tc.WorkingDirectory

				exitCode := importCommand.Run(append(tc.Args, "--config-path=../testdata/configs/tmp/config.json"))
				u.So(t, exitCode, gc.ShouldEqual, tc.ExpectedExitCode)
				u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, tc.ExpectedError)

				cachePath := filepath.Join(filepath.Dir(importCommand.flagConfigPath), utils.HostingCacheFileName)

				assetCache, cErr := hosting.CacheFileToAssetCache(cachePath)
				u.So(t, cErr, gc.ShouldBeNil)

				rErr := os.Remove(cachePath)
				u.So(t, rErr, gc.ShouldBeNil)

				_, ok := assetCache.Get(importCommand.flagAppID, "/asset_file0.json")
				u.So(t, ok, gc.ShouldBeTrue)

				_, ok = assetCache.Get(importCommand.flagAppID, "/ships/nostromo.json")
				u.So(t, ok, gc.ShouldBeTrue)
			})
		}

		// include dependencies
		for _, tc := range []testCase{
			{
				Description:      "it succeeds if given a valid flagAppPath and flagIncludeDependencies",
				Args:             append([]string{"--path=../testdata/full_app", "--include-dependencies"}, validArgs...),
				ExpectedExitCode: 0,
				RealmClient: u.MockRealmClient{
					ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
						return "", u.NewResponseBody(bytes.NewReader([]byte{})), nil
					},
					ImportFn: func(groupID, appID string, appData []byte, strategy string) error {
						return nil
					},
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
				importCommand, mockUI := setup()

				// Mock a "yes" response when we prompt the user to confirm the diff
				mockUI.InputReader = strings.NewReader("y\n")
				importCommand.realmClient = &tc.RealmClient
				importCommand.workingDirectory = tc.WorkingDirectory

				exitCode := importCommand.Run(append(tc.Args, "--config-path=../testdata/configs/tmp/config.json"))
				u.So(t, exitCode, gc.ShouldEqual, tc.ExpectedExitCode)
				u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, tc.ExpectedError)
			})
		}

		t.Run("syncing data after a successful import", func(t *testing.T) {
			t.Run("on success", func(t *testing.T) {
				type testCase struct {
					description            string
					args                   []string
					workingDirectory       string
					expectedDirectory      string
					expectedExportStrategy api.ExportStrategy
				}

				for _, tc := range []testCase{
					{
						description:            "it writes data to the provided directory",
						args:                   append([]string{"--path=../testdata/simple_app"}, validArgs...),
						workingDirectory:       "",
						expectedDirectory:      abs("../testdata/simple_app"),
						expectedExportStrategy: api.ExportStrategyNone,
					},
					{
						description:            "it writes data to the working directory when using app config file data",
						args:                   []string{},
						workingDirectory:       "../testdata/simple_app_with_instance_data",
						expectedDirectory:      abs("../testdata/simple_app_with_instance_data"),
						expectedExportStrategy: api.ExportStrategyNone,
					},
					{
						description:            "it uses the 'source_control' export strategy when importing using the 'replace-by-name' import strategy",
						args:                   append([]string{"--path=../testdata/simple_app", "--strategy=replace-by-name"}, validArgs...),
						workingDirectory:       "",
						expectedDirectory:      abs("../testdata/simple_app"),
						expectedExportStrategy: api.ExportStrategySourceControl,
					},
				} {
					t.Run(tc.description, func(t *testing.T) {
						importCommand, mockUI := setup()
						mockUI.InputReader = strings.NewReader("y\n")
						importCommand.workingDirectory = tc.workingDirectory
						var exportStrategy api.ExportStrategy
						mockRealmClient := &u.MockRealmClient{
							ExportFn: func(groupID, appID string, strategy api.ExportStrategy) (string, io.ReadCloser, error) {
								exportStrategy = strategy
								return "", u.NewResponseBody(strings.NewReader("export response")), nil
							},
							ImportFn: func(groupID, appID string, appData []byte, strategy string) error {
								return nil
							},
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
						importCommand.realmClient = mockRealmClient

						destinationDirectory := ""
						writeContent := ""

						importCommand.writeToDirectory = func(dest string, zipData io.Reader, overwrite bool) error {
							b, err := ioutil.ReadAll(zipData)
							if err != nil {
								return err
							}
							destinationDirectory = dest
							writeContent = string(b)
							return nil
						}

						exitCode := importCommand.Run(tc.args)
						u.So(t, exitCode, gc.ShouldEqual, 0)
						u.So(t, mockUI.ErrorWriter.String(), gc.ShouldBeEmpty)
						u.So(t, abs(destinationDirectory), gc.ShouldEqual, tc.expectedDirectory)
						u.So(t, exportStrategy, gc.ShouldEqual, tc.expectedExportStrategy)
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
