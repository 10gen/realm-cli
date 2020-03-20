package utils

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"path"
	"strings"
	"testing"
)

var files = []struct {
	Name, Body string
}{
	{"folder1/a.txt", "Content of a.txt"},
	{"folder1/b.txt", "Content of b.txt"},
	{"c.txt", "Content of c.txt"},
}

func TestArchiveReader(t *testing.T) {
	for _, testCase := range []struct {
		description   string
		filename      string
		createArchive func(*testing.T) (ArchiveInputReader, int64)
	}{
		{
			description:   "With a .tar archive",
			filename:      "test.tar",
			createArchive: createTarArchive,
		},
		{
			description:   "With a .zip archive",
			filename:      "test.zip",
			createArchive: createZipArchive,
		},
		{
			description:   "With a .tar.gz archive",
			filename:      "test.tar.gz",
			createArchive: createGZArchive,
		},
		{
			description:   "With a .tgz archive",
			filename:      "test.tgz",
			createArchive: createGZArchive,
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {

			t.Run("Reading before advancing to the first entry with Next", func(t *testing.T) {
				archive, size := testCase.createArchive(t)
				ar, err := NewArchiveReader(archive, testCase.filename, size)
				if err != nil {
					t.Fatal(err)
				}

				var b []byte
				n, err := ar.Read(b)
				if n != 0 {
					t.Fatal("expected bytes read to be 0")
				}
				if err != io.EOF {
					t.Fatal("error did not match EOF", err)
				}

			})

			t.Run("Reading after reaching the end of the archive", func(t *testing.T) {
				archive, size := testCase.createArchive(t)
				ar, err := NewArchiveReader(archive, testCase.filename, size)
				if err != nil {
					t.Fatal(err)
				}

				for err == nil {
					_, err = ar.Next()
				}
				var b []byte
				n, err := ar.Read(b)
				if n != 0 {
					t.Fatal("expected bytes read to be 0")
				}
				if err != io.EOF {
					t.Fatal("error did not match EOF", err)
				}
			})

			t.Run("Reading each file", func(t *testing.T) {
				archive, size := testCase.createArchive(t)
				ar, err := NewArchiveReader(archive, testCase.filename, size)
				if err != nil {
					t.Fatal(err)
				}

				for _, file := range files {
					header, err := ar.Next()
					if err != nil {
						t.Fatal(err)
					}
					if header.FullPath != file.Name {
						t.Fatalf("%s did not match expected: %s", header.FullPath, file.Name)
					}
					baseName := path.Base(file.Name)
					if header.FileInfo().Name() != baseName {
						t.Fatalf("%s did not match expected: %s", header.FileInfo().Name(), file.Name)
					}
					var buf bytes.Buffer
					n, err := io.Copy(&buf, ar)
					if err != nil {
						t.Fatal(err)
					}
					if n <= 0 {
						t.Fatal("Expected n to be greater than 0")
					}
					if buf.String() != file.Body {
						t.Fatalf("%s did not match expected: %s", buf.String(), file.Body)
					}

				}

				_, err = ar.Next()
				if err != io.EOF {
					t.Fatal("error did not match EOF", err)
				}
			})

			t.Run("Traversal should work", func(t *testing.T) {
				archive, size := testCase.createArchive(t)
				archiveReader, err := NewArchiveReader(archive, testCase.filename, size)
				if err != nil {
					t.Fatal(err)
				}

				var counter int
				err = TraverseArchiveReader(archiveReader, func(header *FileHeader) error {
					file := files[counter]
					body, err := ioutil.ReadAll(archiveReader)
					if err != nil {
						t.Fatal(err)
					}

					if header.FullPath != file.Name {
						t.Fatalf("%s did not match expected: %s", header.FullPath, file.Name)
					}
					if string(body) != file.Body {
						t.Fatalf("%s did not match expected: %s", string(body), file.Body)
					}

					counter++
					return nil
				})
				if err != nil {
					t.Fatal(err)
				}
				if counter != len(files) {
					t.Fatalf("%d did not match expected: %d", counter, len(files))
				}
			})
		})
	}
}

func createTarArchive(t *testing.T) (ArchiveInputReader, int64) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	for _, file := range files {
		hdr := &tar.Header{
			Name: file.Name,
			Size: int64(len(file.Body)),
		}
		err := tw.WriteHeader(hdr)
		if err != nil {
			t.Fatal(err)
		}
		_, err = tw.Write([]byte(file.Body))
		if err != nil {
			t.Fatal(err)
		}
	}
	err := tw.Close()
	if err != nil {
		t.Fatal(err)
	}
	return bytes.NewReader(buf.Bytes()), int64(buf.Len())
}

func createZipArchive(t *testing.T) (ArchiveInputReader, int64) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	for _, file := range files {
		f, err := w.Create(file.Name)
		if err != nil {
			t.Fatal(err)
		}
		_, err = f.Write([]byte(file.Body))
		if err != nil {
			t.Fatal(err)
		}
	}
	err := w.Close()
	if err != nil {
		t.Fatal(err)
	}

	return bytes.NewReader(buf.Bytes()), int64(buf.Len())
}

func createGZArchive(t *testing.T) (ArchiveInputReader, int64) {
	var tBuf bytes.Buffer
	tw := tar.NewWriter(&tBuf)

	for _, file := range files {
		hdr := &tar.Header{
			Name: file.Name,
			Size: int64(len(file.Body)),
		}
		err := tw.WriteHeader(hdr)
		if err != nil {
			t.Fatal(err)
		}
		_, err = tw.Write([]byte(file.Body))
		if err != nil {
			t.Fatal(err)
		}
	}
	err := tw.Close()
	if err != nil {
		t.Fatal(err)
	}

	var zBuf bytes.Buffer
	zw := gzip.NewWriter(&zBuf)
	zw.Write(tBuf.Bytes())
	err = zw.Close()
	if err != nil {
		t.Fatal(err)
	}

	return bytes.NewReader(zBuf.Bytes()), int64(zBuf.Len())
}

func TestArchiveReaderInvalidExtension(t *testing.T) {
	var buf bytes.Buffer
	r := bytes.NewReader(buf.Bytes())
	for _, filename := range []string{
		"test.zipped",
		"test.gz",
	} {
		_, err := NewArchiveReader(r, filename, 0)
		if err == nil {
			t.Fatal("expected an error")
		}
		if !strings.Contains(err.Error(), "unrecognized archive extension for file") {
			t.Fatalf("expected error to contain  %s but it did not: %s", "unrecognized archive extension for file", err)
		}
	}
}
