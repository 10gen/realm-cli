package local

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// set of supported archive file extensions
const (
	extTar   = ".tar"
	extTarGz = ".tar.gz"
	extTgz   = ".tgz"
	extZip   = ".zip"
)

type archiveMatcher func(path, ext string) bool

func isZip(_, ext string) bool {
	return ext == extZip
}

func isTar(_, ext string) bool {
	return ext == extTar
}

func isTarCompressed(path, ext string) bool {
	return ext == extTgz || strings.HasSuffix(path, extTarGz)
}

var (
	preferredArchiveMatchers = []archiveMatcher{
		isZip,
		isTar,
		isTarCompressed,
	}
)

type archive struct {
	path  string
	isDir bool
}

func selectArchive(archives []string) (archive, error) {
	extCache := map[string]string{}

	for _, archiveMatcher := range preferredArchiveMatchers {
		for _, path := range archives {
			ext := strings.ToLower(filepath.Ext(path))

			if archiveMatcher(path, ext) {
				return archive{path: path}, nil
			}

			extCache[path] = ext
		}
	}

	for _, path := range archives {
		ext, ok := extCache[path]
		if !ok {
			continue
		}
		if ext == "" {
			return archive{path: path, isDir: true}, nil
		}
	}

	return archive{}, fmt.Errorf(
		"make sure file format is one of [%s, %s, %s, %s]",
		extZip, extTar, extTgz, extTarGz,
	)
}

type fileHeader struct {
	path string
	info os.FileInfo
}

type dirReader struct {
	files []fileHeader

	currFileIdx  int
	currOpenFile io.ReadCloser
}

func newDirReader(dir string) (*dirReader, error) {
	if _, err := os.Stat(dir); err != nil {
		return nil, err
	}

	r := dirReader{currFileIdx: -1}

	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		r.files = append(r.files, fileHeader{path, info})
		return nil
	}); err != nil {
		return nil, err
	}

	return &r, nil
}

func (r *dirReader) Next() (fileHeader, error) {
	if r.currOpenFile != nil {
		r.currOpenFile.Close()
		r.currOpenFile = nil
	}

	r.currFileIdx++
	if r.currFileIdx >= len(r.files) {
		return fileHeader{}, io.EOF
	}

	return r.files[r.currFileIdx], nil
}

func (r *dirReader) Read(b []byte) (int, error) {
	if r.currFileIdx < 0 || r.currFileIdx >= len(r.files) {
		return 0, io.EOF
	}

	if r.currOpenFile == nil {
		f, err := os.Open(r.files[r.currFileIdx].path)
		if err != nil {
			return 0, err
		}
		r.currOpenFile = f
	}

	return r.currOpenFile.Read(b)
}
