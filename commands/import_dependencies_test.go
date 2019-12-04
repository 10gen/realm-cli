package commands

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	u "github.com/10gen/stitch-cli/utils/test"
	gc "github.com/smartystreets/goconvey/convey"
)

func TestImportDependencies(t *testing.T) {
	t.Run("should be successful", func(t *testing.T) {
		expectedGroupID := "group-id"
		expectedAppID := "app-id"
		dir := "../testdata/app_with_dependencies/functions"

		excpectedPath, pathErr := filepath.Abs("../testdata/app_with_dependencies/functions/node_modules.tar")
		u.So(t, pathErr, gc.ShouldBeNil)

		stitchClient := &u.MockStitchClient{
			UploadDependenciesFn: func(groupID, appID, fullPath string) error {
				u.So(t, groupID, gc.ShouldEqual, expectedGroupID)
				u.So(t, appID, gc.ShouldEqual, expectedAppID)
				u.So(t, fullPath, gc.ShouldEqual, excpectedPath)
				return nil
			},
		}

		err := ImportDependencies(expectedGroupID, expectedAppID, dir, stitchClient)
		u.So(t, err, gc.ShouldBeNil)
	})
}

func TestFindDependenciesArchive(t *testing.T) {
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
			filename, err := findDependenciesArchive(tc.dir)
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
		{
			desc: "should return an error with an app with too many deps archives",
			dir:  "../testdata/app_with_too_many_deps_archives/functions",
			err:  "found more than one node_modules archive in the '%s' directory",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			dir, dirErr := filepath.Abs(tc.dir)
			u.So(t, dirErr, gc.ShouldBeNil)

			_, err := findDependenciesArchive(dir)
			u.So(t, err.Error(), gc.ShouldEqual, fmt.Sprintf(tc.err, dir))
		})
	}
}

func TestValidateFileFormat(t *testing.T) {
	for _, tc := range []struct {
		desc        string
		file        string
		expectedErr error
	}{
		{
			desc:        "TAR format should be supported",
			file:        "functions/node_modules.tar",
			expectedErr: nil,
		},
		{
			desc:        "ZIP format should be supported",
			file:        "functions/node_modules.zip",
			expectedErr: nil,
		},
		{
			desc:        "GZIP format should be supported (gz)",
			file:        "functions/node_modules.tar.gz",
			expectedErr: nil,
		},
		{
			desc:        "GZIP format should be supported (tgz)",
			file:        "functions/node_modules.tgz",
			expectedErr: nil,
		},
		{
			desc:        "ZIPX should not be supported",
			file:        "functions/node_modules.zipx",
			expectedErr: errors.New("file 'functions/node_modules.zipx' has an unsupported format"),
		},
		{
			desc:        "an extension with a 'gz' suffix should not be supported",
			file:        "node_modules.2gz",
			expectedErr: errors.New("file 'node_modules.2gz' has an unsupported format"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := validateDependenciesFileFormat(tc.file)
			u.So(t, err, gc.ShouldResemble, tc.expectedErr)
		})
	}
}
