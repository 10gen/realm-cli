package pull

import (
	"archive/zip"
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

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestPullHandler(t *testing.T) {
	t.Run("should return an error if the command fails to resolve from", func(t *testing.T) {
		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return nil, errors.New("something bad happened")
		}

		cmd := &Command{inputs{Project: "elsewhere", RemoteApp: "somewhere"}}

		err := cmd.Handler(nil, nil, cli.Clients{Realm: realmClient})
		assert.Equal(t, errors.New("something bad happened"), err)
	})

	t.Run("should return an error if the command fails to do the export", func(t *testing.T) {
		_, ui := mock.NewUI()

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", Name: "appName"}}, nil
		}
		realmClient.ExportFn = func(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error) {
			return "", nil, errors.New("something bad happened")
		}
		realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, bool, error) {
			return nil, false, nil
		}
		cmd := &Command{inputs{Project: "elsewhere", RemoteApp: "somewhere"}}

		err := cmd.Handler(nil, ui, cli.Clients{Realm: realmClient})
		assert.Equal(t, errors.New("something bad happened"), err)
	})

	t.Run("with a successful export", func(t *testing.T) {
		zipPkg, zipErr := zip.OpenReader("testdata/test.zip")
		assert.Nil(t, zipErr)
		defer zipPkg.Close()

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", Name: "appName"}}, nil
		}
		realmClient.ExportFn = func(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error) {
			return "app_20210101", &zipPkg.Reader, nil
		}
		realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, bool, error) {
			return nil, false, nil
		}

		t.Run("should not write any contents to the destination in a dry run", func(t *testing.T) {
			profile := mock.NewProfile(t)

			out, ui := mock.NewUI()

			cmd := &Command{inputs{Project: "elsewhere", DryRun: true, LocalPath: "app"}}

			assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: realmClient}))
			destination := filepath.Join(profile.WorkingDirectory, "app")

			assert.Equal(t, `No changes were written to your file system
Contents would have been written to: app
`, out.String())

			_, err := os.Stat(destination)
			assert.True(t, os.IsNotExist(err), "expected %s to not exist, but instead: %s", err)
		})

		t.Run("should write the received zip package to the destination", func(t *testing.T) {
			profile, teardown := mock.NewProfileFromTmpDir(t, "pull_handler_test")
			defer teardown()

			out, ui := mock.NewUI()

			cmd := &Command{inputs{Project: "elsewhere", LocalPath: "app"}}

			assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: realmClient}))
			destination := filepath.Join(profile.WorkingDirectory, "app")

			assert.Equal(t, `Saved app to disk
Successfully pulled app down: app
`, out.String())

			_, err := os.Stat(destination)
			assert.Nil(t, err)

			testData, readErr := ioutil.ReadFile(filepath.Join(destination, "test.json"))
			assert.Nil(t, readErr)
			assert.Equal(t, `{"egg":"corn"}
`, string(testData))
		})
	})

	t.Run("with a realm client that fails to export dependencies", func(t *testing.T) {
		zipPkg, zipErr := zip.OpenReader("testdata/test.zip")
		assert.Nil(t, zipErr)
		defer zipPkg.Close()

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", Name: "appName"}}, nil
		}
		realmClient.ExportFn = func(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error) {
			return "app_20210101", &zipPkg.Reader, nil
		}
		realmClient.ExportDependenciesFn = func(groupID, appID string) (string, io.ReadCloser, error) {
			return "", nil, errors.New("something bad happened")
		}
		realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, bool, error) {
			return nil, false, nil
		}

		t.Run("should not attempt to export dependencies if the flag is not set", func(t *testing.T) {
			profile, teardown := mock.NewProfileFromTmpDir(t, "pull_handler_test")
			defer teardown()

			out, ui := mock.NewUI()

			cmd := &Command{inputs{Project: "elsewhere", LocalPath: "app"}}

			assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: realmClient}))

			assert.Equal(t, `Saved app to disk
Successfully pulled app down: app
`, out.String())
		})

		t.Run("should return the error when exporting dependencies fails", func(t *testing.T) {
			profile, teardown := mock.NewProfileFromTmpDir(t, "pull_handler_test")
			defer teardown()

			_, ui := mock.NewUI()

			cmd := &Command{inputs{Project: "elsewhere", LocalPath: "app", IncludeDependencies: true}}

			err := cmd.Handler(profile, ui, cli.Clients{Realm: realmClient})
			assert.Equal(t, errors.New("something bad happened"), err)
		})
	})

	t.Run("with a realm client that successfully exports dependencies should write the archive file", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "pull_handler_test")
		defer teardown()

		out, ui := mock.NewUI()

		zipPkg, zipErr := zip.OpenReader("testdata/test.zip")
		assert.Nil(t, zipErr)
		defer zipPkg.Close()

		depsPkg, err := os.Open("testdata/node_modules.zip")
		assert.Nil(t, err)
		defer depsPkg.Close()

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", Name: "appName"}}, nil
		}
		realmClient.ExportFn = func(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error) {
			return "app_20210101", &zipPkg.Reader, nil
		}
		realmClient.ExportDependenciesFn = func(groupID, appID string) (string, io.ReadCloser, error) {
			return "node_modules.zip", depsPkg, nil
		}
		realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, bool, error) {
			return nil, false, nil
		}

		cmd := &Command{inputs{Project: "elsewhere", LocalPath: "app", IncludeDependencies: true}}

		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: realmClient}))
		assert.Equal(t, `Saved app to disk
Fetched dependencies archive
Successfully pulled app down: app
`, out.String())

		_, pkgErr := os.Stat(filepath.Join(profile.WorkingDirectory, "app", local.NameFunctions, "node_modules.zip"))
		assert.Nil(t, pkgErr)
	})

	t.Run("with a realm client that fails to get hosting assets", func(t *testing.T) {
		zipPkg, zipErr := zip.OpenReader("testdata/test.zip")
		assert.Nil(t, zipErr)
		defer zipPkg.Close()

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", Name: "appName"}}, nil
		}
		realmClient.ExportFn = func(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error) {
			return "app_20210101", &zipPkg.Reader, nil
		}
		realmClient.HostingAssetsFn = func(groupID, appID string) ([]realm.HostingAsset, error) {
			return nil, errors.New("something bad happened")
		}
		realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, bool, error) {
			return nil, false, nil
		}

		t.Run("should not attempt to export hosting assets if the flag is not set", func(t *testing.T) {
			profile, teardown := mock.NewProfileFromTmpDir(t, "pull_handler_test")
			defer teardown()

			out, ui := mock.NewUI()

			cmd := &Command{inputs{Project: "elsewhere", LocalPath: "app"}}

			assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: realmClient}))

			assert.Equal(t, `Saved app to disk
Successfully pulled app down: app
`, out.String())
		})

		t.Run("should return the error when getting hosting assets fails", func(t *testing.T) {
			profile, teardown := mock.NewProfileFromTmpDir(t, "pull_handler_test")
			defer teardown()

			_, ui := mock.NewUI()

			cmd := &Command{inputs{Project: "elsewhere", LocalPath: "app", IncludeHosting: true}}

			err := cmd.Handler(profile, ui, cli.Clients{Realm: realmClient})
			assert.Equal(t, errors.New("something bad happened"), err)
		})
	})

	t.Run("with a realm client that successfully gets hosting assets should write the hosting files", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "pull_handler_test")
		defer teardown()

		filesDir := filepath.Join(profile.WorkingDirectory, "app", local.NameHosting, local.NameFiles)
		assert.Nil(t, os.MkdirAll(filesDir, os.ModePerm))

		t.Log("create an existing hosting asset to be left alone")
		assert.Nil(t, ioutil.WriteFile(
			filepath.Join(profile.WorkingDirectory, "app", local.NameHosting, local.NameFiles, "removed.html"),
			[]byte("<html><body>i do not belong here</body></html>"),
			0666,
		))

		t.Log("create an existing hosting asset to be modified")
		assert.Nil(t, ioutil.WriteFile(
			filepath.Join(profile.WorkingDirectory, "app", local.NameHosting, local.NameFiles, "modified.html"),
			[]byte("<html><body>i should be something else</body></html>"),
			0666,
		))

		out := new(bytes.Buffer)
		ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

		zipPkg, zipErr := zip.OpenReader("testdata/test.zip")
		assert.Nil(t, zipErr)
		defer zipPkg.Close()

		var realmClient mock.RealmClient
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", Name: "appName"}}, nil
		}
		realmClient.ExportFn = func(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error) {
			return "app_20210101", &zipPkg.Reader, nil
		}
		realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, bool, error) {
			return nil, false, nil
		}
		realmClient.HostingAssetsFn = func(groupID, appID string) ([]realm.HostingAsset, error) {
			return []realm.HostingAsset{
				{
					HostingAssetData: realm.HostingAssetData{
						FilePath: "/index.html",
						FileHash: "9163ebc83aa75cae0a7e74b4e16af317",
						FileSize: 51,
					},
					Attrs: nil,
					URL:   "http://url.com/index.html",
				},
				{
					HostingAssetData: realm.HostingAssetData{
						FilePath: "/modified.html",
						FileHash: "9163ebc83aa75cae0a7e74b4e16af317",
						FileSize: 51,
					},
					Attrs: nil,
					URL:   "http://url.com/modified.html",
				},
			}, nil
		}

		hostingAssetClient := mockHostingAssetClient{"<html><body>hello world!</body></html>"}

		cmd := &Command{inputs{Project: "elsewhere", LocalPath: "app", IncludeHosting: true}}

		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: realmClient, HostingAsset: hostingAssetClient}))
		assert.Equal(t, `Saved app to disk
Fetched hosting assets
Successfully pulled app down: app
`, out.String())

		t.Log("should have added the new file")
		filename := filepath.Join(profile.WorkingDirectory, "app", local.NameHosting, local.NameFiles, "index.html")
		newData, err := ioutil.ReadFile(filename)
		assert.Nil(t, err)
		assert.Equal(t, "<html><body>hello world!</body></html>", string(newData))
		fileInfo, err := os.Stat(filename)
		assert.Nil(t, err)
		assert.Equal(t, fileInfo.Mode().String(), "-rw-r--r--")

		t.Log("should have preserved the existing file not found on the app")
		_, err = os.Stat(filepath.Join(profile.WorkingDirectory, "app", local.NameHosting, local.NameFiles, "removed.html"))
		assert.Nil(t, err)

		t.Log("should have modified the existing file found on the app")
		modifiedData, err := ioutil.ReadFile(filepath.Join(profile.WorkingDirectory, "app", local.NameHosting, local.NameFiles, "modified.html"))
		assert.Nil(t, err)
		assert.Equal(t, "<html><body>hello world!</body></html>", string(modifiedData))
	})

	t.Run("with a template id to export with", func(t *testing.T) {
		// Needs to have a successful export for templates
		zipPkg, err := zip.OpenReader("testdata/test.zip")
		assert.Nil(t, err)
		defer zipPkg.Close()

		var realmClient mock.RealmClient

		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "appID", Name: "appName"}}, nil
		}
		realmClient.ExportFn = func(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error) {
			return "app_20210101", &zipPkg.Reader, nil
		}

		templateZipPkg1, err := zip.OpenReader("testdata/template_1.zip")
		assert.Nil(t, err)
		defer templateZipPkg1.Close()

		templateZipPkg2, err := zip.OpenReader("testdata/template_2.zip")
		assert.Nil(t, err)
		defer templateZipPkg2.Close()

		t.Run("should export the corresponding template when the template id is passed in", func(t *testing.T) {
			templateID := "template_2"
			templateZip := &templateZipPkg2.Reader
			expectedTemplate := `{
  "egg": "scrambled",
  "pancake": "plain",
  "toast": "none"
}`
			realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, bool, error) {
				return []realm.Template{{ID: templateID, Name: "some template"}}, true, nil
			}

			realmClient.ClientTemplateFn = func(groupID, appID, templateID string) (*zip.Reader, error) {
				return templateZip, nil
			}

			profile, teardown := mock.NewProfileFromTmpDir(t, "profile_name")
			defer teardown()

			_, ui := mock.NewUI()

			cmd := &Command{inputs{Project: "some_project", LocalPath: "app", TemplateID: templateID}}
			assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: realmClient}))

			destination := filepath.Join(profile.WorkingDirectory, "app", local.FrontendPath)

			_, err := os.Stat(destination)
			assert.Nil(t, err)

			testData, readErr := ioutil.ReadFile(filepath.Join(destination, fmt.Sprintf("%s.json", templateID)))
			assert.Nil(t, readErr)
			assert.Equal(t, expectedTemplate, string(testData))
		})

		t.Run("should export all selected compatible templates if no template id is passed in and the app is made with a template and autoConfirm is true", func(t *testing.T) {
			input := inputs{Project: "some_project", LocalPath: "app"}

			realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, bool, error) {
				return []realm.Template{{ID: "template_1", Name: "Template 1"}, {ID: "template_2", Name: "Template 2"}}, true, nil
			}
			realmClient.ClientTemplateFn = func(groupID, appID, templateID string) (*zip.Reader, error) {
				if templateID == "template_1" {
					return &templateZipPkg1.Reader, nil
				}
				if templateID == "template_2" {
					return &templateZipPkg2.Reader, nil
				}
				return nil, nil
			}

			profile, teardown := mock.NewProfileFromTmpDir(t, "profile_name")
			defer teardown()

			console, _, ui, err := mock.NewVT10XConsoleWithOptions(mock.UIOptions{AutoConfirm: true})
			assert.Nil(t, err)
			defer console.Close()

			doneCh := make(chan struct{})
			go func() {
				defer close(doneCh)
				console.ExpectString("Which template(s) would you like to export this app with")
				console.Send("Template 1")
				console.Send(" ")
				console.Send("Template 2")
				console.Send(" ")
				console.SendLine("")
				console.ExpectEOF()
			}()

			cmd := &Command{input}
			assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: realmClient}))
			expected := map[string]string{
				"template_1": `{
  "egg": "over easy",
  "pancake": "blueberry",
  "toast": "french"
}`,
				"template_2": `{
  "egg": "scrambled",
  "pancake": "plain",
  "toast": "none"
}`,
			}
			for templateID, expectedString := range expected {
				destination := filepath.Join(profile.WorkingDirectory, "app", local.FrontendPath)

				_, err = os.Stat(destination)
				assert.Nil(t, err)

				testData, readErr := ioutil.ReadFile(filepath.Join(destination, fmt.Sprintf("%s.json", templateID)))
				assert.Nil(t, readErr)
				assert.Equal(t, expectedString, string(testData))
			}
		})

		t.Run("should return an error if resolving the template returns an error", func(t *testing.T) {
			realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, bool, error) {
				return nil, false, errors.New("some kind of error")
			}

			profile, teardown := mock.NewProfileFromTmpDir(t, "should_fail")
			defer teardown()

			out := new(bytes.Buffer)
			ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

			cmd := &Command{inputs{Project: "elsewhere", LocalPath: "app", TemplateID: "some-template-id"}}

			err := cmd.Handler(profile, ui, cli.Clients{Realm: realmClient})
			assert.Equal(t, errors.New("some kind of error"), err)
		})

		t.Run("should return an error if the template id is not compatible with the app", func(t *testing.T) {
			realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, bool, error) {
				return []realm.Template{{ID: "wrong_id", Name: "should not appear"}}, true, nil
			}

			profile, teardown := mock.NewProfileFromTmpDir(t, "should_fail_also")
			defer teardown()

			out := new(bytes.Buffer)
			ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

			cmd := &Command{inputs{Project: "elsewhere", LocalPath: "app", TemplateID: "some-template-id"}}

			err := cmd.Handler(profile, ui, cli.Clients{Realm: realmClient})
			assert.Equal(t, errors.New("template 'some-template-id' is not compatible with this app"), err)
		})

		t.Run("should return nothing and continue exporting the app if the app is not made with a template", func(t *testing.T) {
			realmClient.CompatibleTemplatesFn = func(groupID, appID string) ([]realm.Template, bool, error) {
				return nil, false, nil
			}

			profile, teardown := mock.NewProfileFromTmpDir(t, "should_fail_also")
			defer teardown()

			out := new(bytes.Buffer)
			ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

			cmd := &Command{inputs{Project: "elsewhere", LocalPath: "app"}}

			err := cmd.Handler(profile, ui, cli.Clients{Realm: realmClient})
			assert.Nil(t, err)

			// Check to see that export went as normal
			destination := filepath.Join(profile.WorkingDirectory, "app")

			_, err = os.Stat(destination)
			assert.Nil(t, err)

			testData, readErr := ioutil.ReadFile(filepath.Join(destination, "test.json"))
			assert.Nil(t, readErr)
			assert.Equal(t, `{"egg":"corn"}
`, string(testData))
		})
	})
}

func TestPullCommandDoExport(t *testing.T) {
	t.Run("should return an error if the export fails", func(t *testing.T) {
		groupID, appID := "groupID", "appID"

		var realmClient mock.RealmClient

		var capturedGroupID, capturedAppID string
		var capturedExportReq realm.ExportRequest
		realmClient.ExportFn = func(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error) {
			capturedGroupID = groupID
			capturedAppID = appID
			capturedExportReq = req
			return "", nil, errors.New("something bad happened")
		}

		cmd := &Command{inputs{AppVersion: realm.AppConfigVersion20210101}}

		_, _, err := cmd.doExport(nil, realmClient, groupID, appID)
		assert.Equal(t, errors.New("something bad happened"), err)

		t.Log("and should properly pass through the expected args")
		assert.Equal(t, groupID, capturedGroupID)
		assert.Equal(t, appID, capturedAppID)
		assert.Equal(t, realm.ExportRequest{ConfigVersion: realm.AppConfigVersion20210101}, capturedExportReq)
	})

	t.Run("should return the expected destination file path", func(t *testing.T) {
		profile := mock.NewProfile(t)
		profile.WorkingDirectory = "/some/system/path"

		for _, tc := range []struct {
			description  string
			flagLocal    string
			zipName      string
			expectedPath string
		}{
			{
				description:  "with a to flag set",
				flagLocal:    "../../my-project",
				expectedPath: "/some/my-project",
			},
			{
				description:  "with no to flag set and the zip file name has a timestamp",
				zipName:      "app_20210101",
				expectedPath: "/some/system/path/app",
			},
			{
				description:  "with an absolute to flag set",
				flagLocal:    "/some/system/path/my-project/app-abcde",
				expectedPath: "/some/system/path/my-project/app-abcde",
			},
			{
				description:  "with no to flag set and the zip file name has no timestamp",
				zipName:      "app-abcde",
				expectedPath: "/some/system/path/app-abcde",
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				var realmClient mock.RealmClient
				realmClient.ExportFn = func(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error) {
					return tc.zipName, &zip.Reader{}, nil
				}

				cmd := &Command{inputs{LocalPath: tc.flagLocal}}

				path, zipPkg, err := cmd.doExport(profile, realmClient, "", "")
				assert.Nil(t, err)
				assert.NotNil(t, zipPkg)
				assert.Equal(t, tc.expectedPath, path)
			})
		}
	})
}

func TestPullCommandCheckAppDestination(t *testing.T) {
	t.Run("should return true early if auto confirm is on", func(*testing.T) {
		ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, new(bytes.Buffer))

		ok, err := checkPathDestination(ui, "")
		assert.Nil(t, err)
		assert.True(t, ok, "should be ok")
	})

	t.Run("should return true early if the path does not exist", func(t *testing.T) {
		_, ui := mock.NewUI()

		ok, err := checkPathDestination(ui, "./not_a_directory")
		assert.Nil(t, err)
		assert.True(t, ok, "should be ok")
	})

	t.Run("should return true if the path does exist but is a file", func(t *testing.T) {
		tmpDir, teardown, err := u.NewTempDir("pull-command")
		assert.Nil(t, err)
		defer teardown()

		file, err := os.Create(filepath.Join(tmpDir, "project"))
		assert.Nil(t, err)
		defer file.Close()

		_, ui := mock.NewUI()

		ok, err := checkPathDestination(ui, filepath.Join(tmpDir, "project"))
		assert.Nil(t, err)
		assert.True(t, ok, "should be ok")
	})

	t.Run("should prompt the user to continue if the directory already exists", func(t *testing.T) {
		tmpDir, teardown, err := u.NewTempDir("pull-command")
		assert.Nil(t, err)
		defer teardown()

		dir := filepath.Join(tmpDir, "project")

		assert.Nil(t, os.MkdirAll(dir, os.ModePerm))

		for _, tc := range []struct {
			description string
			input       string
			answer      bool
		}{
			{"yes", "y", true},
			{description: "no"},
		} {
			t.Run(fmt.Sprintf("and return %t with an answer of '%s'", tc.answer, tc.description), func(t *testing.T) {
				_, console, _, ui, err := mock.NewVT10XConsole()
				assert.Nil(t, err)
				defer console.Close()

				doneCh := make(chan (struct{}))
				go func() {
					defer close(doneCh)
					console.ExpectString(fmt.Sprintf("Directory '%s' already exists, do you still wish to proceed?", dir))
					console.SendLine(tc.input)
					console.ExpectEOF()
				}()

				ok, err := checkPathDestination(ui, dir)
				assert.Nil(t, err)
				assert.Equal(t, tc.answer, ok)
			})
		}
	})
}

type mockHostingAssetClient struct {
	contents string
}

func (client mockHostingAssetClient) Get(url string) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(strings.NewReader(client.contents)),
	}, nil
}
