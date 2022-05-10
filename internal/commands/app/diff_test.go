package app

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/api"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"github.com/Netflix/go-expect"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAppDiffHandler(t *testing.T) {
	groupID1 := primitive.NewObjectID().Hex()
	app1 := realm.App{
		ID:          "app1",
		GroupID:     groupID1,
		ClientAppID: "app1-abcde",
		Name:        "app1",
	}

	apps := []realm.App{app1}

	for _, tc := range []struct {
		description        string
		inputs             diffInputs
		expectedAppFilter  realm.AppFilter
		expectedDiff       []string
		expectedDiffOutput string
		expectedErr        error
		appError           bool
		skipFindApps       bool
		path               string
	}{
		{
			description:        "no project nor app flag set should diff based on input and resolve with app meta",
			expectedDiff:       []string{"diff1"},
			expectedDiffOutput: "The following reflects the proposed changes to your Realm app\ndiff1\n",
			skipFindApps:       true,
			path:               "testdata/diff-meta",
		},
		{
			description:        "no project flag set and an app flag set should show the diff for the app",
			inputs:             diffInputs{RemoteApp: "app1"},
			expectedAppFilter:  realm.AppFilter{App: "app1"},
			expectedDiff:       []string{"diff1"},
			expectedDiffOutput: "The following reflects the proposed changes to your Realm app\ndiff1\n",
			path:               "testdata/diff",
		},
		{
			description:        "no diffs between local and remote app",
			inputs:             diffInputs{RemoteApp: "app1"},
			expectedAppFilter:  realm.AppFilter{App: "app1"},
			expectedDiffOutput: "Deployed app is identical to proposed version\n",
			path:               "testdata/diff",
		},
		{
			description:        "a project flag set and no app flag set should diff based on input",
			inputs:             diffInputs{Project: groupID1},
			expectedAppFilter:  realm.AppFilter{GroupID: groupID1},
			expectedDiff:       []string{"diff1"},
			expectedDiffOutput: "The following reflects the proposed changes to your Realm app\ndiff1\n",
			path:               "testdata/diff",
		},
		{
			description:       "error on the diff",
			inputs:            diffInputs{Project: groupID1, RemoteApp: "app1"},
			expectedAppFilter: realm.AppFilter{GroupID: groupID1, App: "app1"},
			expectedErr:       errors.New("something went wrong"),
			path:              "testdata/diff",
		},
		{
			description:       "error on finding apps",
			inputs:            diffInputs{Project: groupID1, RemoteApp: "app1"},
			expectedAppFilter: realm.AppFilter{GroupID: groupID1, App: "app1"},
			expectedErr:       errors.New("something went wrong"),
			appError:          true,
			path:              "testdata/diff",
		},
	} {
		t.Run("with a local path that exists and "+tc.description, func(t *testing.T) {
			out, ui := mock.NewUI()

			realmClient := mock.RealmClient{}

			var appFilter realm.AppFilter
			var findAppsCalled bool
			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				appFilter = filter
				findAppsCalled = true
				if tc.appError {
					return nil, tc.expectedErr
				}
				return apps, nil
			}
			realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
				return tc.expectedDiff, tc.expectedErr
			}

			tc.inputs.LocalPath = tc.path
			cmd := &CommandDiff{tc.inputs}

			assert.Equal(t, tc.expectedErr, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
			assert.Equal(t, tc.expectedAppFilter, appFilter)
			assert.Equal(t, tc.expectedDiffOutput, out.String())
			assert.Equal(t, tc.skipFindApps, !findAppsCalled)
		})
	}

	t.Run("should return an error if local path does not resolve to an app directory", func(t *testing.T) {
		_, ui := mock.NewUI()

		cmd := &CommandDiff{diffInputs{LocalPath: "./some/path"}}
		assert.Equal(t, errors.New("failed to find app at ./some/path"), cmd.Handler(nil, ui, cli.Clients{}))
	})

	t.Run("diff function dependencies", func(t *testing.T) {

		realmClient := mock.RealmClient{}

		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return apps, nil
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

		diffStr := `The following reflects the proposed changes to your Realm app
diff1
diff2
Added Dependencies
  + twilio@3.35.1
Removed Dependencies
  - debug@4.3.1
Modified Dependencies
  * underscore@1.9.1 -> underscore@1.9.2
`
		t.Run("with include node modules set it should diff function dependencies", func(t *testing.T) {
			out, ui := mock.NewUI()
			cmd := &CommandDiff{diffInputs{LocalPath: "testdata/dependencies", IncludeNodeModules: true}}
			assert.Equal(t, nil, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))

			assert.Equal(t, diffStr, out.String())
		})

		t.Run("with include dependencies set it should diff function dependencies", func(t *testing.T) {
			out, ui := mock.NewUI()
			cmd := &CommandDiff{diffInputs{LocalPath: "testdata/dependencies", IncludeDependencies: true}}
			assert.Equal(t, nil, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))

			assert.Equal(t, diffStr, out.String())
		})

		t.Run("with include package json set it should diff function dependencies", func(t *testing.T) {
			out, ui := mock.NewUI()
			cmd := &CommandDiff{diffInputs{LocalPath: "testdata/dependencies", IncludePackageJSON: true}}
			assert.Equal(t, nil, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))

			assert.Equal(t, diffStr, out.String())
		})
	})

	t.Run("should return an error when more than one dependencies flag is set", func(t *testing.T) {
		t.Run("when include node modules and include package json are both set", func(t *testing.T) {
			cmd := &CommandDiff{diffInputs{LocalPath: "testdata/dependencies", IncludeNodeModules: true, IncludePackageJSON: true}}
			assert.Equal(t, errors.New(`cannot use both "include-node-modules" and "include-package-json" at the same time`), cmd.inputs.Resolve(nil, nil))
		})

		t.Run("when include dependencies and include package json are both set", func(t *testing.T) {
			cmd := &CommandDiff{diffInputs{LocalPath: "testdata/dependencies", IncludeDependencies: true, IncludePackageJSON: true}}
			assert.Equal(t, errors.New(`cannot use both "include-dependencies" and "include-package-json" at the same time`), cmd.inputs.Resolve(nil, nil))
		})
	})

	t.Run("should return an error when diff dependencies returns an error", func(t *testing.T) {
		t.Run("when include node modules is set", func(t *testing.T) {
			_, ui := mock.NewUI()

			realmClient := mock.RealmClient{}

			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				return apps, nil
			}
			realmClient.DiffDependenciesFn = func(groupID, appID, uploadPath string) (realm.DependenciesDiff, error) {
				return realm.DependenciesDiff{}, errors.New("realm client error")
			}
			realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
				return []string{"diff1", "diff2"}, nil
			}

			cmd := &CommandDiff{diffInputs{LocalPath: "testdata/dependencies", IncludeNodeModules: true}}
			assert.Equal(t, errors.New("realm client error"), cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
		})

		t.Run("when include package json is set", func(t *testing.T) {
			_, ui := mock.NewUI()

			realmClient := mock.RealmClient{}

			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				return apps, nil
			}
			realmClient.DiffDependenciesFn = func(groupID, appID, uploadPath string) (realm.DependenciesDiff, error) {
				return realm.DependenciesDiff{}, errors.New("realm client error")
			}
			realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
				return []string{"diff1", "diff2"}, nil
			}

			cmd := &CommandDiff{diffInputs{LocalPath: "testdata/dependencies", IncludePackageJSON: true}}
			assert.Equal(t, errors.New("realm client error"), cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
		})
		t.Run("when include dependencies is set", func(t *testing.T) {
			_, ui := mock.NewUI()

			realmClient := mock.RealmClient{}

			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				return apps, nil
			}
			realmClient.DiffDependenciesFn = func(groupID, appID, uploadPath string) (realm.DependenciesDiff, error) {
				return realm.DependenciesDiff{}, errors.New("realm client error")
			}
			realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
				return []string{"diff1", "diff2"}, nil
			}

			cmd := &CommandDiff{diffInputs{LocalPath: "testdata/dependencies", IncludeDependencies: true}}
			assert.Equal(t, errors.New("realm client error"), cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
		})
	})

	t.Run("with include hosting set should diff hosting assets", func(t *testing.T) {
		profile := mock.NewProfile(t)

		out, ui := mock.NewUI()

		realmClient := mock.RealmClient{}

		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return apps, nil
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
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			return []string{"diff1", "diff2"}, nil
		}

		cmd := &CommandDiff{diffInputs{LocalPath: "testdata/diff", IncludeHosting: true}}
		assert.Equal(t, nil, cmd.Handler(profile, ui, cli.Clients{Realm: realmClient}))

		assert.Equal(t, `The following reflects the proposed changes to your Realm app
diff1
diff2
New hosting files
  + /index.html
Removed hosting files
  - /deleteme.html
Modified hosting files
  * /404.html
`, out.String())
	})
}

func TestAppDiffInputs(t *testing.T) {
	for _, tc := range []struct {
		description    string
		inputs         diffInputs
		prepareProfile func(p *user.Profile)
		procedure      func(c *expect.Console)
		test           func(t *testing.T, i diffInputs, p *user.Profile)
	}{
		{
			description:    "should resolve empty inputs when outside an app directory by prompting for the local path which exists",
			prepareProfile: func(p *user.Profile) {},
			procedure: func(c *expect.Console) {
				c.ExpectString("App filepath (local)")
				c.SendLine("./testdata/diff")
			},
			test: func(t *testing.T, i diffInputs, p *user.Profile) {
				assert.Equal(t, filepath.Join(p.WorkingDirectory, "testdata/diff"), i.LocalPath)
				assert.Equal(t, "eggcorn-abcde", i.RemoteApp)
			},
		},
		{
			description: "should resolve empty inputs when inside an app directory to the app details",
			prepareProfile: func(p *user.Profile) {
				p.WorkingDirectory = filepath.Join(p.WorkingDirectory, "testdata/diff/hosting")
			},
			procedure: func(c *expect.Console) {},
			test: func(t *testing.T, i diffInputs, p *user.Profile) {
				assert.Equal(t, filepath.Join(p.WorkingDirectory, ".."), i.LocalPath)
				assert.Equal(t, "eggcorn-abcde", i.RemoteApp)
			},
		},
		{
			description: "should not override app flag when run inside an app directory",
			inputs:      diffInputs{RemoteApp: "different-app"},
			prepareProfile: func(p *user.Profile) {
				p.WorkingDirectory = filepath.Join(p.WorkingDirectory, "testdata/diff/hosting")
			},
			procedure: func(c *expect.Console) {},
			test: func(t *testing.T, i diffInputs, p *user.Profile) {
				assert.Equal(t, filepath.Join(p.WorkingDirectory, ".."), i.LocalPath)
				assert.Equal(t, "different-app", i.RemoteApp)
			},
		},
		{
			description: "should not override app flag when local path exists",
			inputs: diffInputs{
				RemoteApp: "different-app",
				LocalPath: "./testdata/diff",
			},
			prepareProfile: func(p *user.Profile) {},
			procedure:      func(c *expect.Console) {},
			test: func(t *testing.T, i diffInputs, p *user.Profile) {
				assert.Equal(t, filepath.Join(p.WorkingDirectory, "testdata/diff"), i.LocalPath)
				assert.Equal(t, "different-app", i.RemoteApp)
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			profile := mock.NewProfile(t)

			_, console, _, ui, consoleErr := mock.NewVT10XConsole()
			assert.Nil(t, consoleErr)
			defer console.Close()

			tc.prepareProfile(profile)

			doneCh := make(chan (struct{}))
			go func() {
				defer close(doneCh)
				tc.procedure(console)
			}()

			assert.Nil(t, tc.inputs.Resolve(profile, ui))

			console.Tty().Close() // flush the writers
			<-doneCh              // wait for procedure to complete

			tc.test(t, tc.inputs, profile)
		})
	}

	t.Run("should return an error when specified local path does not exist", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_init_input_test")
		defer teardown()

		localPath := "fakePath"

		i := diffInputs{LocalPath: localPath}
		assert.Equal(t, errProjectInvalid(localPath, false), i.Resolve(profile, nil))
	})

	t.Run("should return an error when specified local path is not a supported Realm app project", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app_init_input_test")
		defer teardown()

		localPath := "./testdata"

		i := diffInputs{LocalPath: localPath}
		assert.Equal(t, errProjectInvalid(localPath, true), i.Resolve(profile, nil))
	})
}
