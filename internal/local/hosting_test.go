package local

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/api"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestHostingFind(t *testing.T) {
	wd, err := os.Getwd()
	assert.Nil(t, err)

	testRoot := filepath.Join(wd, "testdata")

	t.Run("should return an empty data when outside a project directory", func(t *testing.T) {
		hosting, err := FindAppHosting(testRoot)
		assert.Nil(t, err)
		assert.Equal(t, Hosting{}, hosting)
	})

	t.Run("should locate the hosting directory when provided a project directory", func(t *testing.T) {
		hosting, err := FindAppHosting(filepath.Join(testRoot, "hosting"))
		assert.Nil(t, err)
		assert.Equal(t, hosting.RootDir, filepath.Join(testRoot, "hosting/hosting"))

		t.Run("and should compute all diffs", func(t *testing.T) {
			tmpDir, teardown, err := u.NewTempDir("hosting")
			assert.Nil(t, err)
			defer teardown()

			hostingDiffs, err := hosting.Diffs(filepath.Join(tmpDir, user.HostingAssetCacheDir, "test.json"), "", []realm.HostingAsset{
				{HostingAssetData: realm.HostingAssetData{FilePath: "/deleteme.html"}},
				{
					HostingAssetData: realm.HostingAssetData{FilePath: "/404.html", FileHash: "7785338f982ac81219ef449f4943ec89"},
					Attrs:            realm.HostingAssetAttributes{{api.HeaderContentLanguage, "en-US"}},
				},
			})

			expectedAdded := []realm.HostingAsset{{
				HostingAssetData: realm.HostingAssetData{
					FilePath:     "/index.html",
					FileHash:     "daad4fb706d494feb9014e131f6520d4",
					FileSize:     163,
					LastModified: 1614291853,
				},
				Attrs: realm.HostingAssetAttributes{{api.HeaderContentType, "text/html"}},
			}}
			expectedDeleted := []realm.HostingAsset{{
				HostingAssetData: realm.HostingAssetData{
					FilePath: "/deleteme.html",
				},
			}}
			expectedModified := []ModifiedHostingAsset{{
				HostingAsset: realm.HostingAsset{
					HostingAssetData: realm.HostingAssetData{
						FilePath:     "/404.html",
						FileHash:     "7785338f982ac81219ef449f4943ec89",
						FileSize:     36,
						LastModified: 1614291853,
					},
					Attrs: realm.HostingAssetAttributes{},
				},
				AttrsModified: true,
			}}

			assert.Nil(t, err)

			assert.Equal(t, len(expectedAdded), len(hostingDiffs.Added))
			assert.Equal(t, expectedAdded[0].HostingAssetData.FilePath, hostingDiffs.Added[0].HostingAssetData.FilePath)
			assert.Equal(t, expectedAdded[0].HostingAssetData.FileHash, hostingDiffs.Added[0].HostingAssetData.FileHash)
			assert.Equal(t, expectedAdded[0].HostingAssetData.FileSize, hostingDiffs.Added[0].HostingAssetData.FileSize)
			assert.Equal(t, expectedAdded[0].Attrs, hostingDiffs.Added[0].Attrs)

			assert.Equal(t, expectedDeleted, hostingDiffs.Deleted)

			assert.Equal(t, len(expectedModified), len(hostingDiffs.Modified))
			assert.Equal(t, expectedModified[0].HostingAssetData.FilePath, hostingDiffs.Modified[0].HostingAssetData.FilePath)
			assert.Equal(t, expectedModified[0].HostingAssetData.FileHash, hostingDiffs.Modified[0].HostingAssetData.FileHash)
			assert.Equal(t, expectedModified[0].HostingAssetData.FileSize, hostingDiffs.Modified[0].HostingAssetData.FileSize)
			assert.Equal(t, expectedModified[0].Attrs, hostingDiffs.Modified[0].Attrs)
			assert.Equal(t, expectedModified[0].AttrsModified, hostingDiffs.Modified[0].AttrsModified)
			assert.Equal(t, expectedModified[0].BodyModified, hostingDiffs.Modified[0].BodyModified)
		})
	})
}
