package push

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/utils/api"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/Netflix/go-expect"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPushHandler(t *testing.T) {
	wd, wdErr := os.Getwd()
	assert.Nil(t, wdErr)

	testApp := local.App{
		RootDir: filepath.Join(wd, "testdata/project"),
		Config:  local.FileConfig,
		AppData: &local.AppConfigJSON{local.AppDataV1{local.AppStructureV1{
			ConfigVersion:        realm.AppConfigVersion20200603,
			ID:                   "eggcorn-abcde",
			Name:                 "eggcorn",
			Location:             realm.LocationVirginia,
			DeploymentModel:      realm.DeploymentModelGlobal,
			Security:             map[string]interface{}{},
			CustomUserDataConfig: map[string]interface{}{"enabled": true},
			Sync:                 map[string]interface{}{"development_mode_enabled": false},
		}}},
	}

	t.Run("should return an error if the command fails to resolve to", func(t *testing.T) {
		var realmClient mock.RealmClient

		var capturedFilter realm.AppFilter
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			capturedFilter = filter
			return nil, errors.New("something bad happened")
		}

		cmd := &Command{inputs{LocalPath: "testdata/project", Project: "groupID", RemoteApp: "appID"}}

		err := cmd.Handler(nil, nil, cli.Clients{Realm: realmClient})
		assert.Equal(t, errors.New("something bad happened"), err)

		t.Log("and should properly pass through the expected inputs")
		assert.Equal(t, realm.AppFilter{"groupID", "appID", nil}, capturedFilter)
	})

	t.Run("should return an error if the command fails to resolve group id", func(t *testing.T) {
		var atlasClient mock.AtlasClient
		atlasClient.GroupsFn = func(url string, useBaseURL bool) (atlas.Groups, error) {
			return atlas.Groups{}, errors.New("something bad happened")
		}

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return nil, nil
		}

		cmd := &Command{inputs{LocalPath: "testdata/project"}}

		err := cmd.Handler(nil, nil, cli.Clients{
			Realm: realmClient,
			Atlas: atlasClient,
		})
		assert.Equal(t, errors.New("something bad happened"), err)
	})

	t.Run("should return an error if the command fails to create a new app", func(t *testing.T) {
		out := new(bytes.Buffer)
		ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{GroupID: "groupID"}}, nil
		}

		var capturedGroupID, capturedName string
		var capturedMeta realm.AppMeta
		var i int
		realmClient.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			i++
			capturedGroupID = groupID
			capturedName = name
			capturedMeta = meta
			return realm.App{}, errors.New("something bad happened")
		}

		cmd := &Command{inputs{LocalPath: "testdata/project", RemoteApp: "appID"}}

		assert.Equal(t, errors.New("something bad happened"), cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))

		t.Log("and should properly pass through the expected inputs")
		assert.Equal(t, "groupID", capturedGroupID)
		assert.Equal(t, "eggcorn", capturedName)
		assert.Equal(t, realm.AppMeta{
			Location:        realm.LocationVirginia,
			DeploymentModel: realm.DeploymentModelGlobal,
			Environment:     realm.EnvironmentNone,
		}, capturedMeta)
	})

	t.Run("should return an error if the command fails to get the initial diff", func(t *testing.T) {
		_, ui := mock.NewUI()

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", GroupID: "groupID"}}, nil
		}

		var capturedAppData interface{}
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			capturedAppData = appData
			return nil, errors.New("something bad happened")
		}

		cmd := &Command{inputs{LocalPath: "testdata/project", RemoteApp: "appID"}}

		err := cmd.Handler(nil, ui, cli.Clients{Realm: realmClient})
		assert.Equal(t, errors.New("something bad happened"), err)

		t.Log("and should properly pass through the expected inputs")
		assert.Equal(t, testApp.AppData, capturedAppData)
	})

	t.Run("should return an error if the command fails to create a new draft", func(t *testing.T) {
		out := new(bytes.Buffer)
		ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", GroupID: "groupID"}}, nil
		}
		realmClient.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return realm.App{ID: "appID", GroupID: "groupID"}, nil
		}
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			return []string{"diff1"}, nil
		}

		var capturedGroupID, capturedAppID string
		realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) {
			capturedGroupID = groupID
			capturedAppID = appID
			return realm.AppDraft{}, errors.New("something bad happened")
		}

		cmd := &Command{inputs{LocalPath: "testdata/project", RemoteApp: "appID"}}

		err := cmd.Handler(nil, ui, cli.Clients{Realm: realmClient})
		assert.Equal(t, errors.New("something bad happened"), err)

		t.Log("and should properly pass through the expected inputs")
		assert.Equal(t, "groupID", capturedGroupID)
		assert.Equal(t, "appID", capturedAppID)
	})

	t.Run("should return an error if the command fails to import", func(t *testing.T) {
		out := new(bytes.Buffer)
		ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", GroupID: "groupID"}}, nil
		}
		realmClient.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return realm.App{ID: "appID", GroupID: "groupID"}, nil
		}
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			return []string{"diff1"}, nil
		}
		realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) {
			return realm.AppDraft{ID: "draftID"}, nil
		}

		var capturedGroupID, capturedAppID string
		var capturedAppData interface{}
		realmClient.ImportFn = func(groupID, appID string, appData interface{}) error {
			capturedGroupID = groupID
			capturedAppID = appID
			capturedAppData = appData
			return errors.New("something bad happened")
		}

		var discardDraftCalled bool
		realmClient.DiscardDraftFn = func(groupID, appID, draftID string) error {
			discardDraftCalled = true
			return nil
		}

		cmd := &Command{inputs{LocalPath: "testdata/project", RemoteApp: "appID"}}

		err := cmd.Handler(nil, ui, cli.Clients{Realm: realmClient})
		assert.Equal(t, errors.New("something bad happened"), err)

		t.Log("and should properly pass through the expected inputs")
		assert.Equal(t, "groupID", capturedGroupID)
		assert.Equal(t, "appID", capturedAppID)
		assert.Equal(t, testApp.AppData, capturedAppData)

		t.Log("and should attempt to discard the created draft")
		assert.True(t, discardDraftCalled, "expected (realm.Client).DiscardDraft to be called")
	})

	t.Run("should return an error if the command fails to import and report when a draft is not discarded", func(t *testing.T) {
		out := new(bytes.Buffer)
		ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", GroupID: "groupID"}}, nil
		}
		realmClient.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return realm.App{ID: "appID", GroupID: "groupID"}, nil
		}
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			return []string{"diff1"}, nil
		}
		realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) {
			return realm.AppDraft{ID: "draftID"}, nil
		}
		realmClient.ImportFn = func(groupID, appID string, appData interface{}) error {
			return errors.New("something bad happened")
		}
		realmClient.DiscardDraftFn = func(groupID, appID, draftID string) error {
			return errors.New("something worse happened")
		}

		cmd := &Command{inputs{LocalPath: "testdata/project", RemoteApp: "appID"}}

		err := cmd.Handler(nil, ui, cli.Clients{Realm: realmClient})
		assert.Equal(t, errors.New("something bad happened"), err)

		assert.Equal(t, `Determining changes
Creating draft
Pushing changes
Failed to discard the draft created for your deployment
`, out.String())
	})

	t.Run("should return an error if the command fails to deploy the draft", func(t *testing.T) {
		out := new(bytes.Buffer)
		ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", GroupID: "groupID"}}, nil
		}
		realmClient.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return realm.App{ID: "appID", GroupID: "groupID"}, nil
		}
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			return []string{"diff1"}, nil
		}
		realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) {
			return realm.AppDraft{ID: "draftID"}, nil
		}
		realmClient.ImportFn = func(groupID, appID string, appData interface{}) error {
			return nil
		}

		var capturedGroupID, capturedAppID, capturedDraftID string
		realmClient.DeployDraftFn = func(groupID, appID, draftID string) (realm.AppDeployment, error) {
			capturedGroupID = groupID
			capturedAppID = appID
			capturedDraftID = draftID
			return realm.AppDeployment{}, errors.New("something bad happened")
		}

		var discardDraftCalled bool
		realmClient.DiscardDraftFn = func(groupID, appID, draftID string) error {
			discardDraftCalled = true
			return nil
		}

		cmd := &Command{inputs{LocalPath: "testdata/project", RemoteApp: "appID"}}

		err := cmd.Handler(nil, ui, cli.Clients{Realm: realmClient})
		assert.Equal(t, errors.New("something bad happened"), err)

		t.Log("and should properly pass through the expected inputs")
		assert.Equal(t, "groupID", capturedGroupID)
		assert.Equal(t, "appID", capturedAppID)
		assert.Equal(t, "draftID", capturedDraftID)

		t.Log("and should attempt to discard the created draft")
		assert.True(t, discardDraftCalled, "expected (realm.Client).DiscardDraft to be called")
	})

	t.Run("should return an error if the command fails to deploy the draft and report when a draft is not discarded", func(t *testing.T) {
		out := new(bytes.Buffer)
		ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", GroupID: "groupID"}}, nil
		}
		realmClient.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return realm.App{ID: "appID", GroupID: "groupID"}, nil
		}
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			return []string{"diff1"}, nil
		}
		realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) {
			return realm.AppDraft{ID: "draftID"}, nil
		}
		realmClient.ImportFn = func(groupID, appID string, appData interface{}) error {
			return nil
		}
		realmClient.DeployDraftFn = func(groupID, appID, draftID string) (realm.AppDeployment, error) {
			return realm.AppDeployment{}, errors.New("something bad happened")
		}
		realmClient.DiscardDraftFn = func(groupID, appID, draftID string) error {
			return errors.New("something worse happened")
		}

		cmd := &Command{inputs{LocalPath: "testdata/project", RemoteApp: "appID"}}

		err := cmd.Handler(nil, ui, cli.Clients{Realm: realmClient})
		assert.Equal(t, errors.New("something bad happened"), err)

		assert.Equal(t, `Determining changes
Creating draft
Pushing changes
Deploying draft
Failed to discard the draft created for your deployment
`, out.String())
	})

	t.Run("should return an error if the command deploys a draft which fails", func(t *testing.T) {
		out := new(bytes.Buffer)
		ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", GroupID: "groupID"}}, nil
		}
		realmClient.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return realm.App{ID: "appID", GroupID: "groupID"}, nil
		}
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			return []string{"diff1"}, nil
		}
		realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) {
			return realm.AppDraft{ID: "draftID"}, nil
		}
		realmClient.ImportFn = func(groupID, appID string, appData interface{}) error {
			return nil
		}
		realmClient.DeployDraftFn = func(groupID, appID, draftID string) (realm.AppDeployment, error) {
			return realm.AppDeployment{Status: realm.DeploymentStatusFailed, StatusErrorMessage: "something bad happened"}, nil
		}

		var discardDraftCalled bool
		realmClient.DiscardDraftFn = func(groupID, appID, draftID string) error {
			discardDraftCalled = true
			return nil
		}

		cmd := &Command{inputs{LocalPath: "testdata/project", RemoteApp: "appID"}}

		err := cmd.Handler(nil, ui, cli.Clients{Realm: realmClient})
		assert.Equal(t, errors.New("failed to deploy app: something bad happened"), err)

		t.Log("and should attempt to discard the created draft")
		assert.True(t, discardDraftCalled, "expected (realm.Client).DiscardDraft to be called")
	})

	t.Run("should return an error if the command deploys a draft which fails and report when a draft is not discarded", func(t *testing.T) {
		out := new(bytes.Buffer)
		ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", GroupID: "groupID"}}, nil
		}
		realmClient.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return realm.App{ID: "appID", GroupID: "groupID"}, nil
		}
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			return []string{"diff1"}, nil
		}
		realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) {
			return realm.AppDraft{ID: "draftID"}, nil
		}
		realmClient.ImportFn = func(groupID, appID string, appData interface{}) error {
			return nil
		}
		realmClient.DeployDraftFn = func(groupID, appID, draftID string) (realm.AppDeployment, error) {
			return realm.AppDeployment{Status: realm.DeploymentStatusFailed, StatusErrorMessage: "something bad happened"}, nil
		}
		realmClient.DiscardDraftFn = func(groupID, appID, draftID string) error {
			return errors.New("something worse happened")
		}

		cmd := &Command{inputs{LocalPath: "testdata/project", RemoteApp: "appID"}}

		err := cmd.Handler(nil, ui, cli.Clients{Realm: realmClient})
		assert.Equal(t, errors.New("failed to deploy app: something bad happened"), err)

		assert.Equal(t, `Determining changes
Creating draft
Pushing changes
Deploying draft
Deployment failed
Failed to discard the draft created for your deployment
`, out.String())
	})

	t.Run("with a realm client that successfully imports and deploys drafts", func(t *testing.T) {
		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", GroupID: "groupID", ClientAppID: "eggcorn-abcde"}}, nil
		}
		realmClient.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return realm.App{ID: "appID", GroupID: "groupID", ClientAppID: "eggcorn-abcde"}, nil
		}
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			return []string{"diff1"}, nil
		}
		realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) {
			return realm.AppDraft{ID: "draftID"}, nil
		}
		realmClient.ImportFn = func(groupID, appID string, appData interface{}) error {
			return nil
		}
		realmClient.DeployDraftFn = func(groupID, appID, draftID string) (realm.AppDeployment, error) {
			return realm.AppDeployment{Status: realm.DeploymentStatusSuccessful}, nil
		}

		t.Run("should run import successfully", func(t *testing.T) {
			runImport(t, realmClient, "testdata/project")
		})

		t.Run("should run import successfully when modifying a different remote app", func(t *testing.T) {
			runImport(t, realmClient, "testdata/project-alt") // specifies a different app id
		})

		t.Run("but fails to upload a hosting asset", func(t *testing.T) {
			profile, teardown := mock.NewProfileFromTmpDir(t, "push-handler")
			defer teardown()

			realmClient.HostingAssetUploadFn = func(groupID, appID, rootDir string, asset realm.HostingAsset) error {
				return errors.New("something bad happened")
			}

			t.Run("should not return an error when include hosting flag is omitted", func(t *testing.T) {
				runImport(t, realmClient, "testdata/project")
			})

			t.Run("should return an error when adding a hosting asset", func(t *testing.T) {
				out := new(bytes.Buffer)
				ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

				realmClient.HostingAssetsFn = func(groupID, appID string) ([]realm.HostingAsset, error) {
					return nil, nil
				}

				cmd := &Command{inputs{LocalPath: "testdata/hosting", RemoteApp: "appID", IncludeHosting: true}}

				err := cmd.Handler(profile, ui, cli.Clients{Realm: realmClient})
				assert.Equal(t, errors.New("2 error(s) occurred while importing hosting assets"), err)
				output := out.String()
				for _, line := range []string{
					`Determining changes
Creating draft
Pushing changes
Deploying draft
Deployment complete`,
					"An error occurred while uploading hosting assets: failed to add /404.html: something bad happened",
					"An error occurred while uploading hosting assets: failed to add /index.html: something bad happened",
				} {
					assert.True(t, strings.Contains(output, line), fmt.Sprintf("expected output to contain: '%s'\nactual output:\n%s", line, output))
				}
			})

			t.Run("should return an error when modifying the body of a hosting asset", func(t *testing.T) {
				out := new(bytes.Buffer)
				ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

				realmClient.HostingAssetsFn = func(groupID, appID string) ([]realm.HostingAsset, error) {
					return []realm.HostingAsset{
						{HostingAssetData: realm.HostingAssetData{FilePath: "/index.html"}},
						{HostingAssetData: realm.HostingAssetData{FilePath: "/404.html"}},
					}, nil
				}

				cmd := &Command{inputs{LocalPath: "testdata/hosting", RemoteApp: "appID", IncludeHosting: true}}

				err := cmd.Handler(profile, ui, cli.Clients{Realm: realmClient})
				assert.Equal(t, errors.New("2 error(s) occurred while importing hosting assets"), err)
				output := out.String()
				for _, line := range []string{
					`Determining changes
Creating draft
Pushing changes
Deploying draft
Deployment complete`,
					"An error occurred while uploading hosting assets: failed to update /404.html: something bad happened",
					"An error occurred while uploading hosting assets: failed to update /index.html: something bad happened",
				} {
					assert.True(t, strings.Contains(output, line), fmt.Sprintf("expected output to contain: '%s'\nactual output:\n%s", line, output))
				}
			})
		})

		t.Run("but fails to remove a hosting asset", func(t *testing.T) {
			profile, teardown := mock.NewProfileFromTmpDir(t, "push-handler")
			defer teardown()

			realmClient.HostingAssetUploadFn = func(groupID, appID, rootDir string, asset realm.HostingAsset) error {
				return nil
			}

			realmClient.HostingAssetRemoveFn = func(groupID, appID, path string) error {
				return errors.New("something bad happened")
			}

			t.Run("should not return an error when include hosting flag is omitted", func(t *testing.T) {
				runImport(t, realmClient, "testdata/hosting")
			})

			t.Run("should return an error when deleting a hosting asset", func(t *testing.T) {
				out := new(bytes.Buffer)
				ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

				realmClient.HostingAssetsFn = func(groupID, appID string) ([]realm.HostingAsset, error) {
					return []realm.HostingAsset{
						{HostingAssetData: realm.HostingAssetData{FilePath: "/deleteme.html"}},
					}, nil
				}

				cmd := &Command{inputs{LocalPath: "testdata/hosting", RemoteApp: "appID", IncludeHosting: true}}

				err := cmd.Handler(profile, ui, cli.Clients{Realm: realmClient})
				assert.Equal(t, errors.New("1 error(s) occurred while importing hosting assets"), err)
				assert.Equal(t, `Determining changes
Creating draft
Pushing changes
Deploying draft
Deployment complete
An error occurred while uploading hosting assets: failed to remove /deleteme.html: something bad happened
`, out.String())
			})
		})

		t.Run("but fails to update a hosting asset attribute", func(t *testing.T) {
			profile, teardown := mock.NewProfileFromTmpDir(t, "push-handler")
			defer teardown()

			realmClient.HostingAssetAttributesUpdateFn = func(groupID, appID, path string, attrs ...realm.HostingAssetAttribute) error {
				return errors.New("something bad happened")
			}

			t.Run("should not return an error when include hosting flag is omitted", func(t *testing.T) {
				runImport(t, realmClient, "testdata/hosting")
			})

			t.Run("should return an error when modifying the attribute of a hosting asset", func(t *testing.T) {
				out := new(bytes.Buffer)
				ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

				realmClient.HostingAssetsFn = func(groupID, appID string) ([]realm.HostingAsset, error) {
					return []realm.HostingAsset{
						{
							HostingAssetData: realm.HostingAssetData{FilePath: "/index.html", FileHash: "daad4fb706d494feb9014e131f6520d4"},
							Attrs:            realm.HostingAssetAttributes{{api.HeaderContentLanguage, "en-US"}},
						},
						{
							HostingAssetData: realm.HostingAssetData{FilePath: "/404.html", FileHash: "7785338f982ac81219ef449f4943ec89"},
							Attrs:            realm.HostingAssetAttributes{{api.HeaderContentLanguage, "en-US"}},
						},
					}, nil
				}

				cmd := &Command{inputs{LocalPath: "testdata/hosting", RemoteApp: "appID", IncludeHosting: true}}

				err := cmd.Handler(profile, ui, cli.Clients{Realm: realmClient})
				assert.Equal(t, errors.New("2 error(s) occurred while importing hosting assets"), err)

				output := out.String()
				for _, line := range []string{
					`Determining changes
Creating draft
Pushing changes
Deploying draft
Deployment complete`,
					"An error occurred while uploading hosting assets: failed to update attributes for /404.html: something bad happened",
					"An error occurred while uploading hosting assets: failed to update attributes for /index.html: something bad happened",
				} {
					assert.True(t, strings.Contains(output, line), fmt.Sprintf("expected output to contain: '%s'\nactual output:\n%s", line, output))
				}
			})
		})

		t.Run("and can import hosting files should run import successfully", func(t *testing.T) {
			profile, teardown := mock.NewProfileFromTmpDir(t, "push-handler")
			defer teardown()

			out := new(bytes.Buffer)
			ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

			var added, removed, updated []string

			realmClient.HostingAssetUploadFn = func(groupID, appID, rootDir string, asset realm.HostingAsset) error {
				added = append(added, asset.FilePath)
				return nil
			}
			realmClient.HostingAssetRemoveFn = func(groupID, appID, path string) error {
				removed = append(removed, path)
				return nil
			}
			realmClient.HostingAssetAttributesUpdateFn = func(groupID, appID, path string, attrs ...realm.HostingAssetAttribute) error {
				updated = append(updated, path)
				return nil
			}
			realmClient.HostingAssetsFn = func(groupID, appID string) ([]realm.HostingAsset, error) {
				return []realm.HostingAsset{
					{HostingAssetData: realm.HostingAssetData{FilePath: "/deleteme.html"}},
					{
						HostingAssetData: realm.HostingAssetData{FilePath: "/404.html", FileHash: "7785338f982ac81219ef449f4943ec89"},
						Attrs:            realm.HostingAssetAttributes{{api.HeaderContentLanguage, "en-US"}},
					},
				}, nil
			}

			cmd := &Command{inputs{LocalPath: "testdata/hosting", RemoteApp: "appID", IncludeHosting: true}}

			err := cmd.Handler(profile, ui, cli.Clients{Realm: realmClient})
			assert.Nil(t, err)
			assert.Equal(t, `Determining changes
Creating draft
Pushing changes
Deploying draft
Deployment complete
Import hosting assets
Successfully pushed app up: eggcorn-abcde
`, out.String())

			assert.Equal(t, []string{"/index.html"}, added)
			assert.Equal(t, []string{"/deleteme.html"}, removed)
			assert.Equal(t, []string{"/404.html"}, updated)
		})

		t.Run("and can import hosting files but fails to invalidate cdn cache should return an error", func(t *testing.T) {
			profile, teardown := mock.NewProfileFromTmpDir(t, "push-handler")
			defer teardown()

			out := new(bytes.Buffer)
			ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

			realmClient.HostingAssetUploadFn = func(groupID, appID, rootDir string, asset realm.HostingAsset) error {
				return nil
			}
			realmClient.HostingAssetRemoveFn = func(groupID, appID, path string) error {
				return nil
			}
			realmClient.HostingAssetAttributesUpdateFn = func(groupID, appID, path string, attrs ...realm.HostingAssetAttribute) error {
				return nil
			}
			realmClient.HostingAssetsFn = func(groupID, appID string) ([]realm.HostingAsset, error) {
				return nil, nil
			}
			realmClient.HostingCacheInvalidateFn = func(groupID, appID, path string) error {
				return errors.New("something bad happened")
			}

			cmd := &Command{inputs{LocalPath: "testdata/hosting", RemoteApp: "appID", IncludeHosting: true, ResetCDNCache: true}}

			err := cmd.Handler(profile, ui, cli.Clients{Realm: realmClient})
			assert.Equal(t, errors.New("something bad happened"), err)
		})

		t.Run("and can import hosting files and invalidate the cdn cache should import successfully", func(t *testing.T) {
			profile, teardown := mock.NewProfileFromTmpDir(t, "push-handler")
			defer teardown()

			out := new(bytes.Buffer)
			ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

			realmClient.HostingAssetUploadFn = func(groupID, appID, rootDir string, asset realm.HostingAsset) error {
				return nil
			}
			realmClient.HostingAssetRemoveFn = func(groupID, appID, path string) error {
				return nil
			}
			realmClient.HostingAssetAttributesUpdateFn = func(groupID, appID, path string, attrs ...realm.HostingAssetAttribute) error {
				return nil
			}
			realmClient.HostingAssetsFn = func(groupID, appID string) ([]realm.HostingAsset, error) {
				return nil, nil
			}
			realmClient.HostingCacheInvalidateFn = func(groupID, appID, path string) error {
				return nil
			}

			cmd := &Command{inputs{LocalPath: "testdata/hosting", RemoteApp: "appID", IncludeHosting: true, ResetCDNCache: true}}

			err := cmd.Handler(profile, ui, cli.Clients{Realm: realmClient})
			assert.Nil(t, err)
			assert.Equal(t, `Determining changes
Creating draft
Pushing changes
Deploying draft
Deployment complete
Import hosting assets
Reset CDN cache
Successfully pushed app up: eggcorn-abcde
`, out.String())
		})

		t.Run("but fails to import dependencies", func(t *testing.T) {
			realmClient.DiffDependenciesFn = func(groupID, appID, uploadPath string) (realm.DependenciesDiff, error) {
				return realm.DependenciesDiff{}, nil
			}
			realmClient.ImportDependenciesFn = func(groupID, appID, uploadPath string) error {
				return errors.New("something bad happened")
			}

			t.Run("should not return an error when include dependencies flag is omitted", func(t *testing.T) {
				out := new(bytes.Buffer)
				ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

				cmd := &Command{inputs{LocalPath: "testdata/project", RemoteApp: "appID"}}

				assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
				assert.Equal(t, `Determining changes
Creating draft
Pushing changes
Deploying draft
Deployment complete
Successfully pushed app up: eggcorn-abcde
`, out.String())
			})

			t.Run("should return an error when include dependencies flag is set", func(t *testing.T) {
				out := new(bytes.Buffer)
				ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

				cmd := &Command{inputs{LocalPath: "testdata/project", RemoteApp: "appID", IncludeNodeModules: true}}

				err := cmd.Handler(nil, ui, cli.Clients{Realm: realmClient})
				assert.Equal(t, errors.New("something bad happened"), err)
			})
		})

		t.Run("succeeds in importing dependencies but fails during installation", func(t *testing.T) {
			realmClient.DiffDependenciesFn = func(groupID, appID, uploadPath string) (realm.DependenciesDiff, error) {
				return realm.DependenciesDiff{}, nil
			}
			realmClient.ImportDependenciesFn = func(groupID, appID, uploadPath string) error {
				return nil
			}

			t.Run("because of an error", func(t *testing.T) {
				realmClient.DependenciesStatusFn = func(groupID, appID string) (realm.DependenciesStatus, error) {
					return realm.DependenciesStatus{}, errors.New("something bad happened")
				}
				out := new(bytes.Buffer)
				ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

				cmd := &Command{inputs{LocalPath: "testdata/project", RemoteApp: "appID", IncludeNodeModules: true}}

				err := cmd.Handler(nil, ui, cli.Clients{Realm: realmClient})
				assert.Equal(t, errors.New("something bad happened"), err)
			})
			t.Run("because of an installation problem", func(t *testing.T) {
				realmClient.DependenciesStatusFn = func(groupID, appID string) (realm.DependenciesStatus, error) {
					return realm.DependenciesStatus{State: realm.DependenciesStateFailed, Message: "something bad happened"}, nil
				}
				out := new(bytes.Buffer)
				ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

				cmd := &Command{inputs{LocalPath: "testdata/project", RemoteApp: "appID", IncludeNodeModules: true}}

				err := cmd.Handler(nil, ui, cli.Clients{Realm: realmClient})
				assert.Equal(t, errors.New("failed to install dependencies: something bad happened"), err)
			})
		})

		t.Run("and can import dependencies should run import successfully", func(t *testing.T) {
			realmClient.ImportDependenciesFn = func(groupID, appID, uploadPath string) error {
				return nil
			}
			realmClient.DependenciesStatusFn = func(groupID, appID string) (realm.DependenciesStatus, error) {
				return realm.DependenciesStatus{State: realm.DependenciesStateSuccessful}, nil
			}

			out := new(bytes.Buffer)
			ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

			cmd := &Command{inputs{LocalPath: "testdata/project", RemoteApp: "appID", IncludeNodeModules: true}}

			assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
			assert.Equal(t, `Determining changes
Creating draft
Pushing changes
Deploying draft
Deployment complete
Installed dependencies
Successfully pushed app up: eggcorn-abcde
`, out.String())
		})
	})

	t.Run("should exit early in a dry run", func(t *testing.T) {
		for _, tc := range []struct {
			description  string
			groupID      string
			groupsCalled bool
		}{
			{
				description:  "and should fetch group id if to is not resolved",
				groupsCalled: true,
			},
			{
				description: "and should not fetch group id if to is resolved",
				groupID:     "groupID",
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				var atlasClient mock.AtlasClient
				var calledGroups bool
				atlasClient.GroupsFn = func(url string, useBaseURL bool) (atlas.Groups, error) {
					calledGroups = true
					return atlas.Groups{Results: []atlas.Group{{ID: "groupID", Name: "groupName"}}}, nil
				}

				var realmClient mock.RealmClient
				realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
					return []realm.App{{GroupID: tc.groupID}}, nil
				}

				cmd := &Command{inputs{LocalPath: "testdata/project", DryRun: true, RemoteApp: "appID"}}

				out, ui := mock.NewUI()

				err := cmd.Handler(nil, ui, cli.Clients{
					Realm: realmClient,
					Atlas: atlasClient,
				})
				assert.Nil(t, err)

				assert.Equal(t, tc.groupsCalled, calledGroups)
				assert.Equal(t, `This is a new app. To create a new app, you must omit the 'dry-run' flag to proceed
Try instead: realm-cli push --local testdata/project --remote appID
`, out.String())
			})
		}
	})

	t.Run("with a user rejecting the option to create a new app", func(t *testing.T) {
		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{GroupID: "groupID"}}, nil
		}

		_, console, _, ui, consoleErr := mock.NewVT10XConsole()
		assert.Nil(t, consoleErr)
		defer console.Close()

		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)

			console.ExpectString("Do you wish to create a new app?")
			console.SendLine("")
			console.ExpectEOF()
		}()

		cmd := &Command{inputs{LocalPath: "testdata/project", RemoteApp: "appID"}}

		err := cmd.Handler(nil, ui, cli.Clients{Realm: realmClient})

		console.Tty().Close() // flush the writers
		<-doneCh              // wait for procedure to complete

		assert.Nil(t, err)
	})

	t.Run("with no diffs generated from the app", func(t *testing.T) {
		out, ui := mock.NewUI()

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", GroupID: "groupID"}}, nil
		}
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			return []string{}, nil
		}

		cmd := &Command{inputs{LocalPath: "testdata/project", DryRun: true, RemoteApp: "appID"}}

		err := cmd.Handler(nil, ui, cli.Clients{Realm: realmClient})
		assert.Nil(t, err)
		assert.Equal(t, `Determining changes
Deployed app is identical to proposed version, nothing to do
`, out.String())
	})

	t.Run("with app meta should skip resolving app", func(t *testing.T) {
		out, ui := mock.NewUI()

		var realmClient mock.RealmClient
		var findAppsCalled bool
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			findAppsCalled = true
			return []realm.App{{ID: "appID", GroupID: "groupID"}}, nil
		}
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			return []string{}, nil
		}

		var atlasClient mock.AtlasClient
		var groupsCalled bool
		atlasClient.GroupsFn = func(url string, useBaseURL bool) (atlas.Groups, error) {
			groupsCalled = true
			return atlas.Groups{Results: []atlas.Group{{ID: "groupID", Name: "groupName"}}}, nil
		}

		cmd := &Command{inputs{LocalPath: "testdata/project-meta", DryRun: true}}

		err := cmd.Handler(nil, ui, cli.Clients{Realm: realmClient, Atlas: atlasClient})
		assert.Nil(t, err)
		assert.False(t, findAppsCalled, "Expected to skip resolve app ID")
		assert.False(t, groupsCalled, "Expected to skip resolve group ID")
		assert.Equal(t, `Determining changes
Deployed app is identical to proposed version, nothing to do
`, out.String())
	})

	t.Run("with diffs generated from the app but is a dry run", func(t *testing.T) {
		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", GroupID: "groupID"}}, nil
		}
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			return []string{"diff1", "diff2"}, nil
		}

		out, ui := mock.NewUI()

		cmd := &Command{inputs{LocalPath: "testdata/project", DryRun: true, RemoteApp: "appID"}}

		err := cmd.Handler(nil, ui, cli.Clients{Realm: realmClient})

		assert.Nil(t, err)
		assert.Equal(t, `Determining changes
The following reflects the proposed changes to your Realm app
diff1
diff2
To push these changes, you must omit the 'dry-run' flag to proceed
Try instead: realm-cli push --local testdata/project --remote appID
`, out.String())
	})

	t.Run("with diffs including dependencies generated from the app but is a dry run", func(t *testing.T) {
		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", GroupID: "groupID"}}, nil
		}
		realmClient.DiffDependenciesFn = func(groupID, appID, uploadPath string) (realm.DependenciesDiff, error) {
			return realm.DependenciesDiff{
				Added:    []realm.DependencyData{{"twilio", "3.35.1"}},
				Deleted:  []realm.DependencyData{{"debug", "4.3.1"}},
				Modified: []realm.DependencyDiffData{{DependencyData: realm.DependencyData{"underscore", "1.9.2"}, PreviousVersion: "1.9.1"}},
			}, nil
		}
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			return []string{"diff1", "diff2"}, nil
		}

		out, ui := mock.NewUI()

		cmd := &Command{inputs{LocalPath: "testdata/dependencies", DryRun: true, RemoteApp: "appID", IncludeNodeModules: true}}

		err := cmd.Handler(nil, ui, cli.Clients{Realm: realmClient})

		assert.Nil(t, err)
		assert.Equal(t, `Determining changes
The following reflects the proposed changes to your Realm app
diff1
diff2
Added Dependencies
  + twilio@3.35.1
Removed Dependencies
  - debug@4.3.1
Modified Dependencies
  * underscore@1.9.1 -> underscore@1.9.2
To push these changes, you must omit the 'dry-run' flag to proceed
Try instead: realm-cli push --local testdata/dependencies --remote appID --include-node-modules
`, out.String())
	})

	t.Run("should return error when diffs include dependencies and diff dependencies returns an error", func(t *testing.T) {
		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", GroupID: "groupID"}}, nil
		}
		realmClient.DiffDependenciesFn = func(groupID, appID, uploadPath string) (realm.DependenciesDiff, error) {
			return realm.DependenciesDiff{}, errors.New("realm client error")
		}
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			return []string{"diff1", "diff2"}, nil
		}

		_, ui := mock.NewUI()

		cmd := &Command{inputs{LocalPath: "testdata/dependencies", RemoteApp: "appID", IncludeNodeModules: true}}

		assert.Equal(t, errors.New("realm client error"), cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
	})

	t.Run("with diffs generated from the app but the user rejects them", func(t *testing.T) {
		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", GroupID: "groupID"}}, nil
		}
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			return []string{"diff1", "diff2"}, nil
		}

		_, console, _, ui, consoleErr := mock.NewVT10XConsole()
		assert.Nil(t, consoleErr)
		defer console.Close()

		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)

			console.ExpectString("Please confirm the changes shown above")
			console.SendLine("")
			console.ExpectEOF()
		}()

		cmd := &Command{inputs{LocalPath: "testdata/project", RemoteApp: "appID"}}

		err := cmd.Handler(nil, ui, cli.Clients{Realm: realmClient})

		console.Tty().Close() // flush the writers
		<-doneCh              // wait for procedure to complete

		assert.Nil(t, err)
	})
}

func TestPushHandlerCreateNewApp(t *testing.T) {
	groupID := "groupID"
	appID := "eggcorn-abcde"

	realmClient := mock.RealmClient{}
	realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
		return []realm.App{{GroupID: groupID, ClientAppID: appID}}, nil
	}
	realmClient.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
		return realm.App{
			GroupID:     groupID,
			ClientAppID: "eggcorn-abcde",
			ID:          appID,
			Name:        name,
			AppMeta:     meta,
		}, nil
	}
	realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) { return realm.AppDraft{ID: "id"}, nil }
	realmClient.ImportFn = func(groupID, appID string, appData interface{}) error { return nil }
	realmClient.DeployDraftFn = func(groupID, appID, draftID string) (realm.AppDeployment, error) {
		return realm.AppDeployment{Status: realm.DeploymentStatusSuccessful}, nil
	}

	t.Run("should successfully push with local path in a nested directory", func(t *testing.T) {
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			return []string{}, nil
		}

		for _, tc := range []struct {
			appConfig      local.File
			appData        local.AppData
			expectedConfig string
		}{
			{
				appConfig: local.FileRealmConfig,
				appData: &local.AppRealmConfigJSON{local.AppDataV2{local.AppStructureV2{
					ConfigVersion:   realm.AppConfigVersion20210101,
					Name:            "eggcorn",
					Location:        realm.Location("location"),
					DeploymentModel: realm.DeploymentModel("deployment_model"),
					Environment:     realm.Environment("environment"),
				}}},
				expectedConfig: `{
    "config_version": 20210101,
    "app_id": "eggcorn-abcde",
    "name": "eggcorn",
    "location": "location",
    "deployment_model": "deployment_model",
    "environment": "environment"
}
`,
			},
			{
				appConfig: local.FileConfig,
				appData: &local.AppConfigJSON{local.AppDataV1{local.AppStructureV1{
					ConfigVersion:   realm.AppConfigVersion20200603,
					Name:            "eggcorn",
					Location:        realm.Location("location"),
					DeploymentModel: realm.DeploymentModel("deployment_model"),
					Environment:     realm.Environment("environment"),
				}}},
				expectedConfig: `{
    "config_version": 20200603,
    "app_id": "eggcorn-abcde",
    "name": "eggcorn",
    "location": "location",
    "deployment_model": "deployment_model",
    "environment": "environment",
    "security": null,
    "custom_user_data_config": {
        "enabled": false
    },
    "sync": {
        "development_mode_enabled": false
    }
}
`,
			},
			{
				appConfig: local.FileStitch,
				appData: &local.AppStitchJSON{local.AppDataV1{local.AppStructureV1{
					ConfigVersion:   realm.AppConfigVersion20180301,
					Name:            "eggcorn",
					Location:        realm.Location("location"),
					DeploymentModel: realm.DeploymentModel("deployment_model"),
					Environment:     realm.Environment("environment"),
				}}},
				expectedConfig: `{
    "config_version": 20180301,
    "app_id": "eggcorn-abcde",
    "name": "eggcorn",
    "location": "location",
    "deployment_model": "deployment_model",
    "environment": "environment",
    "security": null,
    "custom_user_data_config": {
        "enabled": false
    },
    "sync": {
        "development_mode_enabled": false
    }
}
`,
			},
		} {
			t.Run(fmt.Sprintf("and a %s config file", tc.appConfig.String()), func(t *testing.T) {
				tmpDir, teardown, tmpDirErr := u.NewTempDir("push_handler")
				assert.Nil(t, tmpDirErr)
				defer teardown()

				assert.Nil(t, os.Mkdir(filepath.Join(tmpDir, "nested"), os.ModePerm))

				app := local.App{RootDir: tmpDir, Config: tc.appConfig, AppData: tc.appData}

				app.WriteConfig()

				out := new(bytes.Buffer)
				console, ui, consoleErr := mock.NewConsoleWithOptions(mock.UIOptions{AutoConfirm: true}, out)
				assert.Nil(t, consoleErr)
				defer console.Close()

				cmd := &Command{inputs{LocalPath: filepath.Join(tmpDir, "nested"), RemoteApp: "appID"}}
				assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))

				configData, readErr := ioutil.ReadFile(filepath.Join(tmpDir, tc.appConfig.String()))
				assert.Nil(t, readErr)
				assert.Equal(t, tc.expectedConfig, string(configData))

				_, fileErr := os.Stat(filepath.Join(tmpDir, "nested", tc.appConfig.String()))
				assert.True(t, os.IsNotExist(fileErr), "expected nested config path to not exist, but err was: %s", fileErr)
			})
		}
	})

	t.Run("should compare the local app to the created app", func(t *testing.T) {
		for _, tc := range []struct {
			description string
			diffFn      func(groupId, appID string, appData interface{}) ([]string, error)
			procedure   func(c *expect.Console)
		}{
			{
				description: "with an identical diff",
				diffFn:      func(groupId, appID string, appData interface{}) ([]string, error) { return []string{}, nil },
				procedure: func(c *expect.Console) {
					c.ExpectString("Determining changes")
					c.ExpectString("Deployed app is identical to proposed version, nothing to do")
				},
			},
			{
				description: "with a diff which is accepted by the user",
				diffFn:      func(groupId, appID string, appData interface{}) ([]string, error) { return []string{"diff"}, nil },
				procedure: func(c *expect.Console) {
					c.ExpectString("Please confirm the changes shown above")
					c.SendLine("y")

					c.ExpectString("Determining changes")
					c.ExpectString("Creating draft")
					c.ExpectString("Deploying draft")
					c.ExpectString("Deployment complete")
					c.ExpectString("Successfully pushed app up: eggcorn-abcde")
				},
			},
			{
				description: "with a diff which is rejected by the user",
				diffFn:      func(groupId, appID string, appData interface{}) ([]string, error) { return []string{"diff"}, nil },
				procedure: func(c *expect.Console) {
					c.ExpectString("Please confirm the changes")
					c.SendLine("")
				},
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				realmClient.DiffFn = tc.diffFn
				tmpDir, teardown, tmpDirErr := u.NewTempDir("push_handler")
				assert.Nil(t, tmpDirErr)
				defer teardown()

				app := local.App{RootDir: tmpDir, Config: local.FileRealmConfig, AppData: &local.AppRealmConfigJSON{local.AppDataV2{local.AppStructureV2{
					ConfigVersion:   realm.AppConfigVersion20210101,
					Name:            "eggcorn",
					Location:        realm.Location("location"),
					DeploymentModel: realm.DeploymentModel("deployment_model"),
					Environment:     realm.Environment("environment"),
				}}}}

				app.WriteConfig()

				_, console, _, ui, consoleErr := mock.NewVT10XConsole()
				assert.Nil(t, consoleErr)
				defer console.Close()

				doneCh := make(chan (struct{}))
				go func() {
					defer close(doneCh)
					console.ExpectString("Do you wish to create a new app?")
					console.SendLine("y")

					console.ExpectString("App Name")
					console.SendLine("testApp")

					console.ExpectString("App Location")
					console.Send(string(terminal.KeyArrowDown))
					console.SendLine("")

					console.ExpectString("App Deployment Model")
					console.Send(string(terminal.KeyArrowDown))
					console.SendLine("")

					console.ExpectString("App Environment")
					console.Send(string(terminal.KeyArrowDown))
					console.SendLine("")

					console.ExpectString("Please confirm the new app details shown above")
					console.SendLine("y")

					console.ExpectString("App created successfully")

					tc.procedure(console)

					console.ExpectEOF()
				}()

				cmd := &Command{inputs{LocalPath: tmpDir, RemoteApp: "appID"}}
				assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))

				console.Tty().Close() // flush the writers
				<-doneCh              // wait for procedure to complete

				_, fileErr := os.Stat(filepath.Join(tmpDir, local.FileRealmConfig.String()))
				assert.Nil(t, fileErr)
			})
		}
	})
}

func TestPushCommandCreateNewApp(t *testing.T) {
	groupID := "groupID"
	appID := primitive.NewObjectID().Hex()

	fullPkg := &local.AppConfigJSON{local.AppDataV1{local.AppStructureV1{
		ConfigVersion:   realm.AppConfigVersion20200603,
		Name:            "name",
		Location:        realm.Location("location"),
		DeploymentModel: realm.DeploymentModel("deployment_model"),
		Environment:     realm.Environment("environment"),
	}}}

	t.Run("with a client that successfully creates apps", func(t *testing.T) {
		realmClient := mock.RealmClient{}
		realmClient.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return realm.App{
				GroupID: groupID,
				ID:      appID,
				Name:    name,
				AppMeta: meta,
			}, nil
		}

		t.Run("and a ui that is not set to auto confirm", func(t *testing.T) {
			for _, tc := range []struct {
				description     string
				autoConfirm     bool
				procedure       func(c *expect.Console)
				expectedApp     realm.App
				expectedProceed bool
				test            func(t *testing.T, configPath string)
			}{
				{
					description: "should return empty data if user does not wish to continue",
					procedure: func(c *expect.Console) {
						c.ExpectString("Do you wish to create a new app?")
						c.SendLine("")
						c.ExpectEOF()
					},
					test: func(t *testing.T, configPath string) {
						_, err := os.Stat(configPath)
						assert.True(t, os.IsNotExist(err), "expected config path to not exist, but err was: %s", err)
					},
				},
				{
					description: "should prompt for all missing app info if user does want to continue",
					procedure: func(c *expect.Console) {
						c.ExpectString("Do you wish to create a new app?")
						c.SendLine("y")

						c.ExpectString("App Name")
						c.SendLine("testApp")

						c.ExpectString("App Location")
						c.Send(string(terminal.KeyArrowDown))
						c.SendLine("")

						c.ExpectString("App Deployment Model")
						c.Send(string(terminal.KeyArrowDown))
						c.SendLine("")

						c.ExpectString("App Environment")
						c.Send(string(terminal.KeyArrowDown))
						c.SendLine("")

						c.ExpectString("Please confirm the new app details shown above")
						c.SendLine("y")

						c.ExpectEOF()
					},
					test: func(t *testing.T, configPath string) {
						configData, readErr := ioutil.ReadFile(configPath)
						assert.Nil(t, readErr)
						assert.Equal(t, `{
    "config_version": 20210101,
    "name": "testApp",
    "location": "US-OR",
    "deployment_model": "LOCAL",
    "environment": "testing"
}
`, string(configData))
					},
					expectedApp: realm.App{
						ID:      appID,
						GroupID: groupID,
						Name:    "testApp",
						AppMeta: realm.AppMeta{
							Location:        realm.LocationOregon,
							DeploymentModel: realm.DeploymentModelLocal,
							Environment:     realm.EnvironmentTesting,
						},
					},
					expectedProceed: true,
				},
				{
					description: "should return empty data if user rejects app configuration",
					procedure: func(c *expect.Console) {
						c.ExpectString("Do you wish to create a new app?")
						c.SendLine("y")

						c.ExpectString("App Name")
						c.SendLine("testApp")

						c.ExpectString("App Location")
						c.Send(string(terminal.KeyArrowDown))
						c.SendLine("")

						c.ExpectString("App Deployment Model")
						c.Send(string(terminal.KeyArrowDown))
						c.SendLine("")

						c.ExpectString("App Environment")
						c.Send(string(terminal.KeyArrowDown))
						c.SendLine("")

						c.ExpectString("Please confirm the new app details shown above")
						c.SendLine("")

						c.ExpectEOF()
					},
					test: func(t *testing.T, configPath string) {
						_, err := os.Stat(configPath)
						assert.True(t, os.IsNotExist(err), "expected config path to not exist, but err was: %s", err)
					},
				},
				{
					description: "should still prompt for name with all missing data but auto confirm set to true",
					autoConfirm: true,
					procedure: func(c *expect.Console) {
						c.ExpectString("App Name")
						c.SendLine("testApp")

						c.ExpectEOF()
					},
					expectedApp: realm.App{
						ID:      appID,
						GroupID: groupID,
						Name:    "testApp",
					},
					expectedProceed: true,
					test: func(t *testing.T, configPath string) {
						configData, readErr := ioutil.ReadFile(configPath)
						assert.Nil(t, readErr)
						assert.Equal(t, `{
    "config_version": 20210101,
    "name": "testApp"
}
`, string(configData))
					},
				},
			} {
				t.Run(tc.description, func(t *testing.T) {
					tmpDir, teardown, tmpDirErr := u.NewTempDir("push_handler")
					assert.Nil(t, tmpDirErr)
					defer teardown()

					out := new(bytes.Buffer)
					console, _, ui, consoleErr := mock.NewVT10XConsoleWithOptions(mock.UIOptions{AutoConfirm: tc.autoConfirm}, out)
					assert.Nil(t, consoleErr)
					defer console.Close()

					doneCh := make(chan (struct{}))
					go func() {
						defer close(doneCh)
						tc.procedure(console)
					}()

					app, proceed, err := createNewApp(ui, realmClient, tmpDir, groupID, map[string]interface{}{})

					console.Tty().Close() // flush the writers
					<-doneCh              // wait for procedure to complete

					assert.Nil(t, err)
					assert.Equal(t, tc.expectedProceed, proceed)
					assert.Equal(t, tc.expectedApp, app)

					tc.test(t, filepath.Join(tmpDir, local.FileRealmConfig.String()))
				})
			}
		})

		t.Run("and a static ui that is set to auto confirm", func(t *testing.T) {
			out := new(bytes.Buffer)
			ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

			for _, tc := range []struct {
				description       string
				appData           interface{}
				expectedAppMeta   realm.AppMeta
				expectedAppConfig local.File
			}{
				{
					description:       "should use the package name when present and zero values for app meta",
					appData:           local.AppConfigJSON{local.AppDataV1{local.AppStructureV1{ConfigVersion: realm.AppConfigVersion20200603, Name: "name"}}},
					expectedAppConfig: local.FileConfig,
				},
				{
					description: "should use the package name location deployment model and environment when present",
					appData:     fullPkg,
					expectedAppMeta: realm.AppMeta{
						Location:        realm.Location("location"),
						DeploymentModel: realm.DeploymentModel("deployment_model"),
						Environment:     realm.Environment("environment"),
					},
					expectedAppConfig: local.FileConfig,
				},
			} {
				t.Run(tc.description, func(t *testing.T) {
					tmpDir, teardown, tmpDirErr := u.NewTempDir("push_handler")
					assert.Nil(t, tmpDirErr)
					defer teardown()

					expectedApp := realm.App{
						GroupID: groupID,
						ID:      appID,
						Name:    "name",
						AppMeta: tc.expectedAppMeta,
					}

					app, proceed, err := createNewApp(ui, realmClient, tmpDir, "groupID", tc.appData)
					assert.Nil(t, err)
					assert.True(t, proceed, "should proceed")
					assert.Equal(t, expectedApp, app)

					_, fileErr := os.Stat(filepath.Join(tmpDir, tc.expectedAppConfig.String()))
					assert.Nil(t, fileErr)
				})
			}
		})

		t.Run("and an interactive ui that is set to auto confirm", func(t *testing.T) {
			for _, tc := range []struct {
				description     string
				appData         interface{}
				expectedAppMeta realm.AppMeta
			}{
				{
					description: "should prompt for name if not present in the package",
					appData:     local.AppConfigJSON{local.AppDataV1{local.AppStructureV1{Location: realm.Location("location"), DeploymentModel: realm.DeploymentModel("deployment_model"), Environment: realm.Environment("environment")}}},
					expectedAppMeta: realm.AppMeta{
						Location:        realm.Location("location"),
						DeploymentModel: realm.DeploymentModel("deployment_model"),
						Environment:     realm.Environment("environment"),
					},
				},
				{
					description: "should not prompt for location deployment model and environment even if not present in the package",
					appData:     map[string]interface{}{},
				},
			} {
				t.Run(tc.description, func(t *testing.T) {
					tmpDir, teardown, tmpDirErr := u.NewTempDir("push_handler")
					assert.Nil(t, tmpDirErr)
					defer teardown()

					out := new(bytes.Buffer)
					console, _, ui, consoleErr := mock.NewVT10XConsoleWithOptions(mock.UIOptions{AutoConfirm: true}, out)
					assert.Nil(t, consoleErr)
					defer console.Close()

					doneCh := make(chan (struct{}))
					go func() {
						defer close(doneCh)

						console.ExpectString("App Name")
						console.SendLine("test-app")
						console.ExpectEOF()
					}()

					app, proceed, err := createNewApp(ui, realmClient, tmpDir, groupID, tc.appData)
					assert.Nil(t, err)

					console.Tty().Close() // flush the writers
					<-doneCh              // wait for procedure to complete

					expectedApp := realm.App{
						GroupID: groupID,
						ID:      appID,
						Name:    "test-app",
						AppMeta: tc.expectedAppMeta,
					}

					assert.True(t, proceed, "should proceed")
					assert.Equal(t, expectedApp, app)
				})
			}
		})
	})

	t.Run("with a client that fails to create apps it should return that error", func(t *testing.T) {
		realmClient := mock.RealmClient{}
		realmClient.CreateAppFn = func(groupID, name string, meta realm.AppMeta) (realm.App, error) {
			return realm.App{}, errors.New("something bad happened")
		}

		out := new(bytes.Buffer)
		ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

		_, _, err := createNewApp(ui, realmClient, "", "groupID", fullPkg)
		assert.Equal(t, errors.New("something bad happened"), err)
	})
}

func TestPushCommandCreateNewDraft(t *testing.T) {
	t.Run("should create and return the draft when initially successful", func(t *testing.T) {
		groupID, appID, clientAppID := "groupID", "appID", "client-app-id"
		testDraft := realm.AppDraft{ID: "id"}

		realmClient := mock.RealmClient{}

		var capturedGroupID, capturedAppID string
		realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) {
			capturedGroupID = groupID
			capturedAppID = appID
			return testDraft, nil
		}

		draft, proceed, err := createNewDraft(nil, realmClient, appRemote{groupID, appID, clientAppID})
		assert.Nil(t, err)
		assert.Equal(t, testDraft, draft)
		assert.True(t, proceed, "expected draft to be created successfully")

		t.Log("and should properly pass through the expected inputs")
		assert.Equal(t, groupID, capturedGroupID)
		assert.Equal(t, appID, capturedAppID)
	})

	t.Run("should return the error if client fails to create the draft for reasons other than it already exists", func(t *testing.T) {
		realmClient := mock.RealmClient{}
		realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) {
			return realm.AppDraft{}, errors.New("something bad happened while creating a draft")
		}

		_, _, err := createNewDraft(nil, realmClient, appRemote{})
		assert.Equal(t, errors.New("something bad happened while creating a draft"), err)
	})

	t.Run("with a client that fails to create a draft because it already exists", func(t *testing.T) {
		errDraftAlreadyExists := realm.ServerError{Code: realm.ErrCodeDraftAlreadyExists, Message: "a draft already exists"}

		realmClient := mock.RealmClient{}
		realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) {
			return realm.AppDraft{}, errDraftAlreadyExists
		}

		t.Run("and fails to retrieve the existing draft should return the error", func(t *testing.T) {
			realmClient.DraftFn = func(groupID, appID string) (realm.AppDraft, error) {
				return realm.AppDraft{}, errors.New("something bad happened while getting a draft")
			}

			_, _, err := createNewDraft(nil, realmClient, appRemote{})
			assert.Equal(t, errors.New("something bad happened while getting a draft"), err)
		})

		t.Run("and successfully retrieves and diffs the existing draft", func(t *testing.T) {
			draftID := "draftID"

			realmClient.DraftFn = func(groupID, appID string) (realm.AppDraft, error) {
				return realm.AppDraft{ID: draftID}, nil
			}

			realmClient.DiffDraftFn = func(groupID, appID, draftID string) (realm.AppDraftDiff, error) {
				return realm.AppDraftDiff{}, nil
			}

			t.Run("with a ui set to auto-confirm", func(t *testing.T) {
				out := new(bytes.Buffer)
				ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

				t.Run("and a client that fails to discard the draft should return the error", func(t *testing.T) {
					var capturedDraftID string
					realmClient.DiscardDraftFn = func(groupID, appID, draftID string) error {
						capturedDraftID = draftID
						return errors.New("something bad happened while discarding the draft")
					}

					_, _, err := createNewDraft(ui, realmClient, appRemote{})
					assert.Equal(t, errors.New("something bad happened while discarding the draft"), err)

					t.Log("and should properly pass through the expected inputs")
					assert.Equal(t, draftID, capturedDraftID)
				})

				t.Run("and a client that successfully discards the existing draft", func(t *testing.T) {
					realmClient.DiscardDraftFn = func(groupID, appID, draftID string) error {
						return nil
					}

					t.Run("but still fails to create a new draft should return the error", func(t *testing.T) {
						realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) {
							return realm.AppDraft{}, errDraftAlreadyExists
						}

						_, _, err := createNewDraft(ui, realmClient, appRemote{})
						assert.Equal(t, errDraftAlreadyExists, err)
					})

					t.Run("and successfully creates a new draft should be successful", func(t *testing.T) {
						testDraft := realm.AppDraft{ID: "id"}

						realmClient := mock.RealmClient{}

						realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) {
							return testDraft, nil
						}

						draft, proceed, err := createNewDraft(nil, realmClient, appRemote{})
						assert.Nil(t, err)
						assert.Equal(t, testDraft, draft)
						assert.True(t, proceed, "expected draft to be created successfully")
					})
				})
			})

			t.Run("should prompt the user to accept the diffed changes", func(t *testing.T) {
				t.Run("and mark the draft as kept in command outputs if the user selects no", func(t *testing.T) {
					_, console, _, ui, consoleErr := mock.NewVT10XConsole()
					assert.Nil(t, consoleErr)
					defer console.Close()

					doneCh := make(chan (struct{}))
					go func() {
						defer close(doneCh)

						console.ExpectString("Would you like to discard this draft?")
						console.SendLine("")
						console.ExpectEOF()
					}()

					draft, proceed, err := createNewDraft(ui, realmClient, appRemote{})

					console.Tty().Close() // flush the writers
					<-doneCh              // wait for procedure to complete

					assert.Nil(t, err)
					assert.Equal(t, realm.AppDraft{}, draft)
					assert.False(t, proceed, "expected draft to be rejected")
				})

				t.Run("and return a newly created draft if the user selects yes", func(t *testing.T) {
					testDraft := realm.AppDraft{ID: "id"}

					realmClient.DiscardDraftFn = func(groupID, appID, draftID string) error {
						return nil
					}

					realmClient.CreateDraftFn = func(groupID, appID string) (realm.AppDraft, error) {
						return testDraft, nil
					}

					_, console, _, ui, consoleErr := mock.NewVT10XConsole()
					assert.Nil(t, consoleErr)
					defer console.Close()

					doneCh := make(chan (struct{}))
					go func() {
						defer close(doneCh)

						console.ExpectString("Would you like to discard this draft?")
						console.SendLine("y")
						console.ExpectEOF()
					}()

					draft, proceed, err := createNewDraft(ui, realmClient, appRemote{})

					console.Tty().Close() // flush the writers
					<-doneCh              // wait for procedure to complete

					assert.Nil(t, err)
					assert.Equal(t, testDraft, draft)
					assert.True(t, proceed, "expected draft to be created successfully")
				})
			})
		})
	})
}

func TestPushCommandDiffDraft(t *testing.T) {
	t.Run("with a client that fails to diff the draft should return the error", func(t *testing.T) {
		groupID, appID, clientAppID, draftID := "groupID", "appID", "client-app-id", "draftID"

		var realmClient mock.RealmClient

		var capturedGroupID, capturedAppID, capturedDraftID string
		realmClient.DiffDraftFn = func(groupID, appID, draftID string) (realm.AppDraftDiff, error) {
			capturedGroupID = groupID
			capturedAppID = appID
			capturedDraftID = draftID
			return realm.AppDraftDiff{}, errors.New("something bad happened")
		}

		err := diffDraft(nil, realmClient, appRemote{groupID, appID, clientAppID}, draftID)
		assert.Equal(t, errors.New("something bad happened"), err)

		t.Log("and should properly pass through the expected inputs")
		assert.Equal(t, groupID, capturedGroupID)
		assert.Equal(t, appID, capturedAppID)
		assert.Equal(t, draftID, capturedDraftID)
	})

	t.Run("should print the expected contents", func(t *testing.T) {
		for _, tc := range []struct {
			description      string
			actualDiff       realm.AppDraftDiff
			expectedContents string
		}{
			{
				description:      "with a client that returns an empty diff",
				expectedContents: "An empty draft already exists for your app\n",
			},
			{
				description: "with a client that returns a minimal diff",
				actualDiff:  realm.AppDraftDiff{Diffs: []string{"diff1", "diff2", "diff3"}},
				expectedContents: strings.Join(
					[]string{
						"The following draft already exists for your app...",
						"  diff1",
						"  diff2",
						"  diff3\n",
					},
					"\n",
				),
			},
			{
				description: "with a client that returns a full diff",
				actualDiff: realm.AppDraftDiff{
					Diffs: []string{"diff1", "diff2", "diff3"},
					HostingFilesDiff: realm.HostingFilesDiff{
						Added:   []string{"hosting_added1", "hosting_added2"},
						Deleted: []string{"hosting_deleted1"},
					},
					DependenciesDiff: realm.DependenciesDiff{
						Added: []realm.DependencyData{{"dep_added1", "v1"}},
						Modified: []realm.DependencyDiffData{
							{realm.DependencyData{"dep_modified1", "v1"}, "v2"},
							{realm.DependencyData{"dep_modified2", "v2"}, "v1"},
						},
					},
					GraphQLConfigDiff: realm.GraphQLConfigDiff{[]realm.FieldDiff{{"gql_field1", "previous", "updated"}}},
					SchemaOptionsDiff: realm.SchemaOptionsDiff{
						GraphQLValidationDiffs: []realm.FieldDiff{{"gql_validation_field1", "old", "new"}},
						RestValidationDiffs:    []realm.FieldDiff{{"rest_validation_field1", "old", "new"}},
					},
				},
				expectedContents: strings.Join(
					[]string{
						"The following draft already exists for your app...",
						"  diff1",
						"  diff2",
						"  diff3",
						"With changes to your static hosting files...",
						"  added: hosting_added1",
						"  added: hosting_added2",
						"  deleted: hosting_deleted1",
						"With changes to your app dependencies...",
						"  + dep_added1@v1",
						"  dep_modified1@v2 -> dep_modified1@v1",
						"  dep_modified2@v1 -> dep_modified2@v2",
						"With changes to your GraphQL configuration...",
						"  gql_field1: previous -> updated",
						"With changes to your app schema...",
						"  gql_validation_field1: old -> new",
						"  rest_validation_field1: old -> new",
						"",
					},
					"\n",
				),
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				var realmClient mock.RealmClient
				realmClient.DiffDraftFn = func(groupID, appID, draftID string) (realm.AppDraftDiff, error) {
					return tc.actualDiff, nil
				}

				out, ui := mock.NewUI()

				assert.Nil(t, diffDraft(ui, realmClient, appRemote{}, ""))
				assert.Equal(t, tc.expectedContents, out.String())
			})
		}
	})
}

func TestPushCommandDeployDraftAndWait(t *testing.T) {
	groupID, appID, clientAppID, draftID := "groupID", "appID", "client-app-id", "draftID"
	t.Run("should return an error with a client that fails to deploy the draft", func(t *testing.T) {
		realmClient := mock.RealmClient{}

		var capturedGroupID, capturedAppID, capturedDraftID string
		realmClient.DeployDraftFn = func(groupID, appID, draftID string) (realm.AppDeployment, error) {
			capturedGroupID = groupID
			capturedAppID = appID
			capturedDraftID = draftID
			return realm.AppDeployment{}, errors.New("something bad happened")
		}

		err := deployDraftAndWait(nil, realmClient, appRemote{groupID, appID, clientAppID}, draftID)
		assert.Equal(t, errors.New("something bad happened"), err)

		t.Log("and should properly pass through the expected inputs")
		assert.Equal(t, groupID, capturedGroupID)
		assert.Equal(t, appID, capturedAppID)
		assert.Equal(t, draftID, capturedDraftID)
	})

	t.Run("with a client that successfully deploys a draft", func(t *testing.T) {
		realmClient := mock.RealmClient{}
		realmClient.DeployDraftFn = func(groupID, appID, draftID string) (realm.AppDeployment, error) {
			return realm.AppDeployment{ID: "id", Status: realm.DeploymentStatusCreated}, nil
		}

		t.Run("but fails to get the deployment should return the error", func(t *testing.T) {
			realmClient.DeploymentFn = func(groupID, appID, deploymentID string) (realm.AppDeployment, error) {
				return realm.AppDeployment{}, errors.New("something bad happened")
			}

			_, ui := mock.NewUI()

			err := deployDraftAndWait(ui, realmClient, appRemote{groupID, appID, clientAppID}, draftID)
			assert.Equal(t, errors.New("something bad happened"), err)
		})

		t.Run("and successfully retrieves the deployment should eventually succeed", func(t *testing.T) {
			var polls int

			realmClient.DeploymentFn = func(groupID, appID, deploymentID string) (realm.AppDeployment, error) {
				status := realm.DeploymentStatusPending
				if polls > 1 {
					status = realm.DeploymentStatusSuccessful
				}
				polls++
				return realm.AppDeployment{ID: deploymentID, Status: status}, nil
			}

			out, ui := mock.NewUI()

			err := deployDraftAndWait(ui, realmClient, appRemote{groupID, appID, clientAppID}, draftID)
			assert.Nil(t, err)

			assert.Equal(t, "Deployment complete\n", out.String())
		})
	})
}

func TestPushCommandDisplay(t *testing.T) {
	for _, tc := range []struct {
		description string
		inputs      inputs
		omitDryRun  bool
		display     string
	}{
		{
			description: "should print a minimal command string",
			display:     "realm-cli push",
		},
		{
			description: "should print a minimal dry run command string",
			inputs:      inputs{DryRun: true},
			display:     "realm-cli push --dry-run",
		},
		{
			description: "should print a minimal command string with dry run set but omitted",
			inputs:      inputs{DryRun: true},
			omitDryRun:  true,
			display:     "realm-cli push",
		},
		{
			description: "should print a complete command string",
			inputs: inputs{
				Project:            "project",
				LocalPath:          "directory",
				RemoteApp:          "remote",
				IncludeNodeModules: true,
				IncludeHosting:     true,
				ResetCDNCache:      true,
				DryRun:             true,
			},
			display: "realm-cli push --project project --local directory --remote remote --include-node-modules --include-hosting --reset-cdn-cache --dry-run",
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			cmd := &Command{tc.inputs}
			assert.Equal(t, tc.display, cmd.display(tc.omitDryRun))
		})
	}
}

func runImport(t *testing.T, realmClient realm.Client, appDirectory string) {
	t.Helper()

	out := new(bytes.Buffer)
	ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

	cmd := &Command{inputs{LocalPath: appDirectory, RemoteApp: "appID"}}

	assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
	assert.Equal(t, `Determining changes
Creating draft
Pushing changes
Deploying draft
Deployment complete
Successfully pushed app up: eggcorn-abcde
`, out.String())
}
