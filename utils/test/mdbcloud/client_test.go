package mdbcloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	gc "github.com/smartystreets/goconvey/convey"
)

func TestClientRoot(t *testing.T) {
	gc.Convey("With a client using mocked server endpoints", t, func() {
		mockServer := newMockedServer()
		gc.Reset(func() {
			mockServer.server.Close()
		})

		client := NewClient(mockServer.publicAPIBaseURL, mockServer.atlasAPIBaseURL)
		authedClient := client.WithAuth(mockServer.expectedUser, mockServer.expectedPassword)

		gc.Convey("Without auth should fail", func() {
			_, err := client.Root()
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Failing to auth should fail", func() {
			_, err := client.WithAuth("username", "apiKey").Root()
			gc.So(err, gc.ShouldNotBeNil)
			gc.So(err, gc.ShouldResemble, fmt.Errorf("failed to authenticate with MongoDB Cloud API"))
		})

		gc.Convey("Valid response should work", func() {
			root, err := authedClient.Root()
			gc.So(err, gc.ShouldBeNil)
			gc.So(root.Links, gc.ShouldNotBeNil)
		})

		gc.Convey("Bad response should fail", func() {
			*(mockServer.rootFunc) = func(w http.ResponseWriter, r *http.Request) {}
			_, err := authedClient.Root()
			gc.So(err, gc.ShouldNotBeNil)
		})
	})
}

func TestClientSelf(t *testing.T) {
	gc.Convey("With a client using mocked server endpoints", t, func() {
		mockServer := newMockedServer()
		gc.Reset(func() {
			mockServer.server.Close()
		})

		client := NewClient(mockServer.publicAPIBaseURL, mockServer.atlasAPIBaseURL)
		authedClient := client.WithAuth(mockServer.expectedUser, mockServer.expectedPassword)

		gc.Convey("Without auth should fail", func() {
			_, err := client.Self()
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Failing to auth should fail", func() {
			_, err := client.WithAuth("username", "apiKey").Self()
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Valid response should work", func() {
			self, err := authedClient.Self()
			gc.So(err, gc.ShouldBeNil)
			gc.So(self.Roles, gc.ShouldNotBeEmpty)
			gc.So(self.Roles[1].RoleName, gc.ShouldEqual, RoleGroupOwner)
		})

		gc.Convey("Insufficient response from root should fail", func() {
			*(mockServer.rootFunc) = func(w http.ResponseWriter, r *http.Request) {
				root := Root{}
				enc := json.NewEncoder(w)
				if err := enc.Encode(root); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
			_, err := authedClient.Self()
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Bad response from root should fail", func() {
			*(mockServer.rootFunc) = func(w http.ResponseWriter, r *http.Request) {}
			_, err := authedClient.Self()
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Bad response from self should fail", func() {
			*(mockServer.userFunc) = func(w http.ResponseWriter, r *http.Request) {}
			_, err := authedClient.Self()
			gc.So(err, gc.ShouldNotBeNil)
		})
	})
}

func TestClientGroup(t *testing.T) {
	gc.Convey("With a client using mocked server endpoints", t, func() {
		mockServer := newMockedServer()
		gc.Reset(func() {
			mockServer.server.Close()
		})

		client := NewClient(mockServer.publicAPIBaseURL, mockServer.atlasAPIBaseURL)
		authedClient := client.WithAuth(mockServer.expectedUser, mockServer.expectedPassword)

		gc.Convey("Without auth should fail", func() {
			_, err := client.Group(mockServer.expectedGroupID)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Failing to auth should fail", func() {
			_, err := client.WithAuth("username", "apiKey").Group(mockServer.expectedGroupID)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Valid response should work", func() {
			_, err := authedClient.Group(mockServer.expectedGroupID)
			gc.So(err, gc.ShouldBeNil)
		})

		gc.Convey("Bad response should fail", func() {
			*(mockServer.groupFunc) = func(w http.ResponseWriter, r *http.Request) {}
			_, err := authedClient.Group(mockServer.expectedGroupID)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Group not found should fail", func() {
			_, err := authedClient.Group(mockServer.expectedGroupID + "1")
			gc.So(err, gc.ShouldNotBeNil)
		})
	})
}

func TestClientUser(t *testing.T) {
	gc.Convey("With a client using mocked server endpoints", t, func() {
		mockServer := newMockedServer()
		gc.Reset(func() {
			mockServer.server.Close()
		})

		client := NewClient(mockServer.publicAPIBaseURL, mockServer.atlasAPIBaseURL)
		authedClient := client.WithAuth(mockServer.expectedUser, mockServer.expectedPassword)

		gc.Convey("Without auth should fail", func() {
			_, err := client.User(mockServer.expectedUserID)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Failing to auth should fail", func() {
			_, err := client.WithAuth("username", "apiKey").User(mockServer.expectedUserID)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Valid response should work", func() {
			_, err := authedClient.User(mockServer.expectedUserID)
			gc.So(err, gc.ShouldBeNil)
		})

		gc.Convey("Bad response should fail", func() {
			*(mockServer.userFunc) = func(w http.ResponseWriter, r *http.Request) {}
			_, err := authedClient.User(mockServer.expectedUserID)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("User not found should fail", func() {
			_, err := authedClient.User(mockServer.expectedUserID + "1")
			gc.So(err, gc.ShouldNotBeNil)
		})
	})
}

func TestClientAtlasGroup(t *testing.T) {
	gc.Convey("With a client using mocked server endpoints", t, func() {
		mockServer := newMockedServer()
		gc.Reset(func() {
			mockServer.server.Close()
		})

		client := NewClient(mockServer.publicAPIBaseURL, mockServer.atlasAPIBaseURL)
		authedClient := client.WithAuth(mockServer.expectedUser, mockServer.expectedPassword)

		gc.Convey("Without auth should fail", func() {
			_, err := client.AtlasGroup(mockServer.expectedAtlasGroupID)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Failing to auth should fail", func() {
			_, err := client.WithAuth("username", "apiKey").AtlasGroup(mockServer.expectedAtlasGroupID)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Valid response should work", func() {
			_, err := authedClient.AtlasGroup(mockServer.expectedAtlasGroupID)
			gc.So(err, gc.ShouldBeNil)
		})

		gc.Convey("Bad response should fail", func() {
			*(mockServer.atlasGroupFunc) = func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }
			_, err := authedClient.AtlasGroup(mockServer.expectedAtlasGroupID)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Atlas group not found should fail", func() {
			_, err := authedClient.AtlasGroup(mockServer.expectedAtlasGroupID + "1")
			gc.So(err, gc.ShouldNotBeNil)
		})
	})
}

func TestClientAtlasCluster(t *testing.T) {
	gc.Convey("With a client using mocked server endpoints", t, func() {
		mockServer := newMockedServer()
		gc.Reset(func() {
			mockServer.server.Close()
		})

		client := NewClient(mockServer.publicAPIBaseURL, mockServer.atlasAPIBaseURL)
		authedClient := client.WithAuth(mockServer.expectedUser, mockServer.expectedPassword)

		gc.Convey("Without auth should fail", func() {
			_, err := client.AtlasCluster(mockServer.expectedAtlasGroupID, mockServer.expectedAtlasClusterName)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Failing to auth should fail", func() {
			_, err := client.WithAuth("username", "apiKey").AtlasCluster(mockServer.expectedAtlasGroupID, mockServer.expectedAtlasClusterName)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Valid response should work", func() {
			_, err := authedClient.AtlasCluster(mockServer.expectedAtlasGroupID, mockServer.expectedAtlasClusterName)
			gc.So(err, gc.ShouldBeNil)
		})

		gc.Convey("Bad response should fail", func() {
			*(mockServer.atlasClusterFunc) = func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }
			_, err := authedClient.AtlasCluster(mockServer.expectedAtlasGroupID, mockServer.expectedAtlasClusterName)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Malformed response should fail", func() {
			*(mockServer.atlasClusterFunc) = func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("somebadjson"))
			}
			_, err := authedClient.AtlasCluster(mockServer.expectedAtlasGroupID, mockServer.expectedAtlasClusterName)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Atlas cluster not found should fail", func() {
			_, err := authedClient.AtlasCluster(mockServer.expectedAtlasGroupID, mockServer.expectedAtlasClusterName+"1")
			gc.So(err, gc.ShouldNotBeNil)
		})
	})
}

func TestClientAddAtlasIPWhitelistEntries(t *testing.T) {
	gc.Convey("With a client using mocked server endpoints", t, func() {
		mockServer := newMockedServer()
		gc.Reset(func() {
			mockServer.server.Close()
		})

		client := NewClient(mockServer.publicAPIBaseURL, mockServer.atlasAPIBaseURL)
		authedClient := client.WithAuth(mockServer.expectedUser, mockServer.expectedPassword)

		entries := []AtlasIPWhitelistEntry{
			{CIDRBlock: "127.0.0.1/32", Comment: "comment1"},
			{CIDRBlock: "92.31.47.84/24", Comment: "comment2"},
		}

		gc.Convey("Without auth should fail", func() {
			err := client.AddAtlasIPWhitelistEntries(mockServer.expectedAtlasGroupID, entries...)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Failing to auth should fail", func() {
			err := client.WithAuth("username", "apiKey").AddAtlasIPWhitelistEntries(mockServer.expectedAtlasGroupID, entries...)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Valid response should work", func() {
			*(mockServer.addAtlasIPWhitelistEntriesFunc) = func(w http.ResponseWriter, r *http.Request) {
				var remoteEntries []AtlasIPWhitelistEntry
				dec := json.NewDecoder(r.Body)
				if err := dec.Decode(&remoteEntries); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				for idx, entry := range remoteEntries {
					if entry != entries[idx] {
						w.WriteHeader(http.StatusBadRequest)
						return
					}
				}
				w.WriteHeader(http.StatusCreated)
			}
			err := authedClient.AddAtlasIPWhitelistEntries(mockServer.expectedAtlasGroupID, entries...)
			gc.So(err, gc.ShouldBeNil)
		})

		gc.Convey("Bad response should fail", func() {
			*(mockServer.addAtlasIPWhitelistEntriesFunc) = func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }
			err := authedClient.AddAtlasIPWhitelistEntries(mockServer.expectedAtlasGroupID, entries...)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Atlas group not found should fail", func() {
			err := authedClient.AddAtlasIPWhitelistEntries(mockServer.expectedAtlasGroupID+"1", entries...)
			gc.So(err, gc.ShouldNotBeNil)
		})
	})
}

func TestClientAddAtlasDBUser(t *testing.T) {
	gc.Convey("With a client using mocked server endpoints", t, func() {
		mockServer := newMockedServer()
		gc.Reset(func() {
			mockServer.server.Close()
		})

		client := NewClient(mockServer.publicAPIBaseURL, mockServer.atlasAPIBaseURL)
		authedClient := client.WithAuth(mockServer.expectedUser, mockServer.expectedPassword)

		dbUser := &AtlasDBUser{
			DatabaseName: "db1",
			Roles: []AtlasDBUserRole{
				{DatabaseName: "db1", RoleName: AtlasDBRoleReadWriteAnyDatabase},
				{DatabaseName: "db2"},
			},
			Username: "bob",
			Password: "securepassword",
		}

		gc.Convey("Without auth should fail", func() {
			err := client.AddAtlasDBUser(mockServer.expectedAtlasGroupID, dbUser)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Failing to auth should fail", func() {
			err := client.WithAuth("username", "apiKey").AddAtlasDBUser(mockServer.expectedAtlasGroupID, dbUser)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Valid response should work", func() {
			*(mockServer.addAtlasDBUserFunc) = func(w http.ResponseWriter, r *http.Request) {
				var remoteUser AtlasDBUser
				dec := json.NewDecoder(r.Body)
				if err := dec.Decode(&remoteUser); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				if !reflect.DeepEqual(remoteUser, *dbUser) {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusCreated)
			}
			err := authedClient.AddAtlasDBUser(mockServer.expectedAtlasGroupID, dbUser)
			gc.So(err, gc.ShouldBeNil)
		})

		gc.Convey("Bad response should fail", func() {
			*(mockServer.addAtlasDBUserFunc) = func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }
			err := authedClient.AddAtlasDBUser(mockServer.expectedAtlasGroupID, dbUser)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Atlas group not found should fail", func() {
			err := authedClient.AddAtlasDBUser(mockServer.expectedAtlasGroupID+"1", dbUser)
			gc.So(err, gc.ShouldNotBeNil)
		})
	})
}

func TestClientUpdateAtlasDBUser(t *testing.T) {
	gc.Convey("With a client using mocked server endpoints", t, func() {
		mockServer := newMockedServer()
		gc.Reset(func() {
			mockServer.server.Close()
		})

		client := NewClient(mockServer.publicAPIBaseURL, mockServer.atlasAPIBaseURL)
		authedClient := client.WithAuth(mockServer.expectedUser, mockServer.expectedPassword)

		dbUser := &AtlasDBUser{
			DatabaseName: "db1",
			Roles: []AtlasDBUserRole{
				{DatabaseName: "db1", RoleName: AtlasDBRoleReadWriteAnyDatabase},
				{DatabaseName: "db2"},
			},
			Username: mockServer.expectedAtlasDBUsername,
			Password: "securepassword",
		}

		gc.Convey("Without auth should fail", func() {
			err := client.UpdateAtlasDBUser(mockServer.expectedAtlasGroupID, dbUser)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Failing to auth should fail", func() {
			err := client.WithAuth("username", "apiKey").UpdateAtlasDBUser(mockServer.expectedAtlasGroupID, dbUser)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Valid response should work", func() {
			*(mockServer.updateAtlasDBUserFunc) = func(w http.ResponseWriter, r *http.Request) {
				var remoteUser AtlasDBUser
				dec := json.NewDecoder(r.Body)
				if err := dec.Decode(&remoteUser); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				if !reflect.DeepEqual(remoteUser, *dbUser) {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
			}
			err := authedClient.UpdateAtlasDBUser(mockServer.expectedAtlasGroupID, dbUser)
			gc.So(err, gc.ShouldBeNil)
		})

		gc.Convey("Bad response should fail", func() {
			*(mockServer.updateAtlasDBUserFunc) = func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }
			err := authedClient.UpdateAtlasDBUser(mockServer.expectedAtlasGroupID, dbUser)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("User not found should fail", func() {
			dbUser.Username = "somethingelse"
			err := authedClient.UpdateAtlasDBUser(mockServer.expectedAtlasGroupID, dbUser)
			gc.So(err, gc.ShouldNotBeNil)
		})
	})
}

func TestClientAddOrUpdateAtlasDBUser(t *testing.T) {
	gc.Convey("With a client using mocked server endpoints", t, func() {
		mockServer := newMockedServer()
		gc.Reset(func() {
			mockServer.server.Close()
		})

		client := NewClient(mockServer.publicAPIBaseURL, mockServer.atlasAPIBaseURL)
		authedClient := client.WithAuth(mockServer.expectedUser, mockServer.expectedPassword)

		dbUser := &AtlasDBUser{
			DatabaseName: "db1",
			Roles: []AtlasDBUserRole{
				{DatabaseName: "db1", RoleName: AtlasDBRoleReadWriteAnyDatabase},
				{DatabaseName: "db2"},
			},
			Username: mockServer.expectedAtlasDBUsername,
			Password: "securepassword",
		}

		gc.Convey("Without auth should fail", func() {
			err := client.AddOrUpdateAtlasDBUser(mockServer.expectedAtlasGroupID, dbUser)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Failing to auth should fail", func() {
			err := client.WithAuth("username", "apiKey").AddOrUpdateAtlasDBUser(mockServer.expectedAtlasGroupID, dbUser)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Failing to add user should fail", func() {
			*(mockServer.addAtlasDBUserFunc) = func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }
			err := authedClient.AddOrUpdateAtlasDBUser(mockServer.expectedAtlasGroupID, dbUser)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("User does not exist should work", func() {
			*(mockServer.addAtlasDBUserFunc) = func(w http.ResponseWriter, r *http.Request) {
				var remoteUser AtlasDBUser
				dec := json.NewDecoder(r.Body)
				if err := dec.Decode(&remoteUser); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				if !reflect.DeepEqual(remoteUser, *dbUser) {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusCreated)
			}
			*(mockServer.updateAtlasDBUserFunc) = func(w http.ResponseWriter, r *http.Request) {
				t.FailNow()
			}
			err := authedClient.AddOrUpdateAtlasDBUser(mockServer.expectedAtlasGroupID, dbUser)
			gc.So(err, gc.ShouldBeNil)
		})

		gc.Convey("User already exists but failing to update user should fail", func() {
			*(mockServer.addAtlasDBUserFunc) = func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusConflict) }
			*(mockServer.updateAtlasDBUserFunc) = func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }
			err := authedClient.AddOrUpdateAtlasDBUser(mockServer.expectedAtlasGroupID, dbUser)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("Atlas group not found should fail", func() {
			err := authedClient.AddOrUpdateAtlasDBUser(mockServer.expectedAtlasGroupID+"1", dbUser)
			gc.So(err, gc.ShouldNotBeNil)
		})

		gc.Convey("User not found should fail if user already exists", func() {
			*(mockServer.addAtlasDBUserFunc) = func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }
			dbUser.Username = "somethingelse"
			err := authedClient.AddOrUpdateAtlasDBUser(mockServer.expectedAtlasGroupID, dbUser)
			gc.So(err, gc.ShouldNotBeNil)
		})
	})
}
