package commands

import (
	"fmt"
	"path/filepath"
	"testing"

	u "github.com/10gen/stitch-cli/utils/test"
	"github.com/mitchellh/cli"
	gc "github.com/smartystreets/goconvey/convey"
)

func TestImportDependencies(t *testing.T) {
	t.Run("should be successful", func(t *testing.T) {
		expectedGroupID := "group-id"
		expectedAppID := "app-id"
		dir := "../testdata/app_with_dependencies/functions"

		stitchClient := &u.MockStitchClient{
			UploadDependenciesFn: func(groupID, appID, fullPath string) error {
				u.So(t, groupID, gc.ShouldEqual, expectedGroupID)
				u.So(t, appID, gc.ShouldEqual, expectedAppID)
				u.So(t, fullPath, gc.ShouldContainSubstring, "node_modules.zip")
				return nil
			},
		}

		mockUI := cli.NewMockUi()
		err := ImportDependencies(mockUI, expectedGroupID, expectedAppID, dir, stitchClient)
		u.So(t, err, gc.ShouldBeNil)
	})

}

func TestFindDependenciesLocation(t *testing.T) {
	dirAbsPath, dirErr := filepath.Abs("../testdata/app_with_dependencies/functions")
	u.So(t, dirErr, gc.ShouldBeNil)
	for _, tc := range []struct {
		desc string
		dir  string
		file string
	}{
		{
			desc: "should successfully find the node_modules.tar archive and ignore the node_modules folder",
			dir:  dirAbsPath,
			file: "../testdata/app_with_dependencies/functions/node_modules.tar",
		},
		{
			desc: "should successfully find the node_modules.tar archive with a relative path",
			dir:  "../testdata/app_with_dependencies/functions",
			file: "../testdata/app_with_dependencies/functions/node_modules.tar",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			filename, err := findDependenciesLocation(tc.dir)
			excpectedPath, pathErr := filepath.Abs(tc.file)
			u.So(t, pathErr, gc.ShouldBeNil)
			u.So(t, filename, gc.ShouldEqual, excpectedPath)
			u.So(t, err, gc.ShouldBeNil)
		})
	}

	for _, tc := range []struct {
		desc string
		dir  string
		err  string
	}{
		{
			desc: "should return an error with an app without a node modules archive",
			dir:  "../testdata/app_without_dependencies/functions",
			err:  "node_modules archive not found in the '%s' directory",
		},
		{
			desc: "should return an error with an app without a functions folder",
			dir:  "../testdata/simple_app/functions",
			err:  "node_modules archive not found in the '%s' directory",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			dir, dirErr := filepath.Abs(tc.dir)
			u.So(t, dirErr, gc.ShouldBeNil)

			_, err := findDependenciesLocation(dir)
			u.So(t, err.Error(), gc.ShouldEqual, fmt.Sprintf(tc.err, dir))
		})
	}
}
