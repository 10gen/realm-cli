package local

import (
	"errors"
	"reflect"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"

	"github.com/google/go-cmp/cmp"
)

func TestSelectArchive(t *testing.T) {
	assert.RegisterOpts(reflect.TypeOf(archive{}), cmp.AllowUnexported(archive{}))

	t.Run("should return an error when no archive paths match a supported format", func(t *testing.T) {
		_, err := selectArchive([]string{
			"./node_modules.txt",
			"./node_modules.7z",
		})
		assert.Equal(t, errors.New("make sure file format is one of [.zip, .tar, .tgz, .tar.gz]"), err)
	})

	for _, tc := range []struct {
		description     string
		archives        []string
		expectedArchive archive
	}{
		{
			description: "should select a zip format when other preferences are available",
			archives: []string{
				"./node_modules",
				"./node_modules.tar.gz",
				"./node_modules.tgz",
				"./node_modules.tar",
				"./node_modules.zip",
			},
			expectedArchive: archive{path: "./node_modules.zip"},
		},
		{
			description: "should select a tar format when other preferences are available",
			archives: []string{
				"./node_modules",
				"./node_modules.tar.gz",
				"./node_modules.tgz",
				"./node_modules.tar",
			},
			expectedArchive: archive{path: "./node_modules.tar"},
		},
		{
			description: "should select a tgz format when other preferences are available",
			archives: []string{
				"./node_modules",
				"./node_modules.tgz",
			},
			expectedArchive: archive{path: "./node_modules.tgz"},
		},
		{
			description: "should select a targz format when other preferences are available",
			archives: []string{
				"./node_modules",
				"./node_modules.tar.gz",
			},
			expectedArchive: archive{path: "./node_modules.tar.gz"},
		},
		{
			description: "should select a directory format when no other preferences are available",
			archives: []string{
				"./node_modules.txt",
				"./node_modules.7z",
				"./node_modules",
			},
			expectedArchive: archive{"./node_modules", true},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			archive, err := selectArchive(tc.archives)
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedArchive, archive)
		})
	}
}
