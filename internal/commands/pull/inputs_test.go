package pull

import (
	"archive/zip"
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
	templateZipPkg1, err := zip.OpenReader("testdata/template_1.zip")
	assert.Nil(t, err)
	defer templateZipPkg1.Close()

	templateZipPkg2, err := zip.OpenReader("testdata/template_2.zip")
	assert.Nil(t, err)
	defer templateZipPkg2.Close()

	t.Run("should not do anything if fetching for compatible templates errors", func(t *testing.T) {
		var realmClient mock.RealmClient

		realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, error) {
			return nil, errors.New("something went wrong")
		}

		_, ui := mock.NewUI()
		input := inputs{}
		_, err := input.resolveClientTemplates(ui, realmClient, "some-group-id", "some-app-id")
		assert.Equal(t, errors.New("something went wrong"), err)
	})

	t.Run("with a template id passed in", func(t *testing.T) {
		t.Run("should return an error if fetching the client template for this template fails", func(t *testing.T) {
			var realmClient mock.RealmClient

			realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, error) {
				return []realm.Template{{ID: "some-template-id", Name: "should not be there"}}, nil
			}
			realmClient.ClientTemplateFn = func(groupID, appID, templateID string) (*zip.Reader, bool, error) {
				return nil, false, errors.New("something went wrong with client")
			}
			_, ui := mock.NewUI()
			input := inputs{TemplateID: "some-template-id"}
			_, err := input.resolveClientTemplates(ui, realmClient, "some-group-id", "some-app-id")
			assert.Equal(t, errors.New("something went wrong with client"), err)
		})

		t.Run("should return an error if the template id is not compatible with the app", func(t *testing.T) {
			var realmClient mock.RealmClient

			realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, error) {
				return []realm.Template{{ID: "some-template-id", Name: "some template name"}}, nil
			}

			_, ui := mock.NewUI()
			input := inputs{TemplateID: "wrong-template-id"}
			_, err := input.resolveClientTemplates(ui, realmClient, "some-group-id", "some-app-id")
			assert.Equal(t, errors.New("template 'wrong-template-id' is not compatible with this app"), err)
		})

		t.Run("should return the template client if it is compatible with this app", func(t *testing.T) {
			var realmClient mock.RealmClient

			realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, error) {
				return []realm.Template{{ID: "some-template-id", Name: "some name"}, {ID: "some-other-template-id", Name: "some other name"}}, nil
			}
			realmClient.ClientTemplateFn = func(groupID, appID, templateID string) (*zip.Reader, bool, error) {
				return &templateZipPkg1.Reader, true, nil
			}

			_, ui := mock.NewUI()
			input := inputs{TemplateID: "some-template-id"}
			result, err := input.resolveClientTemplates(ui, realmClient, "some-group-id", "some-app")
			assert.Nil(t, err)
			assert.Equal(t, 1, len(result))
			compareZipPackages(t, templateZipPkg1, result["some-template-id"])
		})
	})

	t.Run("without a template id passed in", func(t *testing.T) {
		input := inputs{}
		t.Run("should return nothing if the app is not made with a template", func(t *testing.T) {
			var realmClient mock.RealmClient

			realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, error) {
				return nil, nil
			}
			_, ui := mock.NewUI()
			result, err := input.resolveClientTemplates(ui, realmClient, "some-group-id", "some-app-id")
			assert.Nil(t, err)
			assert.Equal(t, 0, len(result))

		})

		t.Run("should return nothing if the user refuses to export with compatible templates", func(t *testing.T) {
			var realmClient mock.RealmClient
			realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, error) {
				return []realm.Template{{ID: "some-template-id", Name: "some name"}}, nil
			}

			_, console, _, ui, err := mock.NewVT10XConsole()
			assert.Nil(t, err)
			defer console.Close()

			doneCh := make(chan struct{})
			go func() {
				defer close(doneCh)
				console.ExpectString("Would you like to export with a template?")
				console.SendLine("n")
				console.ExpectEOF()
			}()

			result, err := input.resolveClientTemplates(ui, realmClient, "some-group-id", "some-group-app")
			assert.Nil(t, err)
			assert.Equal(t, 0, len(result))
		})

		t.Run("should prompt the user to export with compatible templates and return the templates selected", func(t *testing.T) {
			var realmClient mock.RealmClient
			realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, error) {
				return []realm.Template{{ID: "some-template-id", Name: "some name"}, {ID: "some-other-id", Name: "some other name"}, {ID: "another-template-id", Name: "another name"}}, nil
			}

			realmClient.ClientTemplateFn = func(groupID, appID, templateID string) (*zip.Reader, bool, error) {
				if templateID == "some-template-id" {
					return &templateZipPkg1.Reader, true, nil
				} else if templateID == "another-template-id" {
					return &templateZipPkg2.Reader, true, nil
				}
				return nil, false, errors.New("should not happen in this test")
			}

			_, console, _, ui, err := mock.NewVT10XConsole()
			assert.Nil(t, err)
			defer console.Close()

			doneCh := make(chan struct{})
			go func() {
				defer close(doneCh)
				console.ExpectString("Would you like to export with a template?")
				console.SendLine("y")
				console.ExpectString("Which template(s) would you like to export this app with")
				console.Send("some name")
				console.Send(" ")
				console.Send("another name")
				console.Send(" ")
				console.SendLine("")
				console.ExpectEOF()
			}()

			result, err := input.resolveClientTemplates(ui, realmClient, "some-group-id", "some-group-app")
			assert.Nil(t, err)
			assert.Equal(t, 2, len(result))
			expected := map[string]*zip.ReadCloser{
				"some-template-id":    templateZipPkg1,
				"another-template-id": templateZipPkg2,
			}
			for templateID, zipPkg := range result {
				compareZipPackages(t, expected[templateID], zipPkg)
			}
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
