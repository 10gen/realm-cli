package realm_test

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/utils/api"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestRealmHosting(t *testing.T) {
	u.SkipUnlessRealmServerRunning(t)

	client := newAuthClient(t)

	groupID := u.CloudGroupID()

	app, teardown := setupTestApp(t, client, groupID, "hosting-test")
	defer teardown()

	t.Run("should initially get empty hosting assets", func(t *testing.T) {
		assets, err := client.HostingAssets(groupID, app.ID)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(assets))
	})

	t.Run("should upload a file successfully", func(t *testing.T) {
		assert.Nil(t, client.HostingAssetUpload(groupID, app.ID, "testdata/hosting", realm.HostingAsset{
			HostingAssetData: realm.HostingAssetData{
				FilePath: "/index.html",
				FileHash: "9163ebc83aa75cae0a7e74b4e16af317",
				FileSize: 51,
			},
			Attrs: nil,
		}))
	})

	t.Run("should then get the hosting asset", func(t *testing.T) {
		assets, err := client.HostingAssets(groupID, app.ID)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(assets))

		asset := assets[0]
		assert.Equal(t, app.ID, asset.AppID)
		assert.Equal(t, "/index.html", asset.FilePath)
		assert.Equal(t, "9163ebc83aa75cae0a7e74b4e16af317", asset.FileHash)
		assert.Equal(t, int64(51), asset.FileSize)
		assert.Equal(t, 0, len(asset.Attrs))
		assert.NotEqual(t, "", asset.URL, "url should not be empty")
		assert.NotEqual(t, 0, asset.LastModified, "last modified should not be zero")

		t.Run("should be able to update the hosting asset attribute", func(t *testing.T) {
			attr := realm.HostingAssetAttribute{api.HeaderContentType, api.MediaTypeJSON}

			assert.Nil(t, client.HostingAssetAttributesUpdate(groupID, app.ID, asset.FilePath, attr))

			appAssets, err := client.HostingAssets(groupID, app.ID)
			assert.Nil(t, err)
			assert.Equal(t, 1, len(appAssets))

			updated := appAssets[0]
			assert.Equal(t, app.ID, updated.AppID)
			assert.Equal(t, "/index.html", updated.FilePath)
			assert.Equal(t, "9163ebc83aa75cae0a7e74b4e16af317", updated.FileHash)
			assert.Equal(t, int64(51), updated.FileSize)
			assert.Equal(t, realm.HostingAssetAttributes{attr}, updated.Attrs)
			assert.Equal(t, asset.URL, updated.URL)
			assert.NotEqual(t, 0, asset.LastModified, "last modified should not be zero")
		})

		t.Run("should be able to remove the hosting asset", func(t *testing.T) {
			assert.Nil(t, client.HostingAssetRemove(groupID, app.ID, asset.FilePath))

			appAssets, err := client.HostingAssets(groupID, app.ID)
			assert.Nil(t, err)
			assert.Equal(t, 0, len(appAssets))
		})
	})

	t.Run("should fail to invalidate the cache with hosting disabled", func(t *testing.T) {
		err := client.HostingCacheInvalidate(groupID, app.ID, "/*")
		assert.Equal(t, realm.ServerError{Message: "hosting is disabled"}, err)
	})

	t.Run("should invalidate the cache for all files successfully with hosting enabled", func(t *testing.T) {
		assert.Nil(t, client.Import(groupID, app.ID, &local.AppRealmConfigJSON{local.AppDataV2{local.AppStructureV2{
			ConfigVersion: realm.AppConfigVersion20210101,
			Hosting:       map[string]interface{}{"enabled": true},
		}}}))
		assert.Nil(t, client.HostingCacheInvalidate(groupID, app.ID, "/*"))
	})
}
