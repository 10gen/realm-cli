package pull

import (
	"archive/zip"
	"bytes"
	"errors"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestPullInputsResolve(t *testing.T) {
	t.Run("should not return an error if run from outside a project directory with no to flag is set", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "pull_input_test")
		defer teardown()

		var i inputs
		assert.Nil(t, i.Resolve(profile, nil))
	})

	t.Run("when run inside a project directory", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "pull_input_test")
		defer teardown()

		assert.Nil(t, ioutil.WriteFile(
			filepath.Join(profile.WorkingDirectory, local.FileRealmConfig.String()),
			[]byte(`{"config_version":20210101,"app_id":"eggcorn-abcde","name":"eggcorn"}`),
			0666,
		))

		t.Run("should set inputs from app if no flags are set", func(t *testing.T) {
			var i inputs
			assert.Nil(t, i.Resolve(profile, nil))

			assert.Equal(t, profile.WorkingDirectory, i.LocalPath)
			assert.Equal(t, "eggcorn-abcde", i.RemoteApp)
			assert.Equal(t, realm.AppConfigVersion20210101, i.AppVersion)
		})

		t.Run("should return an error if app version flag is different from the project value", func(t *testing.T) {
			i := inputs{AppVersion: realm.AppConfigVersion20200603}
			assert.Equal(t, errConfigVersionMismatch, i.Resolve(profile, nil))
		})
	})

	t.Run("resolving the to flag should work", func(t *testing.T) {
		homeDir, teardown := u.SetupHomeDir("")
		defer teardown()

		for _, tc := range []struct {
			description    string
			targetFlag     string
			expectedTarget string
		}{
			{
				description:    "should expand the to flag to include the user home directory",
				targetFlag:     "~/my/project/root",
				expectedTarget: filepath.Join(homeDir, "my/project/root"),
			},
			{
				description:    "should resolve the to flag to account for relative paths",
				targetFlag:     "../../cmd",
				expectedTarget: filepath.Join(homeDir, "../../cmd"),
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				profile := mock.NewProfile(t)

				i := inputs{LocalPath: tc.targetFlag}
				assert.Nil(t, i.Resolve(profile, nil))

				assert.Equal(t, tc.expectedTarget, i.LocalPath)
			})
		}
	})
}

func TestPullInputsResolveRemoteApp(t *testing.T) {
	t.Run("should not resolve group id if project is provided", func(t *testing.T) {
		i := inputs{Project: "some-project", RemoteApp: "some-app"}

		var realmClient mock.RealmClient

		var appFilter realm.AppFilter
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			appFilter = filter
			return []realm.App{{GroupID: "group-id", ID: "app-id"}}, nil
		}

		app, err := i.resolveRemoteApp(nil, cli.Clients{Realm: realmClient})
		assert.Nil(t, err)
		assert.Equal(t, realm.App{GroupID: "group-id", ID: "app-id"}, app)

		assert.Equal(t, realm.AppFilter{GroupID: "some-project", App: "some-app"}, appFilter)
	})

	t.Run("should resolve group id if project is not provided", func(t *testing.T) {
		i := inputs{RemoteApp: "some-app"}

		var atlasClient mock.AtlasClient
		atlasClient.GroupsFn = func() ([]atlas.Group, error) {
			return []atlas.Group{{ID: "group-id", Name: "group-name"}}, nil
		}

		var realmClient mock.RealmClient

		var appFilter realm.AppFilter
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			appFilter = filter
			return []realm.App{{GroupID: "group-id", ID: "app-id"}}, nil
		}

		app, err := i.resolveRemoteApp(nil, cli.Clients{Atlas: atlasClient, Realm: realmClient})
		assert.Nil(t, err)
		assert.Equal(t, realm.App{GroupID: "group-id", ID: "app-id"}, app)

		assert.Equal(t, realm.AppFilter{GroupID: "group-id", App: "some-app"}, appFilter)
	})

	t.Run("should return a project not found error if the app is not found", func(t *testing.T) {
		i := inputs{Project: "some-project"}

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return nil, cli.ErrAppNotFound{}
		}

		_, err := i.resolveRemoteApp(nil, cli.Clients{Realm: realmClient})
		assert.Equal(t, errProjectNotFound{}, err)
	})

	t.Run("should return an error when the atlas client fails to find groups", func(t *testing.T) {
		var i inputs

		var atlasClient mock.AtlasClient
		atlasClient.GroupsFn = func() ([]atlas.Group, error) {
			return nil, errors.New("something bad happened")
		}

		_, err := i.resolveRemoteApp(nil, cli.Clients{Atlas: atlasClient})
		assert.Equal(t, errors.New("something bad happened"), err)
	})

	t.Run("should return an error when the realm client fails to find an app", func(t *testing.T) {
		i := inputs{Project: "some-project"}

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return nil, errors.New("something bad happened")
		}

		_, err := i.resolveRemoteApp(nil, cli.Clients{Realm: realmClient})
		assert.Equal(t, errors.New("something bad happened"), err)
	})
}

func TestPullTemplatesResolve(t *testing.T) {
	t.Run("should not do anything if there is no templateID passed in", func(t *testing.T) {
		i := inputs{Project: "some-project", RemoteApp: "some-app"}
		var realmClient mock.RealmClient
		templateZipPkg, paths, err := i.resolveTemplate(nil, realmClient, "group-id", "app-id")
		assert.Nil(t, err)
		assert.Nil(t, paths)
		assert.Nil(t, templateZipPkg)
	})
	t.Run("when there is a templateID passed in", func(t *testing.T) {
		t.Run("should use frontendPath by default if autoConfirm is on", func(t *testing.T) {
			i := inputs{TemplateID: "some-template-id"}
			var realmClient mock.RealmClient
			realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, error) {
				return []realm.Template{
					{"some-template-id", "the name of some template id"},
					{"some-other-template-id", "the name of the other"},
				}, nil
			}

			expectedZipPkg, err := zip.OpenReader("testdata/template.zip")
			assert.Nil(t, err)
			defer func() {
				err := expectedZipPkg.Close()
				assert.Nil(t, err)
			}()

			realmClient.ClientTemplateFn = func(groupID, appID, templateID string) (*zip.Reader, error) {
				return &expectedZipPkg.Reader, err
			}

			out := new(bytes.Buffer)
			ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

			templateZipPkg, paths, err := i.resolveTemplate(ui, realmClient, "group-id", "app-id")
			assert.Nil(t, err)
			assert.Equal(t, []string{frontendPath}, paths)
			compareZipPackages(t, expectedZipPkg, templateZipPkg)

		})
		for _, tc := range []struct {
			pathDescription string
			chosenPaths     []string
		}{
			{
				pathDescription: "with the frontend path selected",
				chosenPaths:     []string{frontendPath},
			},
			{
				pathDescription: "with the backend path selected",
				chosenPaths:     []string{backendPath},
			},
			{
				pathDescription: "with both paths selected",
				chosenPaths:     []string{frontendPath, backendPath},
			},
		} {
			t.Run(tc.pathDescription, func(t *testing.T) {
				t.Run("should return the zip package and the selected path(s) when the template is compatible with the app's template", func(t *testing.T) {
					i := inputs{TemplateID: "some-template-id"}
					var realmClient mock.RealmClient
					realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, error) {
						return []realm.Template{
							{"some-template-id", "the name of some template id"},
							{"some-other-template-id", "the name of the other"},
						}, nil
					}

					expectedZipPkg, err := zip.OpenReader("testdata/template.zip")
					assert.Nil(t, err)
					defer func() {
						err := expectedZipPkg.Close()
						assert.Nil(t, err)
					}()

					realmClient.ClientTemplateFn = func(groupID, appID, templateID string) (*zip.Reader, error) {
						return &expectedZipPkg.Reader, err
					}

					_, console, _, ui, consoleErr := mock.NewVT10XConsole()
					assert.Nil(t, consoleErr)
					defer console.Close()

					doneCh := make(chan struct{})
					go func() {
						defer close(doneCh)
						console.ExpectString("Where would you like to export the template?")
						for _, path := range tc.chosenPaths {
							console.Send(path)
							console.Send(" ")
						}
						console.SendLine("")
						console.ExpectEOF()
					}()

					templateZipPkg, paths, err := i.resolveTemplate(ui, realmClient, "group-id", "app-id")
					assert.Nil(t, err)
					assert.Equal(t, tc.chosenPaths, paths)
					compareZipPackages(t, expectedZipPkg, templateZipPkg)
				})
			})
		}

		t.Run("should return nil if the requested template has no client", func(t *testing.T) {
			i := inputs{TemplateID: "some-template-id"}
			var realmClient mock.RealmClient
			realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, error) {
				return []realm.Template{
					{"some-template-id", "the name of some template id"},
					{"some-other-template-id", "the name of the other"},
				}, nil
			}
			realmClient.ClientTemplateFn = func(groupID, appID, templateID string) (*zip.Reader, error) {
				return nil, nil
			}

			out := new(bytes.Buffer)
			ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

			templateZipPkg, paths, err := i.resolveTemplate(ui, realmClient, "group-id", "app-id")
			assert.Nil(t, err)
			assert.Nil(t, templateZipPkg)
			assert.Equal(t, []string{frontendPath}, paths)
		})

		t.Run("should return nil if there's an issue with fetching compatible templates", func(t *testing.T) {
			i := inputs{TemplateID: "some-template-id"}
			var realmClient mock.RealmClient
			realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, error) {
				return nil, errors.New("issue with fetching templates")
			}

			out := new(bytes.Buffer)
			ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

			templateZipPkg, paths, err := i.resolveTemplate(ui, realmClient, "group-id", "app-id")
			assert.Nil(t, templateZipPkg)
			assert.Nil(t, paths)
			assert.Equal(t, "issue with fetching templates", err.Error())
		})

		t.Run("should return nil if there's an issue with fetching the client for the requested template", func(t *testing.T) {
			i := inputs{TemplateID: "some-template-id"}
			var realmClient mock.RealmClient
			realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, error) {
				return []realm.Template{
					{"some-template-id", "some-template-name"},
				}, nil
			}
			realmClient.ClientTemplateFn = func(groupID, appID, templateID string) (*zip.Reader, error) {
				return nil, errors.New("issue with fetching client")
			}

			out := new(bytes.Buffer)
			ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

			templateZipPkg, paths, err := i.resolveTemplate(ui, realmClient, "group-id", "app-id")
			assert.Nil(t, templateZipPkg)
			assert.Nil(t, paths)
			assert.Equal(t, "issue with fetching client", err.Error())
		})
		t.Run("should return nil if the requested template is not compatible with the app", func(t *testing.T) {
			i := inputs{TemplateID: "some-requested-template-id"}
			var realmClient mock.RealmClient
			realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, error) {
				return []realm.Template{
					{"some-other-template-id", "some-template-name"},
				}, nil
			}

			out := new(bytes.Buffer)
			ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

			templateZipPkg, paths, err := i.resolveTemplate(ui, realmClient, "group-id", "app-id")
			assert.Nil(t, templateZipPkg)
			assert.Nil(t, paths)
			assert.Equal(t, "templateID some-requested-template-id is not compatible with this app", err.Error())
		})
	})
}

func compareZipPackages(t *testing.T, expectedZipPkg *zip.ReadCloser, actualZipPkg *zip.Reader) {
	expected := make(map[string]*zip.File, len(expectedZipPkg.File))
	for _, f := range actualZipPkg.File {
		expected[f.Name] = f
	}
	for _, f := range actualZipPkg.File {
		assert.Equal(t, expected[f.Name].FileInfo().Name(), f.FileInfo().Name())
		assert.Equal(t, expected[f.Name].FileInfo().ModTime(), f.FileInfo().ModTime())
		assert.Equal(t, expected[f.Name].FileInfo().Size(), f.FileInfo().Size())
	}
}
