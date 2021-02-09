package local

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestArchiveReaderNew(t *testing.T) {
	wd, wdErr := os.Getwd()
	assert.Nil(t, wdErr)

	testRoot := filepath.Join(wd, "testdata/dependencies")

	t.Run("should return an error when a project has an unsupported node_modules archive", func(t *testing.T) {
		archivePath := filepath.Join(testRoot, "7z/functions/node_modules.7z")

		_, err := newArchiveReader(archivePath, nil)
		assert.Equal(t, errUnknownArchiveExtension(archivePath), err)
	})

	for _, tc := range []struct {
		description string
		path        string
		test        func(t *testing.T, r ArchiveReader)
	}{
		{
			description: "should find a zip node_modules archive and return its reader",
			path:        filepath.Join(testRoot, "zip/functions/node_modules.zip"),
			test: func(t *testing.T, r ArchiveReader) {
				_, ok := r.(*zipReader)
				assert.True(t, ok, "should be a zip reader")
			},
		},
		{
			description: "should find find a tar node_modules archive and return its reader",
			path:        filepath.Join(testRoot, "tar/functions/node_modules.tar"),
			test: func(t *testing.T, r ArchiveReader) {
				_, ok := r.(*tarReader)
				assert.True(t, ok, "should be a tar reader")
			},
		},
		{
			description: "should find find a tgz node_modules archive and return its reader",
			path:        filepath.Join(testRoot, "tgz/functions/node_modules.tar.gz"),
			test: func(t *testing.T, r ArchiveReader) {
				_, ok := r.(*tarReader)
				assert.True(t, ok, "should be a tar reader")
			},
		},
		{
			description: "should find find a node_modules directory and return its reader",
			path:        filepath.Join(testRoot, "dir/functions/node_modules"),
			test: func(t *testing.T, r ArchiveReader) {
				_, ok := r.(*dirReader)
				assert.True(t, ok, "should be a dir reader")
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			file, fileErr := os.Open(tc.path)
			assert.Nil(t, fileErr)
			defer file.Close()

			r, readerErr := newArchiveReader(tc.path, file)
			assert.Nil(t, readerErr)

			tc.test(t, r)
		})
	}
}
