package app

import (
	"errors"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/api"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

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
	}{
		{
			description:        "with no project nor app flag set should diff based on input",
			expectedDiff:       []string{"diff1"},
			expectedDiffOutput: "The following reflects the proposed changes to your Realm app\ndiff1\n",
		},
		{
			description:        "with no project flag set and an app flag set should show the diff for the app",
			inputs:             diffInputs{LocalPath: "testdata/project", ProjectInputs: cli.ProjectInputs{App: "app1"}},
			expectedAppFilter:  realm.AppFilter{App: "app1"},
			expectedDiff:       []string{"diff1"},
			expectedDiffOutput: "The following reflects the proposed changes to your Realm app\ndiff1\n",
		},
		{
			description:        "with no diffs between local and remote app",
			inputs:             diffInputs{LocalPath: "testdata/project", ProjectInputs: cli.ProjectInputs{App: "app1"}},
			expectedAppFilter:  realm.AppFilter{App: "app1"},
			expectedDiffOutput: "Deployed app is identical to proposed version\n",
		},
		{
			description:        "with a project flag set and no app flag set should diff based on input",
			inputs:             diffInputs{LocalPath: "testdata/project", ProjectInputs: cli.ProjectInputs{Project: groupID1}},
			expectedAppFilter:  realm.AppFilter{GroupID: groupID1},
			expectedDiff:       []string{"diff1"},
			expectedDiffOutput: "The following reflects the proposed changes to your Realm app\ndiff1\n",
		},
		{
			description:       "error on the diff",
			inputs:            diffInputs{LocalPath: "testdata/project", ProjectInputs: cli.ProjectInputs{Project: groupID1, App: "app1"}},
			expectedAppFilter: realm.AppFilter{GroupID: groupID1, App: "app1"},
			expectedErr:       errors.New("something went wrong"),
		},
		{
			description:       "error on finding apps",
			inputs:            diffInputs{LocalPath: "testdata/project", ProjectInputs: cli.ProjectInputs{Project: groupID1, App: "app1"}},
			expectedAppFilter: realm.AppFilter{GroupID: groupID1, App: "app1"},
			expectedErr:       errors.New("something went wrong"),
			appError:          true,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			out, ui := mock.NewUI()

			realmClient := mock.RealmClient{}

			var appFilter realm.AppFilter
			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				appFilter = filter
				if tc.appError {
					return nil, tc.expectedErr
				}
				return apps, nil
			}
			realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
				return tc.expectedDiff, tc.expectedErr
			}

			cmd := &CommandDiff{inputs: tc.inputs}
			assert.Equal(t, tc.expectedErr, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
			assert.Equal(t, tc.expectedAppFilter, appFilter)
			assert.Equal(t, tc.expectedDiffOutput, out.String())
		})
	}

	t.Run("with include dependencies set should diff function dependencies", func(t *testing.T) {
		out, ui := mock.NewUI()

		realmClient := mock.RealmClient{}

		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return apps, nil
		}
		realmClient.DiffFn = func(groupID, appID string, appData interface{}) ([]string, error) {
			return []string{"diff1", "diff2"}, nil
		}

		cmd := &CommandDiff{diffInputs{IncludeDependencies: true}}
		assert.Equal(t, nil, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))

		assert.Equal(t, `The following reflects the proposed changes to your Realm app
diff1
diff2
+ New function dependencies
`, out.String())
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
