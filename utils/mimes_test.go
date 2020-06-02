package utils_test

import (
	"testing"

	"github.com/10gen/realm-cli/utils"
	u "github.com/10gen/realm-cli/utils/test"
	gc "github.com/smartystreets/goconvey/convey"
)

func TestExtensionToMimeMapping(t *testing.T) {

	getContentTypeByExtension := func(t *testing.T, ext string, cType string, shouldFind bool) {
		mime, found := utils.GetContentTypeByExtension(ext)
		u.So(t, shouldFind, gc.ShouldEqual, found)
		u.So(t, mime, gc.ShouldEqual, cType)
	}

	t.Run("GetDefaultContentType() should return the proper response for the given extension, otherwise it should output the empty string", func(t *testing.T) {
		getContentTypeByExtension(t, "mp4", "video/mp4", true)
		getContentTypeByExtension(t, "mpe", "video/mpeg", true)
		getContentTypeByExtension(t, "qtif", "image/x-quicktime", true)
		getContentTypeByExtension(t, "mesh", "model/mesh", true)
		getContentTypeByExtension(t, "so", "application/octet-stream", true)
		getContentTypeByExtension(t, "lzh", "application/octet-stream", true)
		getContentTypeByExtension(t, "nat", "", false)
		getContentTypeByExtension(t, "CPT", "", false)
	})

	t.Run("GetDefaultContentType() should return the proper response for the given extension, otherwise it should output the empty string", func(t *testing.T) {
		u.So(t, utils.IsDefaultContentType("video/mp4"), gc.ShouldEqual, true)
		u.So(t, utils.IsDefaultContentType("application/mac-binhex40"), gc.ShouldEqual, true)
		u.So(t, utils.IsDefaultContentType("image/prs.btif"), gc.ShouldEqual, true)
		u.So(t, utils.IsDefaultContentType("image/x-macpaint"), gc.ShouldEqual, true)
		u.So(t, utils.IsDefaultContentType("text/vnd.wap.wmlscript"), gc.ShouldEqual, true)
		u.So(t, utils.IsDefaultContentType("text/vnd.dmclientscript"), gc.ShouldEqual, true)
		u.So(t, utils.IsDefaultContentType("video/x-flv"), gc.ShouldEqual, true)
		u.So(t, utils.IsDefaultContentType("text/vnd.wap.wmlscript_notreal"), gc.ShouldEqual, false)
		u.So(t, utils.IsDefaultContentType("text/vnd.dmclientscript_notreal"), gc.ShouldEqual, false)
		u.So(t, utils.IsDefaultContentType("video/x-flv_notreal"), gc.ShouldEqual, false)
		u.So(t, utils.IsDefaultContentType("video/MP$"), gc.ShouldEqual, false)
	})

}
