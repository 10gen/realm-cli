package commands

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/10gen/stitch-cli/api"
	"github.com/10gen/stitch-cli/hosting"
	u "github.com/10gen/stitch-cli/utils/test"
	gc "github.com/smartystreets/goconvey/convey"

	"github.com/mitchellh/cli"
)

func TestImportHosting(t *testing.T) {
	rootDir, rErr := filepath.Abs("../testdata/full_app/hosting/files")
	u.So(t, rErr, gc.ShouldBeNil)
	path0, pErr := filepath.Abs("../testdata/full_app/hosting/files/asset_file0.json")
	u.So(t, pErr, gc.ShouldBeNil)
	relPath0, relErr := filepath.Rel(rootDir, path0)
	u.So(t, relErr, gc.ShouldBeNil)

	path1, pErr := filepath.Abs("../testdata/full_app/hosting/files/ships/nostromo.json")
	u.So(t, pErr, gc.ShouldBeNil)
	relPath1, relErr := filepath.Rel(rootDir, path1)
	u.So(t, relErr, gc.ShouldBeNil)

	assetMetadataDiffs := &hosting.AssetMetadataDiffs{
		[]hosting.AssetMetadata{
			{
				FilePath: fmt.Sprintf("/%s", relPath0),
			},
			{
				FilePath: fmt.Sprintf("/%s", relPath1),
			},
		},
		[]hosting.AssetMetadata{
			{
				FilePath: "/deleteMe",
			},
		},
		[]hosting.ModifiedAssetMetadata{},
	}

	t.Run("should work with a client", func(t *testing.T) {
		testHandler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}
		testServer := httptest.NewServer(http.HandlerFunc(testHandler))
		testClient := api.NewStitchClient(api.NewClient(testServer.URL))
		u.So(t, ImportHosting("groupID", "appID", rootDir, assetMetadataDiffs, false, testClient, cli.NewMockUi()), gc.ShouldBeNil)
	})

	t.Run("should log errors correctly", func(t *testing.T) {
		testHandler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}
		testServer := httptest.NewServer(http.HandlerFunc(testHandler))
		testClient := api.NewStitchClient(api.NewClient(testServer.URL))

		mockUI := cli.NewMockUi()
		importErr := ImportHosting("groupID", "appID", rootDir, assetMetadataDiffs, false, testClient, mockUI)
		u.So(t, importErr, gc.ShouldNotBeNil)
		u.So(t, importErr.Error(), gc.ShouldContainSubstring, "3")
		u.So(t, len(strings.Split(mockUI.ErrorWriter.String(), "\n"))-1, gc.ShouldEqual, 3)
	})
}

func TestHostingOp(t *testing.T) {
	path0, pErr := filepath.Abs("../testdata/full_app/hosting/files/asset_file0.json")
	u.So(t, pErr, gc.ShouldBeNil)
	rootDir, rErr := filepath.Abs("../testdata/full_app/hosting/files")
	u.So(t, rErr, gc.ShouldBeNil)
	relPath, relErr := filepath.Rel(rootDir, path0)
	u.So(t, relErr, gc.ShouldBeNil)

	t.Run("addOp", func(t *testing.T) {
		t.Run("Do should error when an asset file fails to open", func(t *testing.T) {
			add := addOp{
				baseHostingOp{
					"", "", "/some/invalid/root", nil,
				},
				hosting.AssetMetadata{},
			}
			u.So(t, add.Do(), gc.ShouldNotBeNil)
		})

		add := addOp{
			baseHostingOp{
				"groupID", "appID", rootDir, nil,
			},
			hosting.AssetMetadata{
				FilePath: fmt.Sprintf("/%s", relPath),
			},
		}
		t.Run("Do should error when client upload fails", func(t *testing.T) {
			testHandler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}
			testServer := httptest.NewServer(http.HandlerFunc(testHandler))
			testClient := api.NewStitchClient(api.NewClient(testServer.URL))

			add.client = testClient
			u.So(t, add.Do(), gc.ShouldNotBeNil)
		})

		t.Run("Do should work", func(t *testing.T) {
			testHandler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}
			testServer := httptest.NewServer(http.HandlerFunc(testHandler))
			testClient := api.NewStitchClient(api.NewClient(testServer.URL))

			add.client = testClient
			u.So(t, add.Do(), gc.ShouldBeNil)
		})
	})

	t.Run("deleteOp", func(t *testing.T) {
		delete := deleteOp{
			baseHostingOp{
				"groupID", "appID", rootDir, nil,
			},
			hosting.AssetMetadata{
				FilePath: fmt.Sprintf("/%s", relPath),
			},
		}
		t.Run("Do should error when client delete fails", func(t *testing.T) {
			testHandler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}
			testServer := httptest.NewServer(http.HandlerFunc(testHandler))
			testClient := api.NewStitchClient(api.NewClient(testServer.URL))

			delete.client = testClient
			u.So(t, delete.Do(), gc.ShouldNotBeNil)
		})

		t.Run("Do should work", func(t *testing.T) {
			testHandler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}
			testServer := httptest.NewServer(http.HandlerFunc(testHandler))
			testClient := api.NewStitchClient(api.NewClient(testServer.URL))

			delete.client = testClient
			u.So(t, delete.Do(), gc.ShouldBeNil)
		})
	})

	t.Run("modifyOp", func(t *testing.T) {
		t.Run("Do should error when an asset file fails to open", func(t *testing.T) {
			modify := modifyOp{
				baseHostingOp{
					"groupID", "appID", "/some/invalid/root", nil,
				},
				hosting.ModifiedAssetMetadata{
					hosting.AssetMetadata{
						FilePath: "/this/path/is/wrong",
					},
					true,
					false,
				},
			}
			u.So(t, modify.Do(), gc.ShouldNotBeNil)
		})

		bodyModifyOp := modifyOp{
			baseHostingOp{
				"groupID", "appID", rootDir, nil,
			},
			hosting.ModifiedAssetMetadata{
				hosting.AssetMetadata{
					FilePath: fmt.Sprintf("/%s", relPath),
				},
				true,
				false,
			},
		}

		t.Run("Do should error when client upload fails", func(t *testing.T) {
			testHandler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}
			testServer := httptest.NewServer(http.HandlerFunc(testHandler))
			testClient := api.NewStitchClient(api.NewClient(testServer.URL))

			bodyModifyOp.client = testClient
			u.So(t, bodyModifyOp.Do(), gc.ShouldNotBeNil)
		})

		t.Run("Do should work", func(t *testing.T) {
			testHandler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}
			testServer := httptest.NewServer(http.HandlerFunc(testHandler))
			testClient := api.NewStitchClient(api.NewClient(testServer.URL))

			bodyModifyOp.client = testClient
			u.So(t, bodyModifyOp.Do(), gc.ShouldBeNil)
		})

		attrModifyOp := modifyOp{
			baseHostingOp{
				"groupID", "appID", rootDir, nil,
			},
			hosting.ModifiedAssetMetadata{
				hosting.AssetMetadata{
					FilePath: fmt.Sprintf("/%s", relPath),
				},
				false,
				true,
			},
		}

		t.Run("Do should error when attributes modified and client SetAssetAttributes fails", func(t *testing.T) {
			testHandler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}
			testServer := httptest.NewServer(http.HandlerFunc(testHandler))
			testClient := api.NewStitchClient(api.NewClient(testServer.URL))

			attrModifyOp.client = testClient
			u.So(t, attrModifyOp.Do(), gc.ShouldNotBeNil)
		})

		t.Run("Do should work when only attributes are altered", func(t *testing.T) {
			testHandler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}
			testServer := httptest.NewServer(http.HandlerFunc(testHandler))
			testClient := api.NewStitchClient(api.NewClient(testServer.URL))

			attrModifyOp.client = testClient
			u.So(t, attrModifyOp.Do(), gc.ShouldBeNil)
		})
	})

}
