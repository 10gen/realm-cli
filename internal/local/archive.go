package local

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
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

var (
	supportedExts = []string{extZip, extTar, extTgz, extTarGz}
)

// ArchiveReader provides sequential access to the contents of the supported archive formats
type ArchiveReader interface {
	Next() (FileHeader, error)
	Read(b []byte) (int, error)
}

// FileHeader holds the path and file information
type FileHeader struct {
	Path string
	Info os.FileInfo
}

func newArchiveReader(path string, file *os.File) (ArchiveReader, error) {
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(path))

	if ext == extZip {
		return newZipReader(file, fileInfo.Size())
	}

	if ext == extTar {
		return newTarReader(file), nil
	}

	if ext == extTgz || strings.HasSuffix(path, extTarGz) {
		return newGzipReader(file)
	}

	if ext == "" {
		return newDirReader(path)
	}

	return nil, errUnknownArchiveExtension(path)
}

func errUnknownArchiveExtension(path string) error {
	return fmt.Errorf(
		"failed to read archive file at %s: unsupported format, use one of [%s] instead",
		path,
		strings.Join(supportedExts, ", "),
	)
}

type zipReader struct {
	*zip.Reader

	currFileIdx  int
	currOpenFile io.ReadCloser
}

func newZipReader(r io.ReaderAt, size int64) (ArchiveReader, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, err
	}

	return &zipReader{Reader: zr, currFileIdx: -1}, nil
}

func (r *zipReader) Next() (FileHeader, error) {
	if r.currOpenFile != nil {
		r.currOpenFile.Close()
		r.currOpenFile = nil
	}

	r.currFileIdx++
	if r.currFileIdx >= len(r.Reader.File) {
		return FileHeader{}, io.EOF
	}

	f := r.Reader.File[r.currFileIdx]

	return FileHeader{f.Name, f.FileInfo()}, nil
}

func (r *zipReader) Read(b []byte) (int, error) {
	if r.currFileIdx < 0 || r.currFileIdx >= len(r.Reader.File) {
		return 0, io.EOF
	}

	if r.currOpenFile == nil {
		f, err := r.Reader.File[r.currFileIdx].Open()
		if err != nil {
			return 0, err
		}
		r.currOpenFile = f
	}

	return r.currOpenFile.Read(b)
}

type tarReader struct {
	*tar.Reader
}

func newTarReader(r io.Reader) ArchiveReader {
	return &tarReader{Reader: tar.NewReader(r)}
}

func (r *tarReader) Next() (FileHeader, error) {
	h, err := r.Reader.Next()
	if err != nil {
		return FileHeader{}, err
	}

	return FileHeader{h.Name, h.FileInfo()}, nil
}

func newGzipReader(r io.Reader) (ArchiveReader, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	tr := newTarReader(gr)
	if err := gr.Close(); err != nil {
		return nil, err
	}
	return tr, nil
}

type dirReader struct {
	files []FileHeader

	currFileIdx  int
	currOpenFile io.ReadCloser
}

func newDirReader(dir string) (ArchiveReader, error) {
	if _, err := os.Stat(dir); err != nil {
		return nil, err
	}

	r := dirReader{currFileIdx: -1}

	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		r.files = append(r.files, FileHeader{path, info})
		return nil
	}); err != nil {
		return nil, err
	}

	return &r, nil
}

func (r *dirReader) Next() (FileHeader, error) {
	if r.currOpenFile != nil {
		r.currOpenFile.Close()
		r.currOpenFile = nil
	}

	r.currFileIdx++
	if r.currFileIdx >= len(r.files) {
		return FileHeader{}, io.EOF
	}

	return r.files[r.currFileIdx], nil
}

func (r *dirReader) Read(b []byte) (int, error) {
	if r.currFileIdx < 0 || r.currFileIdx >= len(r.files) {
		return 0, io.EOF
	}

	if r.currOpenFile == nil {
		f, err := os.Open(r.files[r.currFileIdx].Path)
		if err != nil {
			return 0, err
		}
		r.currOpenFile = f
	}

	return r.currOpenFile.Read(b)
}
