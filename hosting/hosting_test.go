package hosting_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/10gen/stitch-cli/hosting"
	"github.com/10gen/stitch-cli/utils"
	u "github.com/10gen/stitch-cli/utils/test"

	gc "github.com/smartystreets/goconvey/convey"
)

func localFileToAssetMetadata(t *testing.T, localPath, rootDir string, assetDescriptions map[string]hosting.AssetDescription) *hosting.AssetMetadata {
	file, err := os.Open(localPath)
	u.So(t, err, gc.ShouldBeNil)
	defer file.Close()

	info, statErr := file.Stat()
	u.So(t, statErr, gc.ShouldBeNil)

	fileHashStr, hashErr := utils.GenerateFileHashStr(localPath)
	u.So(t, hashErr, gc.ShouldBeNil)

	appID := "3720"
	relPath, pathErr := filepath.Rel(rootDir, localPath)
	u.So(t, pathErr, gc.ShouldBeNil)
	filePath := fmt.Sprintf("/%s", relPath)
	assetMetadata, famErr := hosting.FileToAssetMetadata(appID, localPath, filePath, info, assetDescriptions[filePath])
	u.So(t, famErr, gc.ShouldBeNil)

	u.So(t, assetMetadata.AppID, gc.ShouldEqual, appID)
	u.So(t, assetMetadata.FilePath, gc.ShouldEqual, filePath)
	u.So(t, assetMetadata.FileHash, gc.ShouldEqual, fileHashStr)
	u.So(t, assetMetadata.FileSize, gc.ShouldEqual, info.Size())

	return assetMetadata
}

func TestListLocalAssetMetadata(t *testing.T) {
	var testData []hosting.AssetMetadata
	path0 := "../testdata/full_app/hosting/files/asset_file0.json"
	path1 := "../testdata/full_app/hosting/files/ships/nostromo.json"
	fp0, fErr := filepath.Abs(path0)
	u.So(t, fErr, gc.ShouldBeNil)
	fp1, fErr := filepath.Abs(path1)
	u.So(t, fErr, gc.ShouldBeNil)

	rootDir, fErr := filepath.Abs("../testdata/full_app/hosting/files")

	jsonAttr := hosting.AssetAttribute{
		Name:  hosting.AttributeContentType,
		Value: "json",
	}
	p0 := fmt.Sprintf("/%s", path0)
	p1 := fmt.Sprintf("/%s", path1)
	assetDescriptions := map[string]hosting.AssetDescription{
		p0: {
			FilePath: p0,
			Attrs:    []hosting.AssetAttribute{jsonAttr},
		},
		p1: {
			FilePath: p1,
			Attrs:    []hosting.AssetAttribute{jsonAttr},
		},
	}

	am0 := localFileToAssetMetadata(t, fp0, rootDir, assetDescriptions)
	am1 := localFileToAssetMetadata(t, fp1, rootDir, assetDescriptions)
	testData = append(testData, *am0)
	testData = append(testData, *am1)

	u.So(t, fErr, gc.ShouldBeNil)
	file, err := os.Open(rootDir)
	u.So(t, err, gc.ShouldBeNil)
	defer file.Close()

	info, statErr := file.Stat()
	u.So(t, statErr, gc.ShouldBeNil)
	u.So(t, info.IsDir(), gc.ShouldBeTrue)

	appID := "3720"
	assetMetadata, listErr := hosting.ListLocalAssetMetadata(appID, rootDir, assetDescriptions)
	u.So(t, listErr, gc.ShouldBeNil)
	u.So(t, assetMetadata, gc.ShouldResemble, testData)
}

var jsonAttr = hosting.AssetAttribute{
	Name:  hosting.AttributeContentType,
	Value: "json",
}
var xmlAttr = hosting.AssetAttribute{
	Name:  hosting.AttributeContentType,
	Value: "xml",
}

func TestGetModifiedAssetMetadata(t *testing.T) {
	for _, tc := range []struct {
		local        hosting.AssetMetadata
		remote       hosting.AssetMetadata
		bodyModified bool
		attrModified bool
	}{
		{
			hosting.AssetMetadata{
				FileHash: "choppedpotato",
				Attrs:    []hosting.AssetAttribute{jsonAttr},
			},
			hosting.AssetMetadata{
				FileHash: "choppedpotato",
				Attrs:    []hosting.AssetAttribute{jsonAttr},
			},
			false,
			false,
		},
		{
			hosting.AssetMetadata{
				FileHash: "choppedpotato",
				Attrs:    []hosting.AssetAttribute{jsonAttr},
			},
			hosting.AssetMetadata{
				FileHash: "dicedpotato",
				Attrs:    []hosting.AssetAttribute{jsonAttr},
			},
			true,
			false,
		},
		{
			hosting.AssetMetadata{
				FileHash: "choppedpotato",
				Attrs:    []hosting.AssetAttribute{jsonAttr},
			},
			hosting.AssetMetadata{
				FileHash: "choppedpotato",
				Attrs:    []hosting.AssetAttribute{xmlAttr},
			},
			false,
			true,
		},
	} {
		u.So(t, hosting.GetModifiedAssetMetadata(tc.local, tc.remote), gc.ShouldResemble, hosting.ModifiedAssetMetadata{
			tc.local,
			tc.bodyModified,
			tc.attrModified,
		})
	}
}

func TestDiffAssetMetadata(t *testing.T) {
	jsonAM := hosting.AssetMetadata{
		FilePath: "/french/fry",
		FileHash: "choppedpotato",
		Attrs:    []hosting.AssetAttribute{jsonAttr},
	}
	xmlAM := hosting.AssetMetadata{
		FilePath: "/philip/j/fry",
		FileHash: "dicedpotato",
		Attrs:    []hosting.AssetAttribute{xmlAttr},
	}

	for _, tc := range []struct {
		local    []hosting.AssetMetadata
		remote   []hosting.AssetMetadata
		added    []hosting.AssetMetadata
		deleted  []hosting.AssetMetadata
		modified []hosting.ModifiedAssetMetadata
	}{
		{
			local: []hosting.AssetMetadata{
				jsonAM,
				xmlAM,
			},
			remote: []hosting.AssetMetadata{
				jsonAM,
				xmlAM,
			},
			added:    nil,
			deleted:  nil,
			modified: nil,
		},
		{
			local: []hosting.AssetMetadata{
				jsonAM,
				xmlAM,
			},
			remote: []hosting.AssetMetadata{
				jsonAM,
			},
			added: []hosting.AssetMetadata{
				xmlAM,
			},
			deleted:  nil,
			modified: nil,
		},
		{
			local: []hosting.AssetMetadata{
				jsonAM,
			},
			remote: []hosting.AssetMetadata{
				jsonAM,
				xmlAM,
			},
			added: nil,
			deleted: []hosting.AssetMetadata{
				xmlAM,
			},
			modified: nil,
		},
		{
			local: []hosting.AssetMetadata{
				jsonAM,
			},
			remote: []hosting.AssetMetadata{
				xmlAM,
			},
			added: []hosting.AssetMetadata{
				jsonAM,
			},
			deleted: []hosting.AssetMetadata{
				xmlAM,
			},
			modified: nil,
		},
		{
			local: []hosting.AssetMetadata{
				jsonAM,
			},
			remote: []hosting.AssetMetadata{
				{
					FilePath: "/french/fry",
					FileHash: "mincedpotato",
					Attrs:    []hosting.AssetAttribute{jsonAttr},
				},
			},
			added:   nil,
			deleted: nil,
			modified: []hosting.ModifiedAssetMetadata{
				{
					AssetMetadata: jsonAM,
					BodyModified:  true,
					AttrModified:  false,
				},
			},
		},
		{
			local: []hosting.AssetMetadata{
				jsonAM,
				xmlAM,
			},
			remote: []hosting.AssetMetadata{
				{
					FilePath: "/french/fry",
					FileHash: "mincedpotato",
					Attrs:    []hosting.AssetAttribute{jsonAttr},
				},
				{
					FilePath: "/philip/j/fry",
					FileHash: "killerpotato",
					Attrs:    []hosting.AssetAttribute{xmlAttr},
				},
			},
			added:   nil,
			deleted: nil,
			modified: []hosting.ModifiedAssetMetadata{
				{
					AssetMetadata: jsonAM,
					BodyModified:  true,
					AttrModified:  false,
				},
				{
					AssetMetadata: xmlAM,
					BodyModified:  true,
					AttrModified:  false,
				},
			},
		},
		{
			local: []hosting.AssetMetadata{
				jsonAM,
			},
			remote: []hosting.AssetMetadata{
				{
					FilePath: "/french/fry",
					FileHash: "choppedpotato",
					Attrs:    []hosting.AssetAttribute{xmlAttr},
				},
			},
			added:   nil,
			deleted: nil,
			modified: []hosting.ModifiedAssetMetadata{
				{
					AssetMetadata: jsonAM,
					BodyModified:  false,
					AttrModified:  true,
				},
			},
		},
		{
			local: []hosting.AssetMetadata{
				jsonAM,
			},
			remote: []hosting.AssetMetadata{
				{
					FilePath: "/french/fry",
					FileHash: "potatopotato",
					Attrs:    []hosting.AssetAttribute{xmlAttr},
				},
			},
			added:   nil,
			deleted: nil,
			modified: []hosting.ModifiedAssetMetadata{
				{
					AssetMetadata: jsonAM,
					BodyModified:  true,
					AttrModified:  true,
				},
			},
		},
	} {
		u.So(t, hosting.DiffAssetMetadata(tc.local, tc.remote), gc.ShouldResemble, hosting.NewAssetMetadataDiffs(tc.added, tc.deleted, tc.modified))
	}
}

func TestAssetAttributesEqual(t *testing.T) {
	for _, tc := range []struct {
		a     []hosting.AssetAttribute
		b     []hosting.AssetAttribute
		equal bool
	}{
		{
			[]hosting.AssetAttribute{{"Han", "Solo"}, {"Lando", "Calrissian"}},
			[]hosting.AssetAttribute{{"Lando", "Calrissian"}, {"Han", "Solo"}},
			true,
		},
		{
			[]hosting.AssetAttribute{{"Han", "Solo"}, {"Lando", "Calrissian"}},
			[]hosting.AssetAttribute{{"Han", "Solo"}, {"Lando", "Calrissian"}},
			true,
		},
		{
			[]hosting.AssetAttribute{{"Han", "Nolo"}, {"Lando", "Calrissian"}},
			[]hosting.AssetAttribute{{"Han", "Solo"}, {"Lando", "Calrissian"}},
			false,
		},
		{
			[]hosting.AssetAttribute{{"Lando", "Calrissian"}},
			[]hosting.AssetAttribute{{"Han", "Solo"}, {"Lando", "Calrissian"}},
			false,
		},
	} {
		u.So(t, hosting.AssetAttributesEqual(tc.a, tc.b), gc.ShouldEqual, tc.equal)
	}
}

func TestMetadataFileToAssetDescriptions(t *testing.T) {
	assetDescriptions, err := hosting.MetadataFileToAssetDescriptions("../testdata/full_app/hosting/metadata.json")
	u.So(t, err, gc.ShouldBeNil)

	u.So(t, len(assetDescriptions), gc.ShouldEqual, 2)
	path0 := "/asset_file0.json"
	path1 := "/ships/nostromo.json"
	u.So(t, assetDescriptions, gc.ShouldResemble, map[string]hosting.AssetDescription{
		path0: {
			path0,
			[]hosting.AssetAttribute{},
		},
		path1: {
			path1,
			[]hosting.AssetAttribute{},
		},
	})
}

func TestAssetMetadataDiff(t *testing.T) {
	a1Path := "/addMe/1"
	a2Path := "/addMe/2"
	added := []hosting.AssetMetadata{
		{
			FilePath: a1Path,
		},
		{
			FilePath: a2Path,
		},
	}

	addDiff := []string{
		"New Files:",
		fmt.Sprintf("\t+ %s", a1Path),
		fmt.Sprintf("\t+ %s", a2Path),
	}

	d1Path := "/deleteMe/1"
	d2Path := "/deleteMe/2"
	deleted := []hosting.AssetMetadata{
		{
			FilePath: d1Path,
		},
		{
			FilePath: d2Path,
		},
	}

	deleteDiff := []string{
		"Removed Files:",
		fmt.Sprintf("\t- %s", d1Path),
		fmt.Sprintf("\t- %s", d2Path),
	}

	m1Path := "/modifyMe/1"
	m2Path := "/modifyMe/2"
	modified := []hosting.ModifiedAssetMetadata{
		{
			AssetMetadata: hosting.AssetMetadata{
				FilePath: m1Path,
			},
		},
		{
			AssetMetadata: hosting.AssetMetadata{
				FilePath: m2Path,
			},
		},
	}

	modifyDiff := []string{
		"Modified Files:",
		fmt.Sprintf("\t* %s", m1Path),
		fmt.Sprintf("\t* %s", m2Path),
	}

	t.Run("with local additions only", func(t *testing.T) {
		amd := hosting.NewAssetMetadataDiffs(added, []hosting.AssetMetadata{}, []hosting.ModifiedAssetMetadata{})
		u.So(t, amd.Diff(), gc.ShouldResemble, addDiff)
	})

	t.Run("with local removals only", func(t *testing.T) {
		amd := hosting.NewAssetMetadataDiffs([]hosting.AssetMetadata{}, deleted, []hosting.ModifiedAssetMetadata{})
		u.So(t, amd.Diff(), gc.ShouldResemble, deleteDiff)
	})

	t.Run("with local modifications only", func(t *testing.T) {
		amd := hosting.NewAssetMetadataDiffs([]hosting.AssetMetadata{}, []hosting.AssetMetadata{}, modified)
		u.So(t, amd.Diff(), gc.ShouldResemble, modifyDiff)
	})

	t.Run("with additions, deletions, and modifcations", func(t *testing.T) {
		amd := hosting.NewAssetMetadataDiffs(added, deleted, modified)
		u.So(t, amd.Diff(), gc.ShouldResemble, append(append(addDiff, deleteDiff...), modifyDiff...))
	})

}
