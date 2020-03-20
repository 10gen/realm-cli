package utils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// List of supported extensions
const (
	extZip = ".zip"
	extTar = ".tar"
	extGz  = ".tar.gz"
	extTgz = ".tgz"
)

// TraverseArchiveReader traverses the archive reader while invoking the provided handler for each new file
func TraverseArchiveReader(archiveReader ArchiveReader, fileHandler func(header *FileHeader) error) error {
	for {
		header, headerErr := archiveReader.Next()
		if headerErr == io.EOF {
			break // End of archive
		}
		if headerErr != nil {
			return fmt.Errorf("failed to advance to the next entry in the archive: %s", headerErr)
		}

		if err := fileHandler(header); err != nil {
			return err
		}
	}
	return nil
}

// ArchiveReader provides sequential access to the contents of the supported
// archive formats.
// Reader.Next advances to the next file in the archive (including the first),
// and then ArchiveReader can be treated as an io.Reader to access the file's data.
type ArchiveReader interface {
	Next() (*FileHeader, error)
	Read(b []byte) (int, error)
}

// NewArchiveReader creates a new ArchiveReader reading from r.
func NewArchiveReader(r ArchiveInputReader, fullPath string, size int64) (ArchiveReader, error) {
	ext := strings.ToLower(path.Ext(fullPath))

	switch {
	case ext == extZip:
		return NewZipReader(r, size)
	case ext == extTar:
		return NewTarReader(r), nil
	case ext == extTgz || strings.HasSuffix(fullPath, extGz):
		return NewGZReader(r)
	case ext == "":
		return NewDirReader(fullPath)
	default:
		return nil, fmt.Errorf("unrecognized archive extension for file: %s", fullPath)
	}
}

// FileHeader holds file header information
type FileHeader struct {
	FullPath string
	fi       os.FileInfo
}

// FileInfo returns an os.FileInfo for the FileHeader.
func (header *FileHeader) FileInfo() os.FileInfo {
	return header.fi
}

// ArchiveInputReader is an interface that wraps Reader and ReaderAt.
type ArchiveInputReader interface {
	io.Reader
	io.ReaderAt
}

type zipReader struct {
	r            *zip.Reader
	currFileIdx  int
	currOpenFile io.ReadCloser
}

// NewZipReader creates a new zip archive reader reading from r
func NewZipReader(r io.ReaderAt, size int64) (ArchiveReader, error) {
	reader, err := zip.NewReader(r, size)
	if err != nil {
		return nil, err
	}

	return &zipReader{
		r:           reader,
		currFileIdx: -1,
	}, nil
}

// Next advances to the next entry in the zip archive.
// io.EOF is returned at the end of the input.
func (zr *zipReader) Next() (*FileHeader, error) {
	if zr.currOpenFile != nil {
		zr.currOpenFile.Close()
		zr.currOpenFile = nil
	}

	zr.currFileIdx++
	if zr.currFileIdx >= len(zr.r.File) {
		return nil, io.EOF
	}

	file := zr.r.File[zr.currFileIdx]
	return &FileHeader{
		FullPath: file.Name,
		fi:       file.FileInfo(),
	}, nil
}

// Read reads from the current file in the zip archive.
// It returns (0, io.EOF) when it reaches the end of that file,
// until Next is called to advance to the next file.
func (zr *zipReader) Read(b []byte) (int, error) {
	if zr.currFileIdx < 0 || zr.currFileIdx >= len(zr.r.File) {
		return 0, io.EOF
	}

	if zr.currOpenFile == nil {
		file := zr.r.File[zr.currFileIdx]
		rc, err := file.Open()
		if err != nil {
			return 0, err
		}
		zr.currOpenFile = rc
	}

	return zr.currOpenFile.Read(b)
}

type dirReader struct {
	currFileIdx  int
	files        []FileHeader
	currOpenFile io.ReadCloser
}

// NewDirReader creates a new dir reader by traversing through all the files in the directory
func NewDirReader(dirName string) (ArchiveReader, error) {
	_, err := os.Stat(dirName)
	if err != nil {
		return nil, err
	}
	files := make([]FileHeader, 0)
	err = filepath.Walk(dirName, func(path string, info os.FileInfo, err error) error {

		if info.IsDir() {
			return nil
		}
		files = append(files, FileHeader{
			FullPath: path,
			fi:       info,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &dirReader{
		files:       files,
		currFileIdx: -1,
	}, nil
}

// Next advances to the next entry in the directory.
// io.EOF is returned at the end of the input.
func (d *dirReader) Next() (*FileHeader, error) {
	if d.currOpenFile != nil {
		d.currOpenFile.Close()
		d.currOpenFile = nil
	}

	d.currFileIdx++
	if d.currFileIdx >= len(d.files) {
		return nil, io.EOF
	}

	fileHeader := d.files[d.currFileIdx]
	return &fileHeader, nil
}

// Read reads from the current file in the directory.
// It returns (0, io.EOF) when it reaches the end of that file,
// until Next is called to advance to the next file.
func (d *dirReader) Read(b []byte) (int, error) {
	if d.currFileIdx < 0 || d.currFileIdx >= len(d.files) {
		return 0, io.EOF
	}

	if d.currOpenFile == nil {
		rc, err := os.Open(d.files[d.currFileIdx].FullPath)
		if err != nil {
			return 0, err
		}
		d.currOpenFile = rc
	}

	return d.currOpenFile.Read(b)
}

type tarReader struct {
	r *tar.Reader
}

// NewTarReader creates a new tar archive reader reading from r
func NewTarReader(r io.Reader) ArchiveReader {
	return &tarReader{
		r: tar.NewReader(r),
	}
}

// Next advances to the next entry in the tar archive.
// io.EOF is returned at the end of the input.
func (tr *tarReader) Next() (*FileHeader, error) {
	header, err := tr.r.Next()

	if err != nil {
		return nil, err
	}

	return &FileHeader{
		FullPath: header.Name,
		fi:       header.FileInfo(),
	}, nil
}

// Read reads from the current file in the tar archive.
// It returns (0, io.EOF) when it reaches the end of that file,
// until Next is called to advance to the next file.
func (tr *tarReader) Read(b []byte) (int, error) {
	return tr.r.Read(b)
}

// NewGZReader creates a new gzip archive reader reading from r
func NewGZReader(r io.Reader) (ArchiveReader, error) {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	archiverReader := NewTarReader(gzr)
	err = gzr.Close()
	if err != nil {
		return nil, err
	}

	return archiverReader, nil
}
