package mdbcloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/abbot/go-http-auth"
	"goji.io"
	"goji.io/pat"
)

func newMockedServer() mockedServer {
	expectedUser := "correctuser"
	expectedUserID := "myuserid"
	expectedPassword := "correctpassword"
	expectedGroupID := "correctgroupid"
	expectedAtlasGroupID := "correctatlasgroupid"
	expectedAtlasClusterName := "correctatlasclustername"
	expectedAtlasDBUsername := "correctdbusername"
	authenticator := auth.NewDigestAuthenticator("domain.com", func(user, realm string) string {
		if user == expectedUser {
			// password is "hello"
			return "98fa7ec27f0c5d12e507db996b6464c5"
		}
		return ""
	})

	publicAPI := goji.NewMux()
	var rootFunc = func(w http.ResponseWriter, r *http.Request) {}
	var groupFunc = func(w http.ResponseWriter, r *http.Request) {}
	var userFunc = func(w http.ResponseWriter, r *http.Request) {}
	var atlasGroupFunc = func(w http.ResponseWriter, r *http.Request) {}
	var atlasClusterFunc = func(w http.ResponseWriter, r *http.Request) {}
	var addAtlasIPWhitelistEntriesFunc = func(w http.ResponseWriter, r *http.Request) {}
	var addAtlasDBUserFunc = func(w http.ResponseWriter, r *http.Request) {}
	var updateAtlasDBUserFunc = func(w http.ResponseWriter, r *http.Request) {}

	publicAPI.Handle(pat.Get("/"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rootFunc(w, r)
	}))
	publicAPI.Handle(pat.Get("/groups/correctgroupid"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		groupFunc(w, r)
	}))
	publicAPI.Handle(pat.Get("/users/myuserid"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userFunc(w, r)
	}))
	publicAPI.Handle(pat.Get("/groups/correctatlasgroupid/clusters"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atlasGroupFunc(w, r)
	}))
	publicAPI.Handle(pat.Get("/groups/correctatlasgroupid/clusters/correctatlasclustername"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atlasClusterFunc(w, r)
	}))
	publicAPI.Handle(pat.Post("/groups/correctatlasgroupid/whitelist"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		addAtlasIPWhitelistEntriesFunc(w, r)
	}))
	publicAPI.Handle(pat.Post("/groups/correctatlasgroupid/databaseUsers"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		addAtlasDBUserFunc(w, r)
	}))
	publicAPI.Handle(pat.Patch("/groups/correctatlasgroupid/databaseUsers/admin/correctdbusername"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		updateAtlasDBUserFunc(w, r)
	}))

	publicAPIServer := httptest.NewServer(authenticator.JustCheck(func(w http.ResponseWriter, r *http.Request) {
		publicAPI.ServeHTTP(w, r)
	}))

	publicAPIBaseURL := publicAPIServer.URL
	atlasAPIBaseURL := publicAPIServer.URL

	rootFunc = func(w http.ResponseWriter, r *http.Request) {
		root := Root{
			Links: []Link{
				{
					HRef: fmt.Sprintf("%s/users/myuserid", publicAPIBaseURL),
					Rel:  RelationUser,
				},
			},
		}
		enc := json.NewEncoder(w)
		if err := enc.Encode(root); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	groupFunc = func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		if err := enc.Encode(Group{}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	userFunc = func(w http.ResponseWriter, r *http.Request) {

		user := User{
			Roles: []RoleAssignment{
				{
					GroupID:  "correctgroupid",
					RoleName: RoleGroupOwner,
				},
				{
					GroupID:  "missinggroupid",
					RoleName: RoleGroupOwner,
				},
				{
					GroupID:  "othergroupid",
					RoleName: "otherrole",
				},
			},
		}
		enc := json.NewEncoder(w)
		if err := enc.Encode(user); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	atlasGroupFunc = func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		if err := enc.Encode(Group{}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	atlasClusterFunc = func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		if err := enc.Encode(AtlasCluster{Name: expectedAtlasClusterName, MongoURI: "someuri"}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	addAtlasIPWhitelistEntriesFunc = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}

	addAtlasDBUserFunc = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}

	return mockedServer{
		server:                         publicAPIServer,
		publicAPIBaseURL:               publicAPIBaseURL,
		atlasAPIBaseURL:                atlasAPIBaseURL,
		expectedUser:                   expectedUser,
		expectedUserID:                 expectedUserID,
		expectedPassword:               expectedPassword,
		expectedGroupID:                expectedGroupID,
		expectedAtlasGroupID:           expectedAtlasGroupID,
		expectedAtlasClusterName:       expectedAtlasClusterName,
		expectedAtlasDBUsername:        expectedAtlasDBUsername,
		rootFunc:                       &rootFunc,
		groupFunc:                      &groupFunc,
		userFunc:                       &userFunc,
		atlasGroupFunc:                 &atlasGroupFunc,
		atlasClusterFunc:               &atlasClusterFunc,
		addAtlasIPWhitelistEntriesFunc: &addAtlasIPWhitelistEntriesFunc,
		addAtlasDBUserFunc:             &addAtlasDBUserFunc,
		updateAtlasDBUserFunc:          &updateAtlasDBUserFunc,
	}
}

type mockedServer struct {
	server                         *httptest.Server
	publicAPIBaseURL               string
	atlasAPIBaseURL                string
	expectedUser                   string
	expectedUserID                 string
	expectedPassword               string
	expectedGroupID                string
	expectedAtlasGroupID           string
	expectedAtlasClusterName       string
	expectedAtlasDBUsername        string
	rootFunc                       *func(w http.ResponseWriter, r *http.Request)
	groupFunc                      *func(w http.ResponseWriter, r *http.Request)
	userFunc                       *func(w http.ResponseWriter, r *http.Request)
	atlasGroupFunc                 *func(w http.ResponseWriter, r *http.Request)
	atlasClusterFunc               *func(w http.ResponseWriter, r *http.Request)
	addAtlasIPWhitelistEntriesFunc *func(w http.ResponseWriter, r *http.Request)
	addAtlasDBUserFunc             *func(w http.ResponseWriter, r *http.Request)
	updateAtlasDBUserFunc          *func(w http.ResponseWriter, r *http.Request)
}
