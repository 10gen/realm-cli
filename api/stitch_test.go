package api_test

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/10gen/stitch-cli/api"

	u "github.com/10gen/stitch-cli/utils/test"
	gc "github.com/smartystreets/goconvey/convey"
)

func TestErrStitchResponse(t *testing.T) {
	t.Run("with a non-JSON response should return the original content", func(t *testing.T) {
		err := api.UnmarshalStitchError(&http.Response{
			Body: u.NewResponseBody(strings.NewReader("not-json")),
		})
		u.So(t, err, gc.ShouldBeError, "error: not-json")
	})

	t.Run("with an empty non-JSON response should respond with the status", func(t *testing.T) {
		err := api.UnmarshalStitchError(&http.Response{
			Status: "418 Toot toot",
			Body:   u.NewResponseBody(strings.NewReader("")),
		})
		u.So(t, err, gc.ShouldBeError, "error: 418 Toot toot")
	})

	t.Run("with a JSON response should decode the error content", func(t *testing.T) {
		err := api.UnmarshalStitchError(&http.Response{
			Body: u.NewResponseBody(strings.NewReader(`{ "error": "something went horribly, horribly wrong" }`)),
		})
		u.So(t, err, gc.ShouldBeError, "error: something went horribly, horribly wrong")
	})
}

// md5Sum returns the md5 hash sum of the input string
func md5Sum(in string) string {
	h := md5.New()
	io.WriteString(h, in)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func TestUploadAsset(t *testing.T) {

	var testHandler http.HandlerFunc

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testHandler(w, r)
	}))

	testContents := "hello world"
	t.Run("uploading an asset should work", func(t *testing.T) {

		var uploadedAssetMetadata api.AssetMetadata
		uploadedFileData := &bytes.Buffer{}

		testHandler = func(w http.ResponseWriter, r *http.Request) {
			mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if !strings.HasPrefix(mediaType, "multipart/") {
				http.Error(w, "invalid multipart header", http.StatusBadRequest)
				return
			}
			mpr := multipart.NewReader(r.Body, params["boundary"])

			// Read the first section from the body which contains metadata about the asset
			metaReader, err := mpr.NextPart()
			if err != nil || metaReader.FormName() != "meta" {
				http.Error(w, "invalid multipart form name", http.StatusBadRequest)
				return
			}

			metaDecoder := json.NewDecoder(metaReader)
			metaDecoderErr := metaDecoder.Decode(&uploadedAssetMetadata)
			if metaDecoderErr != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// Read the file data
			dataReader, err := mpr.NextPart()
			if err != nil || dataReader.FormName() != "file" {
				http.Error(w, "invalid multipart form name", http.StatusBadRequest)
				return
			}

			if _, err = io.Copy(uploadedFileData, dataReader); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		}

		path, hash, size := "/test", md5Sum(testContents), int64(len(testContents))

		testClient := api.NewStitchClient(api.NewClient(testServer.URL))
		testClient.UploadAsset(
			"groupid",
			"appid",
			path,
			hash,
			size,
			strings.NewReader(testContents),
			api.AssetAttribute{
				Name:  "Content-Type",
				Value: "application/json",
			},
		)

		u.So(t, uploadedAssetMetadata, gc.ShouldResemble, api.AssetMetadata{
			AppID:    "appid",
			FilePath: path,
			FileHash: hash,
			FileSize: size,
			Attrs: []api.AssetAttribute{
				{
					Name:  "Content-Type",
					Value: "application/json",
				},
			},
		})
	})
}

func TestListAssetsForAppID(t *testing.T) {

	var testHandler http.HandlerFunc

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testHandler(w, r)
	}))

	testContents := []api.AssetMetadata{
		{
			AppID:    "appid",
			FilePath: "foo.txt",
			URL:      "url/foo.txt",
			FileSize: 20,
			FileHash: "OWEJFOWEF",
		},
		{
			AppID:    "appid",
			FilePath: "bar.txt",
			URL:      "url/bar.txt",
			FileSize: 203,
			FileHash: "OWEJddsdcsFOWEF",
		},
	}
	t.Run("listing assets by AppID should work", func(t *testing.T) {

		testHandler = func(w http.ResponseWriter, r *http.Request) {
			metaJson, err := json.Marshal(&testContents)
			if err != nil {
				http.Error(w, "invalid asset metadata", http.StatusBadRequest)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write(metaJson)
		}

		testClient := api.NewStitchClient(api.NewClient(testServer.URL))
		assetMetadatas, err := testClient.ListAssetsForAppID("groupid", "appID")
		u.So(t, err, gc.ShouldBeNil)
		u.So(t, len(assetMetadatas), gc.ShouldEqual, 2)
		u.So(t, assetMetadatas[0], gc.ShouldResemble, testContents[0])
		u.So(t, assetMetadatas[1], gc.ShouldResemble, testContents[1])
	})
}
