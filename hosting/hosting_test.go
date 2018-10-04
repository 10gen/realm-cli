package hosting

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/10gen/stitch-cli/api"
	"github.com/10gen/stitch-cli/utils"
	u "github.com/10gen/stitch-cli/utils/test"
	gc "github.com/smartystreets/goconvey/convey"
)

func localFileToAssetMetadata(t *testing.T, localPath string) *api.AssetMetadata {
	file, err := os.Open(localPath)
	u.So(t, err, gc.ShouldBeNil)
	defer file.Close()

	info, statErr := file.Stat()
	u.So(t, statErr, gc.ShouldBeNil)

	fmt.Println(localPath)
	fileHash, hashErr := utils.GenerateFileHash(localPath)
	u.So(t, hashErr, gc.ShouldBeNil)

	appID := "3720"
	assetMetadata, famErr := fileToAssetMetadata(appID, localPath, info)
	u.So(t, famErr, gc.ShouldBeNil)

	u.So(t, assetMetadata.AppID, gc.ShouldEqual, appID)
	u.So(t, assetMetadata.FilePath, gc.ShouldEqual, localPath)
	u.So(t, assetMetadata.FileHash, gc.ShouldEqual, utils.FileHashStr(fileHash))
	u.So(t, assetMetadata.FileSize, gc.ShouldEqual, info.Size())

	return assetMetadata
}

func TestListLocalAssetMetadata(t *testing.T) {
	var testData []api.AssetMetadata
	fp0, fErr := filepath.Abs("../testdata/hosting/asset_file0.json")
	u.So(t, fErr, gc.ShouldBeNil)
	fp1, fErr := filepath.Abs("../testdata/hosting/ships/nostromo.json")
	u.So(t, fErr, gc.ShouldBeNil)

	am0 := localFileToAssetMetadata(t, fp0)
	am1 := localFileToAssetMetadata(t, fp1)
	testData = append(testData, *am0)
	testData = append(testData, *am1)

	rootDir, fErr := filepath.Abs("../testdata/hosting")
	u.So(t, fErr, gc.ShouldBeNil)
	file, err := os.Open(rootDir)
	u.So(t, err, gc.ShouldBeNil)
	defer file.Close()

	info, statErr := file.Stat()
	u.So(t, statErr, gc.ShouldBeNil)
	u.So(t, info.IsDir(), gc.ShouldBeTrue)

	appID := "3720"
	assetMetadata, listErr := listLocalAssetMetadata(appID, rootDir)
	u.So(t, listErr, gc.ShouldBeNil)
	u.So(t, assetMetadata, gc.ShouldResemble, testData)
}
