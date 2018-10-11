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
	"github.com/10gen/stitch-cli/hosting"

	u "github.com/10gen/stitch-cli/utils/test"
	gc "github.com/smartystreets/goconvey/convey"
)

const (
	groupID   = "groupID"
	appID     = "appID"
	pathParam = "path"
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
	t.Run("uploading an asset should work", func(t *testing.T) {

		var uploadedAssetMetadata hosting.AssetMetadata
		uploadedFileData := &bytes.Buffer{}

		testContents := "hello world"
		testHandler := func(w http.ResponseWriter, r *http.Request) {
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
		testServer := httptest.NewServer(http.HandlerFunc(testHandler))

		path, hash, size := "/test", md5Sum(testContents), int64(len(testContents))

		testClient := api.NewStitchClient(api.NewClient(testServer.URL))
		testClient.UploadAsset(
			groupID,
			appID,
			path,
			hash,
			size,
			strings.NewReader(testContents),
			hosting.AssetAttribute{
				Name:  "Content-Type",
				Value: "application/json",
			},
		)

		u.So(t, uploadedAssetMetadata, gc.ShouldResemble, hosting.AssetMetadata{
			AppID:    appID,
			FilePath: path,
			FileHash: hash,
			FileSize: size,
			Attrs: []hosting.AssetAttribute{
				{
					Name:  "Content-Type",
					Value: "application/json",
				},
			},
		})
	})
}

func TestListAssetsForAppID(t *testing.T) {
	t.Run("listing assets by AppID should work", func(t *testing.T) {
		testContents := []hosting.AssetMetadata{
			{
				AppID:    appID,
				FilePath: "foo.txt",
				URL:      "url/foo.txt",
				FileSize: 20,
				FileHash: "OWEJFOWEF",
			},
			{
				AppID:    appID,
				FilePath: "bar.txt",
				URL:      "url/bar.txt",
				FileSize: 203,
				FileHash: "OWEJddsdcsFOWEF",
			},
		}
		testHandler := func(w http.ResponseWriter, r *http.Request) {
			metaJson, err := json.Marshal(&testContents)
			if err != nil {
				http.Error(w, "invalid asset metadata", http.StatusBadRequest)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write(metaJson)
		}
		testServer := httptest.NewServer(http.HandlerFunc(testHandler))

		testClient := api.NewStitchClient(api.NewClient(testServer.URL))
		assetMetadatas, err := testClient.ListAssetsForAppID(groupID, appID)
		u.So(t, err, gc.ShouldBeNil)
		u.So(t, len(assetMetadatas), gc.ShouldEqual, 2)
		u.So(t, assetMetadatas[0], gc.ShouldResemble, testContents[0])
		u.So(t, assetMetadatas[1], gc.ShouldResemble, testContents[1])
	})
}

func TestSetAssetAttributes(t *testing.T) {
	t.Run("setting app attributes should work", func(t *testing.T) {
		testContents := []hosting.AssetAttribute{
			{
				Name:  "asset1",
				Value: "value1",
			},
			{
				Name:  "asset2",
				Value: "value2",
			},
		}
		testHandler := func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Query().Get(pathParam)
			if path == "" {
				http.Error(w, "path param required", http.StatusBadRequest)
				return
			}

			payload := struct {
				Attributes []hosting.AssetAttribute `json:"attributes"`
			}{}
			dec := json.NewDecoder(r.Body)
			if err := dec.Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			fmt.Printf("hello\n")

			w.WriteHeader(http.StatusNoContent)
		}
		testServer := httptest.NewServer(http.HandlerFunc(testHandler))
		path := "/foo"

		testClient := api.NewStitchClient(api.NewClient(testServer.URL))
		err := testClient.SetAssetAttributes(groupID, appID, path, testContents...)
		u.So(t, err, gc.ShouldBeNil)
	})
}

func TestPostAsset(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		dec := json.NewDecoder(r.Body)
		defer r.Body.Close()
		payload := struct {
			MoveFrom string `json:"move_from"`
			MoveTo   string `json:"move_to"`
			CopyFrom string `json:"copy_from"`
			CopyTo   string `json:"copy_to"`
		}{}
		if err := dec.Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		isMoveRequest := payload.MoveFrom != "" && payload.MoveTo != ""
		isCopyRequest := payload.CopyFrom != "" && payload.CopyTo != ""

		if isMoveRequest == isCopyRequest {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
	testServer := httptest.NewServer(http.HandlerFunc(testHandler))
	testClient := api.NewStitchClient(api.NewClient(testServer.URL))
	fromPath := "/foo"
	toPath := "/bar"

	t.Run("copying an asset should work", func(t *testing.T) {
		err := testClient.CopyAsset(groupID, appID, fromPath, toPath)
		u.So(t, err, gc.ShouldBeNil)
	})

	t.Run("moving an asset should work", func(t *testing.T) {
		err := testClient.MoveAsset(groupID, appID, fromPath, toPath)
		u.So(t, err, gc.ShouldBeNil)
	})
}

func TestDeleteAsset(t *testing.T) {
	t.Run("deleting an asset should work", func(t *testing.T) {
		testHandler := func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Query().Get(pathParam)
			if path == "" {
				http.Error(w, "path param required", http.StatusBadRequest)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		}
		testServer := httptest.NewServer(http.HandlerFunc(testHandler))
		path := "/foo"

		testClient := api.NewStitchClient(api.NewClient(testServer.URL))
		err := testClient.DeleteAsset(groupID, appID, path)
		u.So(t, err, gc.ShouldBeNil)
	})
}
