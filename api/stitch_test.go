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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/10gen/stitch-cli/api"
	"github.com/10gen/stitch-cli/hosting"
	"github.com/10gen/stitch-cli/models"

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

		testContents := "hello world\r\n"
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
				http.Error(w, metaDecoderErr.Error(), http.StatusBadRequest)
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
			metaJSON, err := json.Marshal(&testContents)
			if err != nil {
				http.Error(w, "invalid asset metadata", http.StatusBadRequest)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write(metaJSON)
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

func TestInvalidateCache(t *testing.T) {
	t.Run("cache invalidation should work", func(t *testing.T) {
		testHandler := func(w http.ResponseWriter, r *http.Request) {
			payload := struct {
				Invalidate bool   `json:"invalidate"`
				Path       string `json:"path"`
			}{}

			dec := json.NewDecoder(r.Body)
			defer r.Body.Close()

			if err := dec.Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}

			if !payload.Invalidate {
				http.Error(w, "'invalidate' param should be true", http.StatusBadRequest)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		}
		testServer := httptest.NewServer(http.HandlerFunc(testHandler))
		path := "foo"

		testClient := api.NewStitchClient(api.NewClient(testServer.URL))
		err := testClient.InvalidateCache(groupID, appID, path)
		u.So(t, err, gc.ShouldBeNil)
	})
}

func TestRequestOrigin(t *testing.T) {
	t.Run("the request origin header should be set", func(t *testing.T) {
		testHandler := func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get(api.StitchRequestOriginHeader) != api.StitchCLIHeaderValue {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		}
		testServer := httptest.NewServer(http.HandlerFunc(testHandler))
		testClient := api.NewClient(testServer.URL)

		resp, err := testClient.ExecuteRequest(http.MethodGet, "", api.RequestOptions{})
		u.So(t, err, gc.ShouldBeNil)
		u.So(t, resp.StatusCode, gc.ShouldEqual, http.StatusNoContent)
	})
}

func TestFetchAppByGroupIDAndClientAppID(t *testing.T) {
	groupID := "group-id"
	atlasAppID := "triggers-stitchapp-abcde"
	standardAppID := "standard-stitchapp-abcde"
	resultTemplate := func(resultAppId string) string {
		return fmt.Sprintf(`[{ "client_app_id": "%v" }]`, resultAppId)
	}
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		productParam := r.URL.Query().Get("product")
		w.WriteHeader(http.StatusOK)
		if productParam == "atlas" {
			w.Write([]byte(resultTemplate(atlasAppID)))
			return
		}
		w.Write([]byte(resultTemplate(standardAppID)))
	}
	testServer := httptest.NewServer(http.HandlerFunc(testHandler))
	testClient := api.NewStitchClient(api.NewClient(testServer.URL))
	t.Run("Should fetch stitch apps", func(t *testing.T) {
		app, err := testClient.FetchAppByGroupIDAndClientAppID(groupID, standardAppID)
		u.So(t, err, gc.ShouldBeNil)
		u.So(t, app.ClientAppID, gc.ShouldEqual, standardAppID)
	})
	t.Run("Should fetch atlas app", func(t *testing.T) {
		app, err := testClient.FetchAppByGroupIDAndClientAppID(groupID, atlasAppID)
		u.So(t, err, gc.ShouldBeNil)
		u.So(t, app.ClientAppID, gc.ShouldEqual, atlasAppID)
	})
}

func TestCreateDraft(t *testing.T) {
	t.Run("CreateDraft should work", func(t *testing.T) {
		testHandler := func(w http.ResponseWriter, r *http.Request) {
			u.So(t, r.URL.Path, gc.ShouldEqual, "/api/admin/v3.0/groups/groupID/apps/appID/drafts")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{ "_id": "test" }`))
		}

		testServer := httptest.NewServer(http.HandlerFunc(testHandler))
		testClient := api.NewStitchClient(api.NewClient(testServer.URL))
		draft, err := testClient.CreateDraft(groupID, appID)
		u.So(t, err, gc.ShouldBeNil)
		u.So(t, draft, gc.ShouldNotBeNil)
		u.So(t, draft.ID, gc.ShouldEqual, "test")
	})
}

func TestDeployDraft(t *testing.T) {
	t.Run("DeployDraft should work", func(t *testing.T) {
		testHandler := func(w http.ResponseWriter, r *http.Request) {
			u.So(t, r.URL.Path, gc.ShouldEqual, "/api/admin/v3.0/groups/groupID/apps/appID/drafts/123/deployment")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{ "_id": "test", "status": "pending" }`))
		}

		testServer := httptest.NewServer(http.HandlerFunc(testHandler))
		testClient := api.NewStitchClient(api.NewClient(testServer.URL))
		deploy, err := testClient.DeployDraft(groupID, appID, "123")
		u.So(t, err, gc.ShouldBeNil)
		u.So(t, deploy, gc.ShouldNotBeNil)
		u.So(t, deploy.ID, gc.ShouldEqual, "test")
		u.So(t, deploy.Status, gc.ShouldEqual, models.DeploymentStatusPending)
	})
}

func TestDiscardDraft(t *testing.T) {
	t.Run("DiscardDraft should work", func(t *testing.T) {
		testHandler := func(w http.ResponseWriter, r *http.Request) {
			u.So(t, r.URL.Path, gc.ShouldEqual, "/api/admin/v3.0/groups/groupID/apps/appID/drafts/123")
			w.WriteHeader(http.StatusNoContent)
		}

		testServer := httptest.NewServer(http.HandlerFunc(testHandler))
		testClient := api.NewStitchClient(api.NewClient(testServer.URL))
		err := testClient.DiscardDraft(groupID, appID, "123")
		u.So(t, err, gc.ShouldBeNil)
	})
}

func TestGetDeployment(t *testing.T) {
	t.Run("GetDeployment should work", func(t *testing.T) {
		testHandler := func(w http.ResponseWriter, r *http.Request) {
			u.So(t, r.URL.Path, gc.ShouldEqual, "/api/admin/v3.0/groups/groupID/apps/appID/deployments/123")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{ "_id": "test", "status": "pending" }`))
		}

		testServer := httptest.NewServer(http.HandlerFunc(testHandler))
		testClient := api.NewStitchClient(api.NewClient(testServer.URL))
		deploy, err := testClient.GetDeployment(groupID, appID, "123")
		u.So(t, err, gc.ShouldBeNil)
		u.So(t, deploy, gc.ShouldNotBeNil)
		u.So(t, deploy.ID, gc.ShouldEqual, "test")
		u.So(t, deploy.Status, gc.ShouldEqual, models.DeploymentStatusPending)
	})
}

func TestGetDrafts(t *testing.T) {
	t.Run("GetDrafts should work", func(t *testing.T) {
		testHandler := func(w http.ResponseWriter, r *http.Request) {
			u.So(t, r.URL.Path, gc.ShouldEqual, "/api/admin/v3.0/groups/groupID/apps/appID/drafts")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{ "_id": "test" }]`))
		}

		testServer := httptest.NewServer(http.HandlerFunc(testHandler))
		testClient := api.NewStitchClient(api.NewClient(testServer.URL))
		drafts, err := testClient.GetDrafts(groupID, appID)
		u.So(t, err, gc.ShouldBeNil)
		u.So(t, drafts, gc.ShouldNotBeNil)
		u.So(t, len(drafts), gc.ShouldEqual, 1)
		u.So(t, drafts[0].ID, gc.ShouldEqual, "test")
	})
}

func TestDraftDiff(t *testing.T) {
	t.Run("DraftDiff should work", func(t *testing.T) {
		testHandler := func(w http.ResponseWriter, r *http.Request) {
			u.So(t, r.URL.Path, gc.ShouldEqual, "/api/admin/v3.0/groups/groupID/apps/appID/drafts/123/diff")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{ "diffs": ["--first diff"] }`))
		}

		testServer := httptest.NewServer(http.HandlerFunc(testHandler))
		testClient := api.NewStitchClient(api.NewClient(testServer.URL))
		diff, err := testClient.DraftDiff(groupID, appID, "123")
		u.So(t, err, gc.ShouldBeNil)
		u.So(t, diff, gc.ShouldNotBeNil)
		u.So(t, len(diff.Diffs), gc.ShouldEqual, 1)
		u.So(t, diff.Diffs[0], gc.ShouldEqual, "--first diff")
	})
}

func TestUploadDependencies(t *testing.T) {
	t.Run("uploading dependencies should work", func(t *testing.T) {
		path, pathErr := filepath.Abs("../testdata/app_with_dependencies/functions/node_modules.tar")
		u.So(t, pathErr, gc.ShouldBeNil)
		file, openErr := os.Open(path)
		u.So(t, openErr, gc.ShouldBeNil)
		defer file.Close()

		expectedFileData := &bytes.Buffer{}
		expectedSize, copyErr := io.Copy(expectedFileData, file)
		u.So(t, copyErr, gc.ShouldBeNil)
		u.So(t, expectedSize, gc.ShouldBeGreaterThan, 0)

		uploadedFileData := &bytes.Buffer{}
		testHandler := func(w http.ResponseWriter, r *http.Request) {
			formFile, header, err := r.FormFile("file")
			u.So(t, err, gc.ShouldBeNil)
			defer formFile.Close()
			u.So(t, header.Filename, gc.ShouldEqual, "node_modules.tar")
			u.So(t, header.Size, gc.ShouldEqual, expectedSize)

			_, err = io.Copy(uploadedFileData, formFile)
			u.So(t, err, gc.ShouldBeNil)

			w.WriteHeader(http.StatusNoContent)
		}
		testServer := httptest.NewServer(http.HandlerFunc(testHandler))

		testClient := api.NewStitchClient(api.NewClient(testServer.URL))
		testClient.UploadDependencies(groupID, appID, path)

		u.So(t, len(uploadedFileData.Bytes()), gc.ShouldResemble, len(expectedFileData.Bytes()))
		u.So(t, uploadedFileData, gc.ShouldResemble, expectedFileData)
	})
}
